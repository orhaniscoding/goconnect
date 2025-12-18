package service

import (
"errors"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

func setupIPAMTestNetwork(t *testing.T, cidr string) (context.Context, *IPAMService, repository.NetworkRepository, repository.MembershipRepository, string) {
	ctx := context.Background()
	netRepo := repository.NewInMemoryNetworkRepository()
	mRepo := repository.NewInMemoryMembershipRepository()
	ipRepo := repository.NewInMemoryIPAM()
	svc := NewIPAMService(netRepo, mRepo, ipRepo)
	// create network (compact form after merge)
	n := &domain.Network{ID: domain.GenerateNetworkID(), TenantID: "t1", Name: "n1", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: cidr, CreatedBy: "u_admin"}
	if err := netRepo.Create(ctx, n); err != nil {
		t.Fatalf("create network: %v", err)
	}
	return ctx, svc, netRepo, mRepo, n.ID
}

func TestIPAMSequentialAllocation(t *testing.T) {
	ctx, svc, _, mRepo, netID := setupIPAMTestNetwork(t, "10.10.0.0/30") // /30 => 4 addresses => usable: 2
	_, _ = mRepo.UpsertApproved(ctx, netID, "user1", domain.RoleMember, time.Now())
	_, _ = mRepo.UpsertApproved(ctx, netID, "user2", domain.RoleMember, time.Now())
	a1, err := svc.AllocateIP(ctx, netID, "user1", "t1")
	if err != nil {
		t.Fatalf("alloc1: %v", err)
	}
	if a1.IP != "10.10.0.1" {
		t.Fatalf("expected first usable 10.10.0.1 got %s", a1.IP)
	}
	a2, err := svc.AllocateIP(ctx, netID, "user2", "t1")
	if err != nil {
		t.Fatalf("alloc2: %v", err)
	}
	if a2.IP != "10.10.0.2" {
		t.Fatalf("expected second usable 10.10.0.2 got %s", a2.IP)
	}
	// Now exhausted
	_, _ = mRepo.UpsertApproved(ctx, netID, "user3", domain.RoleMember, time.Now())
	_, err = svc.AllocateIP(ctx, netID, "user3", "t1")
	if err == nil {
		t.Fatalf("expected exhaustion error")
	}
	var derr *domain.Error; ok := errors.As(err, &derr)
	if !ok || derr.Code != domain.ErrIPExhausted {
		t.Fatalf("expected ERR_IP_EXHAUSTED got %+v", err)
	}
}

func TestIPAMSameUserStable(t *testing.T) {
	ctx, svc, _, mRepo, netID := setupIPAMTestNetwork(t, "10.20.0.0/29")
	_, _ = mRepo.UpsertApproved(ctx, netID, "userX", domain.RoleMember, time.Now())
	a1, err := svc.AllocateIP(ctx, netID, "userX", "t1")
	if err != nil {
		t.Fatalf("alloc1: %v", err)
	}
	a2, err := svc.AllocateIP(ctx, netID, "userX", "t1")
	if err != nil {
		t.Fatalf("alloc2: %v", err)
	}
	if a1.IP != a2.IP {
		t.Fatalf("expected stable allocation, got %s vs %s", a1.IP, a2.IP)
	}
}

func TestIPAMInvalidNetwork(t *testing.T) {
	ctx := context.Background()
	netRepo := repository.NewInMemoryNetworkRepository()
	mRepo := repository.NewInMemoryMembershipRepository()
	ipRepo := repository.NewInMemoryIPAM()
	svc := NewIPAMService(netRepo, mRepo, ipRepo)
	_, err := svc.AllocateIP(ctx, "missing", "user1", "t1")
	if err == nil {
		t.Fatalf("expected error for missing network")
	}
	var derr *domain.Error; ok := errors.As(err, &derr)
	if !ok || derr.Code != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound got %+v", err)
	}
}

func TestIPAMNonMemberDenied(t *testing.T) {
	ctx, svc, _, _, netID := setupIPAMTestNetwork(t, "10.70.0.0/30")
	_, err := svc.AllocateIP(ctx, netID, "userZ", "t1")
	if err == nil {
		t.Fatalf("expected authorization error for non-member")
	}
	var derr *domain.Error; ok := errors.As(err, &derr)
	if !ok || derr.Code != domain.ErrNotAuthorized {
		t.Fatalf("expected ErrNotAuthorized got %+v", err)
	}
}

func TestIPAMReleaseReuseSameIP(t *testing.T) {
	ctx, svc, _, mRepo, netID := setupIPAMTestNetwork(t, "10.80.0.0/30")
	_, _ = mRepo.UpsertApproved(ctx, netID, "userA", domain.RoleMember, time.Now())
	first, err := svc.AllocateIP(ctx, netID, "userA", "t1")
	if err != nil {
		t.Fatalf("alloc1: %v", err)
	}
	if err := svc.ReleaseIP(ctx, netID, "userA", "t1"); err != nil {
		t.Fatalf("release: %v", err)
	}
	second, err := svc.AllocateIP(ctx, netID, "userA", "t1")
	if err != nil {
		t.Fatalf("alloc2: %v", err)
	}
	if first.IP != second.IP {
		t.Fatalf("expected reuse of freed IP got %s vs %s", second.IP, first.IP)
	}
}

