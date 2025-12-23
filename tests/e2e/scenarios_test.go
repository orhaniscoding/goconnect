package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserFlow(t *testing.T) {
	if apiURL == "" {
		t.Skip("API_URL not set")
	}

	// Unique email for every run
	email := fmt.Sprintf("e2e-%d@example.com", time.Now().UnixNano())
	password := "TestPass123!"

	t.Run("Register", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    email,
			"password": password,
			"full_name": "E2E Test User",
		}
		resp := makeRequest(t, "POST", "/v1/auth/register", reqBody, "")
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	var token string
	t.Run("Login", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    email,
			"password": password,
		}
		resp := makeRequest(t, "POST", "/v1/auth/login", reqBody, "")
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]interface{}
		decodeBody(t, resp, &body)
		
		data, ok := body["data"].(map[string]interface{})
		require.True(t, ok, "response data missing")
		
		tok, ok := data["access_token"].(string)
		require.True(t, ok, "access_token missing")
		token = tok
	})

	var networkID string
	t.Run("CreateNetwork", func(t *testing.T) {
		reqBody := map[string]string{
			"name": "E2E Test Network",
		}
		resp := makeRequest(t, "POST", "/v1/networks", reqBody, token)
		// Check idempotency requirements if any (handler says Idempotency-Key required for POST /v1/networks)
		// Wait, handler code says: idempotencyKey := c.GetHeader("Idempotency-Key") ... if empty return error.
		// So I must provide it.
		// Wait, I missed looking at makeRequest... I need to implement it to allow headers.
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var body map[string]interface{}
		decodeBody(t, resp, &body)
		data := body["data"].(map[string]interface{})
		networkID = data["id"].(string)
		t.Logf("Created Network ID: %s", networkID)
	})

	t.Run("ListNetworks", func(t *testing.T) {
		resp := makeRequest(t, "GET", "/v1/networks", nil, token)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// Helper functions

func makeRequest(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, apiURL+path, bodyReader)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	// Add Idempotency-Key for POST/PATCH/DELETE
	if method == "POST" || method == "PATCH" || method == "DELETE" {
		req.Header.Set("Idempotency-Key", fmt.Sprintf("idem-%d", time.Now().UnixNano()))
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

func decodeBody(t *testing.T, resp *http.Response, target interface{}) {
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(target)
	require.NoError(t, err)
}
