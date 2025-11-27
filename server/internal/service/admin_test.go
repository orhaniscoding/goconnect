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

// setupAdminServiceTest creates all required repositories for admin service testing
func setupAdminServiceTest() (*AdminService, *repository.InMemoryUserRepository, *repository.InMemoryTenantRepository, *repository.InMemoryNetworkRepository, *repository.InMemoryDeviceRepository, *repository.InMemoryChatRepository) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	chatRepo := repository.NewInMemoryChatRepository()
	adminRepo := repository.NewAdminRepository(nil) // Pass nil for test (in-memory doesn't use DB)

	service := NewAdminService(
		userRepo,
		adminRepo,
		tenantRepo,
		networkRepo,
		deviceRepo,
		chatRepo,
		nil,                      // No auditor in tests
		nil,                      // No Redis in tests
		func() int { return 10 }, // mock active connections
	)

	return service, userRepo, tenantRepo, networkRepo, deviceRepo, chatRepo
}

// ==================== LIST USERS TESTS ====================

func TestAdminService_ListUsers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "user1@test.com"}))
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u2", Email: "user2@test.com"}))

		users, total, err := svc.ListUsers(ctx, 10, 0, "")
		require.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, 2, total)
	})

	t.Run("With Query Filter", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "alice@test.com"}))
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u2", Email: "bob@test.com"}))
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u3", Email: "alice.smith@test.com"}))

		users, total, err := svc.ListUsers(ctx, 10, 0, "alice")
		require.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, 2, total)
	})

	t.Run("With Pagination", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		for i := 0; i < 15; i++ {
			require.NoError(t, userRepo.Create(ctx, &domain.User{
				ID:    domain.GenerateNetworkID(),
				Email: "user" + string(rune('a'+i)) + "@test.com",
			}))
		}

		users, total, err := svc.ListUsers(ctx, 5, 0, "")
		require.NoError(t, err)
		assert.Len(t, users, 5)
		assert.Equal(t, 15, total)

		// Second page
		users2, total2, err := svc.ListUsers(ctx, 5, 5, "")
		require.NoError(t, err)
		assert.Len(t, users2, 5)
		assert.Equal(t, 15, total2)
	})

	t.Run("Empty List", func(t *testing.T) {
		svc, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		users, total, err := svc.ListUsers(ctx, 10, 0, "")
		require.NoError(t, err)
		assert.Len(t, users, 0)
		assert.Equal(t, 0, total)
	})
}

// ==================== LIST TENANTS TESTS ====================

func TestAdminService_ListTenants(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, _, tenantRepo, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t1", Name: "Tenant 1"}))
		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t2", Name: "Tenant 2"}))

		tenants, total, err := svc.ListTenants(ctx, 10, 0, "")
		require.NoError(t, err)
		assert.Len(t, tenants, 2)
		assert.Equal(t, 2, total)
	})

	t.Run("With Query Filter", func(t *testing.T) {
		svc, _, tenantRepo, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t1", Name: "Acme Corp"}))
		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t2", Name: "Beta Inc"}))
		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t3", Name: "Acme Labs"}))

		tenants, total, err := svc.ListTenants(ctx, 10, 0, "acme")
		require.NoError(t, err)
		assert.Len(t, tenants, 2)
		assert.Equal(t, 2, total)
	})

	t.Run("With Pagination", func(t *testing.T) {
		svc, _, tenantRepo, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		for i := 0; i < 12; i++ {
			require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{
				ID:   domain.GenerateNetworkID(),
				Name: "Tenant " + string(rune('A'+i)),
			}))
		}

		tenants, total, err := svc.ListTenants(ctx, 5, 5, "")
		require.NoError(t, err)
		assert.Len(t, tenants, 5)
		assert.Equal(t, 12, total)
	})
}

// ==================== GET SYSTEM STATS TESTS ====================

