package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func baseValidConfig() Config {
	return Config{
		Server: ServerConfig{Host: "0.0.0.0", Port: "8080"},
		Database: DatabaseConfig{
			Backend:    "sqlite",
			SQLitePath: "/tmp/goconnect.db",
		},
		JWT: JWTConfig{Secret: "this_is_a_very_secure_secret_key_with_at_least_32_chars"},
		WireGuard: WireGuardConfig{
			ServerEndpoint: "vpn.test.com:51820",
			ServerPubKey:   "aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789ABCDEFG=",
		},
	}
}

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

func TestConfigValidate_BackendVariants(t *testing.T) {
	base := func() *Config {
		return &Config{
			Server: ServerConfig{Port: "8080"},
			Database: DatabaseConfig{
				Backend:    "postgres",
				Host:       "localhost",
				Port:       "5432",
				User:       "user",
				DBName:     "db",
				SSLMode:    "disable",
				SQLitePath: "data/test.db",
			},
			JWT: JWTConfig{Secret: "this_is_a_very_secure_secret_key_with_at_least_32_chars"},
			WireGuard: WireGuardConfig{
				ServerEndpoint: "vpn.test.com:51820",
				ServerPubKey:   "aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789ABCDEFG=",
			},
		}
	}

	t.Run("postgres requires host", func(t *testing.T) {
		cfg := base()
		cfg.Database.Host = ""
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DB_HOST")
	})

	t.Run("sqlite requires path", func(t *testing.T) {
		cfg := base()
		cfg.Database.Backend = "sqlite"
		cfg.Database.SQLitePath = ""
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DB_SQLITE_PATH")
	})

	t.Run("memory skips db requirements", func(t *testing.T) {
		cfg := base()
		cfg.Database.Backend = "memory"
		cfg.Database.Host = ""
		cfg.Database.User = ""
		cfg.Database.DBName = ""
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid backend fails fast", func(t *testing.T) {
		cfg := base()
		cfg.Database.Backend = "mongo"
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid DB_BACKEND")
	})

	t.Run("postgres requires dbname", func(t *testing.T) {
		cfg := base()
		cfg.Database.DBName = ""
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DB_NAME")
	})

	t.Run("postgres requires user", func(t *testing.T) {
		cfg := base()
		cfg.Database.User = ""
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DB_USER")
	})

	t.Run("jwt secret too short", func(t *testing.T) {
		cfg := base()
		cfg.JWT.Secret = "tooshort"
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 32")
	})

	t.Run("jwt secret missing", func(t *testing.T) {
		cfg := base()
		cfg.JWT.Secret = ""
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "JWT_SECRET")
	})

	t.Run("wireguard endpoint missing", func(t *testing.T) {
		cfg := base()
		cfg.WireGuard.ServerEndpoint = ""
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "WG_SERVER_ENDPOINT")
	})

	t.Run("wireguard pubkey missing", func(t *testing.T) {
		cfg := base()
		cfg.WireGuard.ServerPubKey = ""
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "WG_SERVER_PUBKEY")
	})

	t.Run("wireguard pubkey wrong length", func(t *testing.T) {
		cfg := base()
		cfg.WireGuard.ServerPubKey = "tooshort"
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "44 characters")
	})

	t.Run("server port missing", func(t *testing.T) {
		cfg := base()
		cfg.Server.Port = ""
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SERVER_PORT")
	})
}

func TestLoadFromFileOrEnv_WithFileAndEnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "goconnect.yaml")
	yamlContent := `
server:
  host: "127.0.0.1"
  port: "9090"
database:
  backend: "sqlite"
  sqlite_path: "/tmp/test.db"
jwt:
  secret: "this_is_a_very_secure_secret_key_with_at_least_32_chars"
wireguard:
  server_endpoint: "vpn.test.com:51820"
  server_pubkey: "aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789ABCDEFG="
`
	require.NoError(t, os.WriteFile(configPath, []byte(yamlContent), 0o600))

	// Env override for server port
	os.Setenv("SERVER_PORT", "9999")
	defer os.Unsetenv("SERVER_PORT")

	cfg, err := LoadFromFileOrEnv(configPath)
	require.NoError(t, err)
	assert.Equal(t, "9999", cfg.Server.Port)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, "sqlite", cfg.Database.Backend)
	assert.Equal(t, "/tmp/test.db", cfg.Database.SQLitePath)
}

