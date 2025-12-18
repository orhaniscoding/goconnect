package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPeerService(t *testing.T) (*PeerService, *repository.InMemoryPeerRepository, *repository.InMemoryDeviceRepository, *repository.InMemoryNetworkRepository) {
	t.Helper()

	peerRepo := repository.NewInMemoryPeerRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()

	return NewPeerService(peerRepo, deviceRepo, networkRepo), peerRepo, deviceRepo, networkRepo
}

func seedNetwork(t *testing.T, repo *repository.InMemoryNetworkRepository, tenantID, id string) *domain.Network {
	t.Helper()

	network := &domain.Network{
		ID:         id,
		TenantID:   tenantID,
		Name:       fmt.Sprintf("Test Network %s", id),
		Visibility: domain.NetworkVisibilityPrivate,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.0.0.0/24",
		CreatedBy:  "tester",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	require.NoError(t, repo.Create(context.Background(), network))
	return network
}

func seedDevice(t *testing.T, repo *repository.InMemoryDeviceRepository, tenantID, userID, id, pubKey string) *domain.Device {
	t.Helper()

	device := &domain.Device{
		ID:        id,
		UserID:    userID,
		TenantID:  tenantID,
		Name:      "Test Device",
		Platform:  "linux",
		PubKey:    pubKey,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	require.NoError(t, repo.Create(context.Background(), device))
	return device
}

func seedPeer(t *testing.T, repo *repository.InMemoryPeerRepository, networkID, deviceID, tenantID, publicKey string) *domain.Peer {
	t.Helper()

	peer := &domain.Peer{
		NetworkID:  networkID,
		DeviceID:   deviceID,
		TenantID:   tenantID,
		PublicKey:  publicKey,
		AllowedIPs: []string{"10.0.0.5/32"},
	}

	require.NoError(t, repo.Create(context.Background(), peer))
	return peer
}

func TestPeerService_CreatePeer(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, peerRepo, deviceRepo, networkRepo := setupPeerService(t)
		network := seedNetwork(t, networkRepo, "tenant-1", "network-1")
		device := seedDevice(t, deviceRepo, "tenant-1", "user-1", "device-1", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")

		req := &domain.CreatePeerRequest{
			NetworkID:           network.ID,
			DeviceID:            device.ID,
			PublicKey:           "cOvbNjH7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pW3nA2aZ=",
			PresharedKey:        "DuplicateKey7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pWX=",
			AllowedIPs:          []string{"10.0.0.2/32"},
			PersistentKeepalive: 15,
		}

		peer, err := svc.CreatePeer(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, network.TenantID, peer.TenantID)
		assert.Equal(t, req.PublicKey, peer.PublicKey)
		assert.Equal(t, req.PresharedKey, peer.PresharedKey)
		assert.False(t, peer.Active)

		// Stored in repository
		saved, err := peerRepo.GetByID(ctx, peer.ID)
		require.NoError(t, err)
		assert.Equal(t, peer.ID, saved.ID)
	})

	t.Run("validation error for public key length", func(t *testing.T) {
		svc, _, deviceRepo, networkRepo := setupPeerService(t)
		network := seedNetwork(t, networkRepo, "tenant-1", "network-1")
		device := seedDevice(t, deviceRepo, "tenant-1", "user-1", "device-1", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")

		req := &domain.CreatePeerRequest{
			NetworkID:  network.ID,
			DeviceID:   device.ID,
			PublicKey:  "short-key",
			AllowedIPs: []string{"10.0.0.2/32"},
		}

		_, err := svc.CreatePeer(ctx, req)

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrValidation, domainErr.Code)
	})

	t.Run("forbidden when tenant mismatch", func(t *testing.T) {
		svc, _, deviceRepo, networkRepo := setupPeerService(t)
		network := seedNetwork(t, networkRepo, "tenant-1", "network-1")
		device := seedDevice(t, deviceRepo, "tenant-2", "user-1", "device-1", "bOtHerKeyXYZ123456789ABCDEFGHIJKLMNOPQRSTUV=")

		req := &domain.CreatePeerRequest{
			NetworkID:  network.ID,
			DeviceID:   device.ID,
			PublicKey:  "cOvbNjH7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pW3nA2aZ=",
			AllowedIPs: []string{"10.0.0.3/32"},
		}

		_, err := svc.CreatePeer(ctx, req)

		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrForbidden, domainErr.Code)
	})

	t.Run("preshared key length validation", func(t *testing.T) {
		svc, _, deviceRepo, networkRepo := setupPeerService(t)
		network := seedNetwork(t, networkRepo, "tenant-1", "network-1")
		device := seedDevice(t, deviceRepo, "tenant-1", "user-1", "device-1", "cOvbNjH7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pW3nA2aZ=")

		req := &domain.CreatePeerRequest{
			NetworkID:    network.ID,
			DeviceID:     device.ID,
			PublicKey:    "bOtHerKeyXYZ123456789ABCDEFGHIJKLMNOPQRSTUV=",
			PresharedKey: "too-short",
			AllowedIPs:   []string{"10.0.0.4/32"},
		}

		_, err := svc.CreatePeer(ctx, req)

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrValidation, domainErr.Code)
	})
}

