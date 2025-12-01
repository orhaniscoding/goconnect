package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/storage"
)

// Client handles communication with the GoConnect server
type Client struct {
	config     *config.Config
	keyring    *storage.KeyringStore
	httpClient *http.Client

	// WebSocket fields
	wsConn          *websocket.Conn
	wsLock          sync.Mutex
	wsWriteLock     sync.Mutex
	signalCallbacks SignalCallbacks
	stopChan        chan struct{}
}

type SignalCallbacks struct {
	onOffer     func(sourceID, ufrag, pwd string)
	onAnswer    func(sourceID, ufrag, pwd string)
	onCandidate func(sourceID, candidate string)
}

// NewClient creates a new API client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config:  cfg,
		keyring: cfg.Keyring,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopChan: make(chan struct{}),
	}
}

// getAuthToken retrieves the authentication token from the keyring.
func (c *Client) getAuthToken() (string, error) {
	if c.keyring == nil {
		return "", fmt.Errorf("keyring not initialized")
	}
	token, err := c.keyring.RetrieveAuthToken()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve auth token: %w", err)
	}
	return token, nil
}

// RegisterDeviceRequest matches the server's request struct
type RegisterDeviceRequest struct {
	Name      string `json:"name"`
	Platform  string `json:"platform"`
	PubKey    string `json:"pubkey"`
	HostName  string `json:"hostname"`
	OSVersion string `json:"os_version"`
	DaemonVer string `json:"daemon_ver"`
}

// RegisterDeviceResponse matches the server's response struct
type RegisterDeviceResponse struct {
	ID string `json:"id"`
	// Other fields ignored for now
}

// Register registers the device with the server
func (c *Client) Register(ctx context.Context, authToken string, req RegisterDeviceRequest) (*RegisterDeviceResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/devices", c.config.Server.URL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+authToken) // authToken is passed here explicitly

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		// Attempt to read error message from response body
		var errorBody struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorBody); err == nil && errorBody.Message != "" {
			return nil, fmt.Errorf("registration failed: %s (status: %d)", errorBody.Message, resp.StatusCode)
		}
		return nil, fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	var result RegisterDeviceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// HeartbeatRequest matches the server's request struct
type HeartbeatRequest struct {
	IPAddress string `json:"ip_address,omitempty"`
	DaemonVer string `json:"daemon_ver,omitempty"`
	OSVersion string `json:"os_version,omitempty"`
}

// SendHeartbeat sends a heartbeat to the server
func (c *Client) SendHeartbeat(ctx context.Context, deviceID string, req HeartbeatRequest) error {
	authToken, err := c.getAuthToken()
	if err != nil {
		return fmt.Errorf("heartbeat failed: %w", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/devices/%s/heartbeat", c.config.Server.URL, deviceID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Attempt to read error message from response body
		var errorBody struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorBody); err == nil && errorBody.Message != "" {
			return fmt.Errorf("heartbeat failed: %s (status: %d)", errorBody.Message, resp.StatusCode)
		}
		return fmt.Errorf("heartbeat failed with status: %d", resp.StatusCode)
	}

	return nil
}

// DeviceConfig represents the WireGuard configuration
type DeviceConfig struct {
	Interface InterfaceConfig `json:"interface"`
	Peers     []PeerConfig    `json:"peers"`
}

type InterfaceConfig struct {
	ListenPort int      `json:"listen_port"`
	Addresses  []string `json:"addresses"`
	DNS        []string `json:"dns"`
	MTU        int      `json:"mtu"`
}

type PeerConfig struct {
	ID                  string   `json:"id"` // Device ID of the peer (for signaling)
	PublicKey           string   `json:"public_key"`
	Endpoint            string   `json:"endpoint"`
	AllowedIPs          []string `json:"allowed_ips"`
	PresharedKey        string   `json:"preshared_key"`
	PersistentKeepalive int      `json:"persistent_keepalive"`
	Name                string   `json:"name"`     // Friendly name
	Hostname            string   `json:"hostname"` // DNS hostname
}

// GetConfig retrieves the device configuration
func (c *Client) GetConfig(ctx context.Context, deviceID string) (*DeviceConfig, error) {
	authToken, err := c.getAuthToken()
	if err != nil {
		return nil, fmt.Errorf("get config failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/devices/%s/config", c.config.Server.URL, deviceID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Attempt to read error message from response body
		var errorBody struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorBody); err == nil && errorBody.Message != "" {
			return nil, fmt.Errorf("failed to get config: %s (status: %d)", errorBody.Message, resp.StatusCode)
		}
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var config DeviceConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &config, nil
}

// --- WebSocket Implementation ---

// StartWebSocket connects to the WebSocket server and starts the read loop
func (c *Client) StartWebSocket(ctx context.Context) error {
	authToken, err := c.getAuthToken()
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}

	u, err := url.Parse(c.config.Server.URL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	scheme := "ws"
	if u.Scheme == "https" {
		scheme = "wss"
	}
	u.Scheme = scheme
	u.Path = "/v1/ws"

	header := http.Header{}
	header.Set("Authorization", "Bearer "+authToken)

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return fmt.Errorf("websocket dial failed: %w", err)
	}

	c.wsLock.Lock()
	c.wsConn = conn
	c.wsLock.Unlock()

	go c.readLoop()

	return nil
}

