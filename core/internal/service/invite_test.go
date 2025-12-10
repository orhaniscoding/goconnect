package service

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

func TestInviteService_CreateInvite(t *testing.T) {
	// Setup repositories
	inviteRepo := repository.NewInMemoryInviteTokenRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()

	svc := NewInviteService(inviteRepo, networkRepo, membershipRepo, "https://app.example.com")
	ctx := context.Background()

	// Create test network with approval policy
	network := &domain.Network{
		ID:         "net-invite-1",
		TenantID:   "tenant-1",
		Name:       "Invite Test Network",
		CIDR:       "10.10.0.0/24",
		Visibility: domain.NetworkVisibilityPrivate,
		JoinPolicy: domain.JoinPolicyInvite,
		CreatedBy:  "user-owner-1",
	}
	if err := networkRepo.Create(ctx, network); err != nil {
		t.Fatalf("failed to create network: %v", err)
	}

	// Create owner membership
	_, err := membershipRepo.UpsertApproved(ctx, "net-invite-1", "user-owner-1", domain.RoleOwner, time.Now())
	if err != nil {
		t.Fatalf("failed to create membership: %v", err)
	}

	t.Run("CreateInvite success", func(t *testing.T) {
		opts := CreateInviteOptions{
			ExpiresIn: 3600,
			UsesMax:   10,
		}
		response, err := svc.CreateInvite(ctx, "net-invite-1", "tenant-1", "user-owner-1", opts)
		if err != nil {
			t.Fatalf("CreateInvite failed: %v", err)
		}

		if response.NetworkID != "net-invite-1" {
			t.Errorf("expected network_id net-invite-1, got %s", response.NetworkID)
		}

		if response.UsesMax != 10 {
			t.Errorf("expected uses_max 10, got %d", response.UsesMax)
		}

		if response.UsesLeft != 10 {
			t.Errorf("expected uses_left 10, got %d", response.UsesLeft)
		}

		if response.Token == "" {
			t.Error("expected non-empty token")
		}

		if response.InviteURL == "" {
			t.Error("expected non-empty invite_url")
		}

		if !response.IsActive {
			t.Error("expected is_active to be true")
		}
	})

	t.Run("CreateInvite requires admin/owner role", func(t *testing.T) {
		// Create member-only membership
		_, _ = membershipRepo.UpsertApproved(ctx, "net-invite-1", "user-member-1", domain.RoleMember, time.Now())

		opts := CreateInviteOptions{}
		_, err := svc.CreateInvite(ctx, "net-invite-1", "tenant-1", "user-member-1", opts)
		if err == nil {
			t.Error("expected error for non-admin user")
		}

		if domErr, ok := err.(*domain.Error); ok {
			if domErr.Code != domain.ErrNotAuthorized {
				t.Errorf("expected ErrNotAuthorized, got %s", domErr.Code)
			}
		}
	})

	t.Run("CreateInvite fails for open networks", func(t *testing.T) {
		// Create open network
		openNetwork := &domain.Network{
			ID:         "net-open-1",
			TenantID:   "tenant-1",
			Name:       "Open Network",
			CIDR:       "10.20.0.0/24",
			Visibility: domain.NetworkVisibilityPublic,
			JoinPolicy: domain.JoinPolicyOpen,
			CreatedBy:  "user-owner-1",
		}
		networkRepo.Create(ctx, openNetwork)
		membershipRepo.UpsertApproved(ctx, "net-open-1", "user-owner-1", domain.RoleOwner, time.Now())

		opts := CreateInviteOptions{}
		_, err := svc.CreateInvite(ctx, "net-open-1", "tenant-1", "user-owner-1", opts)
		if err == nil {
			t.Error("expected error for open network")
		}
	})
}

