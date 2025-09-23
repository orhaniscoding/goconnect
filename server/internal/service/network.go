package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// NetworkService provides business logic for network operations
type NetworkService struct {
	networkRepo      repository.NetworkRepository
	idempotencyRepo  repository.IdempotencyRepository
}

// NewNetworkService creates a new network service
func NewNetworkService(networkRepo repository.NetworkRepository, idempotencyRepo repository.IdempotencyRepository) *NetworkService {
	return &NetworkService{
		networkRepo:     networkRepo,
		idempotencyRepo: idempotencyRepo,
	}
}

// CreateNetwork creates a new network with idempotency and business rule validation
func (s *NetworkService) CreateNetwork(ctx context.Context, req *domain.CreateNetworkRequest, userID string, idempotencyKey string) (*domain.Network, error) {
	// Handle idempotency first before any business logic
	if idempotencyKey != "" {
		if existing, err := s.handleIdempotency(ctx, idempotencyKey, req); err == nil && existing != nil {
			return existing, nil
		} else if err != nil {
			return nil, err
		}
	}

	// Validate CIDR format
	if err := domain.ValidateCIDR(req.CIDR); err != nil {
		return nil, domain.NewError(domain.ErrCIDRInvalid, 
			fmt.Sprintf("Invalid CIDR format: %v", err), 
			map[string]string{"field": "cidr"})
	}

	// Check for CIDR overlap
	overlap, err := s.networkRepo.CheckCIDROverlap(ctx, req.CIDR, "")
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to check CIDR overlap", nil)
	}
	if overlap {
		return nil, domain.NewError(domain.ErrCIDROverlap, 
			"CIDR range overlaps with existing network", 
			map[string]string{"cidr": req.CIDR})
	}

	// Create network
	now := time.Now()
	network := &domain.Network{
		ID:                  domain.GenerateNetworkID(),
		TenantID:           "default", // TODO: Extract from user context
		Name:               req.Name,
		Visibility:         req.Visibility,
		JoinPolicy:         req.JoinPolicy,
		CIDR:               req.CIDR,
		DNS:                req.DNS,
		MTU:                req.MTU,
		SplitTunnel:        req.SplitTunnel,
		CreatedBy:          userID,
		CreatedAt:          now,
		UpdatedAt:          now,
		ModerationRedacted: false,
	}

	// Store network
	if err := s.networkRepo.Create(ctx, network); err != nil {
		return nil, err
	}

	// Store idempotency record
	if idempotencyKey != "" {
		if err := s.storeIdempotencyRecord(ctx, idempotencyKey, req, network); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to store idempotency record: %v\n", err)
		}
	}

	// TODO: Add audit log entry
	s.logAuditEvent(ctx, "network.created", userID, network.ID, map[string]interface{}{
		"name": network.Name,
		"cidr": network.CIDR,
	})

	return network, nil
}

// ListNetworks retrieves networks with filtering and pagination
func (s *NetworkService) ListNetworks(ctx context.Context, req *domain.ListNetworksRequest, userID string, isAdmin bool) ([]*domain.Network, string, error) {
	// Validate visibility parameter
	if req.Visibility != "public" && req.Visibility != "mine" && req.Visibility != "all" {
		req.Visibility = "public"
	}

	// Non-admin users cannot access "all" visibility
	if req.Visibility == "all" && !isAdmin {
		return nil, "", domain.NewError(domain.ErrForbidden, 
			"Only administrators can view all networks", 
			map[string]string{"required_role": "admin"})
	}

	filter := repository.NetworkFilter{
		Visibility: req.Visibility,
		UserID:     userID,
		IsAdmin:    isAdmin,
		Search:     req.Search,
		Limit:      req.Limit,
		Cursor:     req.Cursor,
	}

	networks, nextCursor, err := s.networkRepo.List(ctx, filter)
	if err != nil {
		return nil, "", err
	}

	// TODO: Add audit log for list operations (optional)
	s.logAuditEvent(ctx, "network.listed", userID, "", map[string]interface{}{
		"visibility": req.Visibility,
		"count":      len(networks),
	})

	return networks, nextCursor, nil
}

// handleIdempotency checks and handles idempotency key conflicts
func (s *NetworkService) handleIdempotency(ctx context.Context, key string, req *domain.CreateNetworkRequest) (*domain.Network, error) {
	record, err := s.idempotencyRepo.Get(ctx, key)
	if err != nil {
		// Key not found or expired - this is expected for new requests
		return nil, nil
	}

	// Check if request body matches
	bodyHash, err := domain.HashRequestBody(req)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to hash request body", nil)
	}

	if record.BodyHash != bodyHash {
		return nil, domain.NewError(domain.ErrIdempotencyConflict, 
			"Idempotency key already used with different request body", 
			map[string]string{"key": key})
	}

	// Deserialize stored response
	var network domain.Network
	if err := json.Unmarshal(record.Response, &network); err != nil {
		// If we can't deserialize, treat as new request
		return nil, nil
	}

	return &network, nil
}

// storeIdempotencyRecord stores the response for future idempotency checks
func (s *NetworkService) storeIdempotencyRecord(ctx context.Context, key string, req *domain.CreateNetworkRequest, network *domain.Network) error {
	bodyHash, err := domain.HashRequestBody(req)
	if err != nil {
		return err
	}

	response, err := json.Marshal(network)
	if err != nil {
		return err
	}

	record := domain.NewIdempotencyRecord(key, bodyHash)
	record.Response = response

	return s.idempotencyRepo.Set(ctx, record)
}

// logAuditEvent logs audit events (placeholder implementation)
func (s *NetworkService) logAuditEvent(ctx context.Context, action, actor, objectID string, details map[string]interface{}) {
	// TODO: Implement proper audit logging
	fmt.Printf("AUDIT: action=%s actor=%s object=%s details=%+v\n", action, actor, objectID, details)
}