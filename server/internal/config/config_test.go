package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Save original env
	originalEnv := make(map[string]string)
	envVars := []string{
		"SERVER_PORT", "DB_HOST", "DB_NAME", "DB_USER",
		"JWT_SECRET", "WG_SERVER_ENDPOINT", "WG_SERVER_PUBKEY",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}
	defer func() {
		for key, value := range originalEnv {
			os.Setenv(key, value)
		}
	}()

	t.Run("Success - valid configuration", func(t *testing.T) {
		os.Setenv("SERVER_PORT", "8080")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_USER", "testuser")
		os.Setenv("JWT_SECRET", "this_is_a_very_secure_secret_key_with_at_least_32_chars")
		os.Setenv("WG_SERVER_ENDPOINT", "vpn.test.com:51820")
		os.Setenv("WG_SERVER_PUBKEY", "aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789ABCDEFG=")

		cfg, err := Load()

		require.NoError(t, err)
		assert.Equal(t, "8080", cfg.Server.Port)
		assert.Equal(t, "localhost", cfg.Database.Host)
		assert.Equal(t, "testdb", cfg.Database.DBName)
		assert.Equal(t, "vpn.test.com:51820", cfg.WireGuard.ServerEndpoint)
		assert.Len(t, cfg.WireGuard.ServerPubKey, 44)
	})

	t.Run("Validation - missing SERVER_PORT", func(t *testing.T) {
		// Clear all env vars
		for _, key := range envVars {
			os.Setenv(key, "")
		}

		// Set only required vars except SERVER_PORT (will use default "8080")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_USER", "testuser")
		os.Setenv("JWT_SECRET", "this_is_a_very_secure_secret_key_with_at_least_32_chars")
		os.Setenv("WG_SERVER_ENDPOINT", "vpn.test.com:51820")
		os.Setenv("WG_SERVER_PUBKEY", "aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789ABCDEFG=")

		cfg, err := Load()

		// Should succeed with default port
		require.NoError(t, err)
		assert.Equal(t, "8080", cfg.Server.Port) // Default value
	})

	t.Run("Validation - missing DB_HOST", func(t *testing.T) {
		// Clear all env vars
		for _, key := range envVars {
			os.Setenv(key, "")
		}

		// Set only required vars except DB_HOST (will use default "localhost")
		os.Setenv("SERVER_PORT", "8080")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_USER", "testuser")
		os.Setenv("JWT_SECRET", "this_is_a_very_secure_secret_key_with_at_least_32_chars")
		os.Setenv("WG_SERVER_ENDPOINT", "vpn.test.com:51820")
		os.Setenv("WG_SERVER_PUBKEY", "aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789ABCDEFG=")

		cfg, err := Load()

		// Should succeed with default host
		require.NoError(t, err)
		assert.Equal(t, "localhost", cfg.Database.Host) // Default value
	})

	t.Run("Validation - JWT_SECRET too short", func(t *testing.T) {
		os.Setenv("SERVER_PORT", "8080")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_USER", "testuser")
		os.Setenv("JWT_SECRET", "short")
		os.Setenv("WG_SERVER_ENDPOINT", "vpn.test.com:51820")
		os.Setenv("WG_SERVER_PUBKEY", "aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789ABCDEFG=")

		_, err := Load()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "JWT_SECRET must be at least 32 characters")
	})

	t.Run("Validation - WG_SERVER_PUBKEY invalid length", func(t *testing.T) {
		os.Setenv("SERVER_PORT", "8080")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_USER", "testuser")
		os.Setenv("JWT_SECRET", "this_is_a_very_secure_secret_key_with_at_least_32_chars")
		os.Setenv("WG_SERVER_ENDPOINT", "vpn.test.com:51820")
		os.Setenv("WG_SERVER_PUBKEY", "tooshort")

		_, err := Load()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be exactly 44 characters")
	})
}

func TestDatabaseConfig_ConnectionString(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	connStr := cfg.ConnectionString()

	assert.Contains(t, connStr, "host=localhost")
	assert.Contains(t, connStr, "port=5432")
	assert.Contains(t, connStr, "user=testuser")
	assert.Contains(t, connStr, "password=testpass")
	assert.Contains(t, connStr, "dbname=testdb")
	assert.Contains(t, connStr, "sslmode=disable")
}

func TestServerConfig_Address(t *testing.T) {
	cfg := ServerConfig{
		Host: "0.0.0.0",
		Port: "8080",
	}

	assert.Equal(t, "0.0.0.0:8080", cfg.Address())
}

func TestServerConfig_Environment(t *testing.T) {
	t.Run("Development", func(t *testing.T) {
		cfg := ServerConfig{Environment: "development"}
		assert.True(t, cfg.IsDevelopment())
		assert.False(t, cfg.IsProduction())
	})

	t.Run("Production", func(t *testing.T) {
		cfg := ServerConfig{Environment: "production"}
		assert.False(t, cfg.IsDevelopment())
		assert.True(t, cfg.IsProduction())
	})
}

func TestGetEnvHelpers(t *testing.T) {
	t.Run("getIntEnv", func(t *testing.T) {
		os.Setenv("TEST_INT", "42")
		defer os.Unsetenv("TEST_INT")

		val := getIntEnv("TEST_INT", 10)
		assert.Equal(t, 42, val)

		val = getIntEnv("NON_EXISTENT", 10)
		assert.Equal(t, 10, val)
	})

	t.Run("getBoolEnv", func(t *testing.T) {
		os.Setenv("TEST_BOOL", "true")
		defer os.Unsetenv("TEST_BOOL")

		val := getBoolEnv("TEST_BOOL", false)
		assert.True(t, val)

		val = getBoolEnv("NON_EXISTENT", false)
		assert.False(t, val)
	})

	t.Run("getDurationEnv", func(t *testing.T) {
		os.Setenv("TEST_DURATION", "30s")
		defer os.Unsetenv("TEST_DURATION")

		val := getDurationEnv("TEST_DURATION", 10*time.Second)
		assert.Equal(t, 30*time.Second, val)

		val = getDurationEnv("NON_EXISTENT", 10*time.Second)
		assert.Equal(t, 10*time.Second, val)
	})
}

func TestSplitAndTrim(t *testing.T) {
	result := splitAndTrim("http://localhost:3000, http://localhost:5173 , http://example.com", ",")

	assert.Len(t, result, 3)
	assert.Equal(t, "http://localhost:3000", result[0])
	assert.Equal(t, "http://localhost:5173", result[1])
	assert.Equal(t, "http://example.com", result[2])
}