func TestSaveToFileAndReload(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "saved.yaml")

	cfg := baseValidConfig()
	require.NoError(t, SaveToFile(&cfg, configPath))

	reloaded, err := LoadFromFileOrEnv(configPath)
	require.NoError(t, err)
	assert.Equal(t, cfg.Database.Backend, reloaded.Database.Backend)
	assert.Equal(t, cfg.Database.SQLitePath, reloaded.Database.SQLitePath)
	assert.Equal(t, cfg.WireGuard.ServerPubKey, reloaded.WireGuard.ServerPubKey)
	assert.Equal(t, cfg.Server.Port, reloaded.Server.Port)
}

func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "onlyfile.yaml")
	cfg := baseValidConfig()
	require.NoError(t, SaveToFile(&cfg, configPath))

	loaded, err := LoadFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, cfg.Database.Backend, loaded.Database.Backend)
	assert.Equal(t, cfg.Database.SQLitePath, loaded.Database.SQLitePath)
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.yaml")
	require.Error(t, err)
}

func TestSaveToFile_InvalidPath(t *testing.T) {
	cfg := baseValidConfig()
	err := SaveToFile(&cfg, "/nonexistent/directory/config.yaml")
	require.Error(t, err)
}

func TestDefaultConfigPath(t *testing.T) {
	t.Run("DefaultPath", func(t *testing.T) {
		os.Unsetenv("GOCONNECT_CONFIG_PATH")
		path := DefaultConfigPath()
		assert.Equal(t, "goconnect.yaml", path)
	})

	t.Run("EnvOverride", func(t *testing.T) {
		os.Setenv("GOCONNECT_CONFIG_PATH", "/custom/path/config.yaml")
		defer os.Unsetenv("GOCONNECT_CONFIG_PATH")
		path := DefaultConfigPath()
		assert.Equal(t, "/custom/path/config.yaml", path)
	})
}

func TestApplyEnvOverrides(t *testing.T) {
	t.Run("OverrideServerPort", func(t *testing.T) {
		os.Setenv("SERVER_PORT", "3000")
		defer os.Unsetenv("SERVER_PORT")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "3000", cfg.Server.Port)
	})

	t.Run("OverrideServerHost", func(t *testing.T) {
		os.Setenv("SERVER_HOST", "0.0.0.0")
		defer os.Unsetenv("SERVER_HOST")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	})

	t.Run("OverrideJWTSecret", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "supersecret12345678901234567890ab") // 32+ chars
		defer os.Unsetenv("JWT_SECRET")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "supersecret12345678901234567890ab", cfg.JWT.Secret)
	})

	t.Run("OverrideDatabaseBackend", func(t *testing.T) {
		os.Setenv("DB_BACKEND", "postgres")
		defer os.Unsetenv("DB_BACKEND")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "postgres", cfg.Database.Backend)
	})

	t.Run("OverrideDatabaseHost", func(t *testing.T) {
		os.Setenv("DB_HOST", "db.example.com")
		defer os.Unsetenv("DB_HOST")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "db.example.com", cfg.Database.Host)
	})
}

