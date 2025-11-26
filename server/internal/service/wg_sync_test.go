package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/wireguard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type fakeWGManager struct {
	received []wgtypes.PeerConfig
	err      error
	calls    int
}

func (f *fakeWGManager) SyncPeers(peers []wgtypes.PeerConfig) error {
	f.received = peers
	f.calls++
	return f.err
}

func TestWireGuardSyncService_SyncSuccessAndSkipsInvalid(t *testing.T) {
	peerRepo := repository.NewInMemoryPeerRepository()
	manager := &fakeWGManager{}
	svc := NewWireGuardSyncService(peerRepo, nil)
	svc.wgManager = manager

	validKey, err := wireguard.GenerateKeyPair()
	require.NoError(t, err)

	ctx := context.Background()

	// valid peer with keepalive and allowed IP
	err = peerRepo.Create(ctx, &domain.Peer{
		NetworkID:           "net-1",
		DeviceID:            "dev-1",
		TenantID:            "t1",
		PublicKey:           validKey.PublicKey,
		AllowedIPs:          []string{"10.0.0.2/32"},
		PersistentKeepalive: 25,
		Active:              true,
	})
	require.NoError(t, err)

	// invalid key peer should be skipped
	err = peerRepo.Create(ctx, &domain.Peer{
		NetworkID:  "net-1",
		DeviceID:   "dev-2",
		TenantID:   "t1",
		PublicKey:  "invalid-key",
		AllowedIPs: []string{"10.0.0.3/32"},
		Active:     true,
	})
	require.NoError(t, err)

	require.NoError(t, svc.Sync(ctx))

	require.Len(t, manager.received, 1)
	cfg := manager.received[0]
	assert.Equal(t, validKey.PublicKey, cfg.PublicKey.String())
	require.Len(t, cfg.AllowedIPs, 1)
	assert.Equal(t, "10.0.0.2/32", cfg.AllowedIPs[0].String())
	require.NotNil(t, cfg.PersistentKeepaliveInterval)
	assert.Equal(t, 25*time.Second, *cfg.PersistentKeepaliveInterval)
}

func TestWireGuardSyncService_SyncManagerError(t *testing.T) {
	peerRepo := repository.NewInMemoryPeerRepository()
	manager := &fakeWGManager{err: errors.New("sync failure")}
	svc := NewWireGuardSyncService(peerRepo, nil)
	svc.wgManager = manager

	validKey, err := wireguard.GenerateKeyPair()
	require.NoError(t, err)

	err = peerRepo.Create(context.Background(), &domain.Peer{
		NetworkID:  "net-err",
		DeviceID:   "dev-err",
		TenantID:   "t1",
		PublicKey:  validKey.PublicKey,
		AllowedIPs: []string{"10.1.0.2/32"},
	})
	require.NoError(t, err)

	syncErr := svc.Sync(context.Background())
	require.Error(t, syncErr)
	assert.EqualError(t, syncErr, "sync failure")
	assert.Len(t, manager.received, 1)
}

func TestWireGuardSyncService_StartSyncLoop(t *testing.T) {
	peerRepo := repository.NewInMemoryPeerRepository()
	manager := &fakeWGManager{}
	svc := NewWireGuardSyncService(peerRepo, nil)
	svc.wgManager = manager

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.StartSyncLoop(ctx, 10*time.Millisecond)

	time.Sleep(30 * time.Millisecond)
	cancel()

	assert.GreaterOrEqual(t, manager.calls, 1)
}
