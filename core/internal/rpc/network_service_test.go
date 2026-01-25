package rpc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/auth"
	"github.com/orhaniscoding/goconnect/server/internal/proto"
)

// Mock implementations

type mockBackendClient struct {
	baseURL    string
	httpClient *http.Client
	handler    http.Handler
}

func (m *mockBackendClient) GetBaseURL() string {
	return m.baseURL
}

func (m *mockBackendClient) GetHTTPClient() *http.Client {
	return m.httpClient
}

type mockTokenManager struct {
	session *auth.TokenSession
	err     error
}

func (m *mockTokenManager) SaveSession(session *auth.TokenSession) error {
	m.session = session
	return nil
}

func (m *mockTokenManager) LoadSession() (*auth.TokenSession, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.session, nil
}

func (m *mockTokenManager) ClearSession() error {
	m.session = nil
	return nil
}

// Test helpers

func createMockJWT(userID string) string {
	// Create a simple mock JWT with sub claim
	// Format: header.payload.signature (we only care about payload for testing)
	header := `{"alg":"HS256","typ":"JWT"}`
	payload := `{"sub":"` + userID + `","exp":9999999999}`

	// Base64URL encode (simplified for testing)
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(header))
	payloadB64 := base64.RawURLEncoding.EncodeToString([]byte(payload))

	return headerB64 + "." + payloadB64 + ".signature"
}

// Tests

func TestUpdateNetwork_Success(t *testing.T) {
	// Setup mock backend server
	networkID := "net123"
	userID := "user456"
	newName := "Updated Network Name"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}

		expectedPath := "/v1/networks/" + networkID
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Missing Content-Type header")
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("Missing Authorization header")
		}

		// Parse request body
		var patch map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if patch["name"] != newName {
			t.Errorf("Expected name %s, got %v", newName, patch["name"])
		}

		// Send success response
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"id":           networkID,
				"name":         newName,
				"visibility":   "private",
				"created_by":   userID,
				"peer_count":   5,
				"online_count": 3,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// Setup mocks
	mockBackend := &mockBackendClient{
		baseURL:    mockServer.URL,
		httpClient: mockServer.Client(),
	}

	mockToken := &mockTokenManager{
		session: &auth.TokenSession{
			AccessToken: createMockJWT(userID),
		},
	}

	// Create handler
	handler := NewNetworkServiceHandler(mockBackend, mockToken)

	// Execute test
	req := &proto.UpdateNetworkRequest{
		NetworkId: networkID,
		Name:      newName,
	}

	resp, err := handler.UpdateNetwork(context.Background(), req)

	// Assertions
	if err != nil {
		t.Fatalf("UpdateNetwork failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Response is nil")
	}

	if resp.Id != networkID {
		t.Errorf("Expected network ID %s, got %s", networkID, resp.Id)
	}

	if resp.Name != newName {
		t.Errorf("Expected name %s, got %s", newName, resp.Name)
	}

	if resp.MyRole != proto.NetworkRole_NETWORK_ROLE_OWNER {
		t.Errorf("Expected role OWNER, got %v", resp.MyRole)
	}

	if resp.PeerCount != 5 {
		t.Errorf("Expected peer count 5, got %d", resp.PeerCount)
	}

	if resp.OnlineCount != 3 {
		t.Errorf("Expected online count 3, got %d", resp.OnlineCount)
	}
}

