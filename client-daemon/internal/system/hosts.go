package system

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
)

const (
	startMarker = "# BEGIN GoConnect Managed Block"
	endMarker   = "# END GoConnect Managed Block"
)

// HostsManager handles /etc/hosts file manipulation
type HostsManager struct {
	mu       sync.Mutex
	filePath string
}

// NewHostsManager creates a new hosts manager
func NewHostsManager() *HostsManager {
	path := "/etc/hosts"
	if runtime.GOOS == "windows" {
		path = os.Getenv("SystemRoot") + "\\System32\\drivers\\etc\\hosts"
	}

	return &HostsManager{
		filePath: path,
	}
}

// HostEntry represents a hostname mapping
type HostEntry struct {
	IP       string
	Hostname string
}

// UpdateHosts updates the hosts file with the given entries
func (m *HostsManager) UpdateHosts(entries []HostEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Read existing content
	content, err := os.ReadFile(m.filePath)
	if err != nil {
		return fmt.Errorf("failed to read hosts file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inBlock := false
	blockFound := false

	// Filter out existing block
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == startMarker {
			inBlock = true
			blockFound = true
			continue
		}
		if trimmed == endMarker {
			inBlock = false
			continue
		}
		if !inBlock {
			newLines = append(newLines, line)
		}
	}

	// If we didn't find a block but we are appending, ensure we have a newline at the end of existing content
	if !blockFound && len(newLines) > 0 && newLines[len(newLines)-1] != "" {
		newLines = append(newLines, "")
	}

	// Construct new block
	if len(entries) > 0 {
		newLines = append(newLines, startMarker)
		for _, entry := range entries {
			// Sanitize
			ip := strings.TrimSpace(entry.IP)
			host := strings.TrimSpace(entry.Hostname)
			if ip == "" || host == "" {
				continue
			}
			newLines = append(newLines, fmt.Sprintf("%s %s # GoConnect", ip, host))
		}
		newLines = append(newLines, endMarker)
		newLines = append(newLines, "") // Trailing newline
	}

	// Write back
	output := strings.Join(newLines, "\n")
	
	// On Windows, we might need to handle line endings, but Go's os.WriteFile usually handles \n fine.
	// However, standard Windows hosts file uses CRLF.
	if runtime.GOOS == "windows" {
		output = strings.ReplaceAll(output, "\n", "\r\n")
		// Fix double CR if any
		output = strings.ReplaceAll(output, "\r\r\n", "\r\n")
	}

	// Write to temporary file first
	tmpFile := m.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write temp hosts file: %w", err)
	}

	// Move temp file to actual file (atomic-ish)
	if err := os.Rename(tmpFile, m.filePath); err != nil {
		// Fallback for Windows where Rename might fail if file exists/locked
		if runtime.GOOS == "windows" {
			if err := os.Remove(m.filePath); err == nil {
				if err := os.Rename(tmpFile, m.filePath); err == nil {
					return nil
				}
			}
		}
		os.Remove(tmpFile) // Cleanup
		return fmt.Errorf("failed to replace hosts file: %w", err)
	}

	return nil
}

// ReadHosts reads the current hosts file (helper for debugging)
func (m *HostsManager) ReadHosts() ([]string, error) {
	file, err := os.Open(m.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
