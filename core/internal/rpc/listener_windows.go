package rpc

import (
	"errors"
	"net"
	"runtime"
)

// ListenNamedPipe creates a Windows Named Pipe listener.
// This is a stub for now as UDS is preferred on Unix, but could be used on Windows.
func ListenNamedPipe(pipeName string) (net.Listener, error) {
	if runtime.GOOS != "windows" {
		return nil, errors.New("named pipes are only supported on Windows")
	}
	
	// In the future, we could use github.com/Microsoft/go-winio here.
	// For now, we will fallback to TCP or UDS (Windows 10+ supports UDS).
	return nil, errors.New("named pipe listener not yet implemented")
}
