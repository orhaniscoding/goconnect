package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path == "" {
		t.Error("Expected non-empty default config path")
	}
	// Should contain .goconnect and config.yaml
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("Expected config.yaml, got %s", filepath.Base(path))
	}
}

func TestLoadConfig_NonExistent(t *testing.T) {
	// Loading non-existent file should return default config
	cfg, err := LoadConfig("/non/existent/path/config.yaml")
	if err != nil {
		t.Fatalf("Expected no error for non-existent file, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected default config, got nil")
	}

	// Check default port
	if cfg.Daemon.LocalPort != 34100 {
		t.Errorf("Expected default port 34100, got %d", cfg.Daemon.LocalPort)
	}
}

func TestLoadConfig_ValidYAML(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
server:
  url: "https://api.example.com"
daemon:
  listen_addr: "127.0.0.1:8080"
  local_port: 9999
  health_check_interval: 30s
wireguard:
  interface_name: "wg0"
p2p:
  enabled: true
  stun_server: "stun:stun.example.com:3478"
settings:
  auto_connect: true
  notifications_enabled: false
  download_path: "/tmp/downloads"
  log_level: "debug"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify server settings
	if cfg.Server.URL != "https://api.example.com" {
		t.Errorf("Expected server URL 'https://api.example.com', got %s", cfg.Server.URL)
	}

	// Verify daemon settings
	if cfg.Daemon.LocalPort != 9999 {
		t.Errorf("Expected local port 9999, got %d", cfg.Daemon.LocalPort)
	}
	if cfg.Daemon.ListenAddr != "127.0.0.1:8080" {
		t.Errorf("Expected listen addr '127.0.0.1:8080', got %s", cfg.Daemon.ListenAddr)
	}

	// Verify WireGuard settings
	if cfg.WireGuard.InterfaceName != "wg0" {
		t.Errorf("Expected interface name 'wg0', got %s", cfg.WireGuard.InterfaceName)
	}

	// Verify P2P settings
	if !cfg.P2P.Enabled {
		t.Error("Expected P2P enabled")
	}
	if cfg.P2P.StunServer != "stun:stun.example.com:3478" {
		t.Errorf("Expected STUN server 'stun:stun.example.com:3478', got %s", cfg.P2P.StunServer)
	}

	// Verify settings
	if !cfg.Settings.AutoConnect {
		t.Error("Expected auto_connect true")
	}
	if cfg.Settings.DownloadPath != "/tmp/downloads" {
		t.Errorf("Expected download path '/tmp/downloads', got %s", cfg.Settings.DownloadPath)
	}
	if cfg.Settings.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got %s", cfg.Settings.LogLevel)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Invalid YAML
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Minimal config - should get defaults
	yamlContent := `
server:
  url: "https://api.example.com"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Check default port is set
	if cfg.Daemon.LocalPort != 34100 {
		t.Errorf("Expected default port 34100, got %d", cfg.Daemon.LocalPort)
	}

	// Check default log level
	if cfg.Settings.LogLevel != "info" {
		t.Errorf("Expected default log level 'info', got %s", cfg.Settings.LogLevel)
	}

	// Check default notifications
	if !cfg.Settings.NotificationsEnabled {
		t.Error("Expected notifications enabled by default")
	}

	// Check ConfigPath is set
	if cfg.ConfigPath != configPath {
		t.Errorf("Expected ConfigPath %s, got %s", configPath, cfg.ConfigPath)
	}
}

func TestConfig_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	cfg := &Config{}
	cfg.Server.URL = "https://test.example.com"
	cfg.Daemon.LocalPort = 12345
	cfg.Settings.AutoConnect = true
	cfg.Settings.LogLevel = "debug"
	cfg.IdentityPath = "/path/to/identity.json"

	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load and verify
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.Server.URL != "https://test.example.com" {
		t.Errorf("Expected server URL 'https://test.example.com', got %s", loaded.Server.URL)
	}
	if loaded.Daemon.LocalPort != 12345 {
		t.Errorf("Expected port 12345, got %d", loaded.Daemon.LocalPort)
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{}
	cfg.Server.URL = "https://save.example.com"

	err := SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestConfig_SettingsStruct(t *testing.T) {
	// Test Settings struct serialization
	cfg := &Config{}
	cfg.Settings.AutoConnect = true
	cfg.Settings.NotificationsEnabled = true
	cfg.Settings.DownloadPath = "/downloads"
	cfg.Settings.LogLevel = "warn"

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if !loaded.Settings.AutoConnect {
		t.Error("Expected AutoConnect true")
	}
	if loaded.Settings.DownloadPath != "/downloads" {
		t.Errorf("Expected download path '/downloads', got %s", loaded.Settings.DownloadPath)
	}
	// Note: LogLevel defaults to "info" if empty in yaml, but we saved "warn"
}

func TestConfig_Save_ReadOnlyDirectory(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping as root")
	}
	// Skip on Windows as permissions work differently
	if filepath.Separator == '\\' {
		t.Skip("Skipping on Windows")
	}

	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("Failed to create readonly dir: %v", err)
	}
	defer os.Chmod(readOnlyDir, 0755)

	cfg := &Config{}
	cfg.Server.URL = "https://test.example.com"

	// Try to save config in readonly directory
	configPath := filepath.Join(readOnlyDir, "subdir", "config.yaml")
	err := cfg.Save(configPath)
	if err == nil {
		t.Error("Expected error when saving to readonly directory")
	}
}

func TestConfig_Save_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "a", "b", "c", "config.yaml")

	cfg := &Config{}
	cfg.Server.URL = "https://nested.example.com"

	err := cfg.Save(nestedPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}
