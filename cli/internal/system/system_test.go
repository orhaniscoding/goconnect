package system

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHostsManager(t *testing.T) {
	m := NewHostsManager()
	require.NotNil(t, m, "Expected HostsManager, got nil")

	// Verify path is set based on OS
	if runtime.GOOS == "windows" {
		assert.Contains(t, m.filePath, "System32", "Expected Windows hosts path")
	} else {
		assert.Equal(t, "/etc/hosts", m.filePath, "Expected /etc/hosts")
	}
}

func TestNewHostsManager_WindowsPath(t *testing.T) {
	// We can't truly test Windows path on Linux, but we can verify the structure
	m := NewHostsManager()
	assert.NotEmpty(t, m.filePath, "File path should not be empty")
}

func TestHostEntry_Struct(t *testing.T) {
	entry := HostEntry{
		IP:       "192.168.1.100",
		Hostname: "peer.goconnect.local",
	}

	assert.Equal(t, "192.168.1.100", entry.IP)
	assert.Equal(t, "peer.goconnect.local", entry.Hostname)
}

func TestHostEntry_EmptyValues(t *testing.T) {
	entry := HostEntry{}
	assert.Empty(t, entry.IP)
	assert.Empty(t, entry.Hostname)
}

func TestHostsManager_UpdateHosts(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	// Create initial hosts file
	initialContent := "127.0.0.1 localhost\n::1 localhost\n"
	err := os.WriteFile(hostsPath, []byte(initialContent), 0644)
	require.NoError(t, err, "Failed to create test hosts file")

	m := &HostsManager{filePath: hostsPath}

	entries := []HostEntry{
		{IP: "10.0.0.1", Hostname: "peer1.goconnect"},
		{IP: "10.0.0.2", Hostname: "peer2.goconnect"},
	}

	err = m.UpdateHosts(entries)
	require.NoError(t, err, "UpdateHosts failed")

	// Read back and verify
	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err, "Failed to read hosts file")

	contentStr := string(content)

	// Verify original content preserved
	assert.Contains(t, contentStr, "127.0.0.1", "Expected original localhost entry to be preserved")

	// Verify markers present
	assert.Contains(t, contentStr, startMarker, "Expected start marker in hosts file")
	assert.Contains(t, contentStr, endMarker, "Expected end marker in hosts file")

	// Verify entries added
	assert.Contains(t, contentStr, "10.0.0.1 peer1.goconnect", "Expected peer1 entry in hosts file")
	assert.Contains(t, contentStr, "10.0.0.2 peer2.goconnect", "Expected peer2 entry in hosts file")
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
	err := os.WriteFile(hostsPath, []byte(initialContent), 0644)
	require.NoError(t, err, "Failed to create test hosts file")

	m := &HostsManager{filePath: hostsPath}

	entries := []HostEntry{
		{IP: "10.0.0.1", Hostname: "new-peer.goconnect"},
	}

	err = m.UpdateHosts(entries)
	require.NoError(t, err, "UpdateHosts failed")

	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err, "Failed to read hosts file")

	contentStr := string(content)

	// Old entry should be gone
	assert.NotContains(t, contentStr, "10.0.0.99", "Old peer entry should have been removed")
	assert.NotContains(t, contentStr, "old-peer", "Old peer hostname should have been removed")

	// New entry should be present
	assert.Contains(t, contentStr, "10.0.0.1 new-peer.goconnect", "Expected new peer entry in hosts file")

	// Only one start/end marker pair should exist
	startCount := strings.Count(contentStr, startMarker)
	assert.Equal(t, 1, startCount, "Expected exactly 1 start marker")
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
	err := os.WriteFile(hostsPath, []byte(initialContent), 0644)
	require.NoError(t, err, "Failed to create test hosts file")

	m := &HostsManager{filePath: hostsPath}

	// Update with empty entries to clear block
	err = m.UpdateHosts([]HostEntry{})
	require.NoError(t, err, "UpdateHosts failed")

	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err, "Failed to read hosts file")

	contentStr := string(content)

	// Block should be removed
	assert.NotContains(t, contentStr, startMarker, "Start marker should have been removed when clearing entries")
	assert.NotContains(t, contentStr, endMarker, "End marker should have been removed when clearing entries")

	// Original content should remain
	assert.Contains(t, contentStr, "127.0.0.1 localhost", "Original localhost entry should be preserved")
}

