package service

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/wireguard"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// WireGuardSyncService handles synchronization between DB peers and OS WireGuard interface
type WireGuardSyncService struct {
	peerRepo  repository.PeerRepository
	wgManager peerSyncer
}

// peerSyncer abstracts WireGuard peer synchronization for easier testing.
type peerSyncer interface {
	SyncPeers([]wgtypes.PeerConfig) error
}

// NewWireGuardSyncService creates a new sync service
func NewWireGuardSyncService(peerRepo repository.PeerRepository, wgManager *wireguard.Manager) *WireGuardSyncService {
	return &WireGuardSyncService{
		peerRepo:  peerRepo,
		wgManager: wgManager,
	}
}

// Sync fetches all active peers and applies them to the WireGuard interface
func (s *WireGuardSyncService) Sync(ctx context.Context) error {
	peers, err := s.peerRepo.GetAllActive(ctx)
	if err != nil {
		return err
	}

	var wgPeers []wgtypes.PeerConfig
	for _, p := range peers {
		key, err := wgtypes.ParseKey(p.PublicKey)
		if err != nil {
			slog.Error("Invalid key for peer", "peer_id", p.ID, "error", err)
			continue
		}

		allowedIPs := []net.IPNet{}
		for _, ipStr := range p.AllowedIPs {
			_, ipNet, err := net.ParseCIDR(ipStr)
			if err == nil {
				allowedIPs = append(allowedIPs, *ipNet)
			}
		}

		// Convert keepalive
		ka := 0
		if p.PersistentKeepalive > 0 {
			ka = p.PersistentKeepalive
		}
		kaDuration := time.Duration(ka) * time.Second

		wgPeers = append(wgPeers, wgtypes.PeerConfig{
			PublicKey:                   key,
			AllowedIPs:                  allowedIPs,
			PersistentKeepaliveInterval: &kaDuration,
			ReplaceAllowedIPs:           true,
		})
	}

	return s.wgManager.SyncPeers(wgPeers)
}

// StartSyncLoop starts a background loop to sync peers periodically
func (s *WireGuardSyncService) StartSyncLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Starting WireGuard sync loop", "interval", interval)

	// Initial sync
	if err := s.Sync(ctx); err != nil {
		slog.Error("Initial WireGuard sync failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Sync(ctx); err != nil {
				slog.Error("Failed to sync WireGuard peers", "error", err)
			}
		}
	}
}
