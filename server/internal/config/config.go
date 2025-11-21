package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the application configuration
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	WireGuard WireGuardConfig
	Audit     AuditConfig
	CORS      CORSConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Environment  string // "development" or "production"
}

// DatabaseConfig holds PostgreSQL database configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig holds JWT token configuration
type JWTConfig struct {
	Secret           string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	RefreshSecretKey string // Optional separate key for refresh tokens
}

// WireGuardConfig holds WireGuard server configuration
type WireGuardConfig struct {
	InterfaceName  string // Interface name (default: wg0)
	PrivateKey     string // Server's WireGuard private key
	ServerEndpoint string // Public endpoint (e.g., "vpn.example.com:51820")
	ServerPubKey   string // Server's WireGuard public key (44 chars base64)
	DNS            string // DNS servers for clients (comma-separated)
	MTU            int    // MTU for the interface
	Keepalive      int    // Persistent keepalive in seconds
	Port           int    // Listen port (default: 51820)
}

// AuditConfig holds audit logging configuration
type AuditConfig struct {
	SQLiteDSN     string // SQLite database path for audit logs
	HashSecrets   string // Comma-separated base64 secrets for hashing
	Async         bool   // Enable async audit buffering
	QueueSize     int    // Async queue size
	WorkerCount   int    // Number of async workers
	FlushInterval time.Duration
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string // Whitelist of allowed origins
	AllowCredentials bool
	MaxAge           time.Duration
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
	}

	// Validate critical fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate ensures all required configuration is present
func (c *Config) Validate() error {
	// Server validation
	if c.Server.Port == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}

	// Database validation
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
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
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", "goconnect"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
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
