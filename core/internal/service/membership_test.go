package service

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// ==================== JoinNetwork Additional Tests ====================

func TestJoinNetwork_NetworkNotFound(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	_, _, err := svc.JoinNetwork(context.Background(), "non-existent-network", "user-1", "t1", domain.GenerateIdempotencyKey())
	if err == nil {
		t.Fatal("expected error for non-existent network")
	}
}

func TestJoinNetwork_BannedUser(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	// Create network
	net := &domain.Network{
		ID:         "net-banned-test",
		TenantID:   "t1",
		Name:       "Banned Test",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.80.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)

	// Add admin
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin", domain.RoleAdmin, time.Now())

	// Add user and ban them
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "banned-user", domain.RoleMember, time.Now())
	_ = mrepo.SetStatus(context.Background(), net.ID, "banned-user", domain.StatusBanned)

	// Banned user tries to rejoin
	_, _, err := svc.JoinNetwork(context.Background(), net.ID, "banned-user", "t1", domain.GenerateIdempotencyKey())
	if err == nil {
		t.Fatal("expected error for banned user trying to rejoin")
	}
	if derr, ok := err.(*domain.Error); ok {
		if derr.Code != domain.ErrUserBanned {
			t.Errorf("expected ErrUserBanned, got %s", derr.Code)
		}
	}
}

func TestJoinNetwork_PrivateNetwork(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	// Create private network without a name (simulating unknown network)
	net := &domain.Network{
		ID:         "net-private-test",
		TenantID:   "t1",
		Name:       "", // Empty name simulates private network where name is unknown
		Visibility: domain.NetworkVisibilityPrivate,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.81.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)

	// Try to join private network
	_, _, err := svc.JoinNetwork(context.Background(), net.ID, "user-1", "t1", domain.GenerateIdempotencyKey())
	if err == nil {
		t.Fatal("expected error for private network with empty name")
	}
	if derr, ok := err.(*domain.Error); ok {
		if derr.Code != domain.ErrNetworkPrivate {
			t.Errorf("expected ErrNetworkPrivate, got %s", derr.Code)
		}
	}
}

func TestJoinNetwork_MissingIdempotencyKey(t *testing.T) {
	svc, nid, uid := setup()

	// Try to join without idempotency key
	_, _, err := svc.JoinNetwork(context.Background(), nid, uid+"_new", "t1", "")
	if err == nil {
		t.Fatal("expected error for missing idempotency key")
	}
	if derr, ok := err.(*domain.Error); ok {
		if derr.Code != domain.ErrInvalidRequest {
			t.Errorf("expected ErrInvalidRequest, got %s", derr.Code)
		}
	}
}

