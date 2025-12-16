package metrics

import (
	"testing"
)

func TestSetWSConnections(t *testing.T) {
	SetWSConnections(5)
	summary := GetSummary()
	if summary.WSConnections != 0 {
		// Note: Summary uses cache, not direct Prometheus read
		// This tests the function doesn't panic
	}
}

func TestSetWSRooms(t *testing.T) {
	SetWSRooms(3)
	// Test function executes without panic
}

func TestIncWSMessage(t *testing.T) {
	IncWSMessage("inbound", "chat.send")
	IncWSMessage("outbound", "chat.message")
	// Counters increment without panic
}

func TestSetNetworksActive(t *testing.T) {
	SetNetworksActive(2)
}

func TestSetPeersOnline(t *testing.T) {
	SetPeersOnline(8)
}

func TestSetMembershipTotal(t *testing.T) {
	SetMembershipTotal(15)
}

func TestUpdateSummaryCache(t *testing.T) {
	UpdateSummaryCache(10, 3, 2, 8, 15, 6)
	
	summary := GetSummary()
	
	if summary.WSConnections != 10 {
		t.Errorf("WSConnections = %d, want 10", summary.WSConnections)
	}
	if summary.WSRooms != 3 {
		t.Errorf("WSRooms = %d, want 3", summary.WSRooms)
	}
	if summary.NetworksActive != 2 {
		t.Errorf("NetworksActive = %d, want 2", summary.NetworksActive)
	}
	if summary.PeersOnline != 8 {
		t.Errorf("PeersOnline = %d, want 8", summary.PeersOnline)
	}
	if summary.Memberships != 15 {
		t.Errorf("Memberships = %d, want 15", summary.Memberships)
	}
	if summary.WGPeers != 6 {
		t.Errorf("WGPeers = %d, want 6", summary.WGPeers)
	}
}

func TestGetSummary_Defaults(t *testing.T) {
	// Reset cache
	UpdateSummaryCache(0, 0, 0, 0, 0, 0)
	
	summary := GetSummary()
	
	if summary.WSConnections != 0 {
		t.Errorf("Expected 0 connections, got %d", summary.WSConnections)
	}
}

func TestSummaryStruct(t *testing.T) {
	summary := Summary{
		WSConnections:  12,
		WSRooms:        4,
		NetworksActive: 3,
		PeersOnline:    15,
		Memberships:    20,
		WGPeers:        8,
	}

	if summary.WSConnections != 12 {
		t.Errorf("Expected 12, got %d", summary.WSConnections)
	}
}
