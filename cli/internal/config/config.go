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
	data, err := os.ReadFile(path)
	if err != nil {
		// Return default config if file not found
		if os.IsNotExist(err) {
			return &Config{
				Daemon: struct {
					ListenAddr          string        `yaml:"listen_addr"`
					LocalPort           int           `yaml:"local_port"`
					HealthCheckInterval time.Duration `yaml:"health_check_interval"`
				}{
					LocalPort: 34100, // Default port
				},
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Initialize Keyring
	keyring, err := storage.NewKeyringStore()
	if err != nil {
		logger.Warn("Failed to initialize keyring", "error", err)
	}
	cfg.Keyring = keyring

	// Set defaults if missing
	if cfg.Daemon.LocalPort == 0 {
		cfg.Daemon.LocalPort = 34100
	}
	if cfg.Identity.Path != "" {
		cfg.IdentityPath = cfg.Identity.Path
	} else {
		// Default identity path
		home, _ := os.UserHomeDir()
		cfg.IdentityPath = filepath.Join(home, ".goconnect", "identity.json")
	}
	// Set default settings
	if cfg.Settings.LogLevel == "" {
		cfg.Settings.LogLevel = "info"
	}
	if cfg.Settings.DownloadPath == "" {
		home, _ := os.UserHomeDir()
		cfg.Settings.DownloadPath = filepath.Join(home, "Downloads")
	}
	cfg.Settings.NotificationsEnabled = true // Default to enabled
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