func TestInviteService_ValidateAndUseInvite(t *testing.T) {
	// Setup repositories
	inviteRepo := repository.NewInMemoryInviteTokenRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()

	svc := NewInviteService(inviteRepo, networkRepo, membershipRepo, "https://app.example.com")
	ctx := context.Background()

	// Create test network
	network := &domain.Network{
		ID:         "net-validate-1",
		TenantID:   "tenant-1",
		Name:       "Validate Test Network",
		CIDR:       "10.30.0.0/24",
		Visibility: domain.NetworkVisibilityPrivate,
		JoinPolicy: domain.JoinPolicyInvite,
		CreatedBy:  "user-owner-1",
	}
	networkRepo.Create(ctx, network)
	membershipRepo.UpsertApproved(ctx, "net-validate-1", "user-owner-1", domain.RoleOwner, time.Now())

	// Create invite with 2 uses
	opts := CreateInviteOptions{
		ExpiresIn: 3600,
		UsesMax:   2,
	}
	response, _ := svc.CreateInvite(ctx, "net-validate-1", "tenant-1", "user-owner-1", opts)

	t.Run("ValidateInvite success", func(t *testing.T) {
		token, err := svc.ValidateInvite(ctx, response.Token)
		if err != nil {
			t.Fatalf("ValidateInvite failed: %v", err)
		}

		if !token.IsValid() {
			t.Error("expected valid token")
		}
	})

	t.Run("UseInvite decrements uses_left", func(t *testing.T) {
		token, err := svc.UseInvite(ctx, response.Token, "user-new-1")
		if err != nil {
			t.Fatalf("UseInvite failed: %v", err)
		}

		if token.UsesLeft != 1 {
			t.Errorf("expected uses_left 1, got %d", token.UsesLeft)
		}

		// Use again
		token, err = svc.UseInvite(ctx, response.Token, "user-new-2")
		if err != nil {
			t.Fatalf("UseInvite failed: %v", err)
		}

		if token.UsesLeft != 0 {
			t.Errorf("expected uses_left 0, got %d", token.UsesLeft)
		}

		// Third use should fail
		_, err = svc.UseInvite(ctx, response.Token, "user-new-3")
		if err == nil {
			t.Error("expected error for exhausted token")
		}
	})

	t.Run("ValidateInvite fails for invalid token", func(t *testing.T) {
		_, err := svc.ValidateInvite(ctx, "invalid-token")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})
}

func TestInviteService_RevokeInvite(t *testing.T) {
	// Setup repositories
	inviteRepo := repository.NewInMemoryInviteTokenRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()

	svc := NewInviteService(inviteRepo, networkRepo, membershipRepo, "https://app.example.com")
	ctx := context.Background()

	// Create test network
	network := &domain.Network{
		ID:         "net-revoke-1",
		TenantID:   "tenant-1",
		Name:       "Revoke Test Network",
		CIDR:       "10.40.0.0/24",
		Visibility: domain.NetworkVisibilityPrivate,
		JoinPolicy: domain.JoinPolicyInvite,
		CreatedBy:  "user-owner-1",
	}
	networkRepo.Create(ctx, network)
	membershipRepo.UpsertApproved(ctx, "net-revoke-1", "user-owner-1", domain.RoleOwner, time.Now())

	// Create invite
	opts := CreateInviteOptions{UsesMax: 0} // unlimited
	response, _ := svc.CreateInvite(ctx, "net-revoke-1", "tenant-1", "user-owner-1", opts)

	t.Run("RevokeInvite success", func(t *testing.T) {
		err := svc.RevokeInvite(ctx, response.ID, "net-revoke-1", "tenant-1", "user-owner-1")
		if err != nil {
			t.Fatalf("RevokeInvite failed: %v", err)
		}

		// Validate should fail now
		_, err = svc.ValidateInvite(ctx, response.Token)
		if err == nil {
			t.Error("expected error for revoked token")
		}

		if domErr, ok := err.(*domain.Error); ok {
			if domErr.Code != domain.ErrInviteTokenRevoked {
				t.Errorf("expected ErrInviteTokenRevoked, got %s", domErr.Code)
			}
		}
	})

	t.Run("RevokeInvite requires admin/owner role", func(t *testing.T) {
		// Create another invite
		opts := CreateInviteOptions{}
		response2, _ := svc.CreateInvite(ctx, "net-revoke-1", "tenant-1", "user-owner-1", opts)

		// Add member
		membershipRepo.UpsertApproved(ctx, "net-revoke-1", "user-member-1", domain.RoleMember, time.Now())

		err := svc.RevokeInvite(ctx, response2.ID, "net-revoke-1", "tenant-1", "user-member-1")
		if err == nil {
			t.Error("expected error for non-admin user")
		}
	})
}