func TestLookupEnv(t *testing.T) {
	t.Run("ExistingVar", func(t *testing.T) {
		os.Setenv("TEST_LOOKUP_VAR", "testvalue")
		defer os.Unsetenv("TEST_LOOKUP_VAR")

		val, ok := lookupEnv("TEST_LOOKUP_VAR")
		assert.True(t, ok)
		assert.Equal(t, "testvalue", val)
	})

	t.Run("NonExistingVar", func(t *testing.T) {
		os.Unsetenv("NONEXISTENT_LOOKUP_VAR")

		val, ok := lookupEnv("NONEXISTENT_LOOKUP_VAR")
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("EmptyVar", func(t *testing.T) {
		os.Setenv("EMPTY_LOOKUP_VAR", "")
		defer os.Unsetenv("EMPTY_LOOKUP_VAR")

		val, ok := lookupEnv("EMPTY_LOOKUP_VAR")
		assert.True(t, ok) // lookupEnv returns true even for empty
		assert.Empty(t, val)
	})
}

func TestLookupEnvNonEmpty(t *testing.T) {
	t.Run("ExistingNonEmptyVar", func(t *testing.T) {
		os.Setenv("TEST_NONEMPTY_VAR", "value")
		defer os.Unsetenv("TEST_NONEMPTY_VAR")

		val, ok := lookupEnvNonEmpty("TEST_NONEMPTY_VAR")
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	})

	t.Run("NonExistingVar", func(t *testing.T) {
		os.Unsetenv("NONEXISTENT_NONEMPTY_VAR")

		val, ok := lookupEnvNonEmpty("NONEXISTENT_NONEMPTY_VAR")
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("EmptyVar", func(t *testing.T) {
		os.Setenv("EMPTY_NONEMPTY_VAR", "")
		defer os.Unsetenv("EMPTY_NONEMPTY_VAR")

		val, ok := lookupEnvNonEmpty("EMPTY_NONEMPTY_VAR")
		assert.False(t, ok) // lookupEnvNonEmpty returns false for empty
		assert.Empty(t, val)
	})
}

func TestApplyEnvOverrides_MoreVars(t *testing.T) {
	t.Run("OverrideWireGuardEndpoint", func(t *testing.T) {
		os.Setenv("WG_SERVER_ENDPOINT", "wg.example.com:51820")
		defer os.Unsetenv("WG_SERVER_ENDPOINT")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "wg.example.com:51820", cfg.WireGuard.ServerEndpoint)
	})

	t.Run("OverrideWireGuardPubKey", func(t *testing.T) {
		os.Setenv("WG_SERVER_PUBKEY", "newpubkey123456789012345678901234567890AB=")
		defer os.Unsetenv("WG_SERVER_PUBKEY")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "newpubkey123456789012345678901234567890AB=", cfg.WireGuard.ServerPubKey)
	})

	t.Run("OverrideRedisHost", func(t *testing.T) {
		os.Setenv("REDIS_HOST", "redis.example.com")
		defer os.Unsetenv("REDIS_HOST")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "redis.example.com", cfg.Redis.Host)
	})

	t.Run("OverrideRedisPort", func(t *testing.T) {
		os.Setenv("REDIS_PORT", "6380")
		defer os.Unsetenv("REDIS_PORT")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "6380", cfg.Redis.Port)
	})

	t.Run("OverrideRedisDB", func(t *testing.T) {
		os.Setenv("REDIS_DB", "5")
		defer os.Unsetenv("REDIS_DB")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, 5, cfg.Redis.DB)
	})

	t.Run("OverrideRedisDB_InvalidValue", func(t *testing.T) {
		os.Setenv("REDIS_DB", "invalid")
		defer os.Unsetenv("REDIS_DB")
		cfg := baseValidConfig()
		cfg.Redis.DB = 0 // Set initial value
		applyEnvOverrides(&cfg)
		assert.Equal(t, 0, cfg.Redis.DB) // Should remain unchanged
	})
}

func TestApplyEnvOverrides_ServerTimeouts(t *testing.T) {
	t.Run("OverrideReadTimeout", func(t *testing.T) {
		os.Setenv("SERVER_READ_TIMEOUT", "30s")
		defer os.Unsetenv("SERVER_READ_TIMEOUT")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	})

	t.Run("OverrideWriteTimeout", func(t *testing.T) {
		os.Setenv("SERVER_WRITE_TIMEOUT", "45s")
		defer os.Unsetenv("SERVER_WRITE_TIMEOUT")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, 45*time.Second, cfg.Server.WriteTimeout)
	})

	t.Run("OverrideIdleTimeout", func(t *testing.T) {
		os.Setenv("SERVER_IDLE_TIMEOUT", "120s")
		defer os.Unsetenv("SERVER_IDLE_TIMEOUT")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, 120*time.Second, cfg.Server.IdleTimeout)
	})

	t.Run("OverrideEnvironment", func(t *testing.T) {
		os.Setenv("ENVIRONMENT", "production")
		defer os.Unsetenv("ENVIRONMENT")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "production", cfg.Server.Environment)
	})

	t.Run("InvalidReadTimeout", func(t *testing.T) {
		os.Setenv("SERVER_READ_TIMEOUT", "invalid")
		defer os.Unsetenv("SERVER_READ_TIMEOUT")
		cfg := baseValidConfig()
		cfg.Server.ReadTimeout = 10 * time.Second
		applyEnvOverrides(&cfg)
		assert.Equal(t, 10*time.Second, cfg.Server.ReadTimeout) // Should remain unchanged
	})
}

