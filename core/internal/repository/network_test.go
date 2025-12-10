package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

func mkNet(id, name, cidr, createdBy string, vis domain.NetworkVisibility) *domain.Network {
	now := time.Now()
	return &domain.Network{
		ID:         id,
		TenantID:   "default",
		Name:       name,
		Visibility: vis,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       cidr,
		CreatedBy:  createdBy,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func TestInMemoryNetworkRepository_CreateListAndCursor(t *testing.T) {
	r := NewInMemoryNetworkRepository()
	ctx := context.Background()
	// Create two networks
	if err := r.Create(ctx, mkNet("n1", "Public One", "10.0.0.0/24", "u1", domain.NetworkVisibilityPublic)); err != nil {
		t.Fatalf("create n1: %v", err)
	}
	if err := r.Create(ctx, mkNet("n2", "Mine One", "10.0.1.0/24", "u2", domain.NetworkVisibilityPrivate)); err != nil {
		t.Fatalf("create n2: %v", err)
	}

	// List public (should include only n1)
	got, next, err := r.List(ctx, NetworkFilter{Visibility: "public", TenantID: "default", Limit: 10})
	if err != nil {
		t.Fatalf("list public: %v", err)
	}
	if len(got) != 1 || got[0].ID != "n1" {
		t.Fatalf("unexpected public list: %+v", got)
	}
	if next != "" {
		t.Fatalf("unexpected next cursor: %q", next)
	}

	// Cursor: start after n1, limit 1 over all visible to admin
	got, next, err = r.List(ctx, NetworkFilter{Visibility: "all", TenantID: "default", IsAdmin: true, Limit: 1, Cursor: "n1"})
	if err != nil {
		t.Fatalf("list all cursor: %v", err)
	}
	if len(got) != 1 || got[0].ID != "n2" {
		t.Fatalf("unexpected cursor page: %+v", got)
	}
	if next != "n2" {
		t.Fatalf("unexpected next cursor after paging: %q", next)
	}
}

func TestInMemoryNetworkRepository_NameUniquenessAndSoftDelete(t *testing.T) {
	r := NewInMemoryNetworkRepository()
	ctx := context.Background()
	n1 := mkNet("n1", "SameName", "10.0.0.0/24", "u1", domain.NetworkVisibilityPublic)
	if err := r.Create(ctx, n1); err != nil {
		t.Fatalf("create n1: %v", err)
	}
	// Duplicate name under same tenant -> error
	if err := r.Create(ctx, mkNet("n2", "SameName", "10.0.2.0/24", "u2", domain.NetworkVisibilityPrivate)); err == nil {
		t.Fatalf("expected duplicate name error")
	}
	// Soft delete n1 then allow same name
	if err := r.SoftDelete(ctx, "n1", time.Now()); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	if err := r.Create(ctx, mkNet("n3", "SameName", "10.10.0.0/24", "u2", domain.NetworkVisibilityPublic)); err != nil {
		t.Fatalf("create n3 after delete: %v", err)
	}
}

func TestInMemoryNetworkRepository_CIDROverlap(t *testing.T) {
	r := NewInMemoryNetworkRepository()
	ctx := context.Background()
	if err := r.Create(ctx, mkNet("n1", "Net1", "10.0.0.0/24", "u1", domain.NetworkVisibilityPublic)); err != nil {
		t.Fatalf("create: %v", err)
	}
	// Overlap with 10.0.0.0/24
	ov, err := r.CheckCIDROverlap(ctx, "10.0.0.128/25", "", "default")
	if err != nil {
		t.Fatalf("overlap err: %v", err)
	}
	if !ov {
		t.Fatalf("expected overlap true")
	}
	// No overlap
	ov, err = r.CheckCIDROverlap(ctx, "10.0.1.0/24", "", "default")
	if err != nil {
		t.Fatalf("overlap err: %v", err)
	}
	if ov {
		t.Fatalf("expected overlap false")
	}
}

func TestInMemoryNetworkRepository_GetByID(t *testing.T) {
	r := NewInMemoryNetworkRepository()
	ctx := context.Background()

	// Create a network
	net := mkNet("n1", "TestNet", "10.0.0.0/24", "user1", domain.NetworkVisibilityPublic)
	if err := r.Create(ctx, net); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Get existing network
	got, err := r.GetByID(ctx, "n1")
	if err != nil {
		t.Fatalf("getbyid: %v", err)
	}
	if got.ID != "n1" || got.Name != "TestNet" {
		t.Fatalf("unexpected network: %+v", got)
	}

	// Get non-existent network
	_, err = r.GetByID(ctx, "nonexistent")
	if err == nil {
		t.Fatalf("expected error for non-existent network")
	}

	// Soft delete and verify it's not returned
	if err := r.SoftDelete(ctx, "n1", time.Now()); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	_, err = r.GetByID(ctx, "n1")
	if err == nil {
		t.Fatalf("expected error for deleted network")
	}
}

func TestInMemoryNetworkRepository_Update(t *testing.T) {
	r := NewInMemoryNetworkRepository()
	ctx := context.Background()

	// Create a network
	net := mkNet("n1", "Original", "10.0.0.0/24", "user1", domain.NetworkVisibilityPublic)
	if err := r.Create(ctx, net); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Update the network
	updated, err := r.Update(ctx, "n1", func(n *domain.Network) error {
		n.Name = "Updated"
		n.Visibility = domain.NetworkVisibilityPrivate
		return nil
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Updated" || updated.Visibility != domain.NetworkVisibilityPrivate {
		t.Fatalf("update did not apply: %+v", updated)
	}

	// Verify the update persisted
	got, _ := r.GetByID(ctx, "n1")
	if got.Name != "Updated" {
		t.Fatalf("update not persisted: %+v", got)
	}

	// Update non-existent network
	_, err = r.Update(ctx, "nonexistent", func(n *domain.Network) error {
		return nil
	})
	if err == nil {
		t.Fatalf("expected error for non-existent network")
	}
}

func TestInMemoryNetworkRepository_Update_NameConflict(t *testing.T) {
	r := NewInMemoryNetworkRepository()
	ctx := context.Background()

	// Create two networks
	if err := r.Create(ctx, mkNet("n1", "Net1", "10.0.0.0/24", "u1", domain.NetworkVisibilityPublic)); err != nil {
		t.Fatalf("create n1: %v", err)
	}
	if err := r.Create(ctx, mkNet("n2", "Net2", "10.0.1.0/24", "u1", domain.NetworkVisibilityPublic)); err != nil {
		t.Fatalf("create n2: %v", err)
	}

	// Try to rename n1 to Net2 (conflict)
	_, err := r.Update(ctx, "n1", func(n *domain.Network) error {
		n.Name = "Net2"
		return nil
	})
	if err == nil {
		t.Fatalf("expected error for name conflict")
	}

	// Try to rename n1 to Net1 (no change, should succeed)
	_, err = r.Update(ctx, "n1", func(n *domain.Network) error {
		n.Name = "Net1"
		return nil
	})
	if err != nil {
		t.Fatalf("update same name failed: %v", err)
	}
}

func TestInMemoryNetworkRepository_List_AdvancedFilters(t *testing.T) {
	r := NewInMemoryNetworkRepository()
	ctx := context.Background()

	// Create networks
	// n1: Public, User1, "Alpha"
	// n2: Private, User1, "Beta"
	// n3: Public, User2, "Gamma"
	r.Create(ctx, mkNet("n1", "Alpha", "10.0.0.0/24", "u1", domain.NetworkVisibilityPublic))
	r.Create(ctx, mkNet("n2", "Beta", "10.0.1.0/24", "u1", domain.NetworkVisibilityPrivate))
	r.Create(ctx, mkNet("n3", "Gamma", "10.0.2.0/24", "u2", domain.NetworkVisibilityPublic))

	// Filter: Mine (User1) -> Should get n1, n2
	got, _, err := r.List(ctx, NetworkFilter{Visibility: "mine", TenantID: "default", UserID: "u1", Limit: 10})
	if err != nil {
		t.Fatalf("list mine: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 networks for mine, got %d", len(got))
	}

	// Filter: Search "alp" -> Should get n1
	got, _, err = r.List(ctx, NetworkFilter{Visibility: "public", TenantID: "default", Search: "alp", Limit: 10})
	if err != nil {
		t.Fatalf("list search: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Alpha" {
		t.Fatalf("expected Alpha, got %+v", got)
	}

	// Filter: Search "ta" -> Should get n2 (if searching in mine)
	got, _, err = r.List(ctx, NetworkFilter{Visibility: "mine", TenantID: "default", UserID: "u1", Search: "ta", Limit: 10})
	if err != nil {
		t.Fatalf("list search mine: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Beta" {
		t.Fatalf("expected Beta, got %+v", got)
	}
}

func TestInMemoryNetworkRepository_CIDROverlap_EdgeCases(t *testing.T) {
	r := NewInMemoryNetworkRepository()
	ctx := context.Background()

	r.Create(ctx, mkNet("n1", "Net1", "10.0.0.0/24", "u1", domain.NetworkVisibilityPublic))

	// Check overlap with itself (excludeID) -> Should be false
	ov, err := r.CheckCIDROverlap(ctx, "10.0.0.0/24", "n1", "default")
	if err != nil {
		t.Fatalf("overlap check: %v", err)
	}
	if ov {
		t.Fatalf("expected no overlap when excluding self")
	}

	// Soft delete n1
	r.SoftDelete(ctx, "n1", time.Now())

	// Check overlap with deleted network -> Should be false
	ov, err = r.CheckCIDROverlap(ctx, "10.0.0.0/24", "n2", "default")
	if err != nil {
		t.Fatalf("overlap check: %v", err)
	}
	if ov {
		t.Fatalf("expected no overlap with deleted network")
	}
}

func TestInMemoryNetworkRepository_CIDROverlap_TenantIsolation(t *testing.T) {
	r := NewInMemoryNetworkRepository()
	ctx := context.Background()
	// Create network in tenant "t1"
	n1 := mkNet("n1", "Net1", "10.0.0.0/24", "u1", domain.NetworkVisibilityPublic)
	n1.TenantID = "t1"
	r.Create(ctx, n1)

	// Check overlap in tenant "t1" -> Should be true
	ov, err := r.CheckCIDROverlap(ctx, "10.0.0.0/24", "", "t1")
	if err != nil || !ov {
		t.Fatalf("expected overlap in same tenant")
	}

	// Check overlap in tenant "t2" -> Should be false
	ov, err = r.CheckCIDROverlap(ctx, "10.0.0.0/24", "", "t2")
	if err != nil || ov {
		t.Fatalf("expected NO overlap in different tenant")
	}
}

func TestInMemoryNetworkRepository_List_Search(t *testing.T) {
	r := NewInMemoryNetworkRepository()
	ctx := context.Background()

	networks := []*domain.Network{
		mkNet("n1", "Alpha Network", "10.0.0.0/24", "u1", domain.NetworkVisibilityPublic),
		mkNet("n2", "Beta Network", "10.0.1.0/24", "u1", domain.NetworkVisibilityPublic),
		mkNet("n3", "Gamma Network", "10.0.2.0/24", "u1", domain.NetworkVisibilityPublic),
	}

	for _, n := range networks {
		if err := r.Create(ctx, n); err != nil {
			t.Fatalf("create %s: %v", n.Name, err)
		}
	}

	// Test search "Beta"
	got, _, err := r.List(ctx, NetworkFilter{
		TenantID: "default",
		IsAdmin:  true,
		Limit:    10,
		Search:   "Beta",
	})
	if err != nil {
		t.Fatalf("list search: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Beta Network" {
		t.Fatalf("expected Beta Network, got %+v", got)
	}

	// Test case insensitive "alpha"
	got, _, err = r.List(ctx, NetworkFilter{
		TenantID: "default",
		IsAdmin:  true,
		Limit:    10,
		Search:   "alpha",
	})
	if err != nil {
		t.Fatalf("list search: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Alpha Network" {
		t.Fatalf("expected Alpha Network, got %+v", got)
	}

	// Test empty search
	got, _, err = r.List(ctx, NetworkFilter{
		TenantID: "default",
		IsAdmin:  true,
		Limit:    10,
		Search:   "",
	})
	if err != nil {
		t.Fatalf("list search: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 networks, got %d", len(got))
	}
}

func TestNetworkRepository_Count(t *testing.T) {
	repo := NewInMemoryNetworkRepository()
	ctx := context.Background()

	// Count empty
	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}

	// Add networks
	n1 := mkNet("n1", "net1", "10.0.0.0/24", "user1", domain.NetworkVisibilityPrivate)
	n2 := mkNet("n2", "net2", "10.0.1.0/24", "user1", domain.NetworkVisibilityPrivate)
	n3 := mkNet("n3", "net3", "10.0.2.0/24", "user1", domain.NetworkVisibilityPublic)
	_ = repo.Create(ctx, n1)
	_ = repo.Create(ctx, n2)
	_ = repo.Create(ctx, n3)

	// Count all
	count, err = repo.Count(ctx)
	if err != nil {
		t.Fatalf("count all: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3, got %d", count)
	}
}
