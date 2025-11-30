package repository

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteChatRepository_CreateList(t *testing.T) {
	dir := t.TempDir()
	db, err := database.ConnectSQLite(filepath.Join(dir, "chat.db"))
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	repo := NewSQLiteChatRepository(db)
	msg := &domain.ChatMessage{
		Scope:    "network:net-1",
		TenantID: "tenant-1",
		UserID:   "user-1",
		Body:     "hello",
	}
	require.NoError(t, repo.Create(context.Background(), msg))

	list, _, err := repo.List(context.Background(), domain.ChatMessageFilter{
		Scope: "network:net-1",
		Limit: 10,
	})
	require.NoError(t, err)
	require.NotEmpty(t, list)
	assert.Equal(t, "hello", list[0].Body)
}
