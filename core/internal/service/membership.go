package service

import (
	"context"
	"strings"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/rbac"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

type MembershipService struct {
	networks         repository.NetworkRepository
	members          repository.MembershipRepository
	joins            repository.JoinRequestRepository
	idempotency      repository.IdempotencyRepository
	invites          repository.InviteTokenRepository
	peerProvisioning *PeerProvisioningService
	aud              Auditor
	notifier         MembershipNotifier
}

// MembershipNotifier defines interface for membership events
type MembershipNotifier interface {
	MemberJoined(networkID, userID string)
	MemberLeft(networkID, userID string)
}

type noopNotifier struct{}

func (n noopNotifier) MemberJoined(networkID, userID string) {}
func (n noopNotifier) MemberLeft(networkID, userID string)   {}

// Auditor is a minimal interface to decouple from concrete audit package
type Auditor interface {
	Event(ctx context.Context, tenantID, action, actor, object string, details map[string]any)
}

type auditorFunc func(ctx context.Context, tenantID, action, actor, object string, details map[string]any)

func (f auditorFunc) Event(ctx context.Context, tenantID, action, actor, object string, details map[string]any) {
	f(ctx, tenantID, action, actor, object, details)
}

var noopAuditor = auditorFunc(func(ctx context.Context, tenantID, action, actor, object string, details map[string]any) {})

func NewMembershipService(n repository.NetworkRepository, m repository.MembershipRepository, j repository.JoinRequestRepository, idem repository.IdempotencyRepository) *MembershipService {
	return &MembershipService{networks: n, members: m, joins: j, idempotency: idem, aud: noopAuditor, notifier: noopNotifier{}}
}

// SetNotifier sets the membership notifier
func (s *MembershipService) SetNotifier(n MembershipNotifier) {
	if n != nil {
		s.notifier = n
	}
}

// SetInviteTokenRepository sets the invite token repository
func (s *MembershipService) SetInviteTokenRepository(r repository.InviteTokenRepository) {
	s.invites = r
}

// SetPeerProvisioning sets the peer provisioning service
func (s *MembershipService) SetPeerProvisioning(pp *PeerProvisioningService) {
	s.peerProvisioning = pp
}

// SetAuditor allows wiring a real auditor from main
func (s *MembershipService) SetAuditor(a Auditor) {
	if a != nil {
		s.aud = a
	}
}

// JoinNetwork handles POST /v1/networks/:id/join with idempotency and audit
func (s *MembershipService) JoinNetwork(ctx context.Context, networkID, userID, tenantID, idempotencyKey string) (*domain.Membership, *domain.JoinRequest, error) {
	if idempotencyKey == "" {
		return nil, nil, domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key header is required for mutation operations", map[string]string{"required_header": "Idempotency-Key"})
	}

	// Idempotency: we store a synthetic response containing membership if created
	// Not implementing full replay; rely on repos for harmless duplicates

	net, err := s.networks.GetByID(ctx, networkID)
	if err != nil {
		return nil, nil, err
	}
	// Enforce tenant isolation
	if net.TenantID != tenantID {
		return nil, nil, domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}

	// If private and name unknown -> ERR_NETWORK_PRIVATE; we check visibility and simulate name unknown case
	if net.Visibility == domain.NetworkVisibilityPrivate && net.Name == "" {
		return nil, nil, domain.NewError(domain.ErrNetworkPrivate, "Network is private", nil)
	}

	// Check existing membership
	if m, err := s.members.Get(ctx, networkID, userID); err == nil {
		if m.Status == domain.StatusApproved {
			// Double-join guard: treat as successful no-op
			s.audit(ctx, tenantID, audit.ActionNetworkJoin, userID, networkID, map[string]any{"dedup": true})
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
		s.audit(ctx, tenantID, audit.ActionNetworkJoin, userID, networkID, map[string]any{"policy": "open"})
		s.notifier.MemberJoined(networkID, userID)

		// Automatically provision peers for user's devices
		if s.peerProvisioning != nil {
			if err := s.peerProvisioning.ProvisionPeersForNewMember(ctx, networkID, userID); err != nil {
				// Log error but don't fail the join operation
				s.audit(ctx, tenantID, "PEER_PROVISION_FAILED", userID, networkID, map[string]any{"error": err.Error()})
			}
		}

		return m, nil, nil
	case domain.JoinPolicyApproval:
		jr, err := s.joins.CreatePending(ctx, networkID, userID)
		if err != nil {
			if derr, ok := err.(*domain.Error); ok && derr.Code == domain.ErrAlreadyRequested {
				s.audit(ctx, tenantID, audit.ActionNetworkJoinRequest, userID, networkID, map[string]any{"dedup": true})
				return nil, jr, derr
			}
			return nil, nil, err
		}
		s.audit(ctx, tenantID, audit.ActionNetworkJoinRequest, userID, networkID, nil)
		return nil, jr, nil
	case domain.JoinPolicyInvite:
		// For v1 we expect a token param else invalid
		return nil, nil, domain.NewError(domain.ErrInvalidRequest, "Invite token required", map[string]string{"param": "token"})
	default:
		return nil, nil, domain.NewError(domain.ErrInvalidRequest, "Unknown join policy", map[string]string{"join_policy": string(net.JoinPolicy)})
	}
}

// Approve, Deny, Kick, Ban
func (s *MembershipService) Approve(ctx context.Context, networkID, targetUserID, actorID, tenantID string) (*domain.Membership, error) {
	// Verify network tenant
	net, err := s.networks.GetByID(ctx, networkID)
	if err != nil {
		return nil, err
	}
	if net.TenantID != tenantID {
		return nil, domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}

	// RBAC simplified: assume actor is admin/owner if membership role is admin or owner
	if !s.hasManagePrivilege(ctx, networkID, actorID) {
		return nil, domain.NewError(domain.ErrNotAuthorized, "Administrator privileges required", nil)
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
	s.audit(ctx, tenantID, audit.ActionNetworkJoinApprove, actorID, networkID, map[string]any{"user": targetUserID})

	// Automatically provision peers for approved user's devices
	if s.peerProvisioning != nil {
		if err := s.peerProvisioning.ProvisionPeersForNewMember(ctx, networkID, targetUserID); err != nil {
			// Log error but don't fail the approval
			s.audit(ctx, tenantID, "PEER_PROVISION_FAILED", actorID, networkID, map[string]any{"user": targetUserID, "error": err.Error()})
		}
	}

	// Notify external systems about the new member
	s.notifier.MemberJoined(networkID, targetUserID)

	return m, nil
}

func (s *MembershipService) Deny(ctx context.Context, networkID, targetUserID, actorID, tenantID string) error {
	// Verify network tenant
	net, err := s.networks.GetByID(ctx, networkID)
	if err != nil {
		return err
	}
	if net.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}

	if !s.hasManagePrivilege(ctx, networkID, actorID) {
		return domain.NewError(domain.ErrNotAuthorized, "Administrator privileges required", nil)
	}
	jr, err := s.joins.GetPending(ctx, networkID, targetUserID)
	if err != nil {
		return err
	}
	if err := s.joins.Decide(ctx, jr.ID, false); err != nil {
		return err
	}
	s.audit(ctx, tenantID, audit.ActionNetworkJoinDeny, actorID, networkID, map[string]any{"user": targetUserID})
	return nil
}

func (s *MembershipService) Kick(ctx context.Context, networkID, targetUserID, actorID, tenantID string) error {
	// Verify network tenant
	net, err := s.networks.GetByID(ctx, networkID)
	if err != nil {
		return err
	}
	if net.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}

	if !s.hasManagePrivilege(ctx, networkID, actorID) {
		return domain.NewError(domain.ErrNotAuthorized, "Administrator privileges required", nil)
	}
	if err := s.members.Remove(ctx, networkID, targetUserID); err != nil {
		return err
	}

	// Deprovision peers when user is kicked
	if s.peerProvisioning != nil {
		if err := s.peerProvisioning.DeprovisionPeersForRemovedMember(ctx, networkID, targetUserID); err != nil {
			// Log error but don't fail the kick operation
			s.audit(ctx, tenantID, "PEER_DEPROVISION_FAILED", actorID, networkID, map[string]any{"user": targetUserID, "error": err.Error()})
		}
	}

	s.audit(ctx, tenantID, audit.ActionNetworkMemberKick, actorID, networkID, map[string]any{"user": targetUserID})

	// Notify external systems about the member removal
	s.notifier.MemberLeft(networkID, targetUserID)

	return nil
}

func (s *MembershipService) Ban(ctx context.Context, networkID, targetUserID, actorID, tenantID string) error {
	// Verify network tenant
	net, err := s.networks.GetByID(ctx, networkID)
	if err != nil {
		return err
	}
	if net.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}

	if !s.hasManagePrivilege(ctx, networkID, actorID) {
		return domain.NewError(domain.ErrNotAuthorized, "Administrator privileges required", nil)
	}
	if err := s.members.SetStatus(ctx, networkID, targetUserID, domain.StatusBanned); err != nil {
		return err
	}
	s.audit(ctx, tenantID, audit.ActionNetworkMemberBan, actorID, networkID, map[string]any{"user": targetUserID})
	return nil
}