func TestInviteService_ListInvites(t *testing.T) {
	// Setup repositories
	inviteRepo := repository.NewInMemoryInviteTokenRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()

	svc := NewInviteService(inviteRepo, networkRepo, membershipRepo, "https://app.example.com")
	ctx := context.Background()

	// Create test network
	network := &domain.Network{
		ID:         "net-list-1",
		TenantID:   "tenant-1",
		Name:       "List Test Network",
		CIDR:       "10.50.0.0/24",
		Visibility: domain.NetworkVisibilityPrivate,
		JoinPolicy: domain.JoinPolicyInvite,
		CreatedBy:  "user-owner-1",
	}
	networkRepo.Create(ctx, network)
	membershipRepo.UpsertApproved(ctx, "net-list-1", "user-owner-1", domain.RoleOwner, time.Now())

	// Create multiple invites
	for i := 0; i < 3; i++ {
		svc.CreateInvite(ctx, "net-list-1", "tenant-1", "user-owner-1", CreateInviteOptions{})
	}

	t.Run("ListInvites returns all active invites", func(t *testing.T) {
		invites, err := svc.ListInvites(ctx, "net-list-1", "user-owner-1")
		if err != nil {
			t.Fatalf("ListInvites failed: %v", err)
		}

		if len(invites) != 3 {
			t.Errorf("expected 3 invites, got %d", len(invites))
		}
	})

	t.Run("ListInvites requires admin/owner role", func(t *testing.T) {
		membershipRepo.UpsertApproved(ctx, "net-list-1", "user-member-1", domain.RoleMember, time.Now())

		_, err := svc.ListInvites(ctx, "net-list-1", "user-member-1")
		if err == nil {
			t.Error("expected error for non-admin user")
		}
	})
}

// ==================== SetAuditor Tests ====================

func TestInviteService_SetAuditor(t *testing.T) {
	inviteRepo := repository.NewInMemoryInviteTokenRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	service := NewInviteService(inviteRepo, networkRepo, membershipRepo, "https://app.example.com")

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

func TestInviteService_GetInviteByID(t *testing.T) {
	inviteRepo := repository.NewInMemoryInviteTokenRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	service := NewInviteService(inviteRepo, networkRepo, membershipRepo, "https://app.example.com")

	ctx := context.Background()

	// Create a network first
	network := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "Test Network",
		CIDR:       "10.10.0.0/24",
		Visibility: domain.NetworkVisibilityPrivate,
		JoinPolicy: domain.JoinPolicyInvite,
		CreatedBy:  "user-1",
	}
	_ = networkRepo.Create(ctx, network)

	// Create owner membership
	_, _ = membershipRepo.UpsertApproved(ctx, "net-1", "user-1", domain.RoleOwner, time.Now())

	t.Run("Get Existing Invite", func(t *testing.T) {
		// Create an invite
		resp, err := service.CreateInvite(ctx, "net-1", "tenant-1", "user-1", CreateInviteOptions{
			UsesMax:   10,
			ExpiresIn: 86400,
		})
		if err != nil {
			t.Fatalf("failed to create invite: %v", err)
		}

		// Get the invite by ID
		retrieved, err := service.GetInviteByID(ctx, resp.ID)
		if err != nil {
			t.Fatalf("failed to get invite: %v", err)
		}

		if retrieved.ID != resp.ID {
			t.Errorf("expected ID %s, got %s", resp.ID, retrieved.ID)
		}
		if retrieved.NetworkID != "net-1" {
			t.Errorf("expected network ID net-1, got %s", retrieved.NetworkID)
		}
	})

	t.Run("Get Non-Existent Invite", func(t *testing.T) {
		_, err := service.GetInviteByID(ctx, "non-existent")
		if err == nil {
			t.Fatal("expected error for non-existent invite")
		}
	})
}