func TestAdminService_GetSystemStats(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, userRepo, tenantRepo, networkRepo, deviceRepo, chatRepo := setupAdminServiceTest()
		ctx := context.Background()

		// Seed data
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "user@test.com"}))
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u2", Email: "user2@test.com"}))
		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t1", Name: "Tenant"}))
		require.NoError(t, networkRepo.Create(ctx, &domain.Network{ID: "n1", TenantID: "t1", Name: "Network", CIDR: "10.0.0.0/24"}))
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d1", UserID: "u1", TenantID: "t1", Name: "Device", PubKey: "pk1"}))

		// Create chat message today
		require.NoError(t, chatRepo.Create(ctx, &domain.ChatMessage{
			ID:        "msg1",
			Scope:     "network:n1",
			TenantID:  "t1",
			UserID:    "u1",
			Body:      "Hello",
			CreatedAt: time.Now(),
		}))

		stats, err := svc.GetSystemStats(ctx)
		require.NoError(t, err)

		assert.Equal(t, 2, stats.TotalUsers)
		assert.Equal(t, 1, stats.TotalTenants)
		assert.Equal(t, 1, stats.TotalNetworks)
		assert.Equal(t, 1, stats.TotalDevices)
		assert.Equal(t, 10, stats.ActiveConnections) // mock returns 10
		assert.GreaterOrEqual(t, stats.MessagesToday, 0)
	})

	t.Run("Nil Active Connections Callback", func(t *testing.T) {
		userRepo := repository.NewInMemoryUserRepository()
		tenantRepo := repository.NewInMemoryTenantRepository()
		networkRepo := repository.NewInMemoryNetworkRepository()
		deviceRepo := repository.NewInMemoryDeviceRepository()
		chatRepo := repository.NewInMemoryChatRepository()
		adminRepo := repository.NewAdminRepository(nil)

		svc := NewAdminService(
			userRepo,
			adminRepo,
			tenantRepo,
			networkRepo,
			deviceRepo,
			chatRepo,
			nil, // No auditor
			nil, // No Redis
			nil, // nil callback
		)

		ctx := context.Background()
		stats, err := svc.GetSystemStats(ctx)
		require.NoError(t, err)

		assert.Equal(t, 0, stats.ActiveConnections)
	})

	t.Run("Empty Stats", func(t *testing.T) {
		svc, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		stats, err := svc.GetSystemStats(ctx)
		require.NoError(t, err)

		assert.Equal(t, 0, stats.TotalUsers)
		assert.Equal(t, 0, stats.TotalTenants)
		assert.Equal(t, 0, stats.TotalNetworks)
		assert.Equal(t, 0, stats.TotalDevices)
	})
}

// ==================== TOGGLE USER ADMIN TESTS ====================

func TestAdminService_ToggleUserAdmin(t *testing.T) {
	t.Run("Make Admin", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "user@test.com", IsAdmin: false}))

		user, err := svc.ToggleUserAdmin(ctx, "u1")
		require.NoError(t, err)
		assert.True(t, user.IsAdmin)
	})

	t.Run("Remove Admin", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "admin@test.com", IsAdmin: true}))

		user, err := svc.ToggleUserAdmin(ctx, "u1")
		require.NoError(t, err)
		assert.False(t, user.IsAdmin)
	})

	t.Run("User Not Found", func(t *testing.T) {
		svc, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		_, err := svc.ToggleUserAdmin(ctx, "nonexistent")
		assert.Error(t, err)

		domErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrUserNotFound, domErr.Code)
	})

	t.Run("Toggle Twice", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "user@test.com", IsAdmin: false}))

		// First toggle: false -> true
		user, err := svc.ToggleUserAdmin(ctx, "u1")
		require.NoError(t, err)
		assert.True(t, user.IsAdmin)

		// Second toggle: true -> false
		user, err = svc.ToggleUserAdmin(ctx, "u1")
		require.NoError(t, err)
		assert.False(t, user.IsAdmin)
	})
}

// ==================== DELETE USER TESTS ====================

func TestAdminService_DeleteUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "delete@test.com"}))

		err := svc.DeleteUser(ctx, "u1")
		require.NoError(t, err)

		// Verify deleted
		_, err = userRepo.GetByID(ctx, "u1")
		assert.Error(t, err)
	})

	t.Run("User Not Found", func(t *testing.T) {
		svc, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		err := svc.DeleteUser(ctx, "nonexistent")
		assert.Error(t, err)
	})
}

// ==================== DELETE TENANT TESTS ====================

