package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// peerColumns returns the column names for peer queries
func peerColumns() []string {
	return []string{
		"id", "network_id", "device_id", "tenant_id", "public_key", "preshared_key",
		"endpoint", "allowed_ips", "persistent_keepalive", "last_handshake",
		"rx_bytes", "tx_bytes", "active", "created_at", "updated_at", "disabled_at",
	}
}

func TestNewPostgresPeerRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresPeerRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		peer := &domain.Peer{
			ID:                  "peer-123",
			NetworkID:           "network-456",
			DeviceID:            "device-789",
			TenantID:            "tenant-abc",
			PublicKey:           "publickey123456789012345678901234567890123=",
			PresharedKey:        "presharedkey12345678901234567890123456789=",
			Endpoint:            "192.168.1.1:51820",
			AllowedIPs:          []string{"10.0.0.1/32", "10.0.0.2/32"},
			PersistentKeepalive: 25,
			RxBytes:             1000,
			TxBytes:             2000,
			Active:              true,
		}

		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO peers`)).
			WithArgs(
				peer.ID, peer.NetworkID, peer.DeviceID, peer.TenantID,
				peer.PublicKey, peer.PresharedKey, peer.Endpoint,
				pq.Array(peer.AllowedIPs), peer.PersistentKeepalive,
				peer.LastHandshake, peer.RxBytes, peer.TxBytes, peer.Active,
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(peer.ID, now, now))

		err := repo.Create(ctx, peer)
		require.NoError(t, err)
		assert.Equal(t, now, peer.CreatedAt)
		assert.Equal(t, now, peer.UpdatedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("unique constraint violation", func(t *testing.T) {
		peer := &domain.Peer{
			ID:         "peer-123",
			NetworkID:  "network-456",
			DeviceID:   "device-789",
			TenantID:   "tenant-abc",
			PublicKey:  "duplicatepublickey12345678901234567890123=",
			AllowedIPs: []string{"10.0.0.1/32"},
		}

		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO peers`)).
			WithArgs(
				peer.ID, peer.NetworkID, peer.DeviceID, peer.TenantID,
				peer.PublicKey, peer.PresharedKey, peer.Endpoint,
				pq.Array(peer.AllowedIPs), peer.PersistentKeepalive,
				peer.LastHandshake, peer.RxBytes, peer.TxBytes, peer.Active,
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnError(&pq.Error{Code: "23505", Message: "unique_violation"})

		err := repo.Create(ctx, peer)
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrConflict, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		peer := &domain.Peer{
			ID:         "peer-123",
			NetworkID:  "network-456",
			DeviceID:   "device-789",
			TenantID:   "tenant-abc",
			PublicKey:  "somepublickey12345678901234567890123456789=",
			AllowedIPs: []string{"10.0.0.1/32"},
		}

		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO peers`)).
			WithArgs(
				peer.ID, peer.NetworkID, peer.DeviceID, peer.TenantID,
				peer.PublicKey, peer.PresharedKey, peer.Endpoint,
				pq.Array(peer.AllowedIPs), peer.PersistentKeepalive,
				peer.LastHandshake, peer.RxBytes, peer.TxBytes, peer.Active,
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("database error"))

		err := repo.Create(ctx, peer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresPeerRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()
	now := time.Now()
	handshakeTime := now.Add(-5 * time.Minute)

	t.Run("success", func(t *testing.T) {
		peerID := "peer-123"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(peerID).
			WillReturnRows(sqlmock.NewRows(peerColumns()).
				AddRow(
					peerID, "network-456", "device-789", "tenant-abc",
					"publickey123456789012345678901234567890123=",
					"presharedkey12345678901234567890123456789=",
					"192.168.1.1:51820",
					pq.Array([]string{"10.0.0.1/32"}),
					25, handshakeTime, 1000, 2000, true, now, now, nil,
				))

		peer, err := repo.GetByID(ctx, peerID)
		require.NoError(t, err)
		require.NotNil(t, peer)
		assert.Equal(t, peerID, peer.ID)
		assert.Equal(t, "network-456", peer.NetworkID)
		assert.Equal(t, "device-789", peer.DeviceID)
		assert.Equal(t, "tenant-abc", peer.TenantID)
		assert.Equal(t, []string{"10.0.0.1/32"}, peer.AllowedIPs)
		assert.NotNil(t, peer.LastHandshake)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with null last_handshake", func(t *testing.T) {
		peerID := "peer-456"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(peerID).
			WillReturnRows(sqlmock.NewRows(peerColumns()).
				AddRow(
					peerID, "network-456", "device-789", "tenant-abc",
					"publickey123456789012345678901234567890123=",
					"",
					"",
					pq.Array([]string{}),
					0, nil, 0, 0, false, now, now, nil,
				))

		peer, err := repo.GetByID(ctx, peerID)
		require.NoError(t, err)
		require.NotNil(t, peer)
		assert.Nil(t, peer.LastHandshake)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		peerID := "nonexistent-peer"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(peerID).
			WillReturnError(sql.ErrNoRows)

		peer, err := repo.GetByID(ctx, peerID)
		require.Error(t, err)
		assert.Nil(t, peer)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		peerID := "peer-123"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(peerID).
			WillReturnError(errors.New("database error"))

		peer, err := repo.GetByID(ctx, peerID)
		require.Error(t, err)
		assert.Nil(t, peer)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresPeerRepository_GetByNetworkID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()
	now := time.Now()
	handshakeTime := now.Add(-5 * time.Minute)

	t.Run("success with multiple peers", func(t *testing.T) {
		networkID := "network-456"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(networkID).
			WillReturnRows(sqlmock.NewRows(peerColumns()).
				AddRow(
					"peer-1", networkID, "device-1", "tenant-abc",
					"publickey1", "preshared1", "192.168.1.1:51820",
					pq.Array([]string{"10.0.0.1/32"}),
					25, handshakeTime, 1000, 2000, true, now, now, nil,
				).
				AddRow(
					"peer-2", networkID, "device-2", "tenant-abc",
					"publickey2", "preshared2", "192.168.1.2:51820",
					pq.Array([]string{"10.0.0.2/32"}),
					25, nil, 500, 1000, false, now, now, nil,
				))

		peers, err := repo.GetByNetworkID(ctx, networkID)
		require.NoError(t, err)
		assert.Len(t, peers, 2)
		assert.Equal(t, "peer-1", peers[0].ID)
		assert.Equal(t, "peer-2", peers[1].ID)
		assert.NotNil(t, peers[0].LastHandshake)
		assert.Nil(t, peers[1].LastHandshake)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no peers", func(t *testing.T) {
		networkID := "empty-network"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(networkID).
			WillReturnRows(sqlmock.NewRows(peerColumns()))

		peers, err := repo.GetByNetworkID(ctx, networkID)
		require.NoError(t, err)
		assert.Empty(t, peers)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		networkID := "network-456"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(networkID).
			WillReturnError(errors.New("database error"))

		peers, err := repo.GetByNetworkID(ctx, networkID)
		require.Error(t, err)
		assert.Nil(t, peers)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		networkID := "network-456"

		// Return wrong number of columns to trigger scan error
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(networkID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "network_id"}).
				AddRow("peer-1", networkID))

		peers, err := repo.GetByNetworkID(ctx, networkID)
		require.Error(t, err)
		assert.Nil(t, peers)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresPeerRepository_GetByDeviceID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		deviceID := "device-789"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(deviceID).
			WillReturnRows(sqlmock.NewRows(peerColumns()).
				AddRow(
					"peer-1", "network-1", deviceID, "tenant-abc",
					"publickey1", "preshared1", "192.168.1.1:51820",
					pq.Array([]string{"10.0.0.1/32"}),
					25, nil, 1000, 2000, true, now, now, nil,
				))

		peers, err := repo.GetByDeviceID(ctx, deviceID)
		require.NoError(t, err)
		assert.Len(t, peers, 1)
		assert.Equal(t, deviceID, peers[0].DeviceID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		deviceID := "device-789"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(deviceID).
			WillReturnError(errors.New("database error"))

		peers, err := repo.GetByDeviceID(ctx, deviceID)
		require.Error(t, err)
		assert.Nil(t, peers)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresPeerRepository_GetByNetworkAndDevice(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()
	now := time.Now()
	handshakeTime := now.Add(-5 * time.Minute)

	t.Run("success", func(t *testing.T) {
		networkID := "network-456"
		deviceID := "device-789"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(networkID, deviceID).
			WillReturnRows(sqlmock.NewRows(peerColumns()).
				AddRow(
					"peer-1", networkID, deviceID, "tenant-abc",
					"publickey1", "preshared1", "192.168.1.1:51820",
					pq.Array([]string{"10.0.0.1/32"}),
					25, handshakeTime, 1000, 2000, true, now, now, nil,
				))

		peer, err := repo.GetByNetworkAndDevice(ctx, networkID, deviceID)
		require.NoError(t, err)
		require.NotNil(t, peer)
		assert.Equal(t, networkID, peer.NetworkID)
		assert.Equal(t, deviceID, peer.DeviceID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		networkID := "network-456"
		deviceID := "nonexistent-device"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(networkID, deviceID).
			WillReturnError(sql.ErrNoRows)

		peer, err := repo.GetByNetworkAndDevice(ctx, networkID, deviceID)
		require.Error(t, err)
		assert.Nil(t, peer)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		networkID := "network-456"
		deviceID := "device-789"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(networkID, deviceID).
			WillReturnError(errors.New("database error"))

		peer, err := repo.GetByNetworkAndDevice(ctx, networkID, deviceID)
		require.Error(t, err)
		assert.Nil(t, peer)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresPeerRepository_GetByPublicKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		publicKey := "publickey123456789012345678901234567890123="

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(publicKey).
			WillReturnRows(sqlmock.NewRows(peerColumns()).
				AddRow(
					"peer-1", "network-456", "device-789", "tenant-abc",
					publicKey, "preshared1", "192.168.1.1:51820",
					pq.Array([]string{"10.0.0.1/32"}),
					25, nil, 1000, 2000, true, now, now, nil,
				))

		peer, err := repo.GetByPublicKey(ctx, publicKey)
		require.NoError(t, err)
		require.NotNil(t, peer)
		assert.Equal(t, publicKey, peer.PublicKey)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		publicKey := "nonexistent-key"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(publicKey).
			WillReturnError(sql.ErrNoRows)

		peer, err := repo.GetByPublicKey(ctx, publicKey)
		require.Error(t, err)
		assert.Nil(t, peer)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresPeerRepository_GetActivePeers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()
	now := time.Now()
	handshakeTime := now.Add(-1 * time.Minute)

	t.Run("success with active peers", func(t *testing.T) {
		networkID := "network-456"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(networkID).
			WillReturnRows(sqlmock.NewRows(peerColumns()).
				AddRow(
					"peer-1", networkID, "device-1", "tenant-abc",
					"publickey1", "preshared1", "192.168.1.1:51820",
					pq.Array([]string{"10.0.0.1/32"}),
					25, handshakeTime, 1000, 2000, true, now, now, nil,
				).
				AddRow(
					"peer-2", networkID, "device-2", "tenant-abc",
					"publickey2", "preshared2", "192.168.1.2:51820",
					pq.Array([]string{"10.0.0.2/32"}),
					25, handshakeTime, 500, 1000, true, now, now, nil,
				))

		peers, err := repo.GetActivePeers(ctx, networkID)
		require.NoError(t, err)
		assert.Len(t, peers, 2)
		assert.True(t, peers[0].Active)
		assert.True(t, peers[1].Active)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no active peers", func(t *testing.T) {
		networkID := "network-no-active"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(networkID).
			WillReturnRows(sqlmock.NewRows(peerColumns()))

		peers, err := repo.GetActivePeers(ctx, networkID)
		require.NoError(t, err)
		assert.Empty(t, peers)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		networkID := "network-456"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(networkID).
			WillReturnError(errors.New("database error"))

		peers, err := repo.GetActivePeers(ctx, networkID)
		require.Error(t, err)
		assert.Nil(t, peers)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		networkID := "network-456"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(networkID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("peer-1"))

		peers, err := repo.GetActivePeers(ctx, networkID)
		require.Error(t, err)
		assert.Nil(t, peers)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresPeerRepository_GetAllActive(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WillReturnRows(sqlmock.NewRows(peerColumns()).
				AddRow(
					"peer-1", "network-1", "device-1", "tenant-abc",
					"publickey1", "preshared1", "192.168.1.1:51820",
					pq.Array([]string{"10.0.0.1/32"}),
					25, nil, 1000, 2000, true, now, now, nil,
				).
				AddRow(
					"peer-2", "network-2", "device-2", "tenant-def",
					"publickey2", "preshared2", "192.168.1.2:51820",
					pq.Array([]string{"10.0.0.2/32"}),
					25, nil, 500, 1000, false, now, now, nil,
				))

		peers, err := repo.GetAllActive(ctx)
		require.NoError(t, err)
		assert.Len(t, peers, 2)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WillReturnError(errors.New("database error"))

		peers, err := repo.GetAllActive(ctx)
		require.Error(t, err)
		assert.Nil(t, peers)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("peer-1"))

		peers, err := repo.GetAllActive(ctx)
		require.Error(t, err)
		assert.Nil(t, peers)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresPeerRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()
	now := time.Now()
	handshakeTime := now.Add(-5 * time.Minute)

	t.Run("success", func(t *testing.T) {
		peer := &domain.Peer{
			ID:                  "peer-123",
			NetworkID:           "network-456",
			DeviceID:            "device-789",
			TenantID:            "tenant-abc",
			PublicKey:           "publickey123456789012345678901234567890123=",
			PresharedKey:        "presharedkey12345678901234567890123456789=",
			Endpoint:            "192.168.1.100:51820",
			AllowedIPs:          []string{"10.0.0.1/32", "10.0.0.2/32"},
			PersistentKeepalive: 30,
			LastHandshake:       &handshakeTime,
			RxBytes:             5000,
			TxBytes:             10000,
			Active:              true,
		}

		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE peers`)).
			WithArgs(
				peer.ID, peer.NetworkID, peer.DeviceID, peer.TenantID,
				peer.PublicKey, peer.PresharedKey, peer.Endpoint,
				pq.Array(peer.AllowedIPs), peer.PersistentKeepalive,
				peer.LastHandshake, peer.RxBytes, peer.TxBytes, peer.Active,
				sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(now))

		err := repo.Update(ctx, peer)
		require.NoError(t, err)
		assert.Equal(t, now, peer.UpdatedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		peer := &domain.Peer{
			ID:         "nonexistent-peer",
			NetworkID:  "network-456",
			DeviceID:   "device-789",
			TenantID:   "tenant-abc",
			PublicKey:  "publickey123456789012345678901234567890123=",
			AllowedIPs: []string{"10.0.0.1/32"},
		}

		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE peers`)).
			WithArgs(
				peer.ID, peer.NetworkID, peer.DeviceID, peer.TenantID,
				peer.PublicKey, peer.PresharedKey, peer.Endpoint,
				pq.Array(peer.AllowedIPs), peer.PersistentKeepalive,
				peer.LastHandshake, peer.RxBytes, peer.TxBytes, peer.Active,
				sqlmock.AnyArg(),
			).
			WillReturnError(sql.ErrNoRows)

		err := repo.Update(ctx, peer)
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		peer := &domain.Peer{
			ID:         "peer-123",
			NetworkID:  "network-456",
			DeviceID:   "device-789",
			TenantID:   "tenant-abc",
			PublicKey:  "publickey123456789012345678901234567890123=",
			AllowedIPs: []string{"10.0.0.1/32"},
		}

		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE peers`)).
			WithArgs(
				peer.ID, peer.NetworkID, peer.DeviceID, peer.TenantID,
				peer.PublicKey, peer.PresharedKey, peer.Endpoint,
				pq.Array(peer.AllowedIPs), peer.PersistentKeepalive,
				peer.LastHandshake, peer.RxBytes, peer.TxBytes, peer.Active,
				sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("database error"))

		err := repo.Update(ctx, peer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresPeerRepository_UpdateStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()
	now := time.Now()
	handshakeTime := now.Add(-1 * time.Minute)

	t.Run("success", func(t *testing.T) {
		peerID := "peer-123"
		stats := &domain.UpdatePeerStatsRequest{
			Endpoint:      "192.168.1.100:51820",
			LastHandshake: &handshakeTime,
			RxBytes:       10000,
			TxBytes:       20000,
		}

		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE peers`)).
			WithArgs(
				peerID, stats.Endpoint, stats.LastHandshake,
				stats.RxBytes, stats.TxBytes, sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(now))

		err := repo.UpdateStats(ctx, peerID, stats)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		peerID := "nonexistent-peer"
		stats := &domain.UpdatePeerStatsRequest{
			RxBytes: 10000,
			TxBytes: 20000,
		}

		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE peers`)).
			WithArgs(
				peerID, stats.Endpoint, stats.LastHandshake,
				stats.RxBytes, stats.TxBytes, sqlmock.AnyArg(),
			).
			WillReturnError(sql.ErrNoRows)

		err := repo.UpdateStats(ctx, peerID, stats)
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		peerID := "peer-123"
		stats := &domain.UpdatePeerStatsRequest{
			RxBytes: 10000,
			TxBytes: 20000,
		}

		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE peers`)).
			WithArgs(
				peerID, stats.Endpoint, stats.LastHandshake,
				stats.RxBytes, stats.TxBytes, sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("database error"))

		err := repo.UpdateStats(ctx, peerID, stats)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresPeerRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		peerID := "peer-123"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE peers`)).
			WithArgs(peerID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, peerID)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		peerID := "nonexistent-peer"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE peers`)).
			WithArgs(peerID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, peerID)
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exec error", func(t *testing.T) {
		peerID := "peer-123"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE peers`)).
			WithArgs(peerID, sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		err := repo.Delete(ctx, peerID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		peerID := "peer-123"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE peers`)).
			WithArgs(peerID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.Delete(ctx, peerID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rows affected error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresPeerRepository_HardDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		peerID := "peer-123"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM peers WHERE id = $1`)).
			WithArgs(peerID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.HardDelete(ctx, peerID)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		peerID := "nonexistent-peer"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM peers WHERE id = $1`)).
			WithArgs(peerID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.HardDelete(ctx, peerID)
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exec error", func(t *testing.T) {
		peerID := "peer-123"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM peers WHERE id = $1`)).
			WithArgs(peerID).
			WillReturnError(errors.New("database error"))

		err := repo.HardDelete(ctx, peerID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		peerID := "peer-123"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM peers WHERE id = $1`)).
			WithArgs(peerID).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.HardDelete(ctx, peerID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rows affected error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresPeerRepository_ListByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with peers", func(t *testing.T) {
		tenantID := "tenant-abc"
		limit := 10
		offset := 0

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(tenantID, limit, offset).
			WillReturnRows(sqlmock.NewRows(peerColumns()).
				AddRow(
					"peer-1", "network-1", "device-1", tenantID,
					"publickey1", "preshared1", "192.168.1.1:51820",
					pq.Array([]string{"10.0.0.1/32"}),
					25, nil, 1000, 2000, true, now, now, nil,
				).
				AddRow(
					"peer-2", "network-2", "device-2", tenantID,
					"publickey2", "preshared2", "192.168.1.2:51820",
					pq.Array([]string{"10.0.0.2/32"}),
					25, nil, 500, 1000, false, now, now, nil,
				))

		peers, err := repo.ListByTenant(ctx, tenantID, limit, offset)
		require.NoError(t, err)
		assert.Len(t, peers, 2)
		assert.Equal(t, tenantID, peers[0].TenantID)
		assert.Equal(t, tenantID, peers[1].TenantID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with pagination", func(t *testing.T) {
		tenantID := "tenant-abc"
		limit := 5
		offset := 10

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(tenantID, limit, offset).
			WillReturnRows(sqlmock.NewRows(peerColumns()).
				AddRow(
					"peer-11", "network-1", "device-11", tenantID,
					"publickey11", "preshared11", "192.168.1.11:51820",
					pq.Array([]string{"10.0.0.11/32"}),
					25, nil, 1000, 2000, true, now, now, nil,
				))

		peers, err := repo.ListByTenant(ctx, tenantID, limit, offset)
		require.NoError(t, err)
		assert.Len(t, peers, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no peers", func(t *testing.T) {
		tenantID := "empty-tenant"
		limit := 10
		offset := 0

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(tenantID, limit, offset).
			WillReturnRows(sqlmock.NewRows(peerColumns()))

		peers, err := repo.ListByTenant(ctx, tenantID, limit, offset)
		require.NoError(t, err)
		assert.Empty(t, peers)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		tenantID := "tenant-abc"
		limit := 10
		offset := 0

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(tenantID, limit, offset).
			WillReturnError(errors.New("database error"))

		peers, err := repo.ListByTenant(ctx, tenantID, limit, offset)
		require.Error(t, err)
		assert.Nil(t, peers)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		tenantID := "tenant-abc"
		limit := 10
		offset := 0

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
			WithArgs(tenantID, limit, offset).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("peer-1"))

		peers, err := repo.ListByTenant(ctx, tenantID, limit, offset)
		require.Error(t, err)
		assert.Nil(t, peers)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresPeerRepository_scanPeers_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPeerRepository(db)
	ctx := context.Background()
	now := time.Now()

	// Test that rows.Err() is properly checked
	tenantID := "tenant-abc"
	limit := 10
	offset := 0

	rows := sqlmock.NewRows(peerColumns()).
		AddRow(
			"peer-1", "network-1", "device-1", tenantID,
			"publickey1", "preshared1", "192.168.1.1:51820",
			pq.Array([]string{"10.0.0.1/32"}),
			25, nil, 1000, 2000, true, now, now, nil,
		).
		RowError(0, errors.New("row iteration error"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, device_id, tenant_id, public_key, preshared_key`)).
		WithArgs(tenantID, limit, offset).
		WillReturnRows(rows)

	peers, err := repo.ListByTenant(ctx, tenantID, limit, offset)
	require.Error(t, err)
	assert.Nil(t, peers)
	assert.Contains(t, err.Error(), "row iteration error")
	require.NoError(t, mock.ExpectationsWereMet())
}