func TestApplyEnvOverrides_DatabaseExtended(t *testing.T) {
	t.Run("OverrideDatabasePort", func(t *testing.T) {
		os.Setenv("DB_PORT", "5433")
		defer os.Unsetenv("DB_PORT")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "5433", cfg.Database.Port)
	})

	t.Run("OverrideDatabaseUser", func(t *testing.T) {
		os.Setenv("DB_USER", "appuser")
		defer os.Unsetenv("DB_USER")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "appuser", cfg.Database.User)
	})

	t.Run("OverrideDatabasePassword", func(t *testing.T) {
		os.Setenv("DB_PASSWORD", "secret123")
		defer os.Unsetenv("DB_PASSWORD")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "secret123", cfg.Database.Password)
	})

	t.Run("OverrideDatabaseName", func(t *testing.T) {
		os.Setenv("DB_NAME", "mydb")
		defer os.Unsetenv("DB_NAME")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "mydb", cfg.Database.DBName)
	})

	t.Run("OverrideDatabaseSSLMode", func(t *testing.T) {
		os.Setenv("DB_SSLMODE", "require")
		defer os.Unsetenv("DB_SSLMODE")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "require", cfg.Database.SSLMode)
	})

	t.Run("OverrideMaxOpenConns", func(t *testing.T) {
		os.Setenv("DB_MAX_OPEN_CONNS", "50")
		defer os.Unsetenv("DB_MAX_OPEN_CONNS")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, 50, cfg.Database.MaxOpenConns)
	})

	t.Run("OverrideMaxIdleConns", func(t *testing.T) {
		os.Setenv("DB_MAX_IDLE_CONNS", "10")
		defer os.Unsetenv("DB_MAX_IDLE_CONNS")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, 10, cfg.Database.MaxIdleConns)
	})

	t.Run("OverrideConnMaxLifetime", func(t *testing.T) {
		os.Setenv("DB_CONN_MAX_LIFETIME", "5m")
		defer os.Unsetenv("DB_CONN_MAX_LIFETIME")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, 5*time.Minute, cfg.Database.ConnMaxLifetime)
	})

	t.Run("OverrideSQLitePath", func(t *testing.T) {
		os.Setenv("DB_SQLITE_PATH", "/data/app.db")
		defer os.Unsetenv("DB_SQLITE_PATH")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "/data/app.db", cfg.Database.SQLitePath)
	})

	t.Run("InvalidMaxOpenConns", func(t *testing.T) {
		os.Setenv("DB_MAX_OPEN_CONNS", "invalid")
		defer os.Unsetenv("DB_MAX_OPEN_CONNS")
		cfg := baseValidConfig()
		cfg.Database.MaxOpenConns = 25
		applyEnvOverrides(&cfg)
		assert.Equal(t, 25, cfg.Database.MaxOpenConns) // Should remain unchanged
	})
}

func TestApplyEnvOverrides_JWTExtended(t *testing.T) {
	t.Run("OverrideAccessTTL", func(t *testing.T) {
		os.Setenv("JWT_ACCESS_TTL", "1h")
		defer os.Unsetenv("JWT_ACCESS_TTL")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, time.Hour, cfg.JWT.AccessTokenTTL)
	})

	t.Run("OverrideRefreshTTL", func(t *testing.T) {
		os.Setenv("JWT_REFRESH_TTL", "168h")
		defer os.Unsetenv("JWT_REFRESH_TTL")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, 168*time.Hour, cfg.JWT.RefreshTokenTTL)
	})

	t.Run("OverrideRefreshSecret", func(t *testing.T) {
		os.Setenv("JWT_REFRESH_SECRET", "refresh_secret_key")
		defer os.Unsetenv("JWT_REFRESH_SECRET")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "refresh_secret_key", cfg.JWT.RefreshSecretKey)
	})

	t.Run("InvalidAccessTTL", func(t *testing.T) {
		os.Setenv("JWT_ACCESS_TTL", "invalid")
		defer os.Unsetenv("JWT_ACCESS_TTL")
		cfg := baseValidConfig()
		cfg.JWT.AccessTokenTTL = 15 * time.Minute
		applyEnvOverrides(&cfg)
		assert.Equal(t, 15*time.Minute, cfg.JWT.AccessTokenTTL) // Should remain unchanged
	})
}

