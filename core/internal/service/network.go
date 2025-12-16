package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// NetworkService provides business logic for network operations
type NetworkService struct {
	networkRepo     repository.NetworkRepository
	idempotencyRepo repository.IdempotencyRepository
	aud             Auditor // Reuse the Auditor interface defined in membership.go (same package)
}

// NewNetworkService creates a new network service
func NewNetworkService(networkRepo repository.NetworkRepository, idempotencyRepo repository.IdempotencyRepository) *NetworkService {
	return &NetworkService{
		networkRepo:     networkRepo,
		idempotencyRepo: idempotencyRepo,
		aud:             noopAuditor, // comes from membership.go
	}
}

// SetAuditor allows wiring a real auditor from main
func (s *NetworkService) SetAuditor(a Auditor) {
	if a != nil {
		s.aud = a
	}
}

// CreateNetwork creates a new network with idempotency and business rule validation
func (s *NetworkService) CreateNetwork(ctx context.Context, req *domain.CreateNetworkRequest, userID, tenantID string, idempotencyKey string) (*domain.Network, error) {
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
	overlap, err := s.networkRepo.CheckCIDROverlap(ctx, req.CIDR, "", tenantID)
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
		ID:                 domain.GenerateNetworkID(),
		TenantID:           tenantID,
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
			slog.Error("Failed to store idempotency record", "error", err)
		}
	}

	s.audit(ctx, tenantID, "NETWORK_CREATED", userID, network.ID, map[string]any{"name": network.Name, "cidr": network.CIDR})

	return network, nil
}

// ListNetworks retrieves networks with filtering and pagination
func (s *NetworkService) ListNetworks(ctx context.Context, req *domain.ListNetworksRequest, userID, tenantID string, isAdmin bool) ([]*domain.Network, string, error) {
	// Validate visibility parameter
	if req.Visibility != "public" && req.Visibility != "mine" && req.Visibility != "all" {
		req.Visibility = "public"
	}

	// Non-admin users cannot access "all" visibility
	if req.Visibility == "all" && !isAdmin {
		return nil, "", domain.NewError(domain.ErrNotAuthorized,
			"Only administrators can view all networks",
			map[string]string{"required_role": "admin"})
	}

	filter := repository.NetworkFilter{
		Visibility: req.Visibility,
		UserID:     userID,
		TenantID:   tenantID,
		IsAdmin:    isAdmin,
		Search:     req.Search,
		Limit:      req.Limit,
		Cursor:     req.Cursor,
	}

	networks, nextCursor, err := s.networkRepo.List(ctx, filter)
	if err != nil {
		return nil, "", err
	}

	s.audit(ctx, tenantID, "NETWORK_LIST", userID, "", map[string]any{"visibility": req.Visibility, "count": len(networks)})

	return networks, nextCursor, nil
}

// GetNetwork retrieves a single network by ID
func (s *NetworkService) GetNetwork(ctx context.Context, id, actor, tenantID string) (*domain.Network, error) {
	net, err := s.networkRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Ensure network belongs to the tenant
	if net.TenantID != tenantID {
		return nil, domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}

	// Private networks: future enhancement may require membership check; for now always return if found
	s.audit(ctx, tenantID, "NETWORK_GET", actor, id, nil)
	return net, nil
}

// UpdateNetwork updates mutable fields (name, visibility, join_policy, dns, mtu, split_tunnel) enforcing uniqueness & simple validation
func (s *NetworkService) UpdateNetwork(ctx context.Context, id string, actor, tenantID string, patch map[string]any) (*domain.Network, error) {
	// First check if network exists and belongs to tenant
	existing, err := s.networkRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.TenantID != tenantID {
		return nil, domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}

	updated, err := s.networkRepo.Update(ctx, id, func(n *domain.Network) error {
		if v, ok := patch["name"].(string); ok && v != "" && v != n.Name {
			if len(v) < 3 || len(v) > 64 {
				return domain.NewError(domain.ErrInvalidRequest, "Invalid name length", map[string]string{"field": "name"})
			}
			n.Name = v
		}
		if v, ok := patch["visibility"].(string); ok && v != "" && v != string(n.Visibility) {
			if v != string(domain.NetworkVisibilityPublic) && v != string(domain.NetworkVisibilityPrivate) {
				return domain.NewError(domain.ErrInvalidRequest, "Invalid visibility", map[string]string{"field": "visibility"})
			}
			n.Visibility = domain.NetworkVisibility(v)
		}
		if v, ok := patch["join_policy"].(string); ok && v != "" && v != string(n.JoinPolicy) {
			switch v {
			case string(domain.JoinPolicyOpen), string(domain.JoinPolicyInvite), string(domain.JoinPolicyApproval):
				n.JoinPolicy = domain.JoinPolicy(v)
			default:
				return domain.NewError(domain.ErrInvalidRequest, "Invalid join_policy", map[string]string{"field": "join_policy"})
			}
		}
		// DNS field - can be set to a valid IP or cleared (null/empty)
		if v, exists := patch["dns"]; exists {
			if v == nil || v == "" {
				n.DNS = nil
			} else if dns, ok := v.(string); ok {
				// Validate DNS IP format
				if net.ParseIP(dns) == nil {
					return domain.NewError(domain.ErrInvalidRequest, "Invalid DNS IP address", map[string]string{"field": "dns"})
				}
				n.DNS = &dns
			}
		}
		// MTU field - can be set to a valid range or cleared (null)
		if v, exists := patch["mtu"]; exists {
			if v == nil {
				n.MTU = nil
			} else if mtuFloat, ok := v.(float64); ok {
				mtu := int(mtuFloat)
				if mtu < 576 || mtu > 1500 {
					return domain.NewError(domain.ErrInvalidRequest, "MTU must be between 576 and 1500", map[string]string{"field": "mtu"})
				}
				n.MTU = &mtu
			}
		}
		// Split Tunnel field
		if v, exists := patch["split_tunnel"]; exists {
			if v == nil {
				n.SplitTunnel = nil
			} else if splitTunnel, ok := v.(bool); ok {
				n.SplitTunnel = &splitTunnel
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.audit(ctx, tenantID, "NETWORK_UPDATED", actor, id, patch)
	return updated, nil
}

// DeleteNetwork performs a soft delete
func (s *NetworkService) DeleteNetwork(ctx context.Context, id, actor, tenantID string) error {
	// First check if network exists and belongs to tenant
	existing, err := s.networkRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}

	if err := s.networkRepo.SoftDelete(ctx, id, time.Now()); err != nil {
		return err
	}

	s.audit(ctx, tenantID, "NETWORK_DELETED", actor, id, nil)
	return nil
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
func (s *NetworkService) audit(ctx context.Context, tenantID, action, actor, objectID string, details map[string]any) {
	s.aud.Event(ctx, tenantID, action, actor, objectID, details)
}
