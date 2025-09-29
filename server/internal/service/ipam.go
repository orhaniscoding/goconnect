package service

import (
	"context"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// IPAMService provides IP allocation logic on top of repositories.
type IPAMService struct {
	networks repository.NetworkRepository
	ipam     repository.IPAMRepository
	aud      Auditor
}

func NewIPAMService(n repository.NetworkRepository, ip repository.IPAMRepository) *IPAMService {
	return &IPAMService{networks: n, ipam: ip, aud: noopAuditor}
}

// SetAuditor wires a real auditor.
func (s *IPAMService) SetAuditor(a Auditor) {
	if a != nil {
		s.aud = a
	}
}

// AllocateIP returns existing allocation or assigns the next available IP.
func (s *IPAMService) AllocateIP(ctx context.Context, networkID, userID string) (*domain.IPAllocation, error) {
	netw, err := s.networks.GetByID(ctx, networkID)
	if err != nil {
		return nil, err
	}
	alloc, err := s.ipam.GetOrAllocate(ctx, networkID, userID, netw.CIDR)
	if err != nil {
		return nil, err
	}
	s.aud.Event(ctx, "IP_ALLOCATED", userID, networkID, map[string]any{"ip": alloc.IP})
	return alloc, nil
}
