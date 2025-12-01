// Package daemon provides platform-specific IPC implementations.
//go:build !windows
// +build !windows

package daemon

import (
	"net"
)

// PipeName is not used on non-Windows platforms.
const PipeName = ""

// CreateWindowsListener is not supported on non-Windows platforms.
func CreateWindowsListener() (net.Listener, error) {
	return nil, nil
}

// IsPipeSupported returns false on non-Windows platforms.
func IsPipeSupported() bool {
	return false
}
