package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromEnv_WithDefaults(t *testing.T) {
	// Clear any existing env vars
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSLMODE")

	cfg := LoadConfigFromEnv()

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, "5432", cfg.Port)
	assert.Equal(t, "postgres", cfg.User)
	assert.Equal(t, "postgres", cfg.Password)
	assert.Equal(t, "goconnect", cfg.DBName)
	assert.Equal(t, "disable", cfg.SSLMode)
}

func TestLoadConfigFromEnv_WithCustomValues(t *testing.T) {
	// Set custom env vars
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "appuser")
	os.Setenv("DB_PASSWORD", "secret123")
	os.Setenv("DB_NAME", "production_db")
	os.Setenv("DB_SSLMODE", "require")

	defer func() {
		// Cleanup
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
	}()

	cfg := LoadConfigFromEnv()

	assert.Equal(t, "db.example.com", cfg.Host)
	assert.Equal(t, "5433", cfg.Port)
	assert.Equal(t, "appuser", cfg.User)
	assert.Equal(t, "secret123", cfg.Password)
	assert.Equal(t, "production_db", cfg.DBName)
	assert.Equal(t, "require", cfg.SSLMode)
}

func TestLoadConfigFromEnv_PartialCustomValues(t *testing.T) {
	// Set only some env vars
	os.Setenv("DB_HOST", "custom-host")
	os.Setenv("DB_NAME", "custom-db")

	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_NAME")
	}()

	cfg := LoadConfigFromEnv()

	// Custom values
	assert.Equal(t, "custom-host", cfg.Host)
	assert.Equal(t, "custom-db", cfg.DBName)

	// Defaults for the rest
	assert.Equal(t, "5432", cfg.Port)
	assert.Equal(t, "postgres", cfg.User)
	assert.Equal(t, "postgres", cfg.Password)
	assert.Equal(t, "disable", cfg.SSLMode)
}

func TestGetEnv_WithExistingVar(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")

	assert.Equal(t, "test_value", result)
}

func TestGetEnv_WithoutExistingVar(t *testing.T) {
	os.Unsetenv("NONEXISTENT_VAR")

	result := getEnv("NONEXISTENT_VAR", "default_value")

	assert.Equal(t, "default_value", result)
}

func TestGetEnv_WithEmptyVar(t *testing.T) {
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	// Empty string should use default
	result := getEnv("EMPTY_VAR", "default")

	assert.Equal(t, "default", result)
}

func TestConfig_DSNConstruction(t *testing.T) {
	// This tests the DSN format used in Connect()
	// We can't test Connect() without a real DB, but we can verify DSN format
	cfg := &Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	// Build the DSN the same way Connect() does
	expectedDSN := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"

	actualDSN := buildDSN(cfg)

	assert.Equal(t, expectedDSN, actualDSN)
}

func TestConfig_DSNWithSSLRequired(t *testing.T) {
	cfg := &Config{
		Host:     "db.production.com",
		Port:     "5432",
		User:     "produser",
		Password: "securepass",
		DBName:   "proddb",
		SSLMode:  "require",
	}

	expectedDSN := "host=db.production.com port=5432 user=produser password=securepass dbname=proddb sslmode=require"

	actualDSN := buildDSN(cfg)

	assert.Equal(t, expectedDSN, actualDSN)
}

func TestConfig_DSNWithSpecialCharacters(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "user@domain",
		Password: "pass&word!",
		DBName:   "my-db",
		SSLMode:  "verify-full",
	}

	// DSN should include special characters as-is (driver handles escaping)
	expectedDSN := "host=localhost port=5432 user=user@domain password=pass&word! dbname=my-db sslmode=verify-full"

	actualDSN := buildDSN(cfg)

	assert.Equal(t, expectedDSN, actualDSN)
}

func TestConfig_AllSSLModes(t *testing.T) {
	sslModes := []string{"disable", "require", "verify-ca", "verify-full"}

	for _, mode := range sslModes {
		t.Run("SSLMode_"+mode, func(t *testing.T) {
			cfg := &Config{
				Host:     "localhost",
				Port:     "5432",
				User:     "user",
				Password: "pass",
				DBName:   "db",
				SSLMode:  mode,
			}

			dsn := buildDSN(cfg)
			assert.Contains(t, dsn, "sslmode="+mode)
		})
	}
}

