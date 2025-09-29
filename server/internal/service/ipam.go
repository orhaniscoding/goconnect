package service

import (
	"context"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
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
	s.aud.Event(ctx, "IP_ALLOCATED", userID, networkID, map[string]any{"ip": alloc.IP})
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
