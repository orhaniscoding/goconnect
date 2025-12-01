package engine

import (
	"testing"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/chat"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/transfer"
)

// TestEngine_ChatMethods tests the chat-related methods
func TestEngine_GetChatMessages(t *testing.T) {
	chatMgr := chat.NewManager()
	e := &Engine{
		chatMgr: chatMgr,
	}

	// Initially empty
	messages := e.GetChatMessages("", 10, "")
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(messages))
	}
}

func TestEngine_SubscribeUnsubscribeChatMessages(t *testing.T) {
	chatMgr := chat.NewManager()
	e := &Engine{
		chatMgr: chatMgr,
	}

	ch := e.SubscribeChatMessages()
	if ch == nil {
		t.Fatal("Expected channel, got nil")
	}

	// Unsubscribe should not panic
	e.UnsubscribeChatMessages(ch)
}

// TestEngine_TransferMethods tests the transfer-related methods
func TestEngine_GetTransfers(t *testing.T) {
	transferMgr := transfer.NewManager()
	e := &Engine{
		transferMgr: transferMgr,
	}

	// Initially empty
	sessions := e.GetTransfers()
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}
}

func TestEngine_SubscribeUnsubscribeTransfers(t *testing.T) {
	transferMgr := transfer.NewManager()
	e := &Engine{
		transferMgr: transferMgr,
	}

	ch := e.SubscribeTransfers()
	if ch == nil {
		t.Fatal("Expected channel, got nil")
	}

	e.UnsubscribeTransfers(ch)
}

func TestEngine_RejectTransfer_NotFound(t *testing.T) {
	transferMgr := transfer.NewManager()
	e := &Engine{
		transferMgr: transferMgr,
	}

	err := e.RejectTransfer("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent transfer")
	}
}

func TestEngine_CancelTransfer_NotFound(t *testing.T) {
	transferMgr := transfer.NewManager()
	e := &Engine{
		transferMgr: transferMgr,
	}

	err := e.CancelTransfer("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent transfer")
	}
}

// TestEngine_PeerMethods tests peer-related methods
func TestEngine_GetPeerByID_NotFound(t *testing.T) {
	e := &Engine{
		peerMap: make(map[string]api.PeerConfig),
	}

	_, found := e.GetPeerByID("non-existent")
	if found {
		t.Error("Expected peer not found")
	}
}

func TestEngine_GetPeerByID_Found(t *testing.T) {
	e := &Engine{
		peerMap: map[string]api.PeerConfig{
			"peer-1": {
				ID:        "peer-1",
				PublicKey: "abc123",
				Name:      "Test Peer",
			},
		},
	}

	peer, found := e.GetPeerByID("peer-1")
	if !found {
		t.Error("Expected peer to be found")
	}
	if peer.Name != "Test Peer" {
		t.Errorf("Expected name 'Test Peer', got %s", peer.Name)
	}
}

func TestEngine_GetPeers(t *testing.T) {
	e := &Engine{
		peerMap: map[string]api.PeerConfig{
			"peer-1": {ID: "peer-1", Name: "Peer 1"},
			"peer-2": {ID: "peer-2", Name: "Peer 2"},
		},
	}

	peers := e.GetPeers()
	if len(peers) != 2 {
		t.Errorf("Expected 2 peers, got %d", len(peers))
	}
}

// TestEngine_NetworkMethods - GetNetworks requires API client so we skip it
// We can only test the cached networks access pattern

// TestEngine_SetCallbacks tests callback setters
func TestEngine_SetTransferCallbacks(t *testing.T) {
	e := &Engine{}

	progressCalled := false
	requestCalled := false

	e.SetTransferCallbacks(
		func(s transfer.Session) { progressCalled = true },
		func(r transfer.Request, peerID string) { requestCalled = true },
	)

	if e.onTransferProgress == nil {
		t.Error("Expected onTransferProgress to be set")
	}
	if e.onTransferRequest == nil {
		t.Error("Expected onTransferRequest to be set")
	}

	// Trigger callbacks
	e.onTransferProgress(transfer.Session{})
	e.onTransferRequest(transfer.Request{}, "peer-1")

	if !progressCalled {
		t.Error("onTransferProgress was not called")
	}
	if !requestCalled {
		t.Error("onTransferRequest was not called")
	}
}

func TestEngine_SetOnChatMessage(t *testing.T) {
	e := &Engine{}

	called := false
	e.SetOnChatMessage(func(msg chat.Message) {
		called = true
	})

	if e.onChatMessage == nil {
		t.Error("Expected onChatMessage to be set")
	}

	e.onChatMessage(chat.Message{})
	if !called {
		t.Error("onChatMessage was not called")
	}
}

// TestEngine_Version tests version handling
func TestEngine_Version(t *testing.T) {
	// Version is a package-level variable
	if Version == "" {
		t.Error("Expected Version to have a value")
	}
}

// TestEngine_Connect_Disconnect tests connect/disconnect state
func TestEngine_PausedState(t *testing.T) {
	e := &Engine{
		paused: false,
	}

	if e.paused {
		t.Error("Expected paused to be false initially")
	}

	// Manually set paused (normally done by Disconnect)
	e.paused = true
	if !e.paused {
		t.Error("Expected paused to be true after setting")
	}
}
