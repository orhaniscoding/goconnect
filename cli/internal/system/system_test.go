package system

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewHostsManager(t *testing.T) {
	m := NewHostsManager()
	if m == nil {
		t.Fatal("Expected HostsManager, got nil")
	}

	// Verify path is set based on OS
	if runtime.GOOS == "windows" {
		if !strings.Contains(m.filePath, "System32") {
			t.Errorf("Expected Windows hosts path, got %s", m.filePath)
		}
	} else {
		if m.filePath != "/etc/hosts" {
			t.Errorf("Expected /etc/hosts, got %s", m.filePath)
		}
	}
}

func TestHostEntry_Struct(t *testing.T) {
	entry := HostEntry{
		IP:       "192.168.1.100",
		Hostname: "peer.goconnect.local",
	}

	if entry.IP != "192.168.1.100" {
		t.Errorf("Expected IP '192.168.1.100', got %s", entry.IP)
	}
	if entry.Hostname != "peer.goconnect.local" {
		t.Errorf("Expected Hostname 'peer.goconnect.local', got %s", entry.Hostname)
	}
}

func TestHostsManager_UpdateHosts(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	// Create initial hosts file
	initialContent := "127.0.0.1 localhost\n::1 localhost\n"
	if err := os.WriteFile(hostsPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test hosts file: %v", err)
	}

	m := &HostsManager{filePath: hostsPath}

	entries := []HostEntry{
		{IP: "10.0.0.1", Hostname: "peer1.goconnect"},
		{IP: "10.0.0.2", Hostname: "peer2.goconnect"},
	}

	if err := m.UpdateHosts(entries); err != nil {
		t.Fatalf("UpdateHosts failed: %v", err)
	}

	// Read back and verify
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatalf("Failed to read hosts file: %v", err)
	}

	contentStr := string(content)

	// Verify original content preserved
	if !strings.Contains(contentStr, "127.0.0.1") {
		t.Error("Expected original localhost entry to be preserved")
	}

	// Verify markers present
	if !strings.Contains(contentStr, startMarker) {
		t.Error("Expected start marker in hosts file")
	}
	if !strings.Contains(contentStr, endMarker) {
		t.Error("Expected end marker in hosts file")
	}

	// Verify entries added
	if !strings.Contains(contentStr, "10.0.0.1 peer1.goconnect") {
		t.Error("Expected peer1 entry in hosts file")
	}
	if !strings.Contains(contentStr, "10.0.0.2 peer2.goconnect") {
		t.Error("Expected peer2 entry in hosts file")
	}
}

func TestHostsManager_UpdateHosts_ReplaceExisting(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	// Create hosts file with existing GoConnect block
	initialContent := `127.0.0.1 localhost
# BEGIN GoConnect Managed Block
10.0.0.99 old-peer.goconnect # GoConnect
# END GoConnect Managed Block
::1 localhost
`
	if err := os.WriteFile(hostsPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test hosts file: %v", err)
	}

	m := &HostsManager{filePath: hostsPath}

	entries := []HostEntry{
		{IP: "10.0.0.1", Hostname: "new-peer.goconnect"},
	}

	if err := m.UpdateHosts(entries); err != nil {
		t.Fatalf("UpdateHosts failed: %v", err)
	}

	content, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatalf("Failed to read hosts file: %v", err)
	}

	contentStr := string(content)

	// Old entry should be gone
	if strings.Contains(contentStr, "10.0.0.99") {
		t.Error("Old peer entry should have been removed")
	}
	if strings.Contains(contentStr, "old-peer") {
		t.Error("Old peer hostname should have been removed")
	}

	// New entry should be present
	if !strings.Contains(contentStr, "10.0.0.1 new-peer.goconnect") {
		t.Error("Expected new peer entry in hosts file")
	}

	// Only one start/end marker pair should exist
	startCount := strings.Count(contentStr, startMarker)
	if startCount != 1 {
		t.Errorf("Expected 1 start marker, got %d", startCount)
	}
}