// ==================== REVOKE INVITE EDGE CASE TESTS ====================

func TestInviteService_RevokeInvite_InviteNotFound(t *testing.T) {
	inviteRepo := repository.NewInMemoryInviteTokenRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()

	svc := NewInviteService(inviteRepo, networkRepo, membershipRepo, "https://app.example.com")
	ctx := context.Background()

	err := svc.RevokeInvite(ctx, "non-existent-invite", "net-1", "tenant-1", "user-1")
	if err == nil {
		t.Fatal("expected error for non-existent invite")
	}
}

func TestInviteService_RevokeInvite_WrongNetwork(t *testing.T) {
	inviteRepo := repository.NewInMemoryInviteTokenRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()

	svc := NewInviteService(inviteRepo, networkRepo, membershipRepo, "https://app.example.com")
	ctx := context.Background()

	// Create test network
	network := &domain.Network{
		ID:         "net-wrong-1",
		TenantID:   "tenant-1",
		Name:       "Wrong Network Test",
		CIDR:       "10.60.0.0/24",
		Visibility: domain.NetworkVisibilityPrivate,
		JoinPolicy: domain.JoinPolicyInvite,
		CreatedBy:  "user-owner-1",
	}
	networkRepo.Create(ctx, network)
	membershipRepo.UpsertApproved(ctx, "net-wrong-1", "user-owner-1", domain.RoleOwner, time.Now())

	// Create invite for net-wrong-1
	opts := CreateInviteOptions{}
	response, _ := svc.CreateInvite(ctx, "net-wrong-1", "tenant-1", "user-owner-1", opts)

	// Try to revoke using wrong network ID
	err := svc.RevokeInvite(ctx, response.ID, "different-network", "tenant-1", "user-owner-1")
	if err == nil {
		t.Fatal("expected error for wrong network")
	}
	domErr, ok := err.(*domain.Error)
	if !ok || domErr.Code != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestInviteService_RevokeInvite_UserNotMember(t *testing.T) {
	inviteRepo := repository.NewInMemoryInviteTokenRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()

	svc := NewInviteService(inviteRepo, networkRepo, membershipRepo, "https://app.example.com")
	ctx := context.Background()

	// Create test network
	network := &domain.Network{
		ID:         "net-notmember-1",
		TenantID:   "tenant-1",
		Name:       "Not Member Test",
		CIDR:       "10.61.0.0/24",
		Visibility: domain.NetworkVisibilityPrivate,
		JoinPolicy: domain.JoinPolicyInvite,
		CreatedBy:  "user-owner-1",
	}
	networkRepo.Create(ctx, network)
	membershipRepo.UpsertApproved(ctx, "net-notmember-1", "user-owner-1", domain.RoleOwner, time.Now())

	// Create invite
	opts := CreateInviteOptions{}
	response, _ := svc.CreateInvite(ctx, "net-notmember-1", "tenant-1", "user-owner-1", opts)

	// Try to revoke with a non-member user
	err := svc.RevokeInvite(ctx, response.ID, "net-notmember-1", "tenant-1", "random-user")
	if err == nil {
		t.Fatal("expected error for non-member user")
	}
	domErr, ok := err.(*domain.Error)
	if !ok || domErr.Code != domain.ErrNotAuthorized {
		t.Errorf("expected ErrNotAuthorized, got %v", err)
	}
}