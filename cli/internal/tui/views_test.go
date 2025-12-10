package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== viewDashboard Tests ====================

func TestModel_viewDashboard(t *testing.T) {
	t.Run("Renders With Nil Status", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.status = nil

		view := m.viewDashboard()
		assert.NotEmpty(t, view)
		assert.Contains(t, view, "DISCONNECTED")
	})

	t.Run("Renders Connected Status", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.status = &Status{
			Connected:   true,
			IP:          "10.0.0.5",
			NetworkName: "TestNet",
		}

		view := m.viewDashboard()
		assert.NotEmpty(t, view)
		assert.Contains(t, view, "CONNECTED")
		assert.Contains(t, view, "10.0.0.5")
		assert.Contains(t, view, "TestNet")
	})

	t.Run("Shows HTTP IPC Mode", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.usingGRPC = false

		view := m.viewDashboard()
		assert.Contains(t, view, "HTTP")
	})

	t.Run("Shows gRPC IPC Mode", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.usingGRPC = true

		view := m.viewDashboard()
		assert.Contains(t, view, "gRPC")
	})

	t.Run("Shows Network Statistics", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.status = &Status{
			Connected:     true,
			OnlineMembers: 5,
			Role:          "admin",
			Networks:      []Network{{ID: "n1"}, {ID: "n2"}},
		}

		view := m.viewDashboard()
		assert.Contains(t, view, "Network Statistics")
		assert.Contains(t, view, "Online Members")
		assert.Contains(t, view, "5")
		assert.Contains(t, view, "admin")
	})

	t.Run("Shows Loading Stats When No Status", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.status = nil

		view := m.viewDashboard()
		assert.Contains(t, view, "Loading stats...")
	})
}

// ==================== viewForm Tests ====================

func TestModel_viewForm(t *testing.T) {
	t.Run("Renders Form With Title And Label", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24

		view := m.viewForm("Create Network", "Enter network name:")
		assert.NotEmpty(t, view)
		assert.Contains(t, view, "Create Network")
		assert.Contains(t, view, "Enter network name:")
	})

	t.Run("Shows Help Text", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24

		view := m.viewForm("Test Form", "Test label")
		assert.Contains(t, view, "Enter")
		assert.Contains(t, view, "Esc")
	})

	t.Run("Renders Different Titles", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24

		// Create network form
		view1 := m.viewForm("Create Network", "Enter a name for your new network:")
		assert.Contains(t, view1, "Create Network")

		// Join network form
		view2 := m.viewForm("Join Network", "Enter the invite code or link:")
		assert.Contains(t, view2, "Join Network")
		assert.Contains(t, view2, "invite code")
	})
}

// ==================== viewPeers Tests ====================

func TestModel_viewPeers(t *testing.T) {
	t.Run("Renders Empty Peers List", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.peers = []Peer{}

		view := m.viewPeers()
		assert.NotEmpty(t, view)
		assert.Contains(t, view, "Connected Peers")
		assert.Contains(t, view, "No peers connected")
	})

	t.Run("Renders Peer List", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.peers = []Peer{
			{Name: "Peer1", VirtualIP: "10.0.0.2", Status: "connected"},
			{Name: "Peer2", VirtualIP: "10.0.0.3", Status: "connecting"},
		}

		view := m.viewPeers()
		assert.Contains(t, view, "Connected Peers")
		assert.Contains(t, view, "Peer1")
		assert.Contains(t, view, "Peer2")
		assert.Contains(t, view, "10.0.0.2")
		assert.Contains(t, view, "10.0.0.3")
	})

	t.Run("Shows Connection Type P2P", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.peers = []Peer{
			{Name: "DirectPeer", VirtualIP: "10.0.0.2", Status: "connected", ConnectionType: "direct"},
		}

		view := m.viewPeers()
		assert.Contains(t, view, "P2P")
	})

	t.Run("Shows Connection Type Relay", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.peers = []Peer{
			{Name: "RelayPeer", VirtualIP: "10.0.0.2", Status: "connected", ConnectionType: "relay"},
		}

		view := m.viewPeers()
		assert.Contains(t, view, "Relay")
	})

	t.Run("Shows Latency When Available", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.peers = []Peer{
			{Name: "FastPeer", VirtualIP: "10.0.0.2", Status: "connected", LatencyMs: 25},
		}

		view := m.viewPeers()
		assert.Contains(t, view, "25ms")
	})

	t.Run("Peer Status Connected Shows Green", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.peers = []Peer{
			{Name: "OnlinePeer", Status: "connected", VirtualIP: "10.0.0.2"},
		}

		view := m.viewPeers()
		assert.Contains(t, view, "OnlinePeer")
		// Status indicator should be present (●)
		assert.Contains(t, view, "●")
	})

	t.Run("Peer Status Failed Shows Red", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.peers = []Peer{
			{Name: "FailedPeer", Status: "failed", VirtualIP: "10.0.0.2"},
		}

		view := m.viewPeers()
		assert.Contains(t, view, "FailedPeer")
	})

	t.Run("Peer Status Connecting Shows Yellow", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.peers = []Peer{
			{Name: "ConnectingPeer", Status: "connecting", VirtualIP: "10.0.0.2"},
		}

		view := m.viewPeers()
		assert.Contains(t, view, "ConnectingPeer")
	})
}

// ==================== Nil Safety Tests ====================

func TestViews_NilSafety(t *testing.T) {
	t.Run("Dashboard Handles Nil Values", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.status = nil
		m.networks = nil
		m.peers = nil

		// Should not panic
		view := m.viewDashboard()
		assert.NotEmpty(t, view)
	})

	t.Run("Peers View Handles Nil Peers", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.peers = nil

		// Should not panic
		view := m.viewPeers()
		assert.Contains(t, view, "No peers connected")
	})
}
