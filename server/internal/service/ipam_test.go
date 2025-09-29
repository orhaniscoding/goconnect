package service

import (
    "context"
    "testing"

    "github.com/orhaniscoding/goconnect/server/internal/domain"
    "github.com/orhaniscoding/goconnect/server/internal/repository"
)

func setupIPAMTestNetwork(t *testing.T, cidr string) (context.Context, *IPAMService, repository.NetworkRepository, string) {
    ctx := context.Background()
    netRepo := repository.NewInMemoryNetworkRepository()
    ipRepo := repository.NewInMemoryIPAM()
    svc := NewIPAMService(netRepo, ipRepo)
    // create network
    n := &domain.Network{ID: domain.GenerateNetworkID(), TenantID: "t1", Name: "n1", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: cidr, CreatedBy: "u_admin"}
    if err := netRepo.Create(ctx, n); err != nil { t.Fatalf("create network: %v", err) }
    return ctx, svc, netRepo, n.ID
}

func TestIPAMSequentialAllocation(t *testing.T) {
    ctx, svc, _, netID := setupIPAMTestNetwork(t, "10.10.0.0/30") // /30 => 4 addresses => usable: 2 (10.10.0.1, 10.10.0.2)

    a1, err := svc.AllocateIP(ctx, netID, "user1")
    if err != nil { t.Fatalf("alloc1: %v", err) }
    if a1.IP != "10.10.0.1" { t.Fatalf("expected first usable 10.10.0.1 got %s", a1.IP) }

    a2, err := svc.AllocateIP(ctx, netID, "user2")
    if err != nil { t.Fatalf("alloc2: %v", err) }
    if a2.IP != "10.10.0.2" { t.Fatalf("expected second usable 10.10.0.2 got %s", a2.IP) }

    // Now exhausted
    _, err = svc.AllocateIP(ctx, netID, "user3")
    if err == nil { t.Fatalf("expected exhaustion error") }
    derr, ok := err.(*domain.Error)
    if !ok || derr.Code != domain.ErrIPExhausted { t.Fatalf("expected ERR_IP_EXHAUSTED got %+v", err) }
}

func TestIPAMSameUserStable(t *testing.T) {
    ctx, svc, _, netID := setupIPAMTestNetwork(t, "10.20.0.0/29") // /29 => usable 6
    a1, err := svc.AllocateIP(ctx, netID, "userX")
    if err != nil { t.Fatalf("alloc1: %v", err) }
    a2, err := svc.AllocateIP(ctx, netID, "userX")
    if err != nil { t.Fatalf("alloc2: %v", err) }
    if a1.IP != a2.IP { t.Fatalf("expected stable allocation, got %s vs %s", a1.IP, a2.IP) }
}

func TestIPAMInvalidNetwork(t *testing.T) {
    ctx := context.Background()
    netRepo := repository.NewInMemoryNetworkRepository()
    ipRepo := repository.NewInMemoryIPAM()
    svc := NewIPAMService(netRepo, ipRepo)
    _, err := svc.AllocateIP(ctx, "missing", "user1")
    if err == nil { t.Fatalf("expected error for missing network") }
    derr, ok := err.(*domain.Error)
    if !ok || derr.Code != domain.ErrNotFound { t.Fatalf("expected ErrNotFound got %+v", err) }
}