func TestHostsManager_UpdateHosts_ClearEntries(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	// Create hosts file with existing GoConnect block
	initialContent := `127.0.0.1 localhost
# BEGIN GoConnect Managed Block
10.0.0.1 peer.goconnect # GoConnect
# END GoConnect Managed Block
`
	if err := os.WriteFile(hostsPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test hosts file: %v", err)
	}

	m := &HostsManager{filePath: hostsPath}

	// Update with empty entries to clear block
	if err := m.UpdateHosts([]HostEntry{}); err != nil {
		t.Fatalf("UpdateHosts failed: %v", err)
	}

	content, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatalf("Failed to read hosts file: %v", err)
	}

	contentStr := string(content)

	// Block should be removed
	if strings.Contains(contentStr, startMarker) {
		t.Error("Start marker should have been removed when clearing entries")
	}
	if strings.Contains(contentStr, endMarker) {
		t.Error("End marker should have been removed when clearing entries")
	}

	// Original content should remain
	if !strings.Contains(contentStr, "127.0.0.1 localhost") {
		t.Error("Original localhost entry should be preserved")
	}
}

func TestHostsManager_UpdateHosts_SkipsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	if err := os.WriteFile(hostsPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test hosts file: %v", err)
	}

	m := &HostsManager{filePath: hostsPath}

	entries := []HostEntry{
		{IP: "10.0.0.1", Hostname: "valid.goconnect"},
		{IP: "", Hostname: "no-ip.goconnect"},      // Should be skipped
		{IP: "10.0.0.2", Hostname: ""},             // Should be skipped
		{IP: "  ", Hostname: "  "},                 // Should be skipped (whitespace only)
	}

	if err := m.UpdateHosts(entries); err != nil {
		t.Fatalf("UpdateHosts failed: %v", err)
	}

	content, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatalf("Failed to read hosts file: %v", err)
	}

	contentStr := string(content)

	// Valid entry should be present
	if !strings.Contains(contentStr, "10.0.0.1 valid.goconnect") {
		t.Error("Expected valid entry in hosts file")
	}

	// Invalid entries should not be present
	if strings.Contains(contentStr, "no-ip") {
		t.Error("Entry without IP should not be in hosts file")
	}
	if strings.Contains(contentStr, "10.0.0.2") {
		t.Error("Entry without hostname should not be in hosts file")
	}
}

func TestHostsManager_ReadHosts(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	content := "127.0.0.1 localhost\n10.0.0.1 peer1\n10.0.0.2 peer2\n"
	if err := os.WriteFile(hostsPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test hosts file: %v", err)
	}

	m := &HostsManager{filePath: hostsPath}

	lines, err := m.ReadHosts()
	if err != nil {
		t.Fatalf("ReadHosts failed: %v", err)
	}

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	if lines[0] != "127.0.0.1 localhost" {
		t.Errorf("Expected first line '127.0.0.1 localhost', got %s", lines[0])
	}
}

func TestGetOSVersion(t *testing.T) {
	version := GetOSVersion()
	if version == "" {
		t.Error("Expected non-empty OS version")
	}

	// Should contain OS identifier
	switch runtime.GOOS {
	case "windows":
		// Windows returns "Microsoft Windows [Version ...]" or just "Windows"
		if !strings.Contains(strings.ToLower(version), "windows") {
			t.Errorf("Expected 'windows' in version, got %s", version)
		}
	case "darwin":
		if !strings.Contains(strings.ToLower(version), "macos") {
			t.Errorf("Expected 'macos' in version, got %s", version)
		}
	case "linux":
		// Linux might return distro name or just "Linux"
		if version != "Linux" && !strings.Contains(version, " ") {
			// Should be either "Linux" or a distro name with version
		}
	}
}

func TestHostsManager_UpdateHosts_FileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "nonexistent", "hosts")

	m := &HostsManager{filePath: hostsPath}

	entries := []HostEntry{
		{IP: "10.0.0.1", Hostname: "peer.goconnect"},
	}

	err := m.UpdateHosts(entries)
	if err == nil {
		t.Error("Expected error when hosts file doesn't exist")
	}
}
