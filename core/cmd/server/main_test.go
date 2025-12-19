package main

import (
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestBuildBaseURL(t *testing.T) {
	t.Run("Development", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				Environment: "development",
				Host:        "localhost",
				Port:        "8080",
			},
		}
		assert.Equal(t, "http://localhost:8080", buildBaseURL(cfg))
	})

	t.Run("Production", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				Environment: "production",
				Host:        "api.example.com",
				Port:        "443",
			},
		}
		assert.Equal(t, "https://api.example.com:443", buildBaseURL(cfg))
	})

	t.Run("Zero Host", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				Environment: "development",
				Host:        "0.0.0.0",
				Port:        "8080",
			},
		}
		assert.Equal(t, "http://localhost:8080", buildBaseURL(cfg))
	})
}

func TestGetMigrationsPath(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) {
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Backend: "sqlite",
			},
		}
		// Since we are in cmd/server, migrations_sqlite is usually ../../migrations_sqlite
		path := getMigrationsPath(cfg)
		assert.Contains(t, []string{"migrations_sqlite", "migrations_sqlite"}, path)
	})

	t.Run("Postgres", func(t *testing.T) {
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Backend: "postgres",
			},
		}
		path := getMigrationsPath(cfg)
		assert.Equal(t, "migrations", path)
	})

	t.Run("Memory", func(t *testing.T) {
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Backend: "memory",
			},
		}
		path := getMigrationsPath(cfg)
		assert.Equal(t, "", path)
	})
}

func TestInitDatabase(t *testing.T) {
	t.Run("Memory", func(t *testing.T) {
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Backend: "memory",
			},
		}
		db, err := initDatabase(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, db)
		db.Close()
	})

	t.Run("PostgresError", func(t *testing.T) {
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Backend: "postgres",
				Host:    "localhost",
				Port:    "1", // fail
			},
		}
		db, err := initDatabase(cfg)
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

func TestInitRepositories(t *testing.T) {
	// Use memory backend for testing
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Backend: "memory",
		},
	}
	db, _ := initDatabase(cfg)
	defer db.Close()

	repos := initRepositories(db, cfg)
	assert.NotNil(t, repos)
	assert.NotNil(t, repos.User)
	assert.NotNil(t, repos.Tenant)
}

func TestInitServices(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Backend: "memory",
		},
		JWT: config.JWTConfig{
			Secret: "test-secret",
		},
		Audit: config.AuditConfig{
			SQLiteDSN: "", // stdout
		},
	}
	db, _ := initDatabase(cfg)
	defer db.Close()
	repos := initRepositories(db, cfg)

	svcs, auditor := initServices(repos, cfg)
	assert.NotNil(t, svcs)
	assert.NotNil(t, svcs.Auth)
	assert.NotNil(t, auditor)
}

func TestInitHandlers(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Backend: "memory",
		},
		JWT: config.JWTConfig{
			Secret: "test-secret",
		},
		WireGuard: config.WireGuardConfig{
			ServerEndpoint: "1.2.3.4:51820",
		},
	}
	db, _ := initDatabase(cfg)
	defer db.Close()
	repos := initRepositories(db, cfg)
	svcs, auditor := initServices(repos, cfg)

	handlers := initHandlers(svcs, repos, cfg, auditor)
	assert.NotNil(t, handlers)
	assert.NotNil(t, handlers.Auth)
}

func TestSetupRouter(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Backend: "memory",
		},
		JWT: config.JWTConfig{
			Secret: "test-secret",
		},
	}
	db, _ := initDatabase(cfg)
	defer db.Close()
	repos := initRepositories(db, cfg)
	svcs, auditor := initServices(repos, cfg)
	handlers := initHandlers(svcs, repos, cfg, auditor)

	router := setupRouter(cfg, handlers, svcs, repos)
	assert.NotNil(t, router)
}
