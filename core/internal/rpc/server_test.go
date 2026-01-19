package rpc

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/auth"
	"github.com/orhaniscoding/goconnect/server/internal/backend"
	"github.com/orhaniscoding/goconnect/server/internal/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockDaemon struct {
	stopCalled bool
}

func (m *mockDaemon) Stop() {
	m.stopCalled = true
}

func (m *mockDaemon) GetBackend() interface {
	RequestDeviceCode(context.Context) (*backend.DeviceCodeResponse, error)
	PollDeviceToken(context.Context, string) (*backend.AuthResponse, error)
} {
	return nil // Not used in this test
}

func (m *mockDaemon) GetTokenManager() auth.TokenManager {
	return nil
}

func TestListenUnix_Permissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix sockets permissions test skipped on Windows")
	}

	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	lis, err := ListenUnix(socketPath)
	require.NoError(t, err)
	defer lis.Close()

	info, err := os.Stat(socketPath)
	require.NoError(t, err)

	// Check permissions are 0600 (srw-------)
	// Note: os.FileMode for socket might include the S_IFSOCK bit
	mode := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0600), mode, "Socket should have 0600 permissions")
}

func TestServer_StartStop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix socket test skipped on Windows - named pipes work differently")
	}

	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test_server.sock")

	lis, err := ListenUnix(socketPath)
	require.NoError(t, err)

	srv := NewServer()
	md := &mockDaemon{}
	handler := NewDaemonHandler(md, "test-version")
	proto.RegisterDaemonServiceServer(srv.GetGRPCServer(), handler)

	go func() {
		_ = srv.Start(lis)
	}()

	// Give it a moment to start
	time.Sleep(200 * time.Millisecond)

	// Connect and test GetVersion
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	require.NoError(t, err)
	defer conn.Close()

	client := proto.NewDaemonServiceClient(conn)
	resp, err := client.GetVersion(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	assert.Equal(t, "test-version", resp.Version)

	// Test Shutdown call
	_, err = client.Shutdown(ctx, &emptypb.Empty{})
	require.NoError(t, err)

	// Shutdown is async (go h.daemon.Stop()), wait a bit
	time.Sleep(100 * time.Millisecond)
	assert.True(t, md.stopCalled)

	srv.GracefulStop()
}