func TestUpdateNetwork_MissingNetworkID(t *testing.T) {
	mockBackend := &mockBackendClient{}
	mockToken := &mockTokenManager{
		session: &auth.TokenSession{
			AccessToken: createMockJWT("user123"),
		},
	}

	handler := NewNetworkServiceHandler(mockBackend, mockToken)

	req := &proto.UpdateNetworkRequest{
		NetworkId: "", // Empty network ID
		Name:      "New Name",
	}

	_, err := handler.UpdateNetwork(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for missing network ID, got nil")
	}

	expectedError := "network_id is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestUpdateNetwork_NoAuthentication(t *testing.T) {
	mockBackend := &mockBackendClient{}
	mockToken := &mockTokenManager{
		err: fmt.Errorf("no session found"),
	}

	handler := NewNetworkServiceHandler(mockBackend, mockToken)

	req := &proto.UpdateNetworkRequest{
		NetworkId: "net123",
		Name:      "New Name",
	}

	_, err := handler.UpdateNetwork(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for missing authentication, got nil")
	}

	expectedError := "authentication required"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestUpdateNetwork_NoFieldsToUpdate(t *testing.T) {
	mockBackend := &mockBackendClient{}
	mockToken := &mockTokenManager{
		session: &auth.TokenSession{
			AccessToken: createMockJWT("user123"),
		},
	}

	handler := NewNetworkServiceHandler(mockBackend, mockToken)

	req := &proto.UpdateNetworkRequest{
		NetworkId: "net123",
		Name:      "", // Empty name = no fields to update
	}

	_, err := handler.UpdateNetwork(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for no fields to update, got nil")
	}

	expectedError := "no fields to update"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestUpdateNetwork_BackendHTTPError(t *testing.T) {
	// Setup mock backend server that returns error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "permission denied"}`))
	}))
	defer mockServer.Close()

	mockBackend := &mockBackendClient{
		baseURL:    mockServer.URL,
		httpClient: mockServer.Client(),
	}

	mockToken := &mockTokenManager{
		session: &auth.TokenSession{
			AccessToken: createMockJWT("user123"),
		},
	}

	handler := NewNetworkServiceHandler(mockBackend, mockToken)

	req := &proto.UpdateNetworkRequest{
		NetworkId: "net123",
		Name:      "New Name",
	}

	_, err := handler.UpdateNetwork(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for backend HTTP error, got nil")
	}

	// Should contain status code in error message
	if !contains(err.Error(), "403") {
		t.Errorf("Expected error to contain status code 403, got: %v", err)
	}
}

func TestUpdateNetwork_BackendInvalidJSON(t *testing.T) {
	// Setup mock backend server that returns invalid JSON
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json {`))
	}))
	defer mockServer.Close()

	mockBackend := &mockBackendClient{
		baseURL:    mockServer.URL,
		httpClient: mockServer.Client(),
	}

	mockToken := &mockTokenManager{
		session: &auth.TokenSession{
			AccessToken: createMockJWT("user123"),
		},
	}

	handler := NewNetworkServiceHandler(mockBackend, mockToken)

	req := &proto.UpdateNetworkRequest{
		NetworkId: "net123",
		Name:      "New Name",
	}

	_, err := handler.UpdateNetwork(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for invalid JSON response, got nil")
	}

	// Should contain parse error
	if !contains(err.Error(), "parse") && !contains(err.Error(), "unmarshal") {
		t.Errorf("Expected error to mention JSON parsing, got: %v", err)
	}
}

func TestUpdateNetwork_MemberRole(t *testing.T) {
	// Test that non-owner gets MEMBER role
	networkID := "net123"
	ownerID := "user_owner"
	memberID := "user_member"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"id":           networkID,
				"name":         "Network",
				"created_by":   ownerID, // Different from requester
				"peer_count":   2,
				"online_count": 1,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	mockBackend := &mockBackendClient{
		baseURL:    mockServer.URL,
		httpClient: mockServer.Client(),
	}

	mockToken := &mockTokenManager{
		session: &auth.TokenSession{
			AccessToken: createMockJWT(memberID), // Member, not owner
		},
	}

	handler := NewNetworkServiceHandler(mockBackend, mockToken)

	req := &proto.UpdateNetworkRequest{
		NetworkId: networkID,
		Name:      "Updated Name",
	}

	resp, err := handler.UpdateNetwork(context.Background(), req)

	if err != nil {
		t.Fatalf("UpdateNetwork failed: %v", err)
	}

	if resp.MyRole != proto.NetworkRole_NETWORK_ROLE_MEMBER {
		t.Errorf("Expected role MEMBER for non-owner, got %v", resp.MyRole)
	}
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
