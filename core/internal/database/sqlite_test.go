package database

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectSQLite_EmptyPath(t *testing.T) {
	db, err := ConnectSQLite("")
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestConnectSQLite_CreatesDirectoryAndOpens(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nested", "test.db")

	db, err := ConnectSQLite(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Ensure we can ping and the file path was created
	assert.NoError(t, db.Ping())
}

func TestRunSQLiteMigrations_Placeholder(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "migrations.db")

	db, err := ConnectSQLite(dbPath)
	require.NoError(t, err)
	defer db.Close()

	err = RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite"))
	assert.NoError(t, err)
}

func TestConnectSQLite_WhitespacePath(t *testing.T) {
	db, err := ConnectSQLite("   ")
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "sqlite path is required")
}

func TestConnectSQLite_Success(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := ConnectSQLite(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Verify connection works
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestRunSQLiteMigrations_NilDB(t *testing.T) {
	err := RunSQLiteMigrations(nil, "/some/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db is nil")
}

func TestConnectSQLite_ForeignKeysEnabled(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "fk_test.db")

	db, err := ConnectSQLite(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Check foreign_keys pragma
	var fkEnabled int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	require.NoError(t, err)
	assert.Equal(t, 1, fkEnabled, "foreign_keys should be enabled")
}

func TestConnectSQLite_CanExecuteQueries(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "query_test.db")

	db, err := ConnectSQLite(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create a table
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	// Insert data
	_, err = db.Exec("INSERT INTO test (name) VALUES (?)", "test_value")
	require.NoError(t, err)

	// Query data
	var name string
	err = db.QueryRow("SELECT name FROM test WHERE id = 1").Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "test_value", name)
}

func TestRunSQLiteMigrations_InvalidPath(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := ConnectSQLite(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Invalid migrations path
	err = RunSQLiteMigrations(db, "/nonexistent/migrations/path")
	assert.Error(t, err)
}
