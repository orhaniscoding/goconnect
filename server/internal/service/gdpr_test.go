package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

func TestGDPRService_ExportUserData(t *testing.T) {
	// Setup in-memory repositories
	userRepo := repository.NewInMemoryUserRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()

	svc := NewGDPRService(userRepo, deviceRepo, networkRepo, membershipRepo)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		ID:       "user-gdpr-1",
		TenantID: "tenant-1",
		Email:    "gdpr@example.com",
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Create test devices
	device1 := &domain.Device{
		ID:        "dev-1",
		UserID:    "user-gdpr-1",
		TenantID:  "tenant-1",
		Name:      "Test Device 1",
		Platform:  "windows",
		PubKey:    "pubkey1",
		Active:    true,
		IPAddress: "10.0.0.2",
	}
	device2 := &domain.Device{
		ID:        "dev-2",
		UserID:    "user-gdpr-1",
		TenantID:  "tenant-1",
		Name:      "Test Device 2",
		Platform:  "macos",
		PubKey:    "pubkey2",
		Active:    false,
		IPAddress: "10.0.0.3",
	}
	if err := deviceRepo.Create(ctx, device1); err != nil {
		t.Fatalf("failed to create device1: %v", err)
	}
	if err := deviceRepo.Create(ctx, device2); err != nil {
		t.Fatalf("failed to create device2: %v", err)
	}

	// Create test network (created by this user)
	network := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "Test Network",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CreatedBy:  "user-gdpr-1",
	}
	if err := networkRepo.Create(ctx, network); err != nil {
		t.Fatalf("failed to create network: %v", err)
	}

	// Create membership (using UpsertApproved)
	_, err := membershipRepo.UpsertApproved(ctx, "net-1", "user-gdpr-1", domain.RoleMember, time.Now())
	if err != nil {
		t.Fatalf("failed to create membership: %v", err)
	}

	t.Run("ExportUserData returns correct structure", func(t *testing.T) {
		export, err := svc.ExportUserData(ctx, "user-gdpr-1", "tenant-1")
		if err != nil {
			t.Fatalf("ExportUserData failed: %v", err)
		}

		// Check user data
		if export.User.ID != "user-gdpr-1" {
			t.Errorf("expected user ID user-gdpr-1, got %s", export.User.ID)
		}
		if export.User.Email != "gdpr@example.com" {
			t.Errorf("expected email gdpr@example.com, got %s", export.User.Email)
		}

		// Check devices
		if len(export.Devices) != 2 {
			t.Errorf("expected 2 devices, got %d", len(export.Devices))
		}

		// Check memberships
		if len(export.Memberships) != 1 {
			t.Errorf("expected 1 membership, got %d", len(export.Memberships))
		}

		// Check exported_at
		if export.ExportedAt.IsZero() {
			t.Error("exported_at should not be zero")
		}
	})

	t.Run("ExportUserDataJSON returns valid JSON", func(t *testing.T) {
		jsonData, err := svc.ExportUserDataJSON(ctx, "user-gdpr-1", "tenant-1")
		if err != nil {
			t.Fatalf("ExportUserDataJSON failed: %v", err)
		}

		// Verify it's valid JSON
		var parsed GDPRExportData
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if parsed.User.Email != "gdpr@example.com" {
			t.Errorf("expected email in JSON to be gdpr@example.com, got %s", parsed.User.Email)
		}
	})

	t.Run("ExportUserData for non-existent user returns error", func(t *testing.T) {
		_, err := svc.ExportUserData(ctx, "non-existent-user", "tenant-1")
		if err == nil {
			t.Error("expected error for non-existent user")
		}
	})
}

func TestGDPRService_DeleteUserData(t *testing.T) {
	// Setup in-memory repositories
	userRepo := repository.NewInMemoryUserRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()

	svc := NewGDPRService(userRepo, deviceRepo, networkRepo, membershipRepo)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		ID:       "user-delete-1",
		TenantID: "tenant-1",
		Email:    "delete@example.com",
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	t.Run("RequestDeletion returns pending request", func(t *testing.T) {
		req, err := svc.RequestDeletion(ctx, "user-delete-1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Should return a request object
		if req == nil {
			t.Fatal("expected deletion request to be returned")
		}

		if req.Status != "pending" {
			t.Errorf("expected status pending, got %s", req.Status)
		}

		if req.UserID != "user-delete-1" {
			t.Errorf("expected userID user-delete-1, got %s", req.UserID)
		}
	})

	t.Run("RequestDeletion for non-existent user returns error", func(t *testing.T) {
		_, err := svc.RequestDeletion(ctx, "non-existent-user")
		if err == nil {
			t.Error("expected error for non-existent user")
		}
	})
}

func TestGDPRService_ExportUserData_EmptyData(t *testing.T) {
	// Setup in-memory repositories
	userRepo := repository.NewInMemoryUserRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()

	svc := NewGDPRService(userRepo, deviceRepo, networkRepo, membershipRepo)
	ctx := context.Background()

	// Create user with no devices/networks
	user := &domain.User{
		ID:       "user-empty",
		TenantID: "tenant-1",
		Email:    "empty@example.com",
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	t.Run("ExportUserData with no devices or networks", func(t *testing.T) {
		export, err := svc.ExportUserData(ctx, "user-empty", "tenant-1")
		if err != nil {
			t.Fatalf("ExportUserData failed: %v", err)
		}

		if export.User.ID != "user-empty" {
			t.Errorf("expected user ID user-empty, got %s", export.User.ID)
		}

		if len(export.Devices) != 0 {
			t.Errorf("expected 0 devices, got %d", len(export.Devices))
		}

		if len(export.Networks) != 0 {
			t.Errorf("expected 0 networks, got %d", len(export.Networks))
		}

		if len(export.Memberships) != 0 {
			t.Errorf("expected 0 memberships, got %d", len(export.Memberships))
		}
	})
}