func TestHostsManager_UpdateHosts_SkipsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	err := os.WriteFile(hostsPath, []byte(""), 0644)
	require.NoError(t, err, "Failed to create test hosts file")

	m := &HostsManager{filePath: hostsPath}

	entries := []HostEntry{
		{IP: "10.0.0.1", Hostname: "valid.goconnect"},
		{IP: "", Hostname: "no-ip.goconnect"},      // Should be skipped
		{IP: "10.0.0.2", Hostname: ""},             // Should be skipped
		{IP: "  ", Hostname: "  "},                 // Should be skipped (whitespace only)
	}

	err = m.UpdateHosts(entries)
	require.NoError(t, err, "UpdateHosts failed")

	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err, "Failed to read hosts file")

	contentStr := string(content)

	// Valid entry should be present
	assert.Contains(t, contentStr, "10.0.0.1 valid.goconnect", "Expected valid entry in hosts file")

	// Invalid entries should not be present
	assert.NotContains(t, contentStr, "no-ip", "Entry without IP should not be in hosts file")
	assert.NotContains(t, contentStr, "10.0.0.2", "Entry without hostname should not be in hosts file")
}

func TestHostsManager_ReadHosts(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	content := "127.0.0.1 localhost\n10.0.0.1 peer1\n10.0.0.2 peer2\n"
	err := os.WriteFile(hostsPath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create test hosts file")

	m := &HostsManager{filePath: hostsPath}

	lines, err := m.ReadHosts()
	require.NoError(t, err, "ReadHosts failed")

	assert.Len(t, lines, 3, "Expected 3 lines")
	assert.Equal(t, "127.0.0.1 localhost", lines[0])
}

func TestHostsManager_ReadHosts_FileNotExist(t *testing.T) {
	m := &HostsManager{filePath: "/nonexistent/path/hosts"}
	
	lines, err := m.ReadHosts()
	assert.Error(t, err, "Expected error when file doesn't exist")
	assert.Nil(t, lines, "Lines should be nil on error")
}

func TestHostsManager_ReadHosts_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	err := os.WriteFile(hostsPath, []byte(""), 0644)
	require.NoError(t, err)

	m := &HostsManager{filePath: hostsPath}

	lines, err := m.ReadHosts()
	require.NoError(t, err)
	assert.Empty(t, lines, "Expected empty lines for empty file")
}

func TestGetOSVersion(t *testing.T) {
	version := GetOSVersion()
	assert.NotEmpty(t, version, "Expected non-empty OS version")

	// Should contain OS identifier
	switch runtime.GOOS {
	case "windows":
		// Windows returns "Microsoft Windows [Version ...]" or just "Windows"
		assert.Contains(t, strings.ToLower(version), "windows", "Expected 'windows' in version")
	case "darwin":
		assert.Contains(t, strings.ToLower(version), "macos", "Expected 'macos' in version")
	case "linux":
		// Linux might return distro name or just "Linux"
		assert.NotEmpty(t, version, "Linux version should not be empty")
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
	assert.Error(t, err, "Expected error when hosts file doesn't exist")
	assert.Contains(t, err.Error(), "failed to read hosts file", "Error should mention reading failure")
}

func TestHostsManager_UpdateHosts_MultipleBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	// Create hosts file with malformed block (multiple start markers)
	initialContent := `127.0.0.1 localhost
# BEGIN GoConnect Managed Block
10.0.0.1 peer1.goconnect # GoConnect
# BEGIN GoConnect Managed Block
10.0.0.2 peer2.goconnect # GoConnect
# END GoConnect Managed Block
`
	err := os.WriteFile(hostsPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	m := &HostsManager{filePath: hostsPath}

	entries := []HostEntry{
		{IP: "10.0.0.3", Hostname: "new-peer.goconnect"},
	}

	err = m.UpdateHosts(entries)
	require.NoError(t, err, "UpdateHosts should handle malformed blocks")

	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err)

	contentStr := string(content)
	
	// Should have exactly one pair of markers now
	assert.Equal(t, 1, strings.Count(contentStr, startMarker), "Should have exactly 1 start marker")
	assert.Equal(t, 1, strings.Count(contentStr, endMarker), "Should have exactly 1 end marker")
	assert.Contains(t, contentStr, "10.0.0.3 new-peer.goconnect", "New entry should be present")
}