// CloseWebSocket closes the WebSocket connection
func (c *Client) CloseWebSocket() {
	c.wsLock.Lock()
	defer c.wsLock.Unlock()
	if c.wsConn != nil {
		c.wsConn.Close()
		c.wsConn = nil
	}
}

func (c *Client) readLoop() {
	defer c.CloseWebSocket()

	for {
		select {
		case <-c.stopChan:
			return
		default:
			if c.wsConn == nil {
				return
			}
			_, message, err := c.wsConn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}
			c.handleMessage(message)
		}
	}
}

type wsMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type callSignalData struct {
	TargetID string          `json:"targetId,omitempty"`
	FromUser string          `json:"fromUser,omitempty"`
	Signal   json.RawMessage `json:"signal"`
}

type signalPayload struct {
	Ufrag     string `json:"ufrag,omitempty"`
	Pwd       string `json:"pwd,omitempty"`
	Candidate string `json:"candidate,omitempty"`
}

func (c *Client) handleMessage(data []byte) {
	var msg wsMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to unmarshal websocket message: %v", err)
		return
	}

	switch msg.Type {
	case "call.offer", "call.answer", "call.ice":
		var signalData callSignalData
		if err := json.Unmarshal(msg.Data, &signalData); err != nil {
			log.Printf("Failed to unmarshal signal data: %v", err)
			return
		}

		var payload signalPayload
		if err := json.Unmarshal(signalData.Signal, &payload); err != nil {
			log.Printf("Failed to unmarshal signal payload: %v", err)
			return
		}

		switch msg.Type {
		case "call.offer":
			if c.signalCallbacks.onOffer != nil {
				c.signalCallbacks.onOffer(signalData.FromUser, payload.Ufrag, payload.Pwd)
			}
		case "call.answer":
			if c.signalCallbacks.onAnswer != nil {
				c.signalCallbacks.onAnswer(signalData.FromUser, payload.Ufrag, payload.Pwd)
			}
		case "call.ice":
			if c.signalCallbacks.onCandidate != nil {
				c.signalCallbacks.onCandidate(signalData.FromUser, payload.Candidate)
			}
		}
	}
}

func (c *Client) sendSignal(msgType string, targetID string, payload signalPayload) error {
	c.wsLock.Lock()
	if c.wsConn == nil {
		c.wsLock.Unlock()
		return fmt.Errorf("websocket not connected")
	}
	c.wsLock.Unlock()

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	signalData := callSignalData{
		TargetID: targetID,
		Signal:   payloadBytes,
	}
	signalDataBytes, err := json.Marshal(signalData)
	if err != nil {
		return err
	}

	msg := wsMessage{
		Type: msgType,
		Data: signalDataBytes,
	}

	c.wsWriteLock.Lock()
	defer c.wsWriteLock.Unlock()
	return c.wsConn.WriteJSON(msg)
}

// SendOffer sends a WebRTC offer to a target peer
func (c *Client) SendOffer(targetID string, ufrag, pwd string) error {
	return c.sendSignal("call.offer", targetID, signalPayload{
		Ufrag: ufrag,
		Pwd:   pwd,
	})
}

// SendAnswer sends a WebRTC answer to a target peer
func (c *Client) SendAnswer(targetID string, ufrag, pwd string) error {
	return c.sendSignal("call.answer", targetID, signalPayload{
		Ufrag: ufrag,
		Pwd:   pwd,
	})
}

// SendCandidate sends a WebRTC ICE candidate to a target peer
func (c *Client) SendCandidate(targetID string, candidate string) error {
	return c.sendSignal("call.ice", targetID, signalPayload{
		Candidate: candidate,
	})
}

// OnOffer sets the callback for incoming offers
func (c *Client) OnOffer(f func(sourceID, ufrag, pwd string)) {
	c.signalCallbacks.onOffer = f
}

// OnAnswer sets the callback for incoming answers
func (c *Client) OnAnswer(f func(sourceID, ufrag, pwd string)) {
	c.signalCallbacks.onAnswer = f
}

// OnCandidate sets the callback for incoming candidates
func (c *Client) OnCandidate(f func(sourceID, candidate string)) {
	c.signalCallbacks.onCandidate = f
}

// --- Network Management ---

type CreateNetworkRequest struct {
	Name string `json:"name"`
}

type NetworkResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"invite_code,omitempty"`
	Role       string `json:"role"`
}

func (c *Client) CreateNetwork(ctx context.Context, name string) (*NetworkResponse, error) {
	authToken, err := c.getAuthToken()
	if err != nil {
		return nil, fmt.Errorf("create network failed: %w", err)
	}

	req := CreateNetworkRequest{Name: name}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/networks", c.config.Server.URL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errorBody struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorBody); err == nil && errorBody.Message != "" {
			return nil, fmt.Errorf("create network failed: %s", errorBody.Message)
		}
		return nil, fmt.Errorf("create network failed with status: %d", resp.StatusCode)
	}

	var result NetworkResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

type JoinNetworkRequest struct {
	InviteCode string `json:"invite_code"`
}

func (c *Client) JoinNetwork(ctx context.Context, inviteCode string) (*NetworkResponse, error) {
	authToken, err := c.getAuthToken()
	if err != nil {
		return nil, fmt.Errorf("join network failed: %w", err)
	}

	req := JoinNetworkRequest{InviteCode: inviteCode}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/networks/join", c.config.Server.URL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorBody); err == nil && errorBody.Message != "" {
			return nil, fmt.Errorf("join network failed: %s", errorBody.Message)
		}
		return nil, fmt.Errorf("join network failed with status: %d", resp.StatusCode)
	}

	var result NetworkResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetNetworks(ctx context.Context) ([]NetworkResponse, error) {
	authToken, err := c.getAuthToken()
	if err != nil {
		return nil, fmt.Errorf("get networks failed: %w", err)
	}

	url := fmt.Sprintf("%s/v1/networks", c.config.Server.URL)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get networks failed with status: %d", resp.StatusCode)
	}

	var result []NetworkResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// LeaveNetwork removes the current device from a network
func (c *Client) LeaveNetwork(ctx context.Context, networkID string) error {
	authToken, err := c.getAuthToken()
	if err != nil {
		return fmt.Errorf("leave network failed: %w", err)
	}

	url := fmt.Sprintf("%s/v1/networks/%s/leave", c.config.Server.URL, networkID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("leave network failed with status: %d", resp.StatusCode)
	}

	return nil
}

// --- Invite Management ---

// CreateInviteRequest maps to server domain.CreateInviteRequest
type CreateInviteRequest struct {
	ExpiresIn int `json:"expires_in,omitempty"`
	UsesMax   int `json:"uses_max,omitempty"`
}

// InviteTokenResponse maps to server domain.InviteTokenResponse
type InviteTokenResponse struct {
	ID        string    `json:"id"`
	NetworkID string    `json:"network_id"`
	Token     string    `json:"token"`
	InviteURL string    `json:"invite_url"`
	ExpiresAt time.Time `json:"expires_at"`
	UsesMax   int       `json:"uses_max"`
	UsesLeft  int       `json:"uses_left"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

// GenerateInvite creates an invite token for a network via the core API.
// expiresHours: number of hours from now until expiry (0 => server default)
// maxUses: maximum uses (0 => unlimited)
func (c *Client) GenerateInvite(ctx context.Context, networkID string, maxUses int, expiresHours int) (*InviteTokenResponse, error) {
	if networkID == "" {
		return nil, fmt.Errorf("networkID is required")
	}

	authToken, err := c.getAuthToken()
	if err != nil {
		return nil, fmt.Errorf("generate invite failed: %w", err)
	}

	var expiresIn int
	if expiresHours > 0 {
		expiresIn = expiresHours * 3600
	}

	req := CreateInviteRequest{
		ExpiresIn: expiresIn,
		UsesMax:   maxUses,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/networks/%s/invites", c.config.Server.URL, networkID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+authToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errorBody struct{ Message string `json:"message"` }
		if err := json.NewDecoder(resp.Body).Decode(&errorBody); err == nil && errorBody.Message != "" {
			return nil, fmt.Errorf("generate invite failed: %s", errorBody.Message)
		}
		return nil, fmt.Errorf("generate invite failed with status: %d", resp.StatusCode)
	}

	var result InviteTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// KickPeer removes a peer from a network (admin/owner only)
func (c *Client) KickPeer(ctx context.Context, networkID, peerID, reason string) error {
	authToken, err := c.getAuthToken()
	if err != nil {
		return fmt.Errorf("kick peer failed: %w", err)
	}

	body, _ := json.Marshal(map[string]string{"reason": reason})
	url := fmt.Sprintf("%s/v1/networks/%s/members/%s", c.config.Server.URL, networkID, peerID)
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+authToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("kick peer failed with status: %d", resp.StatusCode)
	}
	return nil
}

// BanPeer bans a peer from a network (admin/owner only)
func (c *Client) BanPeer(ctx context.Context, networkID, peerID, reason string) error {
	authToken, err := c.getAuthToken()
	if err != nil {
		return fmt.Errorf("ban peer failed: %w", err)
	}

	body, _ := json.Marshal(map[string]string{"reason": reason})
	url := fmt.Sprintf("%s/v1/networks/%s/members/%s/ban", c.config.Server.URL, networkID, peerID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+authToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("ban peer failed with status: %d", resp.StatusCode)
	}
	return nil
}

// UnbanPeer unbans a peer from a network (admin/owner only)
func (c *Client) UnbanPeer(ctx context.Context, networkID, peerID string) error {
	authToken, err := c.getAuthToken()
	if err != nil {
		return fmt.Errorf("unban peer failed: %w", err)
	}

	url := fmt.Sprintf("%s/v1/networks/%s/members/%s/ban", c.config.Server.URL, networkID, peerID)
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unban peer failed with status: %d", resp.StatusCode)
	}
	return nil
}
