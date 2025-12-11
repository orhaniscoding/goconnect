package sqlite

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteStore(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Test New
	s, err := New(dbPath)
	require.NoError(t, err)
	defer s.Close()

	// Test Migrate
	err = s.Migrate()
	require.NoError(t, err)

	ctx := context.Background()

	// Test Node Operations
	t.Run("Node Operations", func(t *testing.T) {
		node := &store.Node{
			ID:         "node1",
			PublicKey:  "pubkey1",
			PrivateKey: "privkey1",
			Name:       "test-node",
		}

		err := s.SaveNode(ctx, node)
		require.NoError(t, err)

		fetched, err := s.GetNode(ctx)
		require.NoError(t, err)
		assert.Equal(t, node, fetched)
	})

	// Test Peer Operations
	t.Run("Peer Operations", func(t *testing.T) {
		peer := &store.Peer{
			ID:         "peer1",
			PublicKey:  "peerpub1",
			Endpoint:   "1.2.3.4:51820",
			AllowedIPs: "10.0.0.2/32",
		}

		err := s.SavePeer(ctx, peer)
		require.NoError(t, err)

		peers, err := s.ListPeers(ctx)
		require.NoError(t, err)
		assert.Len(t, peers, 1)
		assert.Equal(t, peer, peers[0])

		err = s.DeletePeer(ctx, peer.PublicKey)
		require.NoError(t, err)

		peers, err = s.ListPeers(ctx)
		require.NoError(t, err)
		assert.Len(t, peers, 0)
	})
}

func TestNew_InvalidPath(t *testing.T) {
	// Test with invalid database path (directory that doesn't exist with no write access)
	_, err := New("/nonexistent/path/to/db.sqlite")
	// Should fail to open or ping
	require.Error(t, err)
}

func TestGetNode_EmptyDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "empty.db")

	s, err := New(dbPath)
	require.NoError(t, err)
	defer s.Close()

	err = s.Migrate()
	require.NoError(t, err)

	ctx := context.Background()

	// GetNode on empty database should return nil, nil
	node, err := s.GetNode(ctx)
	require.NoError(t, err)
	assert.Nil(t, node)
}

func TestMigrate_AlreadyMigrated(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "migrate.db")

	s, err := New(dbPath)
	require.NoError(t, err)
	defer s.Close()

	// First migration
	err = s.Migrate()
	require.NoError(t, err)

	// Second migration should be no-op
	err = s.Migrate()
	require.NoError(t, err)
}

func TestListPeers_EmptyDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "peers.db")

	s, err := New(dbPath)
	require.NoError(t, err)
	defer s.Close()

	err = s.Migrate()
	require.NoError(t, err)

	ctx := context.Background()

	// ListPeers on empty database should return empty slice
	peers, err := s.ListPeers(ctx)
	require.NoError(t, err)
	assert.Empty(t, peers)
}

// ==================== ADDITIONAL EDGE CASE TESTS ====================

func TestSQLiteStore_MultiplePeers(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "multi.db")

	s, err := New(dbPath)
	require.NoError(t, err)
	defer s.Close()

	err = s.Migrate()
	require.NoError(t, err)

	ctx := context.Background()

	// Add multiple peers
	for i := 0; i < 5; i++ {
		peer := &store.Peer{
			ID:         "peer" + string(rune('0'+i)),
			PublicKey:  "pubkey" + string(rune('0'+i)),
			Endpoint:   "10.0.0." + string(rune('0'+i)) + ":51820",
			AllowedIPs: "10.1.0." + string(rune('0'+i)) + "/32",
		}
		err := s.SavePeer(ctx, peer)
		require.NoError(t, err)
	}

	// List should return all peers
	peers, err := s.ListPeers(ctx)
	require.NoError(t, err)
	assert.Len(t, peers, 5)
}

func TestSQLiteStore_UpdateNode(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "update.db")

	s, err := New(dbPath)
	require.NoError(t, err)
	defer s.Close()

	err = s.Migrate()
	require.NoError(t, err)

	ctx := context.Background()

	// Save initial node
	node := &store.Node{
		ID:         "node1",
		PublicKey:  "initial-pub",
		PrivateKey: "initial-priv",
		Name:       "initial-name",
	}
	err = s.SaveNode(ctx, node)
	require.NoError(t, err)

	// Update node with same ID (UPSERT)
	updatedNode := &store.Node{
		ID:         "node1",
		PublicKey:  "updated-pub",
		PrivateKey: "updated-priv",
		Name:       "updated-name",
	}
	err = s.SaveNode(ctx, updatedNode)
	require.NoError(t, err)

	// Fetch should return updated values
	fetched, err := s.GetNode(ctx)
	require.NoError(t, err)
	assert.Equal(t, "updated-pub", fetched.PublicKey)
	assert.Equal(t, "updated-priv", fetched.PrivateKey)
	assert.Equal(t, "updated-name", fetched.Name)
}

func TestSQLiteStore_UpdatePeer(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "update-peer.db")

	s, err := New(dbPath)
	require.NoError(t, err)
	defer s.Close()

	err = s.Migrate()
	require.NoError(t, err)

	ctx := context.Background()

	// Save initial peer
	peer := &store.Peer{
		ID:         "peer1",
		PublicKey:  "initial-pub",
		Endpoint:   "1.2.3.4:51820",
		AllowedIPs: "10.0.0.1/32",
	}
	err = s.SavePeer(ctx, peer)
	require.NoError(t, err)

	// Update peer with same ID (UPSERT)
	updatedPeer := &store.Peer{
		ID:         "peer1",
		PublicKey:  "updated-pub",
		Endpoint:   "5.6.7.8:51820",
		AllowedIPs: "10.0.0.2/32",
	}
	err = s.SavePeer(ctx, updatedPeer)
	require.NoError(t, err)

	// List should return updated values
	peers, err := s.ListPeers(ctx)
	require.NoError(t, err)
	assert.Len(t, peers, 1)
	assert.Equal(t, "updated-pub", peers[0].PublicKey)
	assert.Equal(t, "5.6.7.8:51820", peers[0].Endpoint)
}

func TestSQLiteStore_DeleteNonExistentPeer(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "delete.db")

	s, err := New(dbPath)
	require.NoError(t, err)
	defer s.Close()

	err = s.Migrate()
	require.NoError(t, err)

	ctx := context.Background()

	// Delete non-existent peer should not error
	err = s.DeletePeer(ctx, "non-existent-pubkey")
	require.NoError(t, err)
}

func TestSQLiteStore_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "close.db")

	s, err := New(dbPath)
	require.NoError(t, err)

	// Close should succeed
	err = s.Close()
	require.NoError(t, err)
}
