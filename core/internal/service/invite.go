package service

import (
	"context"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// InviteService handles network invitation operations
type InviteService struct {
	inviteRepo  repository.InviteTokenRepository
	networkRepo repository.NetworkRepository
	memberRepo  repository.MembershipRepository
	auditor     audit.Auditor
	baseURL     string // For generating invite URLs
}

// NewInviteService creates a new invite service
func NewInviteService(
	inviteRepo repository.InviteTokenRepository,
	networkRepo repository.NetworkRepository,
	memberRepo repository.MembershipRepository,
	baseURL string,
) *InviteService {
	return &InviteService{
		inviteRepo:  inviteRepo,
		networkRepo: networkRepo,
		memberRepo:  memberRepo,
		baseURL:     baseURL,
	}
}

// SetAuditor sets the auditor for the service
func (s *InviteService) SetAuditor(a audit.Auditor) {
	s.auditor = a
}

// CreateInviteOptions contains options for creating an invite
type CreateInviteOptions struct {
	ExpiresIn int // seconds, default 24 hours
	UsesMax   int // 0 = unlimited
}

// CreateInvite creates a new invite token for a network
func (s *InviteService) CreateInvite(ctx context.Context, networkID, tenantID, userID string, opts CreateInviteOptions) (*domain.InviteTokenResponse, error) {
	// Verify network exists and user has permission
	network, err := s.networkRepo.GetByID(ctx, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get network for invite: %w", err)
	}

	// Verify user is owner/admin of the network
	membership, err := s.memberRepo.Get(ctx, networkID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get network membership for invite: %w", err)
	}

	if membership.Role != domain.RoleOwner && membership.Role != domain.RoleAdmin {
		return nil, domain.NewError(domain.ErrNotAuthorized, "Only owners and admins can create invites", nil)
	}

	// Check network join_policy (only invite/approval networks need invite tokens)
	if network.JoinPolicy == domain.JoinPolicyOpen {
		return nil, domain.NewError(domain.ErrInvalidRequest, "Open networks don't need invite tokens", nil)
	}

	// Generate token
	tokenStr, err := domain.GenerateInviteToken()
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to generate token", nil)
	}

	// Set defaults
	expiresIn := opts.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 86400 // 24 hours default
	}

	token := &domain.InviteToken{
		ID:        domain.GenerateInviteID(),
		NetworkID: networkID,
		TenantID:  tenantID,
		Token:     tokenStr,
		CreatedBy: userID,
		ExpiresAt: time.Now().Add(time.Duration(expiresIn) * time.Second),
		UsesMax:   opts.UsesMax,
		UsesLeft:  opts.UsesMax, // starts equal to max
		CreatedAt: time.Now(),
	}

	if err := s.inviteRepo.Create(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to create invite token: %w", err)
	}

	// Audit log
	if s.auditor != nil {
		s.auditor.Event(ctx, tenantID, "INVITE_CREATED", userID, token.ID, map[string]interface{}{
			"network_id": networkID,
			"expires_at": token.ExpiresAt,
			"uses_max":   opts.UsesMax,
		})
	}

	return s.toResponse(token), nil
}

// ValidateInvite validates an invite token and returns the token details
func (s *InviteService) ValidateInvite(ctx context.Context, tokenStr string) (*domain.InviteToken, error) {
	token, err := s.inviteRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get invite by token: %w", err)
	}

	if !token.IsValid() {
		if token.RevokedAt != nil {
			return nil, domain.NewError(domain.ErrInviteTokenRevoked, "This invite has been revoked", nil)
		}
		return nil, domain.NewError(domain.ErrInviteTokenExpired, "This invite has expired", nil)
	}

	return token, nil
}

// UseInvite uses an invite token to join a network
func (s *InviteService) UseInvite(ctx context.Context, tokenStr, userID string) (*domain.InviteToken, error) {
	token, err := s.inviteRepo.UseToken(ctx, tokenStr)
	if err != nil {
		return nil, fmt.Errorf("failed to use invite token: %w", err)
	}

	// Audit log
	if s.auditor != nil {
		s.auditor.Event(ctx, token.TenantID, "INVITE_USED", userID, token.ID, map[string]interface{}{
			"network_id": token.NetworkID,
			"uses_left":  token.UsesLeft,
		})
	}

	return token, nil
}

// ListInvites lists all active invites for a network
func (s *InviteService) ListInvites(ctx context.Context, networkID, userID string) ([]*domain.InviteTokenResponse, error) {
	// Verify user has permission to see invites
	membership, err := s.memberRepo.Get(ctx, networkID, userID)
	if err != nil {
		return nil, domain.NewError(domain.ErrNotAuthorized, "You must be a member of the network", nil)
	}

	if membership.Role != domain.RoleOwner && membership.Role != domain.RoleAdmin {
		return nil, domain.NewError(domain.ErrNotAuthorized, "Only owners and admins can view invites", nil)
	}

	tokens, err := s.inviteRepo.ListByNetwork(ctx, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to list network invites: %w", err)
	}

	result := make([]*domain.InviteTokenResponse, 0, len(tokens))
	for _, t := range tokens {
		result = append(result, s.toResponse(t))
	}

	return result, nil
}

// RevokeInvite revokes an invite token
func (s *InviteService) RevokeInvite(ctx context.Context, inviteID, networkID, tenantID, userID string) error {
	// Get the invite to verify network match
	invite, err := s.inviteRepo.GetByID(ctx, inviteID)
	if err != nil {
		return fmt.Errorf("failed to get invite for revocation: %w", err)
	}

	if invite.NetworkID != networkID {
		return domain.NewError(domain.ErrNotFound, "Invite not found in this network", nil)
	}

	// Verify user has permission
	membership, err := s.memberRepo.Get(ctx, networkID, userID)
	if err != nil {
		return fmt.Errorf("failed to get membership for invite revocation: %w", err)
	}

	if membership.Role != domain.RoleOwner && membership.Role != domain.RoleAdmin {
		return domain.NewError(domain.ErrNotAuthorized, "Only owners and admins can revoke invites", nil)
	}

	if err := s.inviteRepo.Revoke(ctx, inviteID); err != nil {
		return fmt.Errorf("failed to revoke invite token: %w", err)
	}

	// Audit log
	if s.auditor != nil {
		s.auditor.Event(ctx, tenantID, "INVITE_REVOKED", userID, inviteID, map[string]interface{}{
			"network_id": networkID,
		})
	}

	return nil
}

// GetInviteByID retrieves an invite by ID
func (s *InviteService) GetInviteByID(ctx context.Context, inviteID string) (*domain.InviteTokenResponse, error) {
	token, err := s.inviteRepo.GetByID(ctx, inviteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invite by ID: %w", err)
	}
	return s.toResponse(token), nil
}

// toResponse converts a domain invite token to response format
func (s *InviteService) toResponse(token *domain.InviteToken) *domain.InviteTokenResponse {
	return &domain.InviteTokenResponse{
		ID:        token.ID,
		NetworkID: token.NetworkID,
		Token:     token.Token,
		InviteURL: fmt.Sprintf("%s/join?invite=%s", s.baseURL, token.Token),
		ExpiresAt: token.ExpiresAt,
		UsesMax:   token.UsesMax,
		UsesLeft:  token.UsesLeft,
		CreatedAt: token.CreatedAt,
		IsActive:  token.IsValid(),
	}
}
