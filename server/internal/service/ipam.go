package service

import (
	"context"

	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/rbac"
)

// IPAMService provides IP allocation logic on top of repositories.
type IPAMService struct {
	networks repository.NetworkRepository
	members  repository.MembershipRepository
	ipam     repository.IPAMRepository
	aud      Auditor
}

// NewIPAMService creates a new IPAM service. Membership repository is required to ensure only approved members get IPs.
func NewIPAMService(n repository.NetworkRepository, m repository.MembershipRepository, ip repository.IPAMRepository) *IPAMService {
	return &IPAMService{networks: n, members: m, ipam: ip, aud: noopAuditor}
}

// SetAuditor wires a real auditor.
func (s *IPAMService) SetAuditor(a Auditor) {
	if a != nil {
		s.aud = a
	}
}

// AllocateIP returns existing allocation or assigns the next available IP.
func (s *IPAMService) AllocateIP(ctx context.Context, networkID, userID string) (*domain.IPAllocation, error) {
	// Retrieve network first
	netw, err := s.networks.GetByID(ctx, networkID)
	if err != nil {
		return nil, err
	}
	// Enforce membership: must exist & be approved
	if s.members != nil { // defensive if nil in some legacy tests
		m, mErr := s.members.Get(ctx, networkID, userID)
		if mErr != nil {
			// hide distinction (not found) behind authorization failure
			return nil, domain.NewError(domain.ErrNotAuthorized, "Membership required to allocate IP", nil)
		}
		if m.Status != domain.StatusApproved {
			return nil, domain.NewError(domain.ErrNotAuthorized, "Membership not approved", map[string]string{"status": string(m.Status)})
		}
	}
	alloc, err := s.ipam.GetOrAllocate(ctx, networkID, userID, netw.CIDR)
	if err != nil {
		return nil, err
	}
	s.aud.Event(ctx, audit.ActionIPAllocated, userID, networkID, map[string]any{"ip": alloc.IP})
	return alloc, nil
}

// ListAllocations returns all allocations for a network (member must be approved; admin/owner not yet distinguished here)
func (s *IPAMService) ListAllocations(ctx context.Context, networkID, userID string) ([]*domain.IPAllocation, error) {
	// network existence
	if _, err := s.networks.GetByID(ctx, networkID); err != nil {
		return nil, err
	}
	if s.members != nil {
		m, err := s.members.Get(ctx, networkID, userID)
		if err != nil {
			return nil, domain.NewError(domain.ErrNotAuthorized, "Membership required", nil)
		}
		if m.Status != domain.StatusApproved {
			return nil, domain.NewError(domain.ErrNotAuthorized, "Membership not approved", map[string]string{"status": string(m.Status)})
		}
	}
	return s.ipam.List(ctx, networkID)
}

// ReleaseIP releases a user's allocation (idempotent). If user has no allocation it still succeeds.
func (s *IPAMService) ReleaseIP(ctx context.Context, networkID, userID string) error {
	// ensure network exists
	if _, err := s.networks.GetByID(ctx, networkID); err != nil {
		return err
	}
	// Enforce membership approval before allowing release (prevents probing existence of allocations by outsiders)
	if s.members != nil {
		m, err := s.members.Get(ctx, networkID, userID)
		if err != nil {
			return domain.NewError(domain.ErrNotAuthorized, "Membership required", nil)
		}
		if m.Status != domain.StatusApproved {
			return domain.NewError(domain.ErrNotAuthorized, "Membership not approved", map[string]string{"status": string(m.Status)})
		}
	}
	if err := s.ipam.Release(ctx, networkID, userID); err != nil {
		return err
	}
	s.aud.Event(ctx, audit.ActionIPReleased, userID, networkID, nil)
	return nil
}

// ReleaseIPForActor allows an admin/owner (actor) to release another target user's allocation.
// Rules:
// - Actor must have approved membership with role admin or owner.
// - Target user must have approved membership (if target membership missing we treat as not authorized to avoid probing user existence).
// - Operation is idempotent: if target has no allocation, succeeds silently.
// - Self release by actor should prefer ReleaseIP, but still allowed here.
func (s *IPAMService) ReleaseIPForActor(ctx context.Context, networkID, actorUserID, targetUserID string) error {
	// ensure network exists
	if _, err := s.networks.GetByID(ctx, networkID); err != nil {
		return err
	}
	if s.members == nil {
		return domain.NewError(domain.ErrInternalServer, "Membership repository unavailable", nil)
	}
	// fetch actor membership
	actorM, err := s.members.Get(ctx, networkID, actorUserID)
	if err != nil {
		return domain.NewError(domain.ErrNotAuthorized, "Actor membership required", nil)
	}
	if actorM.Status != domain.StatusApproved {
		return domain.NewError(domain.ErrNotAuthorized, "Actor membership not approved", map[string]string{"status": string(actorM.Status)})
	}
	if !rbac.CanReleaseOtherIP(actorM.Role) {
		return domain.NewError(domain.ErrNotAuthorized, "Admin or owner role required", map[string]string{"role": string(actorM.Role)})
	}
	// target membership (hide absence distinctness)
	targetM, err := s.members.Get(ctx, networkID, targetUserID)
	if err != nil {
		return domain.NewError(domain.ErrNotAuthorized, "Target membership required", nil)
	}
	if targetM.Status != domain.StatusApproved {
		return domain.NewError(domain.ErrNotAuthorized, "Target membership not approved", map[string]string{"status": string(targetM.Status)})
	}
	if err := s.ipam.Release(ctx, networkID, targetUserID); err != nil {
		return err
	}
	s.aud.Event(ctx, audit.ActionIPReleased, actorUserID, networkID, map[string]any{"released_for": targetUserID})
	return nil
}
