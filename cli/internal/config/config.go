package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/logger"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/storage"
	"gopkg.in/yaml.v3"
)

// Config holds all daemon configuration.
type Config struct {
	Server struct {
		URL string `yaml:"url"`
	} `yaml:"server"`
	Daemon struct {
		ListenAddr          string        `yaml:"listen_addr"`
		LocalPort           int           `yaml:"local_port"`
		HealthCheckInterval time.Duration `yaml:"health_check_interval"`
		SocketPath          string        `yaml:"socket_path"`
		IPCTokenPath        string        `yaml:"ipc_token_path"`
	} `yaml:"daemon"`
	WireGuard struct {
		InterfaceName string `yaml:"interface_name"`
	} `yaml:"wireguard"`
	Identity struct {
		Path string `yaml:"path"`
	} `yaml:"identity"`
	P2P struct {
		Enabled    bool   `yaml:"enabled"`
		StunServer string `yaml:"stun_server"`
	} `yaml:"p2p"`
	// User-configurable settings
	Settings struct {
		AutoConnect          bool   `yaml:"auto_connect"`
		NotificationsEnabled bool   `yaml:"notifications_enabled"`
		DownloadPath         string `yaml:"download_path"`
		LogLevel             string `yaml:"log_level"`
	} `yaml:"settings"`

	// Runtime fields
	DevicePrivateKey string                `yaml:"-"`
	DevicePublicKey  string                `yaml:"-"`
	IdentityPath     string                `yaml:"-"`
	Keyring          *storage.KeyringStore `yaml:"-"`
	ConfigPath       string                `yaml:"-"` // Path to config file for saving
}

// DefaultConfigPath returns the default path for the configuration file.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.yaml"
	}
	return filepath.Join(home, ".goconnect", "config.yaml")
}

// LoadConfig loads the configuration from the given path.
func LoadConfig(path string) (*Config, error) {
	var cfg Config
	
	// Default configuration values
	defaultConfig := func(c *Config) {
		if c.Server.URL == "" {
			c.Server.URL = "https://api.goconnect.io"
		}
		if c.Daemon.LocalPort == 0 {
			c.Daemon.LocalPort = 34100
		}
		if c.Daemon.HealthCheckInterval == 0 {
			c.Daemon.HealthCheckInterval = 30 * time.Second
		}
		if c.Daemon.SocketPath == "" {
			if os.Getenv("GOOS") == "windows" {
				c.Daemon.SocketPath = `\\.\pipe\goconnect`
			} else {
				c.Daemon.SocketPath = "/tmp/goconnect.sock"
			}
		}
		
		// Identity Path
		if c.Identity.Path != "" {
			c.IdentityPath = c.Identity.Path
		} else {
			home, _ := os.UserHomeDir()
			c.IdentityPath = filepath.Join(home, ".goconnect", "identity.json")
		}

		// Settings
		if c.Settings.LogLevel == "" {
			c.Settings.LogLevel = "info"
		}
		if c.Settings.DownloadPath == "" {
			home, _ := os.UserHomeDir()
			c.Settings.DownloadPath = filepath.Join(home, "Downloads")
		}
		// Default to enabled if not explicitly disabled (bool default is false, so we need careful handling if we want default true)
		// Since we can't distinguish false from unset easily without pointers, we assume false means disabled. 
		// If we want default true, we'd need *bool. For now, let's leave it as is or enforce logic elsewhere.
		// However, for first run, we can set it.
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// file not found, return default config
			cfg = Config{}
			defaultConfig(&cfg)
			// Initialize keyring for fresh config
			keyring, err := storage.NewKeyringStore()
			if err != nil {
				logger.Warn("Failed to initialize keyring", "error", err)
			}
			cfg.Keyring = keyring
			cfg.Settings.NotificationsEnabled = true 
			return &cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Initialize Keyring
	keyring, err := storage.NewKeyringStore()
	if err != nil {
		logger.Warn("Failed to initialize keyring", "error", err)
	}
	cfg.Keyring = keyring

	// Apply defaults to loaded config
	defaultConfig(&cfg)
	
	// Special handling for NotificationsEnabled: yaml unmarshal handles it, but default logic is tricky. 
	// We'll trust user config or bool default (false) if missing.

	cfg.ConfigPath = path

	return &cfg, nil
}

// Save saves the configuration to the given path.
func (c *Config) Save(path string) error {
	// Sync runtime fields back to YAML fields if needed
	c.Identity.Path = c.IdentityPath

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// SaveConfig saves the configuration to the given path.
func SaveConfig(c *Config, path string) error {
	return c.Save(path)
}