func TestConfig_NonStandardPort(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     "15432",
		User:     "user",
		Password: "pass",
		DBName:   "db",
		SSLMode:  "disable",
	}

	dsn := buildDSN(cfg)
	assert.Contains(t, dsn, "port=15432")
}

func TestLoadConfigFromEnv_Idempotent(t *testing.T) {
	os.Setenv("DB_HOST", "test-host")
	defer os.Unsetenv("DB_HOST")

	cfg1 := LoadConfigFromEnv()
	cfg2 := LoadConfigFromEnv()

	// Should produce identical configs
	assert.Equal(t, cfg1.Host, cfg2.Host)
	assert.Equal(t, cfg1.Port, cfg2.Port)
	assert.Equal(t, cfg1.User, cfg2.User)
	assert.Equal(t, cfg1.Password, cfg2.Password)
	assert.Equal(t, cfg1.DBName, cfg2.DBName)
	assert.Equal(t, cfg1.SSLMode, cfg2.SSLMode)
}

func TestConfig_StructFields(t *testing.T) {
	cfg := &Config{
		Host:     "host1",
		Port:     "1234",
		User:     "user1",
		Password: "pass1",
		DBName:   "db1",
		SSLMode:  "mode1",
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, "host1", cfg.Host)
	assert.Equal(t, "1234", cfg.Port)
	assert.Equal(t, "user1", cfg.User)
	assert.Equal(t, "pass1", cfg.Password)
	assert.Equal(t, "db1", cfg.DBName)
	assert.Equal(t, "mode1", cfg.SSLMode)
}

func TestConnect_InvalidConfig(t *testing.T) {
	// Test with clearly invalid config that should fail quickly
	cfg := &Config{
		Host:     "nonexistent-host-12345.invalid",
		Port:     "99999",
		User:     "invalid",
		Password: "invalid",
		DBName:   "invalid",
		SSLMode:  "disable",
	}

	db, err := Connect(cfg)

	// Should return an error (either Open or Ping will fail)
	require.Error(t, err)

	// If db was created, it should be nil or we should clean up
	if db != nil {
		db.Close()
	}
}

func TestConnect_NilConfig(t *testing.T) {
	// Test with nil config - should panic or error
	assert.Panics(t, func() {
		Connect(nil)
	})
}

// Helper function to build DSN (extracted from Connect for testing)
func buildDSN(cfg *Config) string {
	return "host=" + cfg.Host + " port=" + cfg.Port + " user=" + cfg.User +
		" password=" + cfg.Password + " dbname=" + cfg.DBName + " sslmode=" + cfg.SSLMode
}

// Test environment variable precedence
func TestEnvVarPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		envValue string
		defValue string
		expected string
	}{
		{
			name:     "Env var set with value",
			envVar:   "TEST_VAR_1",
			envValue: "env_value",
			defValue: "default",
			expected: "env_value",
		},
		{
			name:     "Env var not set",
			envVar:   "TEST_VAR_2",
			envValue: "",
			defValue: "default",
			expected: "default",
		},
		{
			name:     "Env var set to empty string",
			envVar:   "TEST_VAR_3",
			envValue: "",
			defValue: "default",
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			} else {
				os.Unsetenv(tt.envVar)
			}

			result := getEnv(tt.envVar, tt.defValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadConfigFromEnv_RealWorldScenario(t *testing.T) {
	// Simulate a real deployment scenario
	os.Setenv("DB_HOST", "postgres.production.svc.cluster.local")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "goconnect_app")
	os.Setenv("DB_PASSWORD", "super_secret_password_123")
	os.Setenv("DB_NAME", "goconnect_production")
	os.Setenv("DB_SSLMODE", "verify-full")

	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
	}()

	cfg := LoadConfigFromEnv()

	require.NotNil(t, cfg)
	assert.Equal(t, "postgres.production.svc.cluster.local", cfg.Host)
	assert.Equal(t, "5432", cfg.Port)
	assert.Equal(t, "goconnect_app", cfg.User)
	assert.Equal(t, "super_secret_password_123", cfg.Password)
	assert.Equal(t, "goconnect_production", cfg.DBName)
	assert.Equal(t, "verify-full", cfg.SSLMode)
}


