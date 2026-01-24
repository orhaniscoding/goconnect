package rpc

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/orhaniscoding/goconnect/server/internal/auth"
	"github.com/orhaniscoding/goconnect/server/internal/logger"
	"github.com/orhaniscoding/goconnect/server/internal/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// BackendClient defines the interface for backend HTTP operations.
type BackendClient interface {
	GetBaseURL() string
	GetHTTPClient() *http.Client
}

// NetworkServiceHandler implements the NetworkServiceServer gRPC interface.
type NetworkServiceHandler struct {
	proto.UnimplementedNetworkServiceServer
	backend      BackendClient
	tokenManager auth.TokenManager
}

// NewNetworkServiceHandler creates a new handler for the Network service.
func NewNetworkServiceHandler(
	backend BackendClient,
	tokenManager auth.TokenManager,
) *NetworkServiceHandler {
	return &NetworkServiceHandler{
		backend:      backend,
		tokenManager: tokenManager,
	}
}

// UpdateNetwork modifies network properties (owner only).
func (h *NetworkServiceHandler) UpdateNetwork(ctx context.Context, req *proto.UpdateNetworkRequest) (*proto.Network, error) {
	logger.Info("UpdateNetwork gRPC called", "network_id", req.NetworkId)

	// 1. Validate request
	if req.NetworkId == "" {
		return nil, fmt.Errorf("network_id is required")
	}

	// 2. Get authenticated user token
	session, err := h.tokenManager.LoadSession()
	if err != nil {
		logger.Warn("UpdateNetwork: No valid session", "error", err)
		return nil, fmt.Errorf("authentication required")
	}

	// 3. Build patch map from proto request
	patch := make(map[string]any)
	if req.Name != "" {
		patch["name"] = req.Name
	}
	// Note: Description not supported in domain model yet

	// If no fields to update, return error
	if len(patch) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// 4. Call Backend HTTP API: PATCH /v1/networks/:id
	network, err := h.callBackendUpdateNetwork(ctx, req.NetworkId, session.AccessToken, patch)
	if err != nil {
		logger.Error("UpdateNetwork: Backend API call failed",
			"network_id", req.NetworkId,
			"error", err,
		)
		return nil, fmt.Errorf("failed to update network: %w", err)
	}

	logger.Info("UpdateNetwork successful",
		"network_id", req.NetworkId,
		"updated_fields", len(patch),
	)

	return network, nil
}

// callBackendUpdateNetwork calls the backend HTTP API to update a network
func (h *NetworkServiceHandler) callBackendUpdateNetwork(ctx context.Context, networkID, accessToken string, patch map[string]any) (*proto.Network, error) {
	url := fmt.Sprintf("%s/v1/networks/%s", h.backend.GetBaseURL(), networkID)

	// Marshal patch to JSON
	body, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patch: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Idempotency-Key", generateIdempotencyKey(networkID, patch))

	// Execute request
	resp, err := h.backend.GetHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("backend returned %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var apiResp struct {
		Data struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Visibility  string `json:"visibility"`
			CreatedBy   string `json:"created_by"`
			PeerCount   int32  `json:"peer_count"`
			OnlineCount int32  `json:"online_count"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to proto.Network
	// Parse JWT to determine user's role
	userID, _ := extractUserIDFromJWT(accessToken)
	myRole := proto.NetworkRole_NETWORK_ROLE_MEMBER
	if apiResp.Data.CreatedBy == userID {
		myRole = proto.NetworkRole_NETWORK_ROLE_OWNER
	}

	return &proto.Network{
		Id:          apiResp.Data.ID,
		Name:        apiResp.Data.Name,
		Description: "",
		MyRole:      myRole,
		PeerCount:   apiResp.Data.PeerCount,
		OnlineCount: apiResp.Data.OnlineCount,
		IsConnected: false,
	}, nil
}

// generateIdempotencyKey creates an idempotency key for the request
func generateIdempotencyKey(networkID string, patch map[string]any) string {
	data, _ := json.Marshal(patch)
	return fmt.Sprintf("update-network-%s-%x", networkID, len(data))
}

// Stub implementations for other NetworkService RPCs (to be implemented later)

func (h *NetworkServiceHandler) CreateNetwork(ctx context.Context, req *proto.CreateNetworkRequest) (*proto.CreateNetworkResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *NetworkServiceHandler) JoinNetwork(ctx context.Context, req *proto.JoinNetworkRequest) (*proto.JoinNetworkResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *NetworkServiceHandler) LeaveNetwork(ctx context.Context, req *proto.LeaveNetworkRequest) (*proto.LeaveNetworkResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *NetworkServiceHandler) ListNetworks(ctx context.Context, req *emptypb.Empty) (*proto.ListNetworksResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *NetworkServiceHandler) GetNetwork(ctx context.Context, req *proto.GetNetworkRequest) (*proto.Network, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *NetworkServiceHandler) DeleteNetwork(ctx context.Context, req *proto.DeleteNetworkRequest) (*emptypb.Empty, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *NetworkServiceHandler) GenerateInvite(ctx context.Context, req *proto.GenerateInviteRequest) (*proto.GenerateInviteResponse, error) {
	return nil, fmt.Errorf("not implemented")
}


// extractUserIDFromJWT parses JWT token and extracts user ID from claims
func extractUserIDFromJWT(tokenString string) (string, error) {
	// Parse JWT without verification (daemon trusts backend token)
	// In production, this should verify the signature
	parts := splitJWT(tokenString)
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT format")
	}

	// Decode claims (second part)
	claims, err := decodeJWTClaims(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode claims: %w", err)
	}

	// Extract sub (subject) claim as user ID
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		return sub, nil
	}

	// Fallback: try user_id claim
	if userID, ok := claims["user_id"].(string); ok && userID != "" {
		return userID, nil
	}

	return "", fmt.Errorf("no user ID found in token claims")
}

// Helper functions for JWT parsing (minimal, zero-dependency)
func splitJWT(token string) []string {
	parts := make([]string, 0, 3)
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	parts = append(parts, token[start:])
	return parts
}

func decodeJWTClaims(encodedClaims string) (map[string]interface{}, error) {
	// Add padding if needed
	switch len(encodedClaims) % 4 {
	case 2:
		encodedClaims += "=="
	case 3:
		encodedClaims += "="
	}

	// Base64 decode (encoding/base64 already imported at top)
	decoded, err := base64.RawURLEncoding.DecodeString(encodedClaims)
	if err != nil {
		return nil, err
	}

	// JSON unmarshal (encoding/json already imported at top)
	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, err
	}

	return claims, nil
}
