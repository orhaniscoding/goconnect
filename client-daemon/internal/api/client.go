package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client handles communication with the GoConnect server
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
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
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/devices", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	var result RegisterDeviceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
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
func (c *Client) SendHeartbeat(ctx context.Context, deviceID, authToken string, req HeartbeatRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1/devices/%s/heartbeat", c.baseURL, deviceID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
func (c *Client) GetConfig(ctx context.Context, deviceID, authToken string) (*DeviceConfig, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/devices/%s/config", c.baseURL, deviceID), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var config DeviceConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
