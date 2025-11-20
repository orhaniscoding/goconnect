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