func TestPeerService_GetPeer(t *testing.T) {
	ctx := context.Background()
	svc, peerRepo, _, _ := setupPeerService(t)
	peer := seedPeer(t, peerRepo, "network-1", "device-1", "tenant-1", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")

	t.Run("success", func(t *testing.T) {
		found, err := svc.GetPeer(ctx, peer.ID)

		require.NoError(t, err)
		assert.Equal(t, peer.ID, found.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetPeer(ctx, "missing-peer")

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
	})
}

func TestPeerService_GetPeersByNetwork(t *testing.T) {
	ctx := context.Background()
	svc, peerRepo, deviceRepo, networkRepo := setupPeerService(t)
	network := seedNetwork(t, networkRepo, "tenant-1", "network-1")
	otherNetwork := seedNetwork(t, networkRepo, "tenant-1", "network-2")
	device := seedDevice(t, deviceRepo, "tenant-1", "user-1", "device-1", "cOvbNjH7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pW3nA2aZ=")
	otherDevice := seedDevice(t, deviceRepo, "tenant-1", "user-2", "device-2", "bOtHerKeyXYZ123456789ABCDEFGHIJKLMNOPQRSTUV=")

	seedPeer(t, peerRepo, network.ID, device.ID, network.TenantID, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	seedPeer(t, peerRepo, network.ID, otherDevice.ID, network.TenantID, "DuplicateKey7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pWX=")
	seedPeer(t, peerRepo, otherNetwork.ID, device.ID, otherNetwork.TenantID, "DuplicateKey7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pWX=")

	t.Run("success", func(t *testing.T) {
		peers, err := svc.GetPeersByNetwork(ctx, network.ID)

		require.NoError(t, err)
		assert.Len(t, peers, 2)
	})

	t.Run("network not found", func(t *testing.T) {
		_, err := svc.GetPeersByNetwork(ctx, "missing")

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
	})
}

func TestPeerService_GetPeersByDevice(t *testing.T) {
	ctx := context.Background()
	svc, peerRepo, deviceRepo, networkRepo := setupPeerService(t)
	network := seedNetwork(t, networkRepo, "tenant-1", "network-1")
	networkTwo := seedNetwork(t, networkRepo, "tenant-1", "network-2")
	device := seedDevice(t, deviceRepo, "tenant-1", "user-1", "device-1", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	otherDevice := seedDevice(t, deviceRepo, "tenant-1", "user-2", "device-2", "bOtHerKeyXYZ123456789ABCDEFGHIJKLMNOPQRSTUV=")

	seedPeer(t, peerRepo, network.ID, device.ID, network.TenantID, "cOvbNjH7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pW3nA2aZ=")
	seedPeer(t, peerRepo, networkTwo.ID, device.ID, network.TenantID, "DuplicateKey7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pWX=")
	seedPeer(t, peerRepo, network.ID, otherDevice.ID, network.TenantID, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")

	t.Run("success", func(t *testing.T) {
		peers, err := svc.GetPeersByDevice(ctx, device.ID)

		require.NoError(t, err)
		assert.Len(t, peers, 2)
	})

	t.Run("device not found", func(t *testing.T) {
		_, err := svc.GetPeersByDevice(ctx, "missing-device")

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
	})
}

func TestPeerService_UpdatePeer(t *testing.T) {
	ctx := context.Background()
	svc, peerRepo, deviceRepo, networkRepo := setupPeerService(t)
	network := seedNetwork(t, networkRepo, "tenant-1", "network-1")
	device := seedDevice(t, deviceRepo, "tenant-1", "user-1", "device-1", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")

	created, err := svc.CreatePeer(ctx, &domain.CreatePeerRequest{
		NetworkID:  network.ID,
		DeviceID:   device.ID,
		PublicKey:  "cOvbNjH7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pW3nA2aZ=",
		AllowedIPs: []string{"10.0.0.2/32"},
	})
	require.NoError(t, err)

	t.Run("successfully updates peer fields", func(t *testing.T) {
		newEndpoint := "10.0.0.5:51820"
		newAllowed := []string{"10.0.0.5/32", "10.0.0.6/32"}
		newPSK := "DuplicateKey7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pWX="
		newKeepalive := 30

		updated, err := svc.UpdatePeer(ctx, created.ID, &domain.UpdatePeerRequest{
			Endpoint:            &newEndpoint,
			AllowedIPs:          &newAllowed,
			PresharedKey:        &newPSK,
			PersistentKeepalive: &newKeepalive,
		})

		require.NoError(t, err)
		assert.Equal(t, newEndpoint, updated.Endpoint)
		assert.Equal(t, newAllowed, updated.AllowedIPs)
		assert.Equal(t, newPSK, updated.PresharedKey)
		assert.Equal(t, newKeepalive, updated.PersistentKeepalive)

		reloaded, _ := peerRepo.GetByID(ctx, created.ID)
		assert.Equal(t, updated.Endpoint, reloaded.Endpoint)
	})

	t.Run("rejects invalid keepalive value", func(t *testing.T) {
		badKeepalive := 70000
		_, err := svc.UpdatePeer(ctx, created.ID, &domain.UpdatePeerRequest{
			PersistentKeepalive: &badKeepalive,
		})

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrValidation, domainErr.Code)
	})
}

func TestPeerService_DeletePeer(t *testing.T) {
	ctx := context.Background()
	svc, peerRepo, _, _ := setupPeerService(t)
	peer := seedPeer(t, peerRepo, "network-1", "device-1", "tenant-1", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")

	err := svc.DeletePeer(ctx, peer.ID)
	require.NoError(t, err)

	_, err = peerRepo.GetByID(ctx, peer.ID)
	require.Error(t, err)
	var domainErr *domain.Error
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestPeerService_GetPeerStats(t *testing.T) {
	t.Run("recent handshake reports latency", func(t *testing.T) {
		ctx := context.Background()
		svc, peerRepo, _, _ := setupPeerService(t)
		last := time.Now().Add(-10 * time.Second)

		peer := &domain.Peer{
			NetworkID:     "network-1",
			DeviceID:      "device-1",
			TenantID:      "tenant-1",
			PublicKey:     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			AllowedIPs:    []string{"10.0.0.2/32"},
			Endpoint:      "198.51.100.1:51820",
			LastHandshake: &last,
			RxBytes:       1234,
			TxBytes:       5678,
			Active:        true,
		}
		require.NoError(t, peerRepo.Create(ctx, peer))

		stats, err := svc.GetPeerStats(ctx, peer.ID)

		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, peer.ID, stats.PeerID)
		assert.Equal(t, peer.Endpoint, stats.Endpoint)
		assert.Equal(t, peer.RxBytes, stats.RxBytes)
		assert.Equal(t, peer.TxBytes, stats.TxBytes)
		assert.Equal(t, peer.Active, stats.Active)
		assert.Equal(t, 50, stats.Latency)
	})

	t.Run("stale handshake yields zero latency", func(t *testing.T) {
		ctx := context.Background()
		svc, peerRepo, _, _ := setupPeerService(t)
		last := time.Now().Add(-2 * time.Minute)
		peer := &domain.Peer{
			NetworkID:     "network-2",
			DeviceID:      "device-2",
			TenantID:      "tenant-1",
			PublicKey:     "bOtHerKeyXYZ123456789ABCDEFGHIJKLMNOPQRSTUV=",
			AllowedIPs:    []string{"10.0.0.3/32"},
			LastHandshake: &last,
			Active:        true,
		}
		require.NoError(t, peerRepo.Create(ctx, peer))

		stats, err := svc.GetPeerStats(ctx, peer.ID)

		require.NoError(t, err)
		assert.Equal(t, 0, stats.Latency)
	})
}

func TestPeerService_RotatePeerKeys(t *testing.T) {
	ctx := context.Background()
	svc, peerRepo, _, _ := setupPeerService(t)
	peer := seedPeer(t, peerRepo, "network-1", "device-1", "tenant-1", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	peer.PresharedKey = "DuplicateKey7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pWX="
	require.NoError(t, peerRepo.Update(ctx, peer))
	oldPublicKey := peer.PublicKey
	oldPresharedKey := peer.PresharedKey

	rotated, err := svc.RotatePeerKeys(ctx, peer.ID)

	require.NoError(t, err)
	require.NotNil(t, rotated)
	assert.NotEqual(t, oldPublicKey, rotated.PublicKey)
	assert.Len(t, rotated.PublicKey, 44)
	assert.Len(t, rotated.PresharedKey, 44)
	assert.NotEqual(t, oldPresharedKey, rotated.PresharedKey)

	saved, err := peerRepo.GetByID(ctx, peer.ID)
	require.NoError(t, err)
	assert.Equal(t, rotated.PublicKey, saved.PublicKey)
	assert.Equal(t, rotated.PresharedKey, saved.PresharedKey)
}

// ==================== GetPeerByNetworkAndDevice Tests ====================

func TestPeerService_GetPeerByNetworkAndDevice(t *testing.T) {
	ctx := context.Background()
	svc, peerRepo, deviceRepo, networkRepo := setupPeerService(t)

	// Seed network and device
	network := seedNetwork(t, networkRepo, "tenant-1", "network-1")
	device := seedDevice(t, deviceRepo, "tenant-1", "user-1", "device-1", "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBd=")
	peer := seedPeer(t, peerRepo, network.ID, device.ID, "tenant-1", "CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCc=")

	t.Run("Success - get peer by network and device", func(t *testing.T) {
		result, err := svc.GetPeerByNetworkAndDevice(ctx, network.ID, device.ID)
		require.NoError(t, err)
		assert.Equal(t, peer.ID, result.ID)
	})

	t.Run("Network not found", func(t *testing.T) {
		_, err := svc.GetPeerByNetworkAndDevice(ctx, "nonexistent-network", device.ID)
		require.Error(t, err)
	})

	t.Run("Device not found", func(t *testing.T) {
		_, err := svc.GetPeerByNetworkAndDevice(ctx, network.ID, "nonexistent-device")
		require.Error(t, err)
	})
}

// ==================== GetActivePeers Tests ====================

func TestPeerService_GetActivePeers(t *testing.T) {
	ctx := context.Background()
	svc, peerRepo, _, networkRepo := setupPeerService(t)

	network := seedNetwork(t, networkRepo, "tenant-1", "active-peers-net")

	// Create active and inactive peers
	activePeer := seedPeer(t, peerRepo, network.ID, "device-active", "tenant-1", "DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDd=")
	activePeer.Active = true
	require.NoError(t, peerRepo.Update(ctx, activePeer))

	inactivePeer := seedPeer(t, peerRepo, network.ID, "device-inactive", "tenant-1", "EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEe=")
	inactivePeer.Active = false
	require.NoError(t, peerRepo.Update(ctx, inactivePeer))

	t.Run("Success - get active peers", func(t *testing.T) {
		peers, err := svc.GetActivePeers(ctx, network.ID)
		require.NoError(t, err)
		// Should only get active peer
		assert.GreaterOrEqual(t, len(peers), 1)
	})

	t.Run("Network not found", func(t *testing.T) {
		_, err := svc.GetActivePeers(ctx, "nonexistent-network")
		require.Error(t, err)
	})
}

// ==================== GetActivePeersConfig Tests ====================

func TestPeerService_GetActivePeersConfig(t *testing.T) {
	ctx := context.Background()
	svc, peerRepo, deviceRepo, networkRepo := setupPeerService(t)

	network := seedNetwork(t, networkRepo, "tenant-1", "config-net")
	device := seedDevice(t, deviceRepo, "tenant-1", "user-1", "device-cfg", "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFf=")

	peer := seedPeer(t, peerRepo, network.ID, device.ID, "tenant-1", "GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGg=")
	peer.Active = true
	require.NoError(t, peerRepo.Update(ctx, peer))

	t.Run("Success - get active peers config", func(t *testing.T) {
		configs, err := svc.GetActivePeersConfig(ctx, network.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(configs), 1)

		if len(configs) > 0 {
			assert.NotEmpty(t, configs[0].PublicKey)
		}
	})

	t.Run("Network not found", func(t *testing.T) {
		_, err := svc.GetActivePeersConfig(ctx, "nonexistent-network")
		require.Error(t, err)
	})
}

// ==================== UpdatePeerStats Tests ====================

func TestPeerService_UpdatePeerStats(t *testing.T) {
	ctx := context.Background()
	svc, peerRepo, _, _ := setupPeerService(t)

	peer := seedPeer(t, peerRepo, "network-stats", "device-stats", "tenant-1", "HHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHh=")

	t.Run("Success - update peer stats", func(t *testing.T) {
		now := time.Now()
		stats := &domain.UpdatePeerStatsRequest{
			RxBytes:       1024,
			TxBytes:       2048,
			LastHandshake: &now,
		}

		err := svc.UpdatePeerStats(ctx, peer.ID, stats)
		require.NoError(t, err)

		// Verify stats were updated
		updated, err := peerRepo.GetByID(ctx, peer.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1024), updated.RxBytes)
		assert.Equal(t, int64(2048), updated.TxBytes)
	})
}

// ==================== GetNetworkPeerStats Tests ====================

func TestPeerService_GetNetworkPeerStats(t *testing.T) {
	ctx := context.Background()
	svc, peerRepo, _, networkRepo := setupPeerService(t)

	network := seedNetwork(t, networkRepo, "tenant-1", "net-stats")
	seedPeer(t, peerRepo, network.ID, "device-ns-1", "tenant-1", "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIi=")
	seedPeer(t, peerRepo, network.ID, "device-ns-2", "tenant-1", "JJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJj=")

	t.Run("Success - get network peer stats", func(t *testing.T) {
		stats, err := svc.GetNetworkPeerStats(ctx, network.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(stats), 2)
	})

	t.Run("Empty network", func(t *testing.T) {
		emptyNet := seedNetwork(t, networkRepo, "tenant-1", "empty-net")
		stats, err := svc.GetNetworkPeerStats(ctx, emptyNet.ID)
		require.NoError(t, err)
		assert.Len(t, stats, 0)
	})
}