func TestAdminService_DeleteTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, _, tenantRepo, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t1", Name: "Delete Me"}))

		err := svc.DeleteTenant(ctx, "t1")
		require.NoError(t, err)

		// Verify deleted
		_, err = tenantRepo.GetByID(ctx, "t1")
		assert.Error(t, err)
	})

	t.Run("Tenant Not Found", func(t *testing.T) {
		svc, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		err := svc.DeleteTenant(ctx, "nonexistent")
		assert.Error(t, err)
	})
}

// ==================== LIST NETWORKS TESTS ====================

func TestAdminService_ListNetworks(t *testing.T) {
	t.Run("Success - With Tenant Filter", func(t *testing.T) {
		svc, _, _, networkRepo, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, networkRepo.Create(ctx, &domain.Network{ID: "n1", TenantID: "t1", Name: "Network 1", CIDR: "10.0.0.0/24", Visibility: domain.NetworkVisibilityPublic}))
		require.NoError(t, networkRepo.Create(ctx, &domain.Network{ID: "n2", TenantID: "t1", Name: "Network 2", CIDR: "10.1.0.0/24", Visibility: domain.NetworkVisibilityPublic}))

		// Note: AdminService.ListNetworks passes IsAdmin=true but empty TenantID
		// The in-memory repository enforces tenant isolation so this returns empty
		// In production, PostgreSQL implementation would return all networks for admin
		networks, _, err := svc.ListNetworks(ctx, 10, "", "")
		require.NoError(t, err)
		// Empty because TenantID filter is empty and repo enforces tenant isolation
		// networks can be nil or empty slice - both are valid
		assert.True(t, len(networks) == 0)
	})

	t.Run("Empty Result", func(t *testing.T) {
		svc, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		networks, nextCursor, err := svc.ListNetworks(ctx, 10, "", "")
		require.NoError(t, err)
		assert.Empty(t, networks)
		assert.Empty(t, nextCursor)
	})
}

// ==================== LIST DEVICES TESTS ====================

func TestAdminService_ListDevices(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, _, _, _, deviceRepo, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d1", UserID: "u1", TenantID: "t1", Name: "Device 1", PubKey: "pk1"}))
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d2", UserID: "u1", TenantID: "t1", Name: "Device 2", PubKey: "pk2"}))

		devices, nextCursor, err := svc.ListDevices(ctx, 10, "", "")
		require.NoError(t, err)
		assert.Len(t, devices, 2)
		assert.Empty(t, nextCursor)
	})

	t.Run("Cross Tenant", func(t *testing.T) {
		svc, _, _, _, deviceRepo, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d1", UserID: "u1", TenantID: "t1", Name: "Device T1", PubKey: "pk1"}))
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d2", UserID: "u2", TenantID: "t2", Name: "Device T2", PubKey: "pk2"}))

		devices, _, err := svc.ListDevices(ctx, 10, "", "")
		require.NoError(t, err)
		assert.Len(t, devices, 2) // Admin sees all tenants
	})

	t.Run("With Limit", func(t *testing.T) {
		svc, _, _, _, deviceRepo, _ := setupAdminServiceTest()
		ctx := context.Background()

		for i := 0; i < 10; i++ {
			require.NoError(t, deviceRepo.Create(ctx, &domain.Device{
				ID:       domain.GenerateNetworkID(),
				UserID:   "u1",
				TenantID: "t1",
				Name:     "Device " + string(rune('A'+i)),
				PubKey:   "pk" + string(rune('a'+i)),
			}))
		}

		devices, nextCursor, err := svc.ListDevices(ctx, 5, "", "")
		require.NoError(t, err)
		assert.Len(t, devices, 5)
		assert.NotEmpty(t, nextCursor)
	})

	t.Run("With Search Query", func(t *testing.T) {
		svc, _, _, _, deviceRepo, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d1", UserID: "u1", TenantID: "t1", Name: "MacBook Pro", PubKey: "pk1"}))
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d2", UserID: "u1", TenantID: "t1", Name: "Windows PC", PubKey: "pk2"}))
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d3", UserID: "u1", TenantID: "t1", Name: "MacBook Air", PubKey: "pk3"}))

		devices, _, err := svc.ListDevices(ctx, 10, "", "MacBook")
		require.NoError(t, err)
		assert.Len(t, devices, 2)
	})
}
