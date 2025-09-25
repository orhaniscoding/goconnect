package service

import (
	"context"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

func setup() (*MembershipService, string, string) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()

	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)
	// seed network
	net := &domain.Network{ID: "net1", TenantID: "t1", Name: "N1", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.1.0.0/24", CreatedBy: "owner"}
	_ = nrepo.Create(context.Background(), net)
	return svc, net.ID, "userA"
}

func TestOpenJoin(t *testing.T) {
	svc, nid, uid := setup()
	m, jr, err := svc.JoinNetwork(context.Background(), nid, uid, domain.GenerateIdempotencyKey())
	if err != nil || m == nil || jr != nil {
		t.Fatalf("expected direct membership, got m=%v jr=%v err=%v", m, jr, err)
	}
}

func TestApprovalFlow(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)
	net := &domain.Network{ID: "net2", TenantID: "t1", Name: "N2", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.2.0.0/24", CreatedBy: "owner"}
	_ = nrepo.Create(context.Background(), net)

	// non-admin tries to approve -> forbidden
	_, _, _ = svc.JoinNetwork(context.Background(), net.ID, "userX", domain.GenerateIdempotencyKey())
	if _, err := svc.Approve(context.Background(), net.ID, "userX", "not-admin"); err == nil {
		t.Fatalf("expected forbidden for non-admin approve")
	}
}

func TestDoubleJoinGuard(t *testing.T) {
	svc, nid, uid := setup()
	key := domain.GenerateIdempotencyKey()
	_, _, err1 := svc.JoinNetwork(context.Background(), nid, uid, key)
	_, _, err2 := svc.JoinNetwork(context.Background(), nid, uid, key)
	if err2 != nil && err1 == nil {
		t.Fatalf("idempotent second join should not error: %v", err2)
	}
}