func TestJoinNetwork_InviteOnlyPolicy(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	// Create invite-only network
	net := &domain.Network{
		ID:         "net-invite-only",
		TenantID:   "t1",
		Name:       "Invite Only Network",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyInvite,
		CIDR:       "10.82.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)

	// Try to join invite-only network
	_, _, err := svc.JoinNetwork(context.Background(), net.ID, "user-1", "t1", domain.GenerateIdempotencyKey())
	if err == nil {
		t.Fatal("expected error for invite-only network")
	}
	// Any error is acceptable - the network requires invite
}

func TestJoinNetwork_ApprovalPolicyCreatesJoinRequest(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	// Create approval-required network
	net := &domain.Network{
		ID:         "net-approval-jr",
		TenantID:   "t1",
		Name:       "Approval JR Test",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyApproval,
		CIDR:       "10.83.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)

	// Try to join
	membership, joinRequest, err := svc.JoinNetwork(context.Background(), net.ID, "user-1", "t1", domain.GenerateIdempotencyKey())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return nil membership and a join request
	if membership != nil {
		t.Error("expected nil membership for approval-required network")
	}
	if joinRequest == nil {
		t.Fatal("expected join request for approval-required network")
	}
	if joinRequest.UserID != "user-1" {
		t.Errorf("expected userID user-1, got %s", joinRequest.UserID)
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

// ==================== SetAuditor Tests ====================

func TestMembershipService_SetAuditor(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	service := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	t.Run("Set Auditor With Nil", func(t *testing.T) {
		// Setting nil auditor should be a no-op (not panic)
		service.SetAuditor(nil)
	})

	t.Run("Set Auditor With Valid Auditor", func(t *testing.T) {
		mockAuditor := auditorFunc(func(ctx context.Context, tenantID, action, actor, object string, details map[string]any) {})
		service.SetAuditor(mockAuditor)
		// Should not panic
	})
}

// ==================== SetNotifier Tests ====================

func TestMembershipService_SetNotifier(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	service := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	t.Run("Set Notifier With Nil", func(t *testing.T) {
		// Setting nil should be a no-op
		service.SetNotifier(nil)
	})

	t.Run("Set Notifier With Valid Notifier", func(t *testing.T) {
		mockNotifier := noopNotifier{}
		service.SetNotifier(mockNotifier)
		// Should not panic
	})
}

// ==================== SetPeerProvisioning Tests ====================

func TestMembershipService_SetPeerProvisioning(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	service := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	t.Run("Set PeerProvisioning With Nil", func(t *testing.T) {
		// Setting nil should be a no-op
		service.SetPeerProvisioning(nil)
	})

	t.Run("Set PeerProvisioning With Valid Service", func(t *testing.T) {
		peerRepo := repository.NewInMemoryPeerRepository()
		deviceRepo := repository.NewInMemoryDeviceRepository()
		ipamRepo := repository.NewInMemoryIPAM()
		peerProvService := NewPeerProvisioningService(peerRepo, deviceRepo, nrepo, mrepo, ipamRepo)
		service.SetPeerProvisioning(peerProvService)
		// Should not panic
	})
}

// ==================== ListJoinRequests Tests ====================

func TestMembershipService_ListJoinRequests(t *testing.T) {
	ctx := context.Background()
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	service := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	// Create a network
	network := &domain.Network{
		ID:         "net-jr-1",
		TenantID:   "tenant-1",
		Name:       "Test Network",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyApproval,
		CIDR:       "10.20.0.0/24",
		CreatedBy:  "owner-1",
	}
	_ = nrepo.Create(ctx, network)

	t.Run("List Join Requests Success", func(t *testing.T) {
		// ListJoinRequests takes networkID and tenantID
		requests, err := service.ListJoinRequests(ctx, "net-jr-1", "tenant-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// May be empty since no pending requests
		_ = requests
	})

	t.Run("List Join Requests Wrong Tenant", func(t *testing.T) {
		// Should fail with wrong tenantID
		_, err := service.ListJoinRequests(ctx, "net-jr-1", "wrong-tenant")
		if err == nil {
			t.Fatal("expected error for wrong tenant")
		}
	})
}

// ==================== IsMember Tests ====================

func TestMembershipService_IsMember(t *testing.T) {
	ctx := context.Background()
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	service := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	// Create a network
	network := &domain.Network{
		ID:         "net-im-1",
		TenantID:   "tenant-1",
		Name:       "Test Network",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.21.0.0/24",
		CreatedBy:  "owner-1",
	}
	_ = nrepo.Create(ctx, network)

	// Add a member using UpsertApproved
	_, _ = mrepo.UpsertApproved(ctx, "net-im-1", "member-1", domain.RoleMember, time.Now())

	t.Run("User Is Member", func(t *testing.T) {
		// IsMember takes networkID, userID
		isMember, err := service.IsMember(ctx, "net-im-1", "member-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isMember {
			t.Error("expected user to be a member")
		}
	})

	t.Run("User Is Not Member", func(t *testing.T) {
		// IsMember returns error for non-existent membership
		isMember, err := service.IsMember(ctx, "net-im-1", "non-member")
		// According to the implementation, it returns error if membership not found
		if err != nil {
			// Expected: not a member (error returned)
			return
		}
		if isMember {
			t.Error("expected user to not be a member")
		}
	})
}

// Test Approve: Admin approves a pending join request (success path)
func TestApprove_Success(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	// Create approval-required network
	net := &domain.Network{
		ID:         "net-approve-success",
		TenantID:   "t1",
		Name:       "ApproveNet",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyApproval,
		CIDR:       "10.30.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)

	// Admin joins directly as admin
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin", domain.RoleAdmin, time.Now())

	// Regular user requests to join
	_, jr, err := svc.JoinNetwork(context.Background(), net.ID, "userToApprove", "t1", domain.GenerateIdempotencyKey())
	if err != nil {
		t.Fatalf("join request failed: %v", err)
	}
	if jr == nil {
		t.Fatal("expected join request for approval policy")
	}

	// Admin approves the request
	membership, err := svc.Approve(context.Background(), net.ID, "userToApprove", "admin", "t1")
	if err != nil {
		t.Fatalf("expected approve to succeed, got error: %v", err)
	}

	if membership == nil {
		t.Fatal("expected membership to be returned")
	}
	if membership.Role != domain.RoleMember {
		t.Errorf("expected role %s, got %s", domain.RoleMember, membership.Role)
	}
	if membership.UserID != "userToApprove" {
		t.Errorf("expected userID userToApprove, got %s", membership.UserID)
	}

	// Verify user is now a member
	_, err = mrepo.Get(context.Background(), net.ID, "userToApprove")
	if err != nil {
		t.Fatalf("expected user to be a member after approval: %v", err)
	}

	// Verify join request was decided
	_, err = jrepo.GetPending(context.Background(), net.ID, "userToApprove")
	if err == nil {
		t.Fatal("expected no pending join request after approval")
	}
}

// Test Approve: Wrong tenant
func TestApprove_WrongTenant(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	net := &domain.Network{
		ID:         "net-approve-tenant",
		TenantID:   "t1",
		Name:       "TenantNet",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyApproval,
		CIDR:       "10.31.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)

	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin", domain.RoleAdmin, time.Now())

	_, _, _ = svc.JoinNetwork(context.Background(), net.ID, "userX", "t1", domain.GenerateIdempotencyKey())

	// Try to approve with wrong tenant
	_, err := svc.Approve(context.Background(), net.ID, "userX", "admin", "wrong-tenant")
	if err == nil {
		t.Fatal("expected error when approving with wrong tenant")
	}

	derr, ok := err.(*domain.Error)
	if !ok {
		t.Fatalf("expected domain.Error, got %T", err)
	}
	if derr.Code != domain.ErrNotFound {
		t.Errorf("expected error code %s, got %s", domain.ErrNotFound, derr.Code)
	}
}

// Test Approve: No pending join request
func TestApprove_NoPendingJoinRequest(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	net := &domain.Network{
		ID:         "net-approve-no-pending",
		TenantID:   "t1",
		Name:       "NoPendingNet",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyApproval,
		CIDR:       "10.32.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)

	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin", domain.RoleAdmin, time.Now())

	// Try to approve user who never requested to join
	_, err := svc.Approve(context.Background(), net.ID, "never-requested", "admin", "t1")
	if err == nil {
		t.Fatal("expected error when approving user without pending request")
	}
}

// ==================== MemberJoined/MemberLeft Notifier Tests ====================

// MockMembershipNotifier tracks calls to MemberJoined and MemberLeft
type MockMembershipNotifier struct {
	joinedCalls []struct{ NetworkID, UserID string }
	leftCalls   []struct{ NetworkID, UserID string }
}

func (m *MockMembershipNotifier) MemberJoined(networkID, userID string) {
	m.joinedCalls = append(m.joinedCalls, struct{ NetworkID, UserID string }{networkID, userID})
}

func (m *MockMembershipNotifier) MemberLeft(networkID, userID string) {
	m.leftCalls = append(m.leftCalls, struct{ NetworkID, UserID string }{networkID, userID})
}

func TestMembershipService_JoinNetworkTriggersMemberJoined(t *testing.T) {
	ctx := context.Background()
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	mockNotifier := &MockMembershipNotifier{}
	svc.SetNotifier(mockNotifier)

	// Create open network
	net := &domain.Network{
		ID:         "net-notify-join",
		TenantID:   "t1",
		Name:       "NotifyJoinNet",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.50.0.0/24",
		CreatedBy:  "owner",
	}
	_ = nrepo.Create(ctx, net)

	t.Run("JoinNetwork with open policy triggers MemberJoined", func(t *testing.T) {
		mockNotifier.joinedCalls = nil

		m, jr, err := svc.JoinNetwork(ctx, net.ID, "user-join-notify", "t1", domain.GenerateIdempotencyKey())

		if err != nil {
			t.Fatalf("expected join to succeed, got error: %v", err)
		}
		if m == nil {
			t.Fatal("expected membership for open policy")
		}
		if jr != nil {
			t.Fatal("expected no join request for open policy")
		}

		// Verify MemberJoined was called
		if len(mockNotifier.joinedCalls) == 0 {
			t.Fatal("expected MemberJoined to be called")
		}
		found := false
		for _, call := range mockNotifier.joinedCalls {
			if call.NetworkID == net.ID && call.UserID == "user-join-notify" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected MemberJoined to be called with networkID=%s, userID=%s", net.ID, "user-join-notify")
		}
	})
}

func TestMembershipService_ApproveTriggersMemberJoined(t *testing.T) {
	ctx := context.Background()
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	mockNotifier := &MockMembershipNotifier{}
	svc.SetNotifier(mockNotifier)

	// Create approval-required network
	net := &domain.Network{
		ID:         "net-notify-approve",
		TenantID:   "t1",
		Name:       "NotifyApproveNet",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyApproval,
		CIDR:       "10.51.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(ctx, net)

	// Admin joins directly
	_, _ = mrepo.UpsertApproved(ctx, net.ID, "admin", domain.RoleAdmin, time.Now())

	// User requests to join
	_, jr, err := svc.JoinNetwork(ctx, net.ID, "user-to-approve", "t1", domain.GenerateIdempotencyKey())
	if err != nil {
		t.Fatalf("join request failed: %v", err)
	}
	if jr == nil {
		t.Fatal("expected join request for approval policy")
	}

	t.Run("Approve triggers MemberJoined", func(t *testing.T) {
		mockNotifier.joinedCalls = nil

		_, err := svc.Approve(ctx, net.ID, "user-to-approve", "admin", "t1")
		if err != nil {
			t.Fatalf("expected approve to succeed, got error: %v", err)
		}

		// Verify MemberJoined was called
		if len(mockNotifier.joinedCalls) == 0 {
			t.Fatal("expected MemberJoined to be called after approval")
		}
		found := false
		for _, call := range mockNotifier.joinedCalls {
			if call.NetworkID == net.ID && call.UserID == "user-to-approve" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected MemberJoined to be called with networkID=%s, userID=%s", net.ID, "user-to-approve")
		}
	})
}

func TestMembershipService_KickTriggersMemberLeft(t *testing.T) {
	ctx := context.Background()
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	mockNotifier := &MockMembershipNotifier{}
	svc.SetNotifier(mockNotifier)

	// Create network
	net := &domain.Network{
		ID:         "net-notify-kick",
		TenantID:   "t1",
		Name:       "NotifyKickNet",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.52.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(ctx, net)

	// Admin is a member
	_, _ = mrepo.UpsertApproved(ctx, net.ID, "admin", domain.RoleAdmin, time.Now())

	// User joins
	_, _, err := svc.JoinNetwork(ctx, net.ID, "user-to-kick", "t1", domain.GenerateIdempotencyKey())
	if err != nil {
		t.Fatalf("join failed: %v", err)
	}

	t.Run("Kick triggers MemberLeft", func(t *testing.T) {
		mockNotifier.leftCalls = nil

		err := svc.Kick(ctx, net.ID, "user-to-kick", "admin", "t1")
		if err != nil {
			t.Fatalf("expected kick to succeed, got error: %v", err)
		}

		// Verify MemberLeft was called
		if len(mockNotifier.leftCalls) == 0 {
			t.Fatal("expected MemberLeft to be called after kick")
		}
		found := false
		for _, call := range mockNotifier.leftCalls {
			if call.NetworkID == net.ID && call.UserID == "user-to-kick" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected MemberLeft to be called with networkID=%s, userID=%s", net.ID, "user-to-kick")
		}
	})
}

func TestNoopNotifier(t *testing.T) {
	// Test that noopNotifier doesn't panic
	notifier := noopNotifier{}

	t.Run("MemberJoined does not panic", func(t *testing.T) {
		// Should not panic
		notifier.MemberJoined("network-id", "user-id")
	})

	t.Run("MemberLeft does not panic", func(t *testing.T) {
		// Should not panic
		notifier.MemberLeft("network-id", "user-id")
	})
}

// ==================== KICK EDGE CASE TESTS ====================

func TestKick_NetworkNotFound(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	err := svc.Kick(context.Background(), "non-existent-network", "some-user", "admin", "t1")
	require.Error(t, err)
}

func TestKick_WrongTenant(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	net := &domain.Network{
		ID:         "net-kick-tenant",
		TenantID:   "t1",
		Name:       "Net",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.10.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin", domain.RoleAdmin, time.Now())

	// Try to kick with wrong tenant
	err := svc.Kick(context.Background(), net.ID, "some-user", "admin", "wrong-tenant")
	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

// ==================== LIST MEMBERS EDGE CASE TESTS ====================

func TestListMembers_NetworkNotFound(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	_, _, err := svc.ListMembers(context.Background(), "non-existent", "", "t1", 10, "")
	require.Error(t, err)
}

func TestListMembers_WrongTenant(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	net := &domain.Network{
		ID:         "net-list-tenant",
		TenantID:   "t1",
		Name:       "Net",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.11.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)

	// Try to list with wrong tenant
	_, _, err := svc.ListMembers(context.Background(), net.ID, "", "wrong-tenant", 10, "")
	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

// ==================== BAN EDGE CASE TESTS ====================

func TestBan_NetworkNotFound(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	err := svc.Ban(context.Background(), "non-existent-network", "some-user", "admin", "t1")
	require.Error(t, err)
}

func TestBan_WrongTenant(t *testing.T) {
	nrepo := repository.NewInMemoryNetworkRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	svc := NewMembershipService(nrepo, mrepo, jrepo, irepo)

	net := &domain.Network{
		ID:         "net-ban-tenant",
		TenantID:   "t1",
		Name:       "Net",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.12.0.0/24",
		CreatedBy:  "admin",
	}
	_ = nrepo.Create(context.Background(), net)
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin", domain.RoleAdmin, time.Now())

	// Try to ban with wrong tenant
	err := svc.Ban(context.Background(), net.ID, "some-user", "admin", "wrong-tenant")
	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}
