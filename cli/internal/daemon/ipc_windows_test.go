//go:build windows
// +build windows

package daemon

import (
	"io"
	"sync"
	"testing"
	"time"
)

func TestPipeAddr(t *testing.T) {
	addr := &PipeAddr{name: `\\.\pipe\test`}

	if addr.Network() != "pipe" {
		t.Errorf("Network() = %q, want %q", addr.Network(), "pipe")
	}

	if addr.String() != `\\.\pipe\test` {
		t.Errorf("String() = %q, want %q", addr.String(), `\\.\pipe\test`)
	}
}

func TestIsPipeSupported(t *testing.T) {
	// On Windows, pipes should be supported
	if !IsPipeSupported() {
		t.Error("IsPipeSupported() should return true on Windows")
	}
}

func TestPipeName(t *testing.T) {
	// Verify the default pipe name is properly formatted
	expected := `\\.\pipe\goconnect-daemon`
	if PipeName != expected {
		t.Errorf("PipeName = %q, want %q", PipeName, expected)
	}
}

func TestNewPipeListener(t *testing.T) {
	// Use a unique pipe name for testing to avoid conflicts
	testPipeName := `\\.\pipe\goconnect-test-` + time.Now().Format("150405")

	listener, err := NewPipeListener(testPipeName)
	if err != nil {
		t.Fatalf("NewPipeListener() error = %v", err)
	}
	defer listener.Close()

	// Verify the listener address
	addr := listener.Addr()
	if addr.Network() != "pipe" {
		t.Errorf("Addr().Network() = %q, want %q", addr.Network(), "pipe")
	}
	if addr.String() != testPipeName {
		t.Errorf("Addr().String() = %q, want %q", addr.String(), testPipeName)
	}
}

func TestPipeListener_Close(t *testing.T) {
	testPipeName := `\\.\pipe\goconnect-test-close-` + time.Now().Format("150405")

	listener, err := NewPipeListener(testPipeName)
	if err != nil {
		t.Fatalf("NewPipeListener() error = %v", err)
	}

	// Close should succeed
	if err := listener.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Second close should also succeed (idempotent)
	if err := listener.Close(); err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestPipeListener_AcceptAfterClose(t *testing.T) {
	testPipeName := `\\.\pipe\goconnect-test-accept-` + time.Now().Format("150405")

	listener, err := NewPipeListener(testPipeName)
	if err != nil {
		t.Fatalf("NewPipeListener() error = %v", err)
	}
	listener.Close()

	// Accept should return error after close
	_, err = listener.Accept()
	if err == nil {
		t.Error("Accept() after Close() should return error")
	}
}

func TestPipeConnection(t *testing.T) {
	testPipeName := `\\.\pipe\goconnect-test-conn-` + time.Now().Format("150405")

	listener, err := NewPipeListener(testPipeName)
	if err != nil {
		t.Fatalf("NewPipeListener() error = %v", err)
	}
	defer listener.Close()

	var wg sync.WaitGroup
	var clientConn *PipeConn
	var dialErr error

	// Start client in goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Give listener time to start accepting
		time.Sleep(100 * time.Millisecond)
		clientConn, dialErr = DialPipe(testPipeName, 5*time.Second)
	}()

	// Accept connection
	serverConn, err := listener.Accept()
	if err != nil {
		t.Fatalf("Accept() error = %v", err)
	}
	defer serverConn.Close()

	wg.Wait()

	if dialErr != nil {
		t.Fatalf("DialPipe() error = %v", dialErr)
	}
	defer clientConn.Close()

	// Test writing from client to server
	testData := []byte("Hello, Named Pipe!")
	n, err := clientConn.Write(testData)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != len(testData) {
		t.Errorf("Write() = %d bytes, want %d", n, len(testData))
	}

	// Read on server
	buf := make([]byte, 256)
	n, err = serverConn.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Read() error = %v", err)
	}
	if string(buf[:n]) != string(testData) {
		t.Errorf("Read() = %q, want %q", string(buf[:n]), string(testData))
	}
}

func TestPipeConn_LocalRemoteAddr(t *testing.T) {
	testPipeName := `\\.\pipe\goconnect-test-addr-` + time.Now().Format("150405")

	listener, err := NewPipeListener(testPipeName)
	if err != nil {
		t.Fatalf("NewPipeListener() error = %v", err)
	}
	defer listener.Close()

	var wg sync.WaitGroup
	var clientConn *PipeConn

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		clientConn, _ = DialPipe(testPipeName, 5*time.Second)
	}()

	serverConn, err := listener.Accept()
	if err != nil {
		t.Fatalf("Accept() error = %v", err)
	}
	defer serverConn.Close()

	wg.Wait()
	defer clientConn.Close()

	// Check addresses
	localAddr := clientConn.LocalAddr()
	if localAddr.Network() != "pipe" {
		t.Errorf("LocalAddr().Network() = %q, want %q", localAddr.Network(), "pipe")
	}

	remoteAddr := clientConn.RemoteAddr()
	if remoteAddr.Network() != "pipe" {
		t.Errorf("RemoteAddr().Network() = %q, want %q", remoteAddr.Network(), "pipe")
	}
}

