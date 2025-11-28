package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/storage"
)

// Config holds all daemon configuration.
type Config struct {
	Server struct {
		URL string `yaml:"url"`
	} `yaml:"server"`
	Daemon struct {
		LocalPort           int           `yaml:"local_port"`
		HealthCheckInterval time.Duration `yaml:"health_check_interval"`
	} `yaml:"daemon"`
	WireGuard struct {
		InterfaceName string `yaml:"interface_name"`
	} `yaml:"wireguard"`
	IdentityPath string `yaml:"identity_path"` // Path to store device identity

	// Device keys are dynamic and not stored in config file
	DevicePrivateKey string
	DevicePublicKey  string

	// Keyring store for sensitive data like auth tokens
	Keyring *storage.KeyringStore `yaml:"-"` // Ignore in YAML marshalling
}

// DefaultConfigPath returns the default path for the daemon config file.
func DefaultConfigPath() string {
	// TODO: Platform specific paths
	return "./config.yaml"
}

// LoadConfig loads configuration from a YAML file or sets defaults.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		Server: struct {
			URL string `yaml:"url"`
		}{
			URL: "http://localhost:8080", // Default server URL
		},
		Daemon: struct {
			LocalPort           int           `yaml:"local_port"`
			HealthCheckInterval time.Duration `yaml:"health_check_interval"`
		}{
			LocalPort:           12345,
			HealthCheckInterval: 30 * time.Second,
		},
		WireGuard: struct {
			InterfaceName string `yaml:"interface_name"`
		}{
			InterfaceName: "goconnect0", // Default interface name
		},
		IdentityPath: "./identity.json", // Default identity path
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Config file not found at %s, using defaults.", path)
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Initialize keyring store
	kr, err := storage.NewKeyringStore()
	if err != nil {
		log.Printf("Failed to initialize keyring store: %v. Sensitive data will not be persistently stored.", err)
		// Don't return error, allow daemon to run without keyring if it fails
	}
	cfg.Keyring = kr

	return cfg, nil
}