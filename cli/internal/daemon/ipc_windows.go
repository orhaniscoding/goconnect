// Package daemon provides platform-specific IPC implementations.
//go:build windows
// +build windows

package daemon

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// Named Pipe constants for Windows IPC.
const (
	// PipeName is the Windows Named Pipe path for the daemon.
	PipeName = `\\.\pipe\goconnect-daemon`

	// Named Pipe access flags
	PIPE_ACCESS_DUPLEX       = 0x00000003
	PIPE_TYPE_BYTE           = 0x00000000
	PIPE_READMODE_BYTE       = 0x00000000
	PIPE_WAIT                = 0x00000000
	PIPE_UNLIMITED_INSTANCES = 255
	NMPWAIT_WAIT_FOREVER     = 0xFFFFFFFF

	// Buffer sizes
	pipeBufferSize = 65536
)

var (
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	createNamedPipeW            = kernel32.NewProc("CreateNamedPipeW")
	connectNamedPipe            = kernel32.NewProc("ConnectNamedPipe")
	disconnectNamedPipe         = kernel32.NewProc("DisconnectNamedPipe")
	waitNamedPipeW              = kernel32.NewProc("WaitNamedPipeW")
	createFileW                 = kernel32.NewProc("CreateFileW")
	setNamedPipeHandleState     = kernel32.NewProc("SetNamedPipeHandleState")
	getNamedPipeClientProcessId = kernel32.NewProc("GetNamedPipeClientProcessId")
)

// PipeListener implements net.Listener for Windows Named Pipes.
type PipeListener struct {
	pipeName  string
	closed    bool
	closeMu   sync.Mutex
	acceptCh  chan *PipeConn
	closeCh   chan struct{}
	acceptErr error
}

// NewPipeListener creates a new Named Pipe listener.
func NewPipeListener(pipeName string) (*PipeListener, error) {
	pl := &PipeListener{
		pipeName: pipeName,
		acceptCh: make(chan *PipeConn, 1),
		closeCh:  make(chan struct{}),
	}

	// Start the accept loop in a goroutine
	go pl.acceptLoop()

	return pl, nil
}

// acceptLoop continuously accepts connections on the named pipe.
func (pl *PipeListener) acceptLoop() {
	for {
		select {
		case <-pl.closeCh:
			return
		default:
		}

		// Create a new named pipe instance
		handle, err := pl.createPipeInstance()
		if err != nil {
			pl.closeMu.Lock()
			if !pl.closed {
				pl.acceptErr = err
			}
			pl.closeMu.Unlock()
			return
		}

		// Wait for a client to connect
		err = pl.waitForClient(handle)
		if err != nil {
			syscall.CloseHandle(handle)
			pl.closeMu.Lock()
			if !pl.closed {
				continue // Try again unless closed
			}
			pl.closeMu.Unlock()
			return
		}

		// Create connection wrapper
		conn := &PipeConn{
			handle:   handle,
			pipeName: pl.pipeName,
		}

		select {
		case pl.acceptCh <- conn:
		case <-pl.closeCh:
			conn.Close()
			return
		}
	}
}

// createPipeInstance creates a new named pipe instance.
func (pl *PipeListener) createPipeInstance() (syscall.Handle, error) {
	pipeNamePtr, err := syscall.UTF16PtrFromString(pl.pipeName)
	if err != nil {
		return syscall.InvalidHandle, fmt.Errorf("invalid pipe name: %w", err)
	}

	// Create the named pipe with security attributes
	// PIPE_ACCESS_DUPLEX | FILE_FLAG_OVERLAPPED could be used for async
	handle, _, err := createNamedPipeW.Call(
		uintptr(unsafe.Pointer(pipeNamePtr)),
		PIPE_ACCESS_DUPLEX,
		PIPE_TYPE_BYTE|PIPE_READMODE_BYTE|PIPE_WAIT,
		PIPE_UNLIMITED_INSTANCES,
		pipeBufferSize,
		pipeBufferSize,
		0,
		0, // Default security (current user only)
	)

	if handle == uintptr(syscall.InvalidHandle) {
		return syscall.InvalidHandle, fmt.Errorf("CreateNamedPipe failed: %w", err)
	}

	return syscall.Handle(handle), nil
}

// waitForClient waits for a client to connect to the pipe.
func (pl *PipeListener) waitForClient(handle syscall.Handle) error {
	// ConnectNamedPipe returns true if a client connected, or false with ERROR_PIPE_CONNECTED
	ret, _, err := connectNamedPipe.Call(uintptr(handle), 0)
	if ret == 0 {
		// Check if client was already connected
		errno, ok := err.(syscall.Errno)
		if !ok || (errno != 0 && errno != 535) { // ERROR_PIPE_CONNECTED = 535
			return fmt.Errorf("ConnectNamedPipe failed: %w", err)
		}
	}
	return nil
}

// Accept waits for and returns the next connection.
func (pl *PipeListener) Accept() (net.Conn, error) {
	pl.closeMu.Lock()
	if pl.closed {
		pl.closeMu.Unlock()
		return nil, net.ErrClosed
	}
	pl.closeMu.Unlock()

	select {
	case conn := <-pl.acceptCh:
		return conn, nil
	case <-pl.closeCh:
		return nil, net.ErrClosed
	}
}