func TestApplyEnvOverrides_WireGuardExtended(t *testing.T) {
	t.Run("OverrideInterfaceName", func(t *testing.T) {
		os.Setenv("WG_INTERFACE_NAME", "wg1")
		defer os.Unsetenv("WG_INTERFACE_NAME")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "wg1", cfg.WireGuard.InterfaceName)
	})

	t.Run("OverridePrivateKey", func(t *testing.T) {
		os.Setenv("WG_PRIVATE_KEY", "privatekey1234567890123456789012345678901234=")
		defer os.Unsetenv("WG_PRIVATE_KEY")
		cfg := baseValidConfig()
		applyEnvOverrides(&cfg)
		assert.Equal(t, "privatekey1234567890123456789012345678901234=", cfg.WireGuard.PrivateKey)
	})
}

func TestLoadFromFileOrEnv_FileNotFound(t *testing.T) {
	// Test LoadFromFileOrEnv when file doesn't exist
	// It should fall back to Load()

	// Set required env vars for Load() to succeed
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("JWT_SECRET", "this_is_a_very_secure_secret_key_with_at_least_32_chars")
	os.Setenv("WG_SERVER_ENDPOINT", "vpn.test.com:51820")
	os.Setenv("WG_SERVER_PUBKEY", "aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789ABCDEFG=")

	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_USER")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("WG_SERVER_ENDPOINT")
		os.Unsetenv("WG_SERVER_PUBKEY")
	}()

	cfg, err := LoadFromFileOrEnv("/nonexistent/path/config.yaml")
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Server.Port)
}

func TestLoadFromFileOrEnv_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	require.NoError(t, os.WriteFile(configPath, []byte("{{invalid yaml content"), 0o600))

	_, err := LoadFromFileOrEnv(configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestLoadFromFileOrEnv_ValidationFailure(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid_config.yaml")

	// Write valid YAML but invalid config (missing required fields)
	yamlContent := `
server:
  port: ""
database:
  backend: "postgres"
`
	require.NoError(t, os.WriteFile(configPath, []byte(yamlContent), 0o600))

	_, err := LoadFromFileOrEnv(configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SERVER_PORT")
}

func TestLoadFromFile_IsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to load a directory as a config file
	_, err := LoadFromFile(tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSaveToFile_NilConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nil.yaml")

	err := SaveToFile(nil, configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestSaveToFile_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Config with invalid values
	cfg := &Config{
		Server: ServerConfig{Port: ""}, // Missing required field
	}

	err := SaveToFile(cfg, configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SERVER_PORT")
}

func TestBackendOrDefault(t *testing.T) {
	t.Run("PostgresBackend", func(t *testing.T) {
		cfg := DatabaseConfig{Backend: "postgres"}
		assert.Equal(t, "postgres", cfg.backendOrDefault())
	})

	t.Run("SQLiteBackend", func(t *testing.T) {
		cfg := DatabaseConfig{Backend: "sqlite"}
		assert.Equal(t, "sqlite", cfg.backendOrDefault())
	})

	t.Run("MemoryBackend", func(t *testing.T) {
		cfg := DatabaseConfig{Backend: "memory"}
		assert.Equal(t, "memory", cfg.backendOrDefault())
	})

	t.Run("EmptyBackend_DefaultsToPostgres", func(t *testing.T) {
		cfg := DatabaseConfig{Backend: ""}
		assert.Equal(t, "postgres", cfg.backendOrDefault())
	})
}
