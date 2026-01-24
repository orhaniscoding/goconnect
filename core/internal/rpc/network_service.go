package rpc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/orhaniscoding/goconnect/server/internal/auth"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/logger"
	"github.com/orhaniscoding/goconnect/server/internal/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// NetworkServiceInterface defines the interface for network operations.
// This prevents import cycle with service package.
type NetworkServiceInterface interface {
	UpdateNetwork(ctx context.Context, id, actor, tenantID string, patch map[string]any) (*domain.Network, error)
}

// NetworkServiceHandler implements the NetworkServiceServer gRPC interface.
type NetworkServiceHandler struct {
	proto.UnimplementedNetworkServiceServer
	networkService  NetworkServiceInterface
	tokenManager    auth.TokenManager
	defaultTenantID string
}

// NewNetworkServiceHandler creates a new handler for the Network service.
func NewNetworkServiceHandler(
	networkService NetworkServiceInterface,
	tokenManager auth.TokenManager,
	defaultTenantID string,
) *NetworkServiceHandler {
	return &NetworkServiceHandler{
		networkService:  networkService,
		tokenManager:    tokenManager,
		defaultTenantID: defaultTenantID,
	}
}

// UpdateNetwork modifies network properties (owner only).
func (h *NetworkServiceHandler) UpdateNetwork(ctx context.Context, req *proto.UpdateNetworkRequest) (*proto.Network, error) {
	logger.Info("UpdateNetwork gRPC called", "network_id", req.NetworkId)

	// 1. Validate request
	if req.NetworkId == "" {
		return nil, fmt.Errorf("network_id is required")
	}

	// 2. Get authenticated user from token
	session, err := h.tokenManager.LoadSession()
	if err != nil {
		logger.Warn("UpdateNetwork: No valid session", "error", err)
		return nil, fmt.Errorf("authentication required")
	}

	// Parse JWT to get UserID
	userID, err := extractUserIDFromJWT(session.AccessToken)
	if err != nil {
		logger.Warn("UpdateNetwork: Failed to extract UserID from token", "error", err)
		return nil, fmt.Errorf("authentication required")
	}

	tenantID := h.defaultTenantID
	if tenantID == "" {
		tenantID = "default"
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

	// 4. Call existing HTTP service layer (reuse business logic)
	updated, err := h.networkService.UpdateNetwork(ctx, req.NetworkId, userID, tenantID, patch)
	if err != nil {
		// Check if it's a domain error
		if domainErr, ok := err.(*domain.Error); ok {
			logger.Warn("UpdateNetwork: Domain error",
				"network_id", req.NetworkId,
				"actor_id", userID,
				"error_code", domainErr.Code,
				"error", domainErr.Message,
			)
			return nil, fmt.Errorf("%s", domainErr.Message)
		}

		logger.Error("UpdateNetwork: Service error",
			"network_id", req.NetworkId,
			"actor_id", userID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to update network: %w", err)
	}

	// 5. Convert domain.Network to proto.Network
	protoNetwork := domainNetworkToProto(updated, userID)

	logger.Info("UpdateNetwork successful",
		"network_id", req.NetworkId,
		"actor_id", userID,
		"updated_fields", len(patch),
	)

	return protoNetwork, nil
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

// Helper: Convert domain.Network to proto.Network
func domainNetworkToProto(n *domain.Network, currentUserID string) *proto.Network {
	if n == nil {
		return nil
	}

	// Determine user's role in the network
	myRole := proto.NetworkRole_NETWORK_ROLE_MEMBER
	if n.CreatedBy == currentUserID {
		myRole = proto.NetworkRole_NETWORK_ROLE_OWNER
	}

	return &proto.Network{
		Id:          n.ID,
		Name:        n.Name,
		Description: "", // Not supported in domain model yet
		InviteCode:  "", // Not exposed in basic Network object
		MyRole:      myRole,
		PeerCount:   0, // TODO: Get from membership count
		OnlineCount: 0, // TODO: Get from active device count
		IsConnected: false, // TODO: Check if current device is connected
		// CreatedAt and JoinedAt would need timestamp conversion
	}
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
