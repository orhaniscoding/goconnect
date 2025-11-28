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