func TestHostsManager_UpdateHosts_NoTrailingNewline(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	// Create hosts file without trailing newline
	initialContent := "127.0.0.1 localhost"
	err := os.WriteFile(hostsPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	m := &HostsManager{filePath: hostsPath}

	entries := []HostEntry{
		{IP: "10.0.0.1", Hostname: "peer.goconnect"},
	}

	err = m.UpdateHosts(entries)
	require.NoError(t, err)

	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err)

	contentStr := string(content)
	
	// Verify markers and entries are properly separated
	assert.Contains(t, contentStr, "127.0.0.1 localhost", "Original entry should be preserved")
	assert.Contains(t, contentStr, startMarker, "Start marker should be present")
	assert.Contains(t, contentStr, "10.0.0.1 peer.goconnect", "New entry should be present")
}

func TestHostsManager_UpdateHosts_LargeNumberOfEntries(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	err := os.WriteFile(hostsPath, []byte("127.0.0.1 localhost\n"), 0644)
	require.NoError(t, err)

	m := &HostsManager{filePath: hostsPath}

	// Create many entries
	entries := make([]HostEntry, 100)
	for i := 0; i < 100; i++ {
		entries[i] = HostEntry{
			IP:       "10.0.0." + string(rune('0'+i%10)),
			Hostname: "peer" + string(rune('0'+i%10)) + ".goconnect",
		}
	}

	err = m.UpdateHosts(entries)
	require.NoError(t, err, "Should handle large number of entries")
}

func TestNewConfigurator(t *testing.T) {
	c := NewConfigurator()
	assert.NotNil(t, c, "NewConfigurator should return a non-nil configurator")
}

func TestNewProtocolHandler(t *testing.T) {
	h := NewProtocolHandler()
	assert.NotNil(t, h, "NewProtocolHandler should return a non-nil handler")
}

func TestHostsManager_Concurrency(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	err := os.WriteFile(hostsPath, []byte("127.0.0.1 localhost\n"), 0644)
	require.NoError(t, err)

	m := &HostsManager{filePath: hostsPath}

	// Test that mutex prevents race conditions
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			entries := []HostEntry{
				{IP: "10.0.0.1", Hostname: "peer.goconnect"},
			}
			m.UpdateHosts(entries)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify file is not corrupted
	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "127.0.0.1 localhost", "Original content should be preserved")
}

func TestHostsManager_UpdateHosts_SpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	err := os.WriteFile(hostsPath, []byte("127.0.0.1 localhost\n"), 0644)
	require.NoError(t, err)

	m := &HostsManager{filePath: hostsPath}

	// Test with various valid hostnames
	entries := []HostEntry{
		{IP: "10.0.0.1", Hostname: "peer-1.goconnect.local"},
		{IP: "10.0.0.2", Hostname: "peer_2.goconnect.local"},
		{IP: "fc00::1", Hostname: "ipv6-peer.goconnect"},
	}

	err = m.UpdateHosts(entries)
	require.NoError(t, err)

	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "10.0.0.1 peer-1.goconnect.local", "Hyphenated hostname should work")
	assert.Contains(t, contentStr, "10.0.0.2 peer_2.goconnect.local", "Underscore hostname should work")
	assert.Contains(t, contentStr, "fc00::1 ipv6-peer.goconnect", "IPv6 address should work")
}
