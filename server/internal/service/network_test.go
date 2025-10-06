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
			network, err := service.CreateNetwork(context.Background(), tt.req, tt.userID, tt.idempotencyKey)
			
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
	
	// First request
	network1, err := service.CreateNetwork(context.Background(), req, userID, idempotencyKey)
	if err != nil {
		t.Fatalf("First CreateNetwork() failed: %v", err)
	}
	
	// Second request with same idempotency key and body should return cached result
	network2, err := service.CreateNetwork(context.Background(), req, userID, idempotencyKey)
	if err != nil {
		t.Fatalf("Second CreateNetwork() failed: %v", err)
	}
	
	// Should return the same network
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
	
	_, err = service.CreateNetwork(context.Background(), reqDifferent, userID, idempotencyKey)
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
				repo.Create(context.Background(), &domain.Network{
					ID:         "net1",
					TenantID:   "default",
					Name:       "Public Network",
					Visibility: domain.NetworkVisibilityPublic,
					CIDR:       "10.0.0.0/24",
					CreatedBy:  "user456",
				})
				// Add private network (should not be returned)
				repo.Create(context.Background(), &domain.Network{
					ID:         "net2",
					TenantID:   "default",
					Name:       "Private Network",
					Visibility: domain.NetworkVisibilityPrivate,
					CIDR:       "10.0.1.0/24",
					CreatedBy:  "user456",
				})
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
				repo.Create(context.Background(), &domain.Network{
					ID:         "net1",
					TenantID:   "default",
					Name:       "My Network",
					Visibility: domain.NetworkVisibilityPrivate,
					CIDR:       "10.0.0.0/24",
					CreatedBy:  "user123",
				})
				// Add other user's network (should not be returned)
				repo.Create(context.Background(), &domain.Network{
					ID:         "net2",
					TenantID:   "default",
					Name:       "Other Network",
					Visibility: domain.NetworkVisibilityPublic,
					CIDR:       "10.0.1.0/24",
					CreatedBy:  "user456",
				})
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
				repo.Create(context.Background(), &domain.Network{
					ID:         "net1",
					TenantID:   "default",
					Name:       "Network 1",
					Visibility: domain.NetworkVisibilityPublic,
					CIDR:       "10.0.0.0/24",
					CreatedBy:  "user123",
				})
				repo.Create(context.Background(), &domain.Network{
					ID:         "net2",
					TenantID:   "default",
					Name:       "Network 2",
					Visibility: domain.NetworkVisibilityPrivate,
					CIDR:       "10.0.1.0/24",
					CreatedBy:  "user456",
				})
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
			networks, _, err := service.ListNetworks(context.Background(), tt.req, tt.userID, tt.isAdmin)
			
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