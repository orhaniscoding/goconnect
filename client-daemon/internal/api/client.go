package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/storage"
)

// Client handles communication with the GoConnect server
type Client struct {
	config     *config.Config
	keyring    *storage.KeyringStore
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config:  cfg,
		keyring: cfg.Keyring,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
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