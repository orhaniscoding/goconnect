package service

import (
	"context"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

func TestNetworkService_CreateNetwork(t *testing.T) {
	tests := []struct {
		name           string
		req            *domain.CreateNetworkRequest
		userID         string
		idempotencyKey string
		setupRepo      func(*repository.InMemoryNetworkRepository)
		wantErr        bool
		wantErrCode    string
	}{
		{
			name: "valid network creation",
			req: &domain.CreateNetworkRequest{
				Name:       "Test Network",
				Visibility: domain.NetworkVisibilityPublic,
				JoinPolicy: domain.JoinPolicyOpen,
				CIDR:       "10.0.0.0/24",
			},
			userID:         "user123",
			idempotencyKey: "test-key-1",
			setupRepo:      func(repo *repository.InMemoryNetworkRepository) {},
			wantErr:        false,
		},
		{
			name: "invalid CIDR format",
			req: &domain.CreateNetworkRequest{
				Name:       "Test Network",
				Visibility: domain.NetworkVisibilityPublic,
				JoinPolicy: domain.JoinPolicyOpen,
				CIDR:       "invalid-cidr",
			},
			userID:         "user123",
			idempotencyKey: "test-key-2",
			setupRepo:      func(repo *repository.InMemoryNetworkRepository) {},
			wantErr:        true,
			wantErrCode:    domain.ErrCIDRInvalid,
		},
		{
			name: "CIDR overlap with existing network",
			req: &domain.CreateNetworkRequest{
				Name:       "Overlapping Network",
				Visibility: domain.NetworkVisibilityPublic,
				JoinPolicy: domain.JoinPolicyOpen,
				CIDR:       "10.0.0.0/24",
			},
			userID:         "user123",
			idempotencyKey: "test-key-3",
			setupRepo: func(repo *repository.InMemoryNetworkRepository) {
				existingNetwork := &domain.Network{
					ID:         "existing-net",
					TenantID:   "default",
					Name:       "Existing Network",
					Visibility: domain.NetworkVisibilityPublic,
					JoinPolicy: domain.JoinPolicyOpen,
					CIDR:       "10.0.0.0/24",
					CreatedBy:  "user456",
				}
				if err := repo.Create(context.Background(), existingNetwork); err != nil {
					t.Fatalf("failed to create network: %v", err)
				}
			},
			wantErr:     true,
			wantErrCode: domain.ErrCIDROverlap,
		},
		{
			name: "CIDR overlap allowed for different tenants",
			req: &domain.CreateNetworkRequest{
				Name:       "Overlapping Network Different Tenant",
				Visibility: domain.NetworkVisibilityPublic,
				JoinPolicy: domain.JoinPolicyOpen,
				CIDR:       "10.0.0.0/24",
			},
			userID:         "user123",
			idempotencyKey: "test-key-4",
			setupRepo: func(repo *repository.InMemoryNetworkRepository) {
				existingNetwork := &domain.Network{
					ID:         "existing-net",
					TenantID:   "other-tenant",
					Name:       "Existing Network",
					Visibility: domain.NetworkVisibilityPublic,
					JoinPolicy: domain.JoinPolicyOpen,
					CIDR:       "10.0.0.0/24",
					CreatedBy:  "user456",
				}
				if err := repo.Create(context.Background(), existingNetwork); err != nil {
					t.Fatalf("failed to create network: %v", err)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup repositories
			networkRepo := repository.NewInMemoryNetworkRepository()
			idempotencyRepo := repository.NewInMemoryIdempotencyRepository()

			// Setup test data
			tt.setupRepo(networkRepo)

			// Create service
			service := NewNetworkService(networkRepo, idempotencyRepo)

			// Execute test
			network, err := service.CreateNetwork(context.Background(), tt.req, tt.userID, "default", tt.idempotencyKey)

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateNetwork() expected error but got none")
					return
				}

				if domainErr, ok := err.(*domain.Error); ok {
					if domainErr.Code != tt.wantErrCode {
						t.Errorf("CreateNetwork() error code = %v, want %v", domainErr.Code, tt.wantErrCode)
					}
				} else {
					t.Errorf("CreateNetwork() expected domain.Error but got %T", err)
				}
				return
			}

			// Check success expectations
			if err != nil {
				t.Errorf("CreateNetwork() unexpected error = %v", err)
				return
			}

			if network == nil {
				t.Error("CreateNetwork() returned nil network")
				return
			}

			// Validate network fields
			if network.Name != tt.req.Name {
				t.Errorf("CreateNetwork() network.Name = %v, want %v", network.Name, tt.req.Name)
			}

			if network.CIDR != tt.req.CIDR {
				t.Errorf("CreateNetwork() network.CIDR = %v, want %v", network.CIDR, tt.req.CIDR)
			}

			if network.CreatedBy != tt.userID {
				t.Errorf("CreateNetwork() network.CreatedBy = %v, want %v", network.CreatedBy, tt.userID)
			}
		})
	}
}

func TestNetworkService_CreateNetwork_Idempotency(t *testing.T) {
	networkRepo := repository.NewInMemoryNetworkRepository()
	idempotencyRepo := repository.NewInMemoryIdempotencyRepository()
	service := NewNetworkService(networkRepo, idempotencyRepo)

	req := &domain.CreateNetworkRequest{
		Name:       "Test Network",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.0.0.0/24",
	}

	userID := "user123"
	idempotencyKey := "test-idempotency-key"

	// First call
	network1, err := service.CreateNetwork(context.Background(), req, userID, "default", idempotencyKey)
	if err != nil {
		t.Fatalf("First CreateNetwork() failed: %v", err)
	}

	// Second call with same idempotency key
	network2, err := service.CreateNetwork(context.Background(), req, userID, "default", idempotencyKey)
	if err != nil {
		t.Fatalf("Second CreateNetwork() failed: %v", err)
	} // Should return the same network
	if network1.ID != network2.ID {
		t.Errorf("Idempotent requests returned different network IDs: %v vs %v", network1.ID, network2.ID)
	}

	// Third request with same key but different body
	reqDifferent := &domain.CreateNetworkRequest{
		Name:       "Different Network",
		Visibility: domain.NetworkVisibilityPrivate,
		JoinPolicy: domain.JoinPolicyApproval,
		CIDR:       "10.0.1.0/24", // Different CIDR to avoid overlap
	}

	_, err = service.CreateNetwork(context.Background(), reqDifferent, userID, "default", idempotencyKey)
	if err == nil {
		t.Error("Expected idempotency conflict error but got none")
	}

	if domainErr, ok := err.(*domain.Error); ok {
		if domainErr.Code != domain.ErrIdempotencyConflict {
			t.Errorf("Expected ErrIdempotencyConflict but got %v", domainErr.Code)
		}
	}
}

