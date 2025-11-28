package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	JWT       JWTConfig       `yaml:"jwt"`
	WireGuard WireGuardConfig `yaml:"wireguard"`
	Audit     AuditConfig     `yaml:"audit"`
	CORS      CORSConfig      `yaml:"cors"`
	Redis     RedisConfig     `yaml:"redis"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         string        `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
	Environment  string        `yaml:"environment"` // "development" or "production"
}

// DatabaseConfig holds database configuration (supports postgres|sqlite|memory)
type DatabaseConfig struct {
	Backend         string        `yaml:"backend"`
	Host            string        `yaml:"host"`
	Port            string        `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	DBName          string        `yaml:"dbname"`
	SSLMode         string        `yaml:"sslmode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	SQLitePath      string        `yaml:"sqlite_path"`
}

// JWTConfig holds JWT token configuration
type JWTConfig struct {
	Secret           string        `yaml:"secret"`
	AccessTokenTTL   time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL  time.Duration `yaml:"refresh_token_ttl"`
	RefreshSecretKey string        `yaml:"refresh_secret_key"` // Optional separate key for refresh tokens
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// WireGuardConfig holds WireGuard server configuration
type WireGuardConfig struct {
	InterfaceName  string `yaml:"interface_name"`  // Interface name (default: wg0)
	PrivateKey     string `yaml:"private_key"`     // Server's WireGuard private key
	ServerEndpoint string `yaml:"server_endpoint"` // Public endpoint (e.g., "vpn.example.com:51820")
	ServerPubKey   string `yaml:"server_pubkey"`   // Server's WireGuard public key (44 chars base64)
	DNS            string `yaml:"dns"`             // DNS servers for clients (comma-separated)
	MTU            int    `yaml:"mtu"`             // MTU for the interface
	Keepalive      int    `yaml:"keepalive"`       // Persistent keepalive in seconds
	Port           int    `yaml:"port"`            // Listen port (default: 51820)
}

// AuditConfig holds audit logging configuration
type AuditConfig struct {
	SQLiteDSN     string        `yaml:"sqlite_dsn"`   // SQLite database path for audit logs
	HashSecrets   string        `yaml:"hash_secrets"` // Comma-separated base64 secrets for hashing
	Async         bool          `yaml:"async"`        // Enable async audit buffering
	QueueSize     int           `yaml:"queue_size"`   // Async queue size
	WorkerCount   int           `yaml:"worker_count"` // Number of async workers
	FlushInterval time.Duration `yaml:"flush_interval"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string      `yaml:"allowed_origins"` // Whitelist of allowed origins
	AllowCredentials bool          `yaml:"allow_credentials"`
	MaxAge           time.Duration `yaml:"max_age"`
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server:    loadServerConfig(),
		Database:  loadDatabaseConfig(),
		JWT:       loadJWTConfig(),
		WireGuard: loadWireGuardConfig(),
		Audit:     loadAuditConfig(),
		CORS:      loadCORSConfig(),
		Redis:     loadRedisConfig(),
	}

	// Validate critical fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// DefaultConfigPath returns the default config file path (env override allowed).
func DefaultConfigPath() string {
	if val := strings.TrimSpace(os.Getenv("GOCONNECT_CONFIG_PATH")); val != "" {
		return val
	}
	return "goconnect.yaml"
}

// LoadFromFileOrEnv loads configuration from a YAML file if it exists, then applies environment variable overrides.
// If the file does not exist, it falls back to the existing environment-based Load().
// Environment overrides are only applied when the variable is explicitly set (no default injection).
func LoadFromFileOrEnv(path string) (*Config, error) {
	fileCfg := Config{}
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
		}
		if err := yaml.Unmarshal(content, &fileCfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
		}
		applyEnvOverrides(&fileCfg)
		if err := fileCfg.Validate(); err != nil {
			return nil, err
		}
		return &fileCfg, nil
	}

	// File missing: keep current env-based behavior
	return Load()
}

// SaveToFile writes the given config to a YAML file at the provided path.
func SaveToFile(cfg *Config, path string) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	return os.WriteFile(path, data, 0o600)
}

// LoadFromFile reads a YAML config file without environment overrides.
func LoadFromFile(path string) (*Config, error) {
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		return nil, fmt.Errorf("config file not found: %s", path)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	cfg := Config{}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate ensures all required configuration is present
func (c *Config) Validate() error {
	// Server validation
	if c.Server.Port == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}

	// Database validation
	switch c.Database.backendOrDefault() {
	case "postgres":
		if c.Database.Host == "" {
			return fmt.Errorf("DB_HOST is required for postgres backend")
		}
		if c.Database.DBName == "" {
			return fmt.Errorf("DB_NAME is required for postgres backend")
		}
		if c.Database.User == "" {
			return fmt.Errorf("DB_USER is required for postgres backend")
		}
	case "sqlite":
		if strings.TrimSpace(c.Database.SQLitePath) == "" {
			return fmt.Errorf("DB_SQLITE_PATH is required for sqlite backend")
		}
	case "memory":
		// No required fields
	default:
		return fmt.Errorf("invalid DB_BACKEND: %s (expected postgres|sqlite|memory)", c.Database.Backend)
	}

	// JWT validation
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required (use a strong random key)")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	// WireGuard validation
	if c.WireGuard.ServerEndpoint == "" {
		return fmt.Errorf("WG_SERVER_ENDPOINT is required")
	}
	if c.WireGuard.ServerPubKey == "" {
		return fmt.Errorf("WG_SERVER_PUBKEY is required")
	}
	if len(c.WireGuard.ServerPubKey) != 44 {
		return fmt.Errorf("WG_SERVER_PUBKEY must be exactly 44 characters (base64 encoded)")
	}

	return nil
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		Host:         getEnv("SERVER_HOST", "0.0.0.0"),
		Port:         getEnv("SERVER_PORT", "8080"),
		ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
		WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		Environment:  getEnv("ENVIRONMENT", "development"),
	}
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Backend:         strings.ToLower(getEnv("DB_BACKEND", "postgres")),
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", "goconnect"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		SQLitePath:      getEnv("DB_SQLITE_PATH", "data/goconnect.db"),
	}
}

func loadJWTConfig() JWTConfig {
	return JWTConfig{
		Secret:           getEnv("JWT_SECRET", ""),
		AccessTokenTTL:   getDurationEnv("JWT_ACCESS_TTL", 15*time.Minute),
		RefreshTokenTTL:  getDurationEnv("JWT_REFRESH_TTL", 7*24*time.Hour),
		RefreshSecretKey: getEnv("JWT_REFRESH_SECRET", ""), // Optional
	}
}

func loadWireGuardConfig() WireGuardConfig {
	return WireGuardConfig{
		InterfaceName:  getEnv("WG_INTERFACE_NAME", "wg0"),
		PrivateKey:     getEnv("WG_PRIVATE_KEY", ""),
		ServerEndpoint: getEnv("WG_SERVER_ENDPOINT", ""),
		ServerPubKey:   getEnv("WG_SERVER_PUBKEY", ""),
		DNS:            getEnv("WG_DNS", "1.1.1.1, 1.0.0.1"),
		MTU:            getIntEnv("WG_MTU", 1420),
		Keepalive:      getIntEnv("WG_KEEPALIVE", 25),
		Port:           getIntEnv("WG_PORT", 51820),
	}
}

func loadAuditConfig() AuditConfig {
	return AuditConfig{
		SQLiteDSN:     getEnv("AUDIT_SQLITE_DSN", ""),
		HashSecrets:   getEnv("AUDIT_HASH_SECRETS_B64", ""),
		Async:         getBoolEnv("AUDIT_ASYNC", true),
		QueueSize:     getIntEnv("AUDIT_QUEUE_SIZE", 1024),
		WorkerCount:   getIntEnv("AUDIT_WORKER_COUNT", 1),
		FlushInterval: getDurationEnv("AUDIT_FLUSH_INTERVAL", 1*time.Second),
	}
}

func loadCORSConfig() CORSConfig {
	originsStr := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	origins := []string{}
	if originsStr != "" {
		origins = splitAndTrim(originsStr, ",")
	}

	return CORSConfig{
		AllowedOrigins:   origins,
		AllowCredentials: getBoolEnv("CORS_ALLOW_CREDENTIALS", true),
		MaxAge:           getDurationEnv("CORS_MAX_AGE", 12*time.Hour),
	}
}

func loadRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getIntEnv("REDIS_DB", 0),
	}
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func applyEnvOverrides(cfg *Config) {
	// Server
	if v, ok := lookupEnvNonEmpty("SERVER_HOST"); ok {
		cfg.Server.Host = v
	}
	if v, ok := lookupEnvNonEmpty("SERVER_PORT"); ok {
		cfg.Server.Port = v
	}
	if v, ok := lookupEnvNonEmpty("SERVER_READ_TIMEOUT"); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.ReadTimeout = d
		}
	}
	if v, ok := lookupEnvNonEmpty("SERVER_WRITE_TIMEOUT"); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.WriteTimeout = d
		}
	}
	if v, ok := lookupEnvNonEmpty("SERVER_IDLE_TIMEOUT"); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.IdleTimeout = d
		}
	}
	if v, ok := lookupEnvNonEmpty("ENVIRONMENT"); ok {
		cfg.Server.Environment = v
	}

	// Database
	if v, ok := lookupEnvNonEmpty("DB_BACKEND"); ok {
		cfg.Database.Backend = strings.ToLower(v)
	}
	if v, ok := lookupEnvNonEmpty("DB_HOST"); ok {
		cfg.Database.Host = v
	}
	if v, ok := lookupEnvNonEmpty("DB_PORT"); ok {
		cfg.Database.Port = v
	}
	if v, ok := lookupEnvNonEmpty("DB_USER"); ok {
		cfg.Database.User = v
	}
	if v, ok := lookupEnvNonEmpty("DB_PASSWORD"); ok {
		cfg.Database.Password = v
	}
	if v, ok := lookupEnvNonEmpty("DB_NAME"); ok {
		cfg.Database.DBName = v
	}
	if v, ok := lookupEnvNonEmpty("DB_SSLMODE"); ok {
		cfg.Database.SSLMode = v
	}
	if v, ok := lookupEnvNonEmpty("DB_MAX_OPEN_CONNS"); ok {
		if iv, err := strconv.Atoi(v); err == nil {
			cfg.Database.MaxOpenConns = iv
		}
	}
	if v, ok := lookupEnvNonEmpty("DB_MAX_IDLE_CONNS"); ok {
		if iv, err := strconv.Atoi(v); err == nil {
			cfg.Database.MaxIdleConns = iv
		}
	}
	if v, ok := lookupEnvNonEmpty("DB_CONN_MAX_LIFETIME"); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Database.ConnMaxLifetime = d
		}
	}
	if v, ok := lookupEnvNonEmpty("DB_SQLITE_PATH"); ok {
		cfg.Database.SQLitePath = v
	}

	// JWT
	if v, ok := lookupEnvNonEmpty("JWT_SECRET"); ok {
		cfg.JWT.Secret = v
	}
	if v, ok := lookupEnvNonEmpty("JWT_ACCESS_TTL"); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.JWT.AccessTokenTTL = d
		}
	}
	if v, ok := lookupEnvNonEmpty("JWT_REFRESH_TTL"); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.JWT.RefreshTokenTTL = d
		}
	}
	if v, ok := lookupEnvNonEmpty("JWT_REFRESH_SECRET"); ok {
		cfg.JWT.RefreshSecretKey = v
	}

	// WireGuard
	if v, ok := lookupEnvNonEmpty("WG_INTERFACE_NAME"); ok {
		cfg.WireGuard.InterfaceName = v
	}
	if v, ok := lookupEnvNonEmpty("WG_PRIVATE_KEY"); ok {
		cfg.WireGuard.PrivateKey = v
	}
	if v, ok := lookupEnvNonEmpty("WG_SERVER_ENDPOINT"); ok {
		cfg.WireGuard.ServerEndpoint = v
	}
	if v, ok := lookupEnvNonEmpty("WG_SERVER_PUBKEY"); ok {
		cfg.WireGuard.ServerPubKey = v
	}
	if v, ok := lookupEnvNonEmpty("WG_DNS"); ok {
		cfg.WireGuard.DNS = v
	}
	if v, ok := lookupEnvNonEmpty("WG_MTU"); ok {
		if iv, err := strconv.Atoi(v); err == nil {
			cfg.WireGuard.MTU = iv
		}
	}
	if v, ok := lookupEnvNonEmpty("WG_KEEPALIVE"); ok {
		if iv, err := strconv.Atoi(v); err == nil {
			cfg.WireGuard.Keepalive = iv
		}
	}
	if v, ok := lookupEnvNonEmpty("WG_PORT"); ok {
		if iv, err := strconv.Atoi(v); err == nil {
			cfg.WireGuard.Port = iv
		}
	}

	// Audit
	if v, ok := lookupEnvNonEmpty("AUDIT_SQLITE_DSN"); ok {
		cfg.Audit.SQLiteDSN = v
	}
	if v, ok := lookupEnvNonEmpty("AUDIT_HASH_SECRETS"); ok {
		cfg.Audit.HashSecrets = v
	}
	if v, ok := lookupEnvNonEmpty("AUDIT_ASYNC"); ok {
		cfg.Audit.Async = strings.ToLower(v) == "true"
	}
	if v, ok := lookupEnvNonEmpty("AUDIT_QUEUE_SIZE"); ok {
		if iv, err := strconv.Atoi(v); err == nil {
			cfg.Audit.QueueSize = iv
		}
	}
	if v, ok := lookupEnvNonEmpty("AUDIT_WORKER_COUNT"); ok {
		if iv, err := strconv.Atoi(v); err == nil {
			cfg.Audit.WorkerCount = iv
		}
	}
	if v, ok := lookupEnvNonEmpty("AUDIT_FLUSH_INTERVAL"); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Audit.FlushInterval = d
		}
	}

	// CORS
	if v, ok := lookupEnvNonEmpty("CORS_ALLOWED_ORIGINS"); ok {
		cfg.CORS.AllowedOrigins = splitAndTrim(v, ",")
	}
	if v, ok := lookupEnvNonEmpty("CORS_ALLOW_CREDENTIALS"); ok {
		cfg.CORS.AllowCredentials = strings.ToLower(v) == "true"
	}
	if v, ok := lookupEnvNonEmpty("CORS_MAX_AGE"); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.CORS.MaxAge = d
		}
	}

	// Redis
	if v, ok := lookupEnvNonEmpty("REDIS_HOST"); ok {
		cfg.Redis.Host = v
	}
	if v, ok := lookupEnvNonEmpty("REDIS_PORT"); ok {
		cfg.Redis.Port = v
	}
	if v, ok := lookupEnvNonEmpty("REDIS_PASSWORD"); ok {
		cfg.Redis.Password = v
	}
	if v, ok := lookupEnvNonEmpty("REDIS_DB"); ok {
		if iv, err := strconv.Atoi(v); err == nil {
			cfg.Redis.DB = iv
		}
	}
}

func lookupEnv(key string) (string, bool) {
	val, ok := os.LookupEnv(key)
	return val, ok
}

func lookupEnvNonEmpty(key string) (string, bool) {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return "", false
	}
	return val, true
}

func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range strings.Split(s, sep) {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// ConnectionString generates PostgreSQL connection string
func (d DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// Address returns the full server address
func (s ServerConfig) Address() string {
	return s.Host + ":" + s.Port
}

// IsDevelopment returns true if environment is development
func (s ServerConfig) IsDevelopment() bool {
	return s.Environment == "development"
}

// IsProduction returns true if environment is production
func (s ServerConfig) IsProduction() bool {
	return s.Environment == "production"
}

// backendOrDefault normalizes backend selection with a safe default
func (d DatabaseConfig) backendOrDefault() string {
	if d.Backend == "" {
		return "postgres"
	}
	return strings.ToLower(d.Backend)
}