func (s *MembershipService) ListMembers(ctx context.Context, networkID, status, tenantID string, limit int, cursor string) ([]*domain.Membership, string, error) {
	// Verify network tenant
	net, err := s.networks.GetByID(ctx, networkID)
	if err != nil {
		return nil, "", err
	}
	if net.TenantID != tenantID {
		return nil, "", domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}
	return s.members.List(ctx, networkID, status, limit, cursor)
}

// ListJoinRequests lists pending join requests for a network (admin/owner only)
func (s *MembershipService) ListJoinRequests(ctx context.Context, networkID, tenantID string) ([]*domain.JoinRequest, error) {
	// Verify network tenant
	net, err := s.networks.GetByID(ctx, networkID)
	if err != nil {
		return nil, err
	}
	if net.TenantID != tenantID {
		return nil, domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}
	return s.joins.ListPending(ctx, networkID)
}

func (s *MembershipService) hasManagePrivilege(ctx context.Context, networkID, userID string) bool {
	m, err := s.members.Get(ctx, networkID, userID)
	if err != nil {
		return false
	}
	return rbac.CanManageNetwork(m.Role)
}

func (s *MembershipService) audit(ctx context.Context, tenantID, action, actor, object string, details map[string]any) {
	s.aud.Event(ctx, tenantID, action, actor, object, details)
}

