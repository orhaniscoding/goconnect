package service

import (
	"context"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

type MembershipService struct {
	networks    repository.NetworkRepository
	members     repository.MembershipRepository
	joins       repository.JoinRequestRepository
	idempotency repository.IdempotencyRepository
}

func NewMembershipService(n repository.NetworkRepository, m repository.MembershipRepository, j repository.JoinRequestRepository, idem repository.IdempotencyRepository) *MembershipService {
	return &MembershipService{networks: n, members: m, joins: j, idempotency: idem}
}

// JoinNetwork handles POST /v1/networks/:id/join with idempotency and audit
func (s *MembershipService) JoinNetwork(ctx context.Context, networkID, userID, idempotencyKey string) (*domain.Membership, *domain.JoinRequest, error) {
	if idempotencyKey == "" {
		return nil, nil, domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key header is required for mutation operations", map[string]string{"required_header": "Idempotency-Key"})
	}

	// Idempotency: we store a synthetic response containing membership if created
	// Not implementing full replay; rely on repos for harmless duplicates

	net, err := s.networks.GetByID(ctx, networkID)
	if err != nil {
		return nil, nil, err
	}

	// If private and name unknown -> ERR_NETWORK_PRIVATE; we check visibility and simulate name unknown case
	if net.Visibility == domain.NetworkVisibilityPrivate && net.Name == "" {
		return nil, nil, domain.NewError(domain.ErrNetworkPrivate, "Network is private", nil)
	}

	// Check existing membership
	if m, err := s.members.Get(ctx, networkID, userID); err == nil {
		if m.Status == domain.StatusApproved {
			// Double-join guard: treat as successful no-op
			s.audit(ctx, "NETWORK_JOIN", userID, networkID, map[string]any{"dedup": true})
			return m, nil, nil
		}
		if m.Status == domain.StatusBanned {
			return nil, nil, domain.NewError(domain.ErrUserBanned, "User is banned", nil)
		}
	}

	now := time.Now()
	switch net.JoinPolicy {
	case domain.JoinPolicyOpen:
		m, err := s.members.UpsertApproved(ctx, networkID, userID, domain.RoleMember, now)
		if err != nil {
			return nil, nil, err
		}
		s.audit(ctx, "NETWORK_JOIN", userID, networkID, map[string]any{"policy": "open"})
		return m, nil, nil
	case domain.JoinPolicyApproval:
		jr, err := s.joins.CreatePending(ctx, networkID, userID)
		if err != nil {
			if derr, ok := err.(*domain.Error); ok && derr.Code == domain.ErrAlreadyRequested {
				s.audit(ctx, "NETWORK_JOIN_REQUEST", userID, networkID, map[string]any{"dedup": true})
				return nil, jr, derr
			}
			return nil, nil, err
		}
		s.audit(ctx, "NETWORK_JOIN_REQUEST", userID, networkID, nil)
		return nil, jr, nil
	case domain.JoinPolicyInvite:
		// For v1 we expect a token param else invalid
		return nil, nil, domain.NewError(domain.ErrInvalidRequest, "Invite token required", map[string]string{"param": "token"})
	default:
		return nil, nil, domain.NewError(domain.ErrInvalidRequest, "Unknown join policy", map[string]string{"join_policy": string(net.JoinPolicy)})
	}
}

// Approve, Deny, Kick, Ban
func (s *MembershipService) Approve(ctx context.Context, networkID, targetUserID, actorID string) (*domain.Membership, error) {
	// RBAC simplified: assume actor is admin/owner if membership role is admin or owner
	if !s.isAdmin(ctx, networkID, actorID) {
		return nil, domain.NewError(domain.ErrForbidden, "Administrator privileges required", nil)
	}
	// find pending join
	jr, err := s.joins.GetPending(ctx, networkID, targetUserID)
	if err != nil {
		return nil, err
	}
	if err := s.joins.Decide(ctx, jr.ID, true); err != nil {
		return nil, err
	}
	m, err := s.members.UpsertApproved(ctx, networkID, targetUserID, domain.RoleMember, time.Now())
	if err != nil {
		return nil, err
	}
	s.audit(ctx, "NETWORK_JOIN_APPROVE", actorID, networkID, map[string]any{"user": targetUserID})
	return m, nil
}

func (s *MembershipService) Deny(ctx context.Context, networkID, targetUserID, actorID string) error {
	if !s.isAdmin(ctx, networkID, actorID) {
		return domain.NewError(domain.ErrForbidden, "Administrator privileges required", nil)
	}
	jr, err := s.joins.GetPending(ctx, networkID, targetUserID)
	if err != nil {
		return err
	}
	if err := s.joins.Decide(ctx, jr.ID, false); err != nil {
		return err
	}
	s.audit(ctx, "NETWORK_JOIN_DENY", actorID, networkID, map[string]any{"user": targetUserID})
	return nil
}

func (s *MembershipService) Kick(ctx context.Context, networkID, targetUserID, actorID string) error {
	if !s.isAdmin(ctx, networkID, actorID) {
		return domain.NewError(domain.ErrForbidden, "Administrator privileges required", nil)
	}
	if err := s.members.Remove(ctx, networkID, targetUserID); err != nil {
		return err
	}
	s.audit(ctx, "NETWORK_MEMBER_KICK", actorID, networkID, map[string]any{"user": targetUserID})
	return nil
}

func (s *MembershipService) Ban(ctx context.Context, networkID, targetUserID, actorID string) error {
	if !s.isAdmin(ctx, networkID, actorID) {
		return domain.NewError(domain.ErrForbidden, "Administrator privileges required", nil)
	}
	if err := s.members.SetStatus(ctx, networkID, targetUserID, domain.StatusBanned); err != nil {
		return err
	}
	s.audit(ctx, "NETWORK_MEMBER_BAN", actorID, networkID, map[string]any{"user": targetUserID})
	return nil
}

func (s *MembershipService) ListMembers(ctx context.Context, networkID, status string, limit int, cursor string) ([]*domain.Membership, string, error) {
	return s.members.List(ctx, networkID, status, limit, cursor)
}

func (s *MembershipService) isAdmin(ctx context.Context, networkID, userID string) bool {
	m, err := s.members.Get(ctx, networkID, userID)
	if err != nil {
		return false
	}
	return m.Role == domain.RoleAdmin || m.Role == domain.RoleOwner
}

func (s *MembershipService) audit(ctx context.Context, action, actor, object string, details map[string]any) {
	// TODO: proper audit pipeline
	fmt.Printf("AUDIT: %s actor=%s object=%s details=%+v\n", action, actor, object, details)
}
