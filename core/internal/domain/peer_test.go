package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== CreatePeerRequest.Validate Tests ====================

func TestCreatePeerRequest_Validate(t *testing.T) {
	validKey := "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY=" // 44 chars

	t.Run("Valid Request", func(t *testing.T) {
		req := &CreatePeerRequest{
			NetworkID:           "net123",
			DeviceID:            "dev123",
			PublicKey:           validKey,
			AllowedIPs:          []string{"10.0.0.1/32"},
			PersistentKeepalive: 25,
		}
		err := req.Validate()
		assert.NoError(t, err)
	})

	t.Run("Missing NetworkID", func(t *testing.T) {
		req := &CreatePeerRequest{
			NetworkID:  "",
			DeviceID:   "dev123",
			PublicKey:  validKey,
			AllowedIPs: []string{"10.0.0.1/32"},
		}
		err := req.Validate()
		require.Error(t, err)
		domErr, ok := err.(*Error)
		require.True(t, ok)
		assert.Equal(t, ErrValidation, domErr.Code)
	})

	t.Run("Missing DeviceID", func(t *testing.T) {
		req := &CreatePeerRequest{
			NetworkID:  "net123",
			DeviceID:   "",
			PublicKey:  validKey,
			AllowedIPs: []string{"10.0.0.1/32"},
		}
		err := req.Validate()
		require.Error(t, err)
	})

	t.Run("Invalid PublicKey Length", func(t *testing.T) {
		req := &CreatePeerRequest{
			NetworkID:  "net123",
			DeviceID:   "dev123",
			PublicKey:  "shortkey",
			AllowedIPs: []string{"10.0.0.1/32"},
		}
		err := req.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "44 characters")
	})

	t.Run("Empty AllowedIPs", func(t *testing.T) {
		req := &CreatePeerRequest{
			NetworkID:  "net123",
			DeviceID:   "dev123",
			PublicKey:  validKey,
			AllowedIPs: []string{},
		}
		err := req.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "allowed IP")
	})

	t.Run("Invalid PersistentKeepalive", func(t *testing.T) {
		req := &CreatePeerRequest{
			NetworkID:           "net123",
			DeviceID:            "dev123",
			PublicKey:           validKey,
			AllowedIPs:          []string{"10.0.0.1/32"},
			PersistentKeepalive: -1,
		}
		err := req.Validate()
		require.Error(t, err)
	})

	t.Run("PersistentKeepalive Too Large", func(t *testing.T) {
		req := &CreatePeerRequest{
			NetworkID:           "net123",
			DeviceID:            "dev123",
			PublicKey:           validKey,
			AllowedIPs:          []string{"10.0.0.1/32"},
			PersistentKeepalive: 65536,
		}
		err := req.Validate()
		require.Error(t, err)
	})
}

// ==================== Peer.IsHandshakeRecent Tests ====================

func TestPeer_IsHandshakeRecent(t *testing.T) {
	t.Run("Returns False When No Handshake", func(t *testing.T) {
		peer := &Peer{LastHandshake: nil}
		assert.False(t, peer.IsHandshakeRecent(5*time.Minute))
	})

	t.Run("Returns True When Handshake Is Recent", func(t *testing.T) {
		recent := time.Now().Add(-time.Minute)
		peer := &Peer{LastHandshake: &recent}
		assert.True(t, peer.IsHandshakeRecent(5*time.Minute))
	})

	t.Run("Returns False When Handshake Is Old", func(t *testing.T) {
		old := time.Now().Add(-10 * time.Minute)
		peer := &Peer{LastHandshake: &old}
		assert.False(t, peer.IsHandshakeRecent(5*time.Minute))
	})
}

// ==================== Peer.IsActive Tests ====================

func TestPeer_IsActive(t *testing.T) {
	t.Run("Active When Active And Recent Handshake", func(t *testing.T) {
		recent := time.Now().Add(-time.Minute)
		peer := &Peer{
			Active:        true,
			LastHandshake: &recent,
		}
		assert.True(t, peer.IsActive(5*time.Minute))
	})

	t.Run("Inactive When Not Active", func(t *testing.T) {
		recent := time.Now().Add(-time.Minute)
		peer := &Peer{
			Active:        false,
			LastHandshake: &recent,
		}
		assert.False(t, peer.IsActive(5*time.Minute))
	})

	t.Run("Inactive When Handshake Too Old", func(t *testing.T) {
		old := time.Now().Add(-10 * time.Minute)
		peer := &Peer{
			Active:        true,
			LastHandshake: &old,
		}
		assert.False(t, peer.IsActive(5*time.Minute))
	})
}

// ==================== Peer.GetTrafficTotal Tests ====================

func TestPeer_GetTrafficTotal(t *testing.T) {
	t.Run("Returns Sum Of Rx And Tx", func(t *testing.T) {
		peer := &Peer{
			RxBytes: 1000,
			TxBytes: 2000,
		}
		assert.Equal(t, int64(3000), peer.GetTrafficTotal())
	})

	t.Run("Returns Zero When No Traffic", func(t *testing.T) {
		peer := &Peer{
			RxBytes: 0,
			TxBytes: 0,
		}
		assert.Equal(t, int64(0), peer.GetTrafficTotal())
	})
}

// ==================== Struct Tests ====================

func TestPeer(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		now := time.Now()
		peer := Peer{
			ID:                  "peer123",
			NetworkID:           "net123",
			DeviceID:            "dev123",
			TenantID:            "tenant123",
			PublicKey:           "pubkey",
			PresharedKey:        "psk",
			Endpoint:            "1.2.3.4:51820",
			AllowedIPs:          []string{"10.0.0.1/32"},
			PersistentKeepalive: 25,
			LastHandshake:       &now,
			RxBytes:             1000,
			TxBytes:             2000,
			Active:              true,
			CreatedAt:           now,
			UpdatedAt:           now,
		}

		assert.Equal(t, "peer123", peer.ID)
		assert.Equal(t, int64(1000), peer.RxBytes)
	})
}

func TestPeerStats(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		now := time.Now()
		stats := PeerStats{
			PeerID:        "peer123",
			Endpoint:      "1.2.3.4:51820",
			LastHandshake: &now,
			RxBytes:       1000,
			TxBytes:       2000,
			Active:        true,
			Latency:       25,
		}

		assert.Equal(t, 25, stats.Latency)
	})
}

func TestUpdatePeerRequest(t *testing.T) {
	t.Run("Has All Optional Fields", func(t *testing.T) {
		endpoint := "1.2.3.4:51820"
		allowedIPs := []string{"10.0.0.0/24"}
		psk := "newpsk"
		keepalive := 30

		req := UpdatePeerRequest{
			Endpoint:            &endpoint,
			AllowedIPs:          &allowedIPs,
			PresharedKey:        &psk,
			PersistentKeepalive: &keepalive,
		}

		assert.Equal(t, "1.2.3.4:51820", *req.Endpoint)
	})
}
