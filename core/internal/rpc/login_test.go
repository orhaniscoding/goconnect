package rpc_test

import (
	"context"
	"errors"
	"testing"

	"net"

	"github.com/orhaniscoding/goconnect/server/internal/auth"
	"github.com/orhaniscoding/goconnect/server/internal/backend"
	"github.com/orhaniscoding/goconnect/server/internal/proto"
	"github.com/orhaniscoding/goconnect/server/internal/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// MockDaemon implements the daemon interface required by DaemonHandler
type MockDaemon struct {
	mock.Mock
}

func (m *MockDaemon) Stop() {
	m.Called()
}

func (m *MockDaemon) GetBackend() interface {
	RequestDeviceCode(context.Context) (*backend.DeviceCodeResponse, error)
	PollDeviceToken(context.Context, string) (*backend.AuthResponse, error)
} {
	args := m.Called()
	if b, ok := args.Get(0).(interface {
		RequestDeviceCode(context.Context) (*backend.DeviceCodeResponse, error)
		PollDeviceToken(context.Context, string) (*backend.AuthResponse, error)
	}); ok {
		return b
	}
	return nil
}

func (m *MockDaemon) GetTokenManager() auth.TokenManager {
	args := m.Called()
	if tm, ok := args.Get(0).(auth.TokenManager); ok {
		return tm
	}
	return nil
}

// MockTokenManager implements auth.TokenManager
type MockTokenManager struct {
	mock.Mock
}

func (m *MockTokenManager) SaveSession(session *auth.TokenSession) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockTokenManager) LoadSession() (*auth.TokenSession, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenSession), args.Error(1)
}

func (m *MockTokenManager) ClearSession() error {
	return m.Called().Error(0)
}

// MockBackend implements the backend logic
type MockBackend struct {
	mock.Mock
}

func (m *MockBackend) RequestDeviceCode(ctx context.Context) (*backend.DeviceCodeResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backend.DeviceCodeResponse), args.Error(1)
}

func (m *MockBackend) PollDeviceToken(ctx context.Context, deviceCode string) (*backend.AuthResponse, error) {
	args := m.Called(ctx, deviceCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backend.AuthResponse), args.Error(1)
}

func TestLogin_Flow(t *testing.T) {
	// Setup
	lis := bufconn.Listen(1024 * 1024)
	s := rpc.NewServer()

	mockBackend := new(MockBackend)
	mockTokenMgr := new(MockTokenManager)
	mockDaemon := new(MockDaemon)

	mockDaemon.On("GetBackend").Return(mockBackend)
	mockDaemon.On("GetTokenManager").Return(mockTokenMgr)

	handler := rpc.NewDaemonHandler(mockDaemon, "test")
	proto.RegisterDaemonServiceServer(s.GetGRPCServer(), handler)

	go func() {
		if err := s.Start(lis); err != nil {
			panic(err)
		}
	}()
	defer s.GracefulStop()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := proto.NewDaemonServiceClient(conn)

	// Expectations
	mockBackend.On("RequestDeviceCode", mock.Anything).Return(&backend.DeviceCodeResponse{
		DeviceCode:      "dc-123",
		UserCode:        "ABCD-1234",
		VerificationURI: "http://test.com/activate",
		ExpiresIn:       300,
		Interval:        1, // 1 second
	}, nil)

	// Poll 1: Pending
	mockBackend.On("PollDeviceToken", mock.Anything, "dc-123").Return(nil, errors.New("authorization_pending")).Once()

	// Poll 2: Success
	mockBackend.On("PollDeviceToken", mock.Anything, "dc-123").Return(&backend.AuthResponse{
		AccessToken:  "acc-token",
		RefreshToken: "ref-token",
		ExpiresIn:    3600,
	}, nil).Once()

	// Token Save Expectation
	mockTokenMgr.On("SaveSession", mock.MatchedBy(func(s *auth.TokenSession) bool {
		return s.AccessToken == "acc-token" && s.RefreshToken == "ref-token"
	})).Return(nil)

	// Execute
	stream, err := client.Login(context.Background(), &proto.LoginRequest{ClientName: "test-client"})
	require.NoError(t, err)

	// 1. Instructions
	update, err := stream.Recv()
	require.NoError(t, err)
	instr, ok := update.Update.(*proto.LoginUpdate_Instructions)
	require.True(t, ok)
	assert.Equal(t, "ABCD-1234", instr.Instructions.Code)

	// 2. Success (after polling)
	update, err = stream.Recv()
	require.NoError(t, err)
	success, ok := update.Update.(*proto.LoginUpdate_Success)
	require.True(t, ok)
	assert.Equal(t, "authenticated", success.Success.UserId) // Placeholder until userinfo endpoint integration

	mockBackend.AssertExpectations(t)
	mockTokenMgr.AssertExpectations(t)
}
