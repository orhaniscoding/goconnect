package service

import (
	"context"
	"testing"
	"time"

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
	m, jr, err := svc.JoinNetwork(context.Background(), nid, uid, "t1", domain.GenerateIdempotencyKey())
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
	_, _, _ = svc.JoinNetwork(context.Background(), net.ID, "userX", "t1", domain.GenerateIdempotencyKey())
	if _, err := svc.Approve(context.Background(), net.ID, "userX", "not-admin", "t1"); err == nil {
		t.Fatalf("expected forbidden for non-admin approve")
	}
}

func TestDoubleJoinGuard(t *testing.T) {
	svc, nid, uid := setup()
	key := domain.GenerateIdempotencyKey()
	_, _, err1 := svc.JoinNetwork(context.Background(), nid, uid, "t1", key)
	_, _, err2 := svc.JoinNetwork(context.Background(), nid, uid, "t1", key)
	if err2 != nil && err1 == nil {
		t.Fatalf("idempotent second join should not error: %v", err2)
	}
}

func TestMembershipTenantIsolation(t *testing.T) {
	svc, nid, uid := setup()
	// Try to join with wrong tenant
	_, _, err := svc.JoinNetwork(context.Background(), nid, uid, "wrong-tenant", domain.GenerateIdempotencyKey())
	if err == nil {
		t.Fatal("expected error when joining with wrong tenant")
	}
	if derr, ok := err.(*domain.Error); !ok || derr.Code != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// Test Deny: Admin denies a pending join request
func TestDeny_Success(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	// Create approval-required network
	net := &domain.Network{
		ID:         "net-approval",
		TenantID:   "t1",
		Name:       "ApprovalNet",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyApproval,
		CIDR:       "10.3.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)

	// Admin joins directly
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin", domain.RoleAdmin, time.Now())

	// Regular user requests to join
	_, _, err := svc.JoinNetwork(context.Background(), net.ID, "userX", "t1", domain.GenerateIdempotencyKey())
	if err != nil {
		t.Fatalf("join request failed: %v", err)
	}

	// Admin denies the request
	err = svc.Deny(context.Background(), net.ID, "userX", "admin", "t1")
	if err != nil {
		t.Fatalf("expected deny to succeed, got error: %v", err)
	}

	// Verify join request was decided
	_, err = jrepo.GetPending(context.Background(), net.ID, "userX")
	if err == nil {
		t.Fatal("expected no pending join request after deny")
	}
}

// Test Deny: Non-admin tries to deny
func TestDeny_Unauthorized(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	net := &domain.Network{
		ID:         "net-deny-unauth",
		TenantID:   "t1",
		Name:       "Net",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyApproval,
		CIDR:       "10.4.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)

	// User requests to join
	_, _, _ = svc.JoinNetwork(context.Background(), net.ID, "userY", "t1", domain.GenerateIdempotencyKey())

	// Another regular user (not admin) tries to deny
	err := svc.Deny(context.Background(), net.ID, "userY", "non-admin", "t1")
	if err == nil {
		t.Fatal("expected error when non-admin tries to deny")
	}

	derr, ok := err.(*domain.Error)
	if !ok {
		t.Fatalf("expected domain.Error, got %T", err)
	}
	if derr.Code != domain.ErrNotAuthorized {
		t.Errorf("expected error code %s, got %s", domain.ErrNotAuthorized, derr.Code)
	}
}

// Test Kick: Admin removes a member
func TestKick_Success(t *testing.T) {
	svc, nid, _ := setup()

	// User joins the network
	userID := "member-to-kick"
	_, _, err := svc.JoinNetwork(context.Background(), nid, userID, "t1", domain.GenerateIdempotencyKey())
	if err != nil {
		t.Fatalf("join failed: %v", err)
	}

	// Create admin membership
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc = NewMembershipService(nrepo, mrepo, jrepo, irepo)

	net := &domain.Network{
		ID:         "net-kick",
		TenantID:   "t1",
		Name:       "KickNet",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.5.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin", domain.RoleAdmin, time.Now())
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "member-to-kick", domain.RoleMember, time.Now())

	// Admin kicks the member
	err = svc.Kick(context.Background(), net.ID, "member-to-kick", "admin", "t1")
	if err != nil {
		t.Fatalf("expected kick to succeed, got error: %v", err)
	}

	// Verify member is removed
	_, err = mrepo.Get(context.Background(), net.ID, "member-to-kick")
	if err == nil {
		t.Fatal("expected member to be removed after kick")
	}
}

// Test Kick: Non-admin tries to kick
func TestKick_Unauthorized(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	net := &domain.Network{
		ID:         "net-kick-unauth",
		TenantID:   "t1",
		Name:       "Net",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.6.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "memberA", domain.RoleMember, time.Now())
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "memberB", domain.RoleMember, time.Now())

	// Non-admin tries to kick another member
	err := svc.Kick(context.Background(), net.ID, "memberB", "memberA", "t1")
	if err == nil {
		t.Fatal("expected error when non-admin tries to kick")
	}

	derr, ok := err.(*domain.Error)
	if !ok {
		t.Fatalf("expected domain.Error, got %T", err)
	}
	if derr.Code != domain.ErrNotAuthorized {
		t.Errorf("expected error code %s, got %s", domain.ErrNotAuthorized, derr.Code)
	}
}

// Test Ban: Admin bans a member
func TestBan_Success(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	net := &domain.Network{
		ID:         "net-ban",
		TenantID:   "t1",
		Name:       "BanNet",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.7.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin", domain.RoleAdmin, time.Now())
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "bad-user", domain.RoleMember, time.Now())

	// Admin bans the member
	err := svc.Ban(context.Background(), net.ID, "bad-user", "admin", "t1")
	if err != nil {
		t.Fatalf("expected ban to succeed, got error: %v", err)
	}

	// Verify member status is banned
	member, err := mrepo.Get(context.Background(), net.ID, "bad-user")
	if err != nil {
		t.Fatalf("expected to find banned member: %v", err)
	}
	if member.Status != domain.StatusBanned {
		t.Errorf("expected status %s, got %s", domain.StatusBanned, member.Status)
	}
}

// Test Ban: Non-admin tries to ban
func TestBan_Unauthorized(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	net := &domain.Network{
		ID:         "net-ban-unauth",
		TenantID:   "t1",
		Name:       "Net",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.8.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "memberX", domain.RoleMember, time.Now())
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "memberY", domain.RoleMember, time.Now())

	// Non-admin tries to ban another member
	err := svc.Ban(context.Background(), net.ID, "memberY", "memberX", "t1")
	if err == nil {
		t.Fatal("expected error when non-admin tries to ban")
	}

	derr, ok := err.(*domain.Error)
	if !ok {
		t.Fatalf("expected domain.Error, got %T", err)
	}
	if derr.Code != domain.ErrNotAuthorized {
		t.Errorf("expected error code %s, got %s", domain.ErrNotAuthorized, derr.Code)
	}
}

// Test ListMembers: List all approved members
func TestListMembers_Success(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	net := &domain.Network{
		ID:         "net-list",
		TenantID:   "t1",
		Name:       "ListNet",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.9.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)

	// Add multiple members
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user1", domain.RoleMember, time.Now())
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user2", domain.RoleMember, time.Now())
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user3", domain.RoleMember, time.Now())

	// List all approved members
	members, cursor, err := svc.ListMembers(context.Background(), net.ID, string(domain.StatusApproved), "t1", 10, "")
	if err != nil {
		t.Fatalf("expected list to succeed, got error: %v", err)
	}

	if len(members) != 3 {
		t.Errorf("expected 3 members, got %d", len(members))
	}
	if cursor != "" {
		t.Errorf("expected empty cursor, got %s", cursor)
	}
}

// Test ListMembers: Filter by banned status
func TestListMembers_FilterBanned(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	net := &domain.Network{
		ID:         "net-banned",
		TenantID:   "t1",
		Name:       "BannedNet",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.10.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin", domain.RoleAdmin, time.Now())

	// Add members and ban one
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user1", domain.RoleMember, time.Now())
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user2", domain.RoleMember, time.Now())
	_ = svc.Ban(context.Background(), net.ID, "user2", "admin", "t1")

	// List only banned members
	members, _, err := svc.ListMembers(context.Background(), net.ID, string(domain.StatusBanned), "t1", 10, "")
	if err != nil {
		t.Fatalf("expected list to succeed, got error: %v", err)
	}

	if len(members) != 1 {
		t.Errorf("expected 1 banned member, got %d", len(members))
	}
	if len(members) > 0 && members[0].Status != domain.StatusBanned {
		t.Errorf("expected banned status, got %s", members[0].Status)
	}
}