func TestIPAMReleaseIdempotent(t *testing.T) {
	ctx, svc, _, mRepo, netID := setupIPAMTestNetwork(t, "10.81.0.0/30")
	_, _ = mRepo.UpsertApproved(ctx, netID, "userA", domain.RoleMember, time.Now())
	if err := svc.ReleaseIP(ctx, netID, "userA", "t1"); err != nil {
		t.Fatalf("first release: %v", err)
	}
	if err := svc.ReleaseIP(ctx, netID, "userA", "t1"); err != nil {
		t.Fatalf("second release: %v", err)
	}
}

func TestIPAMReleaseNonMemberDenied(t *testing.T) {
	ctx, svc, _, _, netID := setupIPAMTestNetwork(t, "10.82.0.0/30")
	err := svc.ReleaseIP(ctx, netID, "userB", "t1")
	if err == nil {
		t.Fatalf("expected authorization error")
	}
	var derr *domain.Error; ok := errors.As(err, &derr)
	if !ok || derr.Code != domain.ErrNotAuthorized {
		t.Fatalf("expected ErrNotAuthorized got %+v", err)
	}
}

func TestIPAMAdminReleaseOtherUser(t *testing.T) {
	ctx, svc, _, mRepo, netID := setupIPAMTestNetwork(t, "10.90.0.0/29")
	now := time.Now()
	_, _ = mRepo.UpsertApproved(ctx, netID, "admin1", domain.RoleAdmin, now)
	_, _ = mRepo.UpsertApproved(ctx, netID, "user1", domain.RoleMember, now)
	if _, err := svc.AllocateIP(ctx, netID, "user1", "t1"); err != nil {
		t.Fatalf("alloc target: %v", err)
	}
	if err := svc.ReleaseIPForActor(ctx, netID, "admin1", "user1", "t1"); err != nil {
		t.Fatalf("admin release: %v", err)
	}
	a2, err := svc.AllocateIP(ctx, netID, "user1", "t1")
	if err != nil {
		t.Fatalf("realloc: %v", err)
	}
	if a2.IP != "10.90.0.1" {
		t.Fatalf("expected reuse first host got %s", a2.IP)
	}
}

func TestIPAMAdminReleaseForbiddenForMember(t *testing.T) {
	ctx, svc, _, mRepo, netID := setupIPAMTestNetwork(t, "10.91.0.0/29")
	now := time.Now()
	_, _ = mRepo.UpsertApproved(ctx, netID, "member1", domain.RoleMember, now)
	_, _ = mRepo.UpsertApproved(ctx, netID, "user1", domain.RoleMember, now)
	if _, err := svc.AllocateIP(ctx, netID, "user1", "t1"); err != nil {
		t.Fatalf("alloc target: %v", err)
	}
	err := svc.ReleaseIPForActor(ctx, netID, "member1", "user1", "t1")
	if err == nil {
		t.Fatalf("expected not authorized error")
	}
	var derr *domain.Error; ok := errors.As(err, &derr)
	if !ok || derr.Code != domain.ErrNotAuthorized {
		t.Fatalf("expected ErrNotAuthorized got %+v", err)
	}
}

func TestIPAMAdminReleaseIdempotent(t *testing.T) {
	ctx, svc, _, mRepo, netID := setupIPAMTestNetwork(t, "10.92.0.0/29")
	now := time.Now()
	_, _ = mRepo.UpsertApproved(ctx, netID, "admin1", domain.RoleAdmin, now)
	_, _ = mRepo.UpsertApproved(ctx, netID, "user1", domain.RoleMember, now)
	if err := svc.ReleaseIPForActor(ctx, netID, "admin1", "user1", "t1"); err != nil {
		t.Fatalf("first idempotent admin release: %v", err)
	}
	if _, err := svc.AllocateIP(ctx, netID, "user1", "t1"); err != nil {
		t.Fatalf("alloc: %v", err)
	}
	if err := svc.ReleaseIPForActor(ctx, netID, "admin1", "user1", "t1"); err != nil {
		t.Fatalf("release: %v", err)
	}
	if err := svc.ReleaseIPForActor(ctx, netID, "admin1", "user1", "t1"); err != nil {
		t.Fatalf("second release idempotent: %v", err)
	}
}

func TestIPAMConcurrentAllocRelease(t *testing.T) {
	ctx, svc, _, mRepo, netID := setupIPAMTestNetwork(t, "10.93.0.0/26")
	now := time.Now()
	userCount := 30
	for i := 0; i < userCount; i++ {
		uid := fmt.Sprintf("userC%d", i)
		_, _ = mRepo.UpsertApproved(ctx, netID, uid, domain.RoleMember, now)
	}
	var wg sync.WaitGroup
	errs := make(chan error, userCount)
	for i := 0; i < userCount; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			uid := fmt.Sprintf("userC%d", i)
			if _, err := svc.AllocateIP(ctx, netID, uid, "t1"); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatalf("alloc error: %v", e)
		}
	}
	allocs, err := svc.ListAllocations(ctx, netID, "userC0", "t1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	seen := make(map[string]bool)
	for _, a := range allocs {
		if seen[a.IP] {
			t.Fatalf("duplicate ip %s", a.IP)
		}
		seen[a.IP] = true
	}
	errs = make(chan error, userCount)
	for i := 0; i < userCount; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			uid := fmt.Sprintf("userC%d", i)
			if err := svc.ReleaseIP(ctx, netID, uid, "t1"); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatalf("release error: %v", e)
		}
	}
}