// Close stops accepting connections.
func (pl *PipeListener) Close() error {
	pl.closeMu.Lock()
	defer pl.closeMu.Unlock()

	if pl.closed {
		return nil
	}
	pl.closed = true
	close(pl.closeCh)
	return nil
}

// Addr returns the listener's network address.
func (pl *PipeListener) Addr() net.Addr {
	return &PipeAddr{name: pl.pipeName}
}

// PipeConn implements net.Conn for a Windows Named Pipe connection.
type PipeConn struct {
	handle   syscall.Handle
	pipeName string
	readMu   sync.Mutex
	writeMu  sync.Mutex
	closed   bool
	closeMu  sync.Mutex
}

// Read reads data from the pipe.
func (pc *PipeConn) Read(b []byte) (int, error) {
	pc.readMu.Lock()
	defer pc.readMu.Unlock()

	pc.closeMu.Lock()
	if pc.closed {
		pc.closeMu.Unlock()
		return 0, net.ErrClosed
	}
	pc.closeMu.Unlock()

	var n uint32
	err := syscall.ReadFile(pc.handle, b, &n, nil)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// Write writes data to the pipe.
func (pc *PipeConn) Write(b []byte) (int, error) {
	pc.writeMu.Lock()
	defer pc.writeMu.Unlock()

	pc.closeMu.Lock()
	if pc.closed {
		pc.closeMu.Unlock()
		return 0, net.ErrClosed
	}
	pc.closeMu.Unlock()

	var n uint32
	err := syscall.WriteFile(pc.handle, b, &n, nil)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// Close closes the pipe connection.
func (pc *PipeConn) Close() error {
	pc.closeMu.Lock()
	defer pc.closeMu.Unlock()

	if pc.closed {
		return nil
	}
	pc.closed = true

	// Disconnect and close the handle
	disconnectNamedPipe.Call(uintptr(pc.handle))
	return syscall.CloseHandle(pc.handle)
}

// LocalAddr returns the local network address.
func (pc *PipeConn) LocalAddr() net.Addr {
	return &PipeAddr{name: pc.pipeName}
}

// RemoteAddr returns the remote network address.
func (pc *PipeConn) RemoteAddr() net.Addr {
	return &PipeAddr{name: pc.pipeName + "-client"}
}

// SetDeadline sets the read and write deadlines.
func (pc *PipeConn) SetDeadline(t time.Time) error {
	// Named pipes don't support deadline natively
	// Would need overlapped I/O for proper timeout support
	return nil
}

// SetReadDeadline sets the deadline for future Read calls.
func (pc *PipeConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls.
func (pc *PipeConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// GetClientProcessID returns the process ID of the connected client.
func (pc *PipeConn) GetClientProcessID() (uint32, error) {
	var pid uint32
	ret, _, err := getNamedPipeClientProcessId.Call(
		uintptr(pc.handle),
		uintptr(unsafe.Pointer(&pid)),
	)
	if ret == 0 {
		return 0, fmt.Errorf("GetNamedPipeClientProcessId failed: %w", err)
	}
	return pid, nil
}

// PipeAddr represents a Windows Named Pipe address.
type PipeAddr struct {
	name string
}

// Network returns the address's network name.
func (pa *PipeAddr) Network() string {
	return "pipe"
}

// String returns the address as a string.
func (pa *PipeAddr) String() string {
	return pa.name
}

// DialPipe connects to a Windows Named Pipe.
func DialPipe(pipeName string, timeout time.Duration) (*PipeConn, error) {
	pipeNamePtr, err := syscall.UTF16PtrFromString(pipeName)
	if err != nil {
		return nil, fmt.Errorf("invalid pipe name: %w", err)
	}

	// Wait for the pipe to be available
	timeoutMs := uint32(NMPWAIT_WAIT_FOREVER)
	if timeout > 0 {
		timeoutMs = uint32(timeout.Milliseconds())
	}

	ret, _, _ := waitNamedPipeW.Call(
		uintptr(unsafe.Pointer(pipeNamePtr)),
		uintptr(timeoutMs),
	)
	if ret == 0 {
		return nil, errors.New("pipe not available")
	}

	// Open the pipe
	handle, err := syscall.CreateFile(
		pipeNamePtr,
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0,
		nil,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open pipe: %w", err)
	}

	// Set the pipe to message mode if needed
	var mode uint32 = PIPE_READMODE_BYTE
	ret, _, err = setNamedPipeHandleState.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&mode)),
		0,
		0,
	)
	if ret == 0 {
		syscall.CloseHandle(handle)
		return nil, fmt.Errorf("SetNamedPipeHandleState failed: %w", err)
	}

	return &PipeConn{
		handle:   handle,
		pipeName: pipeName,
	}, nil
}

// CreateWindowsListener creates a Named Pipe listener for Windows IPC.
// This provides better security than TCP localhost as it uses Windows
// security descriptors and can verify client process identity.
func CreateWindowsListener() (net.Listener, error) {
	return NewPipeListener(PipeName)
}

// IsPipeSupported returns true if Named Pipes are supported.
func IsPipeSupported() bool {
	return true
}
