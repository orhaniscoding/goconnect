package daemon

import (
	"context"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/chat"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/transfer"
)

// DaemonEngine defines the interface for the P2P engine used by the daemon.
type DaemonEngine interface {
	Start()
	Stop()
	Connect()
	Disconnect()
	GetStatus() map[string]interface{}
	ManualConnect(peerID string) error
	SendChatMessage(peerID, content string) error
	SendFileRequest(peerID, filePath string) (*transfer.Session, error)
	AcceptFile(requestID, savePath string) error
	SetOnChatMessage(handler func(chat.Message))
	SetTransferCallbacks(onProgress func(session transfer.Session), onRequest func(req transfer.Request, senderID string))
	GetPeerByID(peerID string) (*api.PeerConfig, bool)
	GenerateInvite(networkID string, maxUses int, expiresHours int) (*api.InviteTokenResponse, error)
	KickPeer(networkID, peerID, reason string) error
	BanPeer(networkID, peerID, reason string) error
	UnbanPeer(networkID, peerID string) error
	GetChatMessages(networkID string, limit int, beforeID string) []chat.Message
	RejectTransfer(transferID string) error
	CancelTransfer(transferID string) error
	GetTransfers() []transfer.Session
	SubscribeTransfers() chan transfer.Session
	UnsubscribeTransfers(ch chan transfer.Session)
	SubscribeChatMessages() chan chat.Message
	UnsubscribeChatMessages(ch chan chat.Message)

	// Network management through engine
	CreateNetwork(name string) (*api.NetworkResponse, error)
	JoinNetwork(inviteCode string) (*api.NetworkResponse, error)
	LeaveNetwork(networkID string) error
	GetNetworks() ([]api.NetworkResponse, error)
	GetNetwork(networkID string) (*api.NetworkResponse, error)
	DeleteNetwork(networkID string) error
}

// DaemonAPIClient defines the interface for the API client used by the daemon.
type DaemonAPIClient interface {
	Register(ctx context.Context, authToken string, req api.RegisterDeviceRequest) (*api.RegisterDeviceResponse, error)
	GetNetworks(ctx context.Context) ([]api.NetworkResponse, error)
	CreateNetwork(ctx context.Context, name string) (*api.NetworkResponse, error)
	JoinNetwork(ctx context.Context, inviteCode string) (*api.NetworkResponse, error)
}
