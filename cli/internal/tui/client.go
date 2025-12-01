package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const DaemonURL = "http://localhost:12345"

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    DaemonURL,
	}
}

// CheckDaemonStatus verifies if the daemon is reachable
func (c *Client) CheckDaemonStatus() bool {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Network represents a network in the list
type Network struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"` // "host" or "client"
}

// Status represents the current connection status
type Status struct {
	Connected     bool      `json:"connected"`
	NetworkName   string    `json:"network_name"`
	InviteCode    string    `json:"invite_code"`
	IP            string    `json:"ip"`
	OnlineMembers int       `json:"online_members"`
	Role          string    `json:"role"`
	Networks      []Network `json:"networks"`
}

// GetStatus fetches the current status from the daemon
// Note: This assumes the daemon exposes a /status endpoint returning this structure.
// If not, we might need to aggregate data from multiple endpoints.
func (c *Client) GetStatus() (*Status, error) {
	// For now, let's mock this or implement a real endpoint in the daemon later.
	// Assuming /status exists.
	resp, err := c.httpClient.Get(c.baseURL + "/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("daemon returned status %d", resp.StatusCode)
	}

	var status Status
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}
	return &status, nil
}

// CreateNetwork creates a new network
func (c *Client) CreateNetwork(name string) (*Network, error) {
	req := map[string]string{"name": name}
	body, _ := json.Marshal(req)

	resp, err := c.httpClient.Post(c.baseURL+"/networks/create", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create network: status %d", resp.StatusCode)
	}

	var network Network
	if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
		return nil, err
	}

	return &network, nil
}

// JoinNetwork joins an existing network
func (c *Client) JoinNetwork(inviteCode string) (*Network, error) {
	req := map[string]string{"invite_code": inviteCode}
	body, _ := json.Marshal(req)

	resp, err := c.httpClient.Post(c.baseURL+"/networks/join", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to join network: status %d", resp.StatusCode)
	}

	var network Network
	if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
		return nil, err
	}

	return &network, nil
}

// GetNetworks fetches the list of networks
func (c *Client) GetNetworks() ([]Network, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/networks")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get networks: status %d", resp.StatusCode)
	}

	var networks []Network
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return nil, err
	}

	return networks, nil
}