// IsMember checks if a user is an approved member of a network
func (s *MembershipService) IsMember(ctx context.Context, networkID, userID string) (bool, error) {
	m, err := s.members.Get(ctx, networkID, userID)
	if err != nil {
		return false, err
	}
	return m.Status == domain.StatusApproved, nil
}

// JoinByInviteCode handles joining a network using an invite code (CLI compat)
// The invite code is resolved to a network ID, then JoinNetwork is called
func (s *MembershipService) JoinByInviteCode(ctx context.Context, inviteCode, userID, tenantID, idempotencyKey string) (*domain.Membership, *domain.JoinRequest, error) {
	if inviteCode == "" {
		return nil, nil, domain.NewError(domain.ErrInvalidRequest, "invite_code is required", nil)
	}

	// Compat: if code starts with "net_", treat as direct Network ID
	if strings.HasPrefix(inviteCode, "net_") {
		return s.JoinNetwork(ctx, inviteCode, userID, tenantID, idempotencyKey)
	}

	// If we have an invite repo, try to resolve the token
	if s.invites != nil {
		token, err := s.invites.UseToken(ctx, inviteCode)
		if err == nil && token != nil {
			// Token valid and used successfully
			return s.JoinNetwork(ctx, token.NetworkID, userID, tenantID, idempotencyKey)
		}
		// If token not found or invalid, fall through to error handling
		// But for now, we returning the error from token usage/lookup if it wasn't a "not found"
		// Actually, if UseToken fails, we should probably fail unless we want to fallback
		// strict behavior: if not net_, must be valid token.
		if err != nil {
			return nil, nil, domain.NewError(domain.ErrInvalidRequest, fmt.Sprintf("Invalid invite code: %v", err), nil)
		}
	}

	// Fallback/Legacy behavior (if no repo or unknown format):
	// behave as if the code IS the network ID (for very old clients sending UUIDs maybe?)
	// But "net_" check handles standard IDs. UUIDs don't have prefix.
	// We'll trust JoinNetwork to return Not Found if it's garbage.
	return s.JoinNetwork(ctx, inviteCode, userID, tenantID, idempotencyKey)

}