func TestIPAMTenantIsolation(t *testing.T) {
	ctx, svc, _, mRepo, netID := setupIPAMTestNetwork(t, "10.94.0.0/29")
	_, _ = mRepo.UpsertApproved(ctx, netID, "user1", domain.RoleMember, time.Now())

	// Try to allocate with wrong tenant
	_, err := svc.AllocateIP(ctx, netID, "user1", "wrong-tenant")
	if err == nil {
		t.Fatalf("expected error for wrong tenant")
	}
	var derr *domain.Error; ok := errors.As(err, &derr)
	if !ok || derr.Code != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound got %+v", err)
	}
}

// ==================== SetAuditor Tests ====================

func TestIPAMService_SetAuditor(t *testing.T) {
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	ipamRepo := repository.NewInMemoryIPAM()
	service := NewIPAMService(networkRepo, membershipRepo, ipamRepo)

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

// ==================== ListAllocations Edge Case Tests ====================

func TestIPAMService_ListAllocations_NetworkNotFound(t *testing.T) {
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	ipamRepo := repository.NewInMemoryIPAM()
	service := NewIPAMService(networkRepo, membershipRepo, ipamRepo)

	_, err := service.ListAllocations(context.Background(), "non-existent", "user1", "t1")
	if err == nil {
		t.Fatal("expected error for non-existent network")
	}
}

func TestIPAMService_ListAllocations_WrongTenant(t *testing.T) {
	ctx, svc, _, mRepo, netID := setupIPAMTestNetwork(t, "10.95.0.0/28")
	_, _ = mRepo.UpsertApproved(ctx, netID, "user1", domain.RoleMember, time.Now())

	_, err := svc.ListAllocations(ctx, netID, "user1", "wrong-tenant")
	if err == nil {
		t.Fatal("expected error for wrong tenant")
	}
	var derr *domain.Error; ok := errors.As(err, &derr)
	if !ok || derr.Code != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound got %+v", err)
	}
}

// ==================== ReleaseIPForActor Edge Case Tests ====================

func TestIPAMService_ReleaseIPForActor_NetworkNotFound(t *testing.T) {
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	ipamRepo := repository.NewInMemoryIPAM()
	service := NewIPAMService(networkRepo, membershipRepo, ipamRepo)

	err := service.ReleaseIPForActor(context.Background(), "non-existent", "admin1", "user1", "t1")
	if err == nil {
		t.Fatal("expected error for non-existent network")
	}
}

func TestIPAMService_ReleaseIPForActor_WrongTenant(t *testing.T) {
	ctx, svc, _, mRepo, netID := setupIPAMTestNetwork(t, "10.96.0.0/28")
	_, _ = mRepo.UpsertApproved(ctx, netID, "admin1", domain.RoleAdmin, time.Now())
	_, _ = mRepo.UpsertApproved(ctx, netID, "user1", domain.RoleMember, time.Now())

	err := svc.ReleaseIPForActor(ctx, netID, "admin1", "user1", "wrong-tenant")
	if err == nil {
		t.Fatal("expected error for wrong tenant")
	}
	var derr *domain.Error; ok := errors.As(err, &derr)
	if !ok || derr.Code != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound got %+v", err)
	}
}

func TestIPAMService_ReleaseIPForActor_ActorNotMember(t *testing.T) {
	ctx, svc, _, _, netID := setupIPAMTestNetwork(t, "10.97.0.0/28")

	err := svc.ReleaseIPForActor(ctx, netID, "non-member", "user1", "t1")
	if err == nil {
		t.Fatal("expected error for actor not member")
	}
	var derr *domain.Error; ok := errors.As(err, &derr)
	if !ok || derr.Code != domain.ErrNotAuthorized {
		t.Fatalf("expected ErrNotAuthorized got %+v", err)
	}
}

func TestIPAMService_ReleaseIPForActor_TargetNotMember(t *testing.T) {
	ctx, svc, _, mRepo, netID := setupIPAMTestNetwork(t, "10.98.0.0/28")
	_, _ = mRepo.UpsertApproved(ctx, netID, "admin1", domain.RoleAdmin, time.Now())

	err := svc.ReleaseIPForActor(ctx, netID, "admin1", "non-member", "t1")
	if err == nil {
		t.Fatal("expected error for target not member")
	}
	var derr *domain.Error; ok := errors.As(err, &derr)
	if !ok || derr.Code != domain.ErrNotAuthorized {
		t.Fatalf("expected ErrNotAuthorized got %+v", err)
	}
}
