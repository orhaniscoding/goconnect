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
	got, next, err := r.List(ctx, NetworkFilter{Visibility: "public", Limit: 10})
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
	got, next, err = r.List(ctx, NetworkFilter{Visibility: "all", IsAdmin: true, Limit: 1, Cursor: "n1"})
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
	ov, err := r.CheckCIDROverlap(ctx, "10.0.0.128/25", "")
	if err != nil {
		t.Fatalf("overlap err: %v", err)
	}
	if !ov {
		t.Fatalf("expected overlap true")
	}
	// No overlap
	ov, err = r.CheckCIDROverlap(ctx, "10.0.1.0/24", "")
	if err != nil {
		t.Fatalf("overlap err: %v", err)
	}
	if ov {
		t.Fatalf("expected overlap false")
	}
}
