package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client handles communication with the GoConnect backend API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new backend client.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RequestDeviceCode initiates the OIDC Device Flow.
func (c *Client) RequestDeviceCode(ctx context.Context) (*DeviceCodeResponse, error) {
	reqBody := map[string]string{"client_id": "daemon"}
	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/v1/auth/device/code", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error status: %d", resp.StatusCode)
	}

	var res DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &res, nil
}

// PollDeviceToken checks if the user has authorized the device.
// Returns (response, retryable_error, fatal_error).
// If retryable_error is set (e.g., authorization_pending), the caller should wait and retry.
func (c *Client) PollDeviceToken(ctx context.Context, deviceCode string) (*AuthResponse, error) {
	reqBody := map[string]string{"device_code": deviceCode}
	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/v1/auth/device/token", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)

		if errResp.Message == "authorization_pending" {
			return nil, fmt.Errorf("authorization_pending")
		}
		if errResp.Message == "slow_down" {
			return nil, fmt.Errorf("slow_down")
		}
		
		return nil, fmt.Errorf("server error: %s (status: %d)", errResp.Message, resp.StatusCode)
	}

	var res AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &res, nil
}

// RefreshToken exchanges a refresh token for a new access token.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	reqBody := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}
	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/v1/auth/refresh", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s (status: %d)", errResp.Message, resp.StatusCode)
	}

	var res AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &res, nil
}
