package daemon

import (
	"context"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/chat"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/transfer"
	"github.com/stretchr/testify/mock"
)

// MockEngine is a mock implementation of DaemonEngine
type MockEngine struct {
	mock.Mock
}

func (m *MockEngine) Start() {
	m.Called()
}

func (m *MockEngine) Stop() {
	m.Called()
}

func (m *MockEngine) Connect() {
	m.Called()
}

func (m *MockEngine) Disconnect() {
	m.Called()
}

func (m *MockEngine) GetStatus() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockEngine) ManualConnect(peerID string) error {
	args := m.Called(peerID)
	return args.Error(0)
}

func (m *MockEngine) SendChatMessage(peerID, content string) error {
	args := m.Called(peerID, content)
	return args.Error(0)
}

func (m *MockEngine) SendFileRequest(peerID, filePath string) (*transfer.Session, error) {
	args := m.Called(peerID, filePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transfer.Session), args.Error(1)
}

func (m *MockEngine) AcceptFile(requestID, savePath string) error {
	args := m.Called(requestID, savePath)
	return args.Error(0)
}

func (m *MockEngine) SetOnChatMessage(handler func(chat.Message)) {
	// m.Called(handler) // Hard to match functions, usually ignored in generic mocks
}

func (m *MockEngine) SetTransferCallbacks(onProgress func(session transfer.Session), onRequest func(req transfer.Request, senderID string)) {
	// m.Called(onProgress, onRequest)
}

func (m *MockEngine) GetPeerByID(peerID string) (*api.PeerConfig, bool) {
	args := m.Called(peerID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*api.PeerConfig), args.Bool(1)
}

func (m *MockEngine) GenerateInvite(networkID string, maxUses int, expiresHours int) (*api.InviteTokenResponse, error) {
	args := m.Called(networkID, maxUses, expiresHours)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.InviteTokenResponse), args.Error(1)
}

func (m *MockEngine) KickPeer(networkID, peerID, reason string) error {
	args := m.Called(networkID, peerID, reason)
	return args.Error(0)
}

func (m *MockEngine) BanPeer(networkID, peerID, reason string) error {
	args := m.Called(networkID, peerID, reason)
	return args.Error(0)
}

func (m *MockEngine) UnbanPeer(networkID, peerID string) error {
	args := m.Called(networkID, peerID)
	return args.Error(0)
}

func (m *MockEngine) GetChatMessages(networkID string, limit int, beforeID string) []chat.Message {
	args := m.Called(networkID, limit, beforeID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]chat.Message)
}

func (m *MockEngine) RejectTransfer(transferID string) error {
	args := m.Called(transferID)
	return args.Error(0)
}

func (m *MockEngine) CancelTransfer(transferID string) error {
	args := m.Called(transferID)
	return args.Error(0)
}

func (m *MockEngine) GetTransfers() []transfer.Session {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]transfer.Session)
}

func (m *MockEngine) SubscribeTransfers() chan transfer.Session {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(chan transfer.Session)
}

func (m *MockEngine) UnsubscribeTransfers(ch chan transfer.Session) {
	m.Called(ch)
}

func (m *MockEngine) SubscribeChatMessages() chan chat.Message {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(chan chat.Message)
}

func (m *MockEngine) UnsubscribeChatMessages(ch chan chat.Message) {
	m.Called(ch)
}

func (m *MockEngine) CreateNetwork(name string) (*api.NetworkResponse, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.NetworkResponse), args.Error(1)
}

func (m *MockEngine) JoinNetwork(inviteCode string) (*api.NetworkResponse, error) {
	args := m.Called(inviteCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.NetworkResponse), args.Error(1)
}

func (m *MockEngine) LeaveNetwork(networkID string) error {
	args := m.Called(networkID)
	return args.Error(0)
}

func (m *MockEngine) GetNetworks() ([]api.NetworkResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]api.NetworkResponse), args.Error(1)
}

func (m *MockEngine) GetNetwork(networkID string) (*api.NetworkResponse, error) {
	args := m.Called(networkID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.NetworkResponse), args.Error(1)
}

func (m *MockEngine) DeleteNetwork(networkID string) error {
	args := m.Called(networkID)
	return args.Error(0)
}

// MockAPIClient is a mock implementation of DaemonAPIClient
type MockAPIClient struct {
	mock.Mock
}

func (m *MockAPIClient) Register(ctx context.Context, authToken string, req api.RegisterDeviceRequest) (*api.RegisterDeviceResponse, error) {
	args := m.Called(ctx, authToken, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.RegisterDeviceResponse), args.Error(1)
}

func (m *MockAPIClient) GetNetworks(ctx context.Context) ([]api.NetworkResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]api.NetworkResponse), args.Error(1)
}

func (m *MockAPIClient) CreateNetwork(ctx context.Context, name, cidr string) (*api.NetworkResponse, error) {
	args := m.Called(ctx, name, cidr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.NetworkResponse), args.Error(1)
}

func (m *MockAPIClient) JoinNetwork(ctx context.Context, inviteCode string) (*api.NetworkResponse, error) {
	args := m.Called(ctx, inviteCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.NetworkResponse), args.Error(1)
}