func TestNetworkService_ListNetworks(t *testing.T) {
	tests := []struct {
		name      string
		req       *domain.ListNetworksRequest
		userID    string
		isAdmin   bool
		setupRepo func(*repository.InMemoryNetworkRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name: "list public networks",
			req: &domain.ListNetworksRequest{
				Visibility: "public",
				Limit:      10,
			},
			userID:  "user123",
			isAdmin: false,
			setupRepo: func(repo *repository.InMemoryNetworkRepository) {
				// Add public network
				if err := repo.Create(context.Background(), &domain.Network{
					ID:         "net1",
					TenantID:   "default",
					Name:       "Public Network",
					Visibility: domain.NetworkVisibilityPublic,
					CIDR:       "10.0.0.0/24",
					CreatedBy:  "user456",
				}); err != nil {
					t.Fatalf("create net1: %v", err)
				}
				// Add private network (should not be returned)
				if err := repo.Create(context.Background(), &domain.Network{
					ID:         "net2",
					TenantID:   "default",
					Name:       "Private Network",
					Visibility: domain.NetworkVisibilityPrivate,
					CIDR:       "10.0.1.0/24",
					CreatedBy:  "user456",
				}); err != nil {
					t.Fatalf("create net2: %v", err)
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "list mine networks",
			req: &domain.ListNetworksRequest{
				Visibility: "mine",
				Limit:      10,
			},
			userID:  "user123",
			isAdmin: false,
			setupRepo: func(repo *repository.InMemoryNetworkRepository) {
				// Add user's network
				if err := repo.Create(context.Background(), &domain.Network{
					ID:         "net1",
					TenantID:   "default",
					Name:       "My Network",
					Visibility: domain.NetworkVisibilityPrivate,
					CIDR:       "10.0.0.0/24",
					CreatedBy:  "user123",
				}); err != nil {
					t.Fatalf("create my net: %v", err)
				}
				// Add other user's network (should not be returned)
				if err := repo.Create(context.Background(), &domain.Network{
					ID:         "net2",
					TenantID:   "default",
					Name:       "Other Network",
					Visibility: domain.NetworkVisibilityPublic,
					CIDR:       "10.0.1.0/24",
					CreatedBy:  "user456",
				}); err != nil {
					t.Fatalf("create other net: %v", err)
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "list all networks as admin",
			req: &domain.ListNetworksRequest{
				Visibility: "all",
				Limit:      10,
			},
			userID:  "admin123",
			isAdmin: true,
			setupRepo: func(repo *repository.InMemoryNetworkRepository) {
				if err := repo.Create(context.Background(), &domain.Network{
					ID:         "net1",
					TenantID:   "default",
					Name:       "Network 1",
					Visibility: domain.NetworkVisibilityPublic,
					CIDR:       "10.0.0.0/24",
					CreatedBy:  "user123",
				}); err != nil {
					t.Fatalf("create net1: %v", err)
				}
				if err := repo.Create(context.Background(), &domain.Network{
					ID:         "net2",
					TenantID:   "default",
					Name:       "Network 2",
					Visibility: domain.NetworkVisibilityPrivate,
					CIDR:       "10.0.1.0/24",
					CreatedBy:  "user456",
				}); err != nil {
					t.Fatalf("create net2: %v", err)
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "list all networks as non-admin should fail",
			req: &domain.ListNetworksRequest{
				Visibility: "all",
				Limit:      10,
			},
			userID:    "user123",
			isAdmin:   false,
			setupRepo: func(repo *repository.InMemoryNetworkRepository) {},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup repositories
			networkRepo := repository.NewInMemoryNetworkRepository()
			idempotencyRepo := repository.NewInMemoryIdempotencyRepository()

			// Setup test data
			tt.setupRepo(networkRepo)

			// Create service
			service := NewNetworkService(networkRepo, idempotencyRepo)

			// Execute test
			networks, _, err := service.ListNetworks(context.Background(), tt.req, tt.userID, "default", tt.isAdmin)

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Error("ListNetworks() expected error but got none")
				}
				return
			}

			// Check success expectations
			if err != nil {
				t.Errorf("ListNetworks() unexpected error = %v", err)
				return
			}

			if len(networks) != tt.wantCount {
				t.Errorf("ListNetworks() returned %v networks, want %v", len(networks), tt.wantCount)
			}
		})
	}
}

func TestNetworkService_GetNetwork(t *testing.T) {
	repo := repository.NewInMemoryNetworkRepository()
	service := NewNetworkService(repo, repository.NewInMemoryIdempotencyRepository())
	ctx := context.Background()

	// Create test network
	net := &domain.Network{
		ID:         "net1",
		TenantID:   "default",
		Name:       "Test Network",
		Visibility: domain.NetworkVisibilityPublic,
		CIDR:       "10.0.0.0/24",
		CreatedBy:  "user1",
	}
	repo.Create(ctx, net)

	// Test Success
	got, err := service.GetNetwork(ctx, "net1", "user1", "default")
	if err != nil {
		t.Fatalf("GetNetwork failed: %v", err)
	}
	if got.ID != "net1" {
		t.Errorf("GetNetwork returned wrong ID: %s", got.ID)
	}

	// Test Not Found
	_, err = service.GetNetwork(ctx, "nonexistent", "user1", "default")
	if err == nil {
		t.Error("GetNetwork expected error for non-existent ID")
	}

	// Test Tenant Mismatch
	_, err = service.GetNetwork(ctx, "net1", "user1", "other-tenant")
	if err == nil {
		t.Error("GetNetwork expected error for tenant mismatch")
	}
	if domainErr, ok := err.(*domain.Error); !ok || domainErr.Code != domain.ErrNotFound {
		t.Errorf("GetNetwork expected ErrNotFound for tenant mismatch, got %v", err)
	}
}

func TestNetworkService_UpdateNetwork(t *testing.T) {
	repo := repository.NewInMemoryNetworkRepository()
	service := NewNetworkService(repo, repository.NewInMemoryIdempotencyRepository())
	ctx := context.Background()

	// Create test network
	net := &domain.Network{
		ID:         "net1",
		TenantID:   "default",
		Name:       "Original Name",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.0.0.0/24",
		CreatedBy:  "user1",
	}
	repo.Create(ctx, net)

	// Test Success
	patch := map[string]any{
		"name":        "Updated Name",
		"visibility":  "private",
		"join_policy": "approval",
	}
	updated, err := service.UpdateNetwork(ctx, "net1", "user1", "default", patch)
	if err != nil {
		t.Fatalf("UpdateNetwork failed: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("UpdateNetwork name mismatch: %s", updated.Name)
	}
	if updated.Visibility != domain.NetworkVisibilityPrivate {
		t.Errorf("UpdateNetwork visibility mismatch: %s", updated.Visibility)
	}
	if updated.JoinPolicy != domain.JoinPolicyApproval {
		t.Errorf("UpdateNetwork join_policy mismatch: %s", updated.JoinPolicy)
	}

	// Test Invalid Name
	_, err = service.UpdateNetwork(ctx, "net1", "user1", "default", map[string]any{"name": "ab"}) // Too short
	if err == nil {
		t.Error("UpdateNetwork expected error for short name")
	}

	// Test Invalid Visibility
	_, err = service.UpdateNetwork(ctx, "net1", "user1", "default", map[string]any{"visibility": "invalid"})
	if err == nil {
		t.Error("UpdateNetwork expected error for invalid visibility")
	}

	// Test Invalid JoinPolicy
	_, err = service.UpdateNetwork(ctx, "net1", "user1", "default", map[string]any{"join_policy": "invalid"})
	if err == nil {
		t.Error("UpdateNetwork expected error for invalid join_policy")
	}

	// Test Not Found
	_, err = service.UpdateNetwork(ctx, "nonexistent", "user1", "default", patch)
	if err == nil {
		t.Error("UpdateNetwork expected error for non-existent ID")
	}

	// Test Tenant Mismatch
	_, err = service.UpdateNetwork(ctx, "net1", "user1", "other-tenant", patch)
	if err == nil {
		t.Error("UpdateNetwork expected error for tenant mismatch")
	}
	if domainErr, ok := err.(*domain.Error); !ok || domainErr.Code != domain.ErrNotFound {
		t.Errorf("UpdateNetwork expected ErrNotFound for tenant mismatch, got %v", err)
	}
}

func TestNetworkService_DeleteNetwork(t *testing.T) {
	repo := repository.NewInMemoryNetworkRepository()
	service := NewNetworkService(repo, repository.NewInMemoryIdempotencyRepository())
	ctx := context.Background()

	// Create test network
	net := &domain.Network{
		ID:         "net1",
		TenantID:   "default",
		Name:       "To Delete",
		Visibility: domain.NetworkVisibilityPublic,
		CIDR:       "10.0.0.0/24",
		CreatedBy:  "user1",
	}
	repo.Create(ctx, net)

	// Test Success
	err := service.DeleteNetwork(ctx, "net1", "user1", "default")
	if err != nil {
		t.Fatalf("DeleteNetwork failed: %v", err)
	}

	// Verify deletion
	_, err = service.GetNetwork(ctx, "net1", "user1", "default")
	if err == nil {
		t.Error("GetNetwork should fail after deletion")
	}

	// Test Not Found
	err = service.DeleteNetwork(ctx, "nonexistent", "user1", "default")
	if err == nil {
		t.Error("DeleteNetwork expected error for non-existent ID")
	}

	// Test Tenant Mismatch
	// Re-create network for mismatch test
	net2 := &domain.Network{
		ID:         "net2",
		TenantID:   "default",
		Name:       "Mismatch Test",
		Visibility: domain.NetworkVisibilityPublic,
		CIDR:       "10.0.1.0/24",
		CreatedBy:  "user1",
	}
	repo.Create(ctx, net2)

	err = service.DeleteNetwork(ctx, "net2", "user1", "other-tenant")
	if err == nil {
		t.Error("DeleteNetwork expected error for tenant mismatch")
	}
	if domainErr, ok := err.(*domain.Error); !ok || domainErr.Code != domain.ErrNotFound {
		t.Errorf("DeleteNetwork expected ErrNotFound for tenant mismatch, got %v", err)
	}
}

// ==================== SetAuditor Tests ====================

func TestNetworkService_SetAuditor(t *testing.T) {
	networkRepo := repository.NewInMemoryNetworkRepository()
	idempotencyRepo := repository.NewInMemoryIdempotencyRepository()
	service := NewNetworkService(networkRepo, idempotencyRepo)

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
