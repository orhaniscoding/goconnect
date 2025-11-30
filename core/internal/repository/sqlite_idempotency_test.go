package repository

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteIdempotencyRepository_SaveGet(t *testing.T) {
	dir := t.TempDir()
	db, err := database.ConnectSQLite(filepath.Join(dir, "idempotency.db"))
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	repo := NewSQLiteIdempotencyRepository(db)
	err = repo.Save(context.Background(), "key-1", `{"ok":true}`, 200)
	require.NoError(t, err)

	body, status, found, err := repo.Get(context.Background(), "key-1")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, 200, status)
	assert.Contains(t, body, "ok")
}
