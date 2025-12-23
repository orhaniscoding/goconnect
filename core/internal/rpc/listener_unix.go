package rpc

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/orhaniscoding/goconnect/server/internal/logger"
)

// ListenUnix creates a Unix domain socket listener at the specified path.
// It ensures stale socket files are removed and sets strict 0600 permissions.
func ListenUnix(socketPath string) (net.Listener, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(socketPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Remove stale socket file if it exists
	if _, err := os.Stat(socketPath); err == nil {
		logger.Debug("Removing stale socket file", "path", socketPath)
		if err := os.Remove(socketPath); err != nil {
			return nil, fmt.Errorf("failed to remove stale socket: %w", err)
		}
	}

	// Listen on Unix domain socket
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on unix socket: %w", err)
	}

	// Set permissions to 0600 (owner read/write only)
	if err := os.Chmod(socketPath, 0600); err != nil {
		lis.Close()
		return nil, fmt.Errorf("failed to set socket permissions: %w", err)
	}

	return lis, nil
}
