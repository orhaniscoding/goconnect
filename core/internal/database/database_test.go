package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	key := "TEST_DB_ENV_VAR"
	def := "default_val"
	
	// Test default
	assert.Equal(t, def, getEnv(key, def))
	
	// Test set val
	os.Setenv(key, "real_val")
	defer os.Unsetenv(key)
	assert.Equal(t, "real_val", getEnv(key, def))
}

func TestLoadConfigFromEnv(t *testing.T) {
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "9999")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
	}()
	
	cfg := LoadConfigFromEnv()
	assert.Equal(t, "db.example.com", cfg.Host)
	assert.Equal(t, "9999", cfg.Port)
	assert.Equal(t, "postgres", cfg.User) // default
}

func TestConnect_Error(t *testing.T) {
	// Test failure to connect with junk config
	cfg := &Config{
		Host: "localhost",
		Port: "1", // invalid port
	}
	db, err := Connect(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestConnectSQLite(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	t.Run("Success", func(t *testing.T) {
		db, err := ConnectSQLite(dbPath)
		assert.NoError(t, err)
		assert.NotNil(t, db)
		db.Close()
	})

	t.Run("EmptyPath", func(t *testing.T) {
		db, err := ConnectSQLite(" ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sqlite path is required")
		assert.Nil(t, db)
	})

	t.Run("DirectoryCreation", func(t *testing.T) {
		nestedPath := filepath.Join(tmpDir, "some", "nested", "dir", "test.db")
		db, err := ConnectSQLite(nestedPath)
		assert.NoError(t, err)
		assert.NotNil(t, db)
		db.Close()
		
		info, err := os.Stat(filepath.Dir(nestedPath))
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

func TestRunSQLiteMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "migrate.db")
	db, _ := ConnectSQLite(dbPath)
	defer db.Close()

	// Create dummy migrations
	migrationsPath := filepath.Join(tmpDir, "migrations")
	err := os.MkdirAll(migrationsPath, 0755)
	assert.NoError(t, err)

	upSQL := "CREATE TABLE accounts (id INTEGER PRIMARY KEY, name TEXT);"
	err = os.WriteFile(filepath.Join(migrationsPath, "001_init.up.sql"), []byte(upSQL), 0644)
	assert.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		err := RunSQLiteMigrations(db, migrationsPath)
		assert.NoError(t, err)

		// Verify table exists
		var name string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='accounts'").Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, "accounts", name)
	})

	t.Run("NilDB", func(t *testing.T) {
		err := RunSQLiteMigrations(nil, migrationsPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db is nil")
	})

	t.Run("InvalidPath", func(t *testing.T) {
		err := RunSQLiteMigrations(db, "/non/existent/path")
		assert.Error(t, err)
	})
}

func TestNewRedisClient(t *testing.T) {
	t.Run("ConnectionError", func(t *testing.T) {
		cfg := config.RedisConfig{
			Host: "localhost",
			Port: "1", // invalid port
		}

		client, err := NewRedisClient(cfg)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}
