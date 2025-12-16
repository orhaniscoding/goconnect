package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite" // SQLite driver (CGO-free); encryption to be added in follow-up
)

// ConnectSQLite opens or creates a SQLite database at the given path.
// Future work: add SQLCipher/field-level encryption and key management.
func ConnectSQLite(path string) (*sql.DB, error) {
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "" {
		return nil, fmt.Errorf("sqlite path is required")
	}

	if dir := filepath.Dir(cleanPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, fmt.Errorf("create sqlite directory: %w", err)
		}
	}

	// busy_timeout to avoid immediate "database is locked", enable foreign keys by default
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)", cleanPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite is file-based; a single writer is safest
	db.SetMaxIdleConns(1) // Keep one handle warm
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	return db, nil
}

// RunSQLiteMigrations is a placeholder to wire golang-migrate with SQLite.
func RunSQLiteMigrations(db *sql.DB, migrationsPath string) error {
	if db == nil {
		return fmt.Errorf("sqlite migrations: db is nil")
	}
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("sqlite migrations: resolve path: %w", err)
	}
	normalizedPath := filepath.ToSlash(absPath)

	driver, err := sqlite.WithInstance(db, &sqlite.Config{
		NoTxWrap: false,
	})
	if err != nil {
		return fmt.Errorf("sqlite migrations: driver init: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", normalizedPath),
		"sqlite",
		driver,
	)
	if err != nil {
		return fmt.Errorf("sqlite migrations: instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("sqlite migrations: up: %w", err)
	}

	slog.Info("SQLite migrations applied", "path", migrationsPath)
	return nil
}