func TestPipeConn_SetDeadlines(t *testing.T) {
	// Deadlines are no-ops but should not error
	conn := &PipeConn{}

	if err := conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		t.Errorf("SetDeadline() error = %v", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Errorf("SetReadDeadline() error = %v", err)
	}

	if err := conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		t.Errorf("SetWriteDeadline() error = %v", err)
	}
}

func TestPipeConn_CloseIdempotent(t *testing.T) {
	testPipeName := `\\.\pipe\goconnect-test-closeidem-` + time.Now().Format("150405")

	listener, err := NewPipeListener(testPipeName)
	if err != nil {
		t.Fatalf("NewPipeListener() error = %v", err)
	}
	defer listener.Close()

	var wg sync.WaitGroup
	var clientConn *PipeConn

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		clientConn, _ = DialPipe(testPipeName, 5*time.Second)
	}()

	serverConn, err := listener.Accept()
	if err != nil {
		t.Fatalf("Accept() error = %v", err)
	}

	wg.Wait()

	// Close should be idempotent
	if err := serverConn.Close(); err != nil {
		t.Errorf("First Close() error = %v", err)
	}
	if err := serverConn.Close(); err != nil {
		t.Errorf("Second Close() error = %v", err)
	}

	if err := clientConn.Close(); err != nil {
		t.Errorf("Client Close() error = %v", err)
	}
}

func TestCreateWindowsListener(t *testing.T) {
	// This uses the default PipeName, so close quickly to avoid conflicts
	listener, err := CreateWindowsListener()
	if err != nil {
		// On CI or if pipe is in use, this might fail - skip
		t.Skipf("CreateWindowsListener() error = %v (may be in use)", err)
	}
	defer listener.Close()

	addr := listener.Addr()
	if addr.String() != PipeName {
		t.Errorf("Addr().String() = %q, want %q", addr.String(), PipeName)
	}
}

func TestDialPipe_Timeout(t *testing.T) {
	// Try to dial a non-existent pipe
	nonExistentPipe := `\\.\pipe\goconnect-nonexistent-` + time.Now().Format("150405")

	start := time.Now()
	_, err := DialPipe(nonExistentPipe, 100*time.Millisecond)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("DialPipe() should fail for non-existent pipe")
	}

	// Should fail within reasonable time
	if elapsed > 2*time.Second {
		t.Errorf("DialPipe() took too long: %v", elapsed)
	}
}

func TestPipeConn_ReadAfterClose(t *testing.T) {
	testPipeName := `\\.\pipe\goconnect-test-readclose-` + time.Now().Format("150405")

	listener, err := NewPipeListener(testPipeName)
	if err != nil {
		t.Fatalf("NewPipeListener() error = %v", err)
	}
	defer listener.Close()

	var wg sync.WaitGroup
	var clientConn *PipeConn

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		clientConn, _ = DialPipe(testPipeName, 5*time.Second)
	}()

	serverConn, err := listener.Accept()
	if err != nil {
		t.Fatalf("Accept() error = %v", err)
	}

	wg.Wait()

	// Close the server connection
	serverConn.Close()

	// Reading from closed connection should error
	buf := make([]byte, 10)
	_, err = serverConn.(*PipeConn).Read(buf)
	if err == nil {
		t.Error("Read() after Close() should return error")
	}

	clientConn.Close()
}

func TestPipeConn_WriteAfterClose(t *testing.T) {
	testPipeName := `\\.\pipe\goconnect-test-writeclose-` + time.Now().Format("150405")

	listener, err := NewPipeListener(testPipeName)
	if err != nil {
		t.Fatalf("NewPipeListener() error = %v", err)
	}
	defer listener.Close()

	var wg sync.WaitGroup
	var clientConn *PipeConn

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		clientConn, _ = DialPipe(testPipeName, 5*time.Second)
	}()

	serverConn, err := listener.Accept()
	if err != nil {
		t.Fatalf("Accept() error = %v", err)
	}

	wg.Wait()

	// Close the client connection
	clientConn.Close()

	// Writing to closed connection should error
	_, err = clientConn.Write([]byte("test"))
	if err == nil {
		t.Error("Write() after Close() should return error")
	}

	serverConn.Close()
}
