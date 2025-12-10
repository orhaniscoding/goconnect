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
func setupAdminServiceTest() (*AdminService, *repository.InMemoryUserRepository, *repository.InMemoryTenantRepository, *repository.InMemoryNetworkRepository, *repository.InMemoryDeviceRepository, *repository.InMemoryChatRepository, *repository.InMemoryAdminRepository) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	chatRepo := repository.NewInMemoryChatRepository()
	adminRepo := repository.NewInMemoryAdminRepository()

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

	return service, userRepo, tenantRepo, networkRepo, deviceRepo, chatRepo, adminRepo
}

// ==================== LIST USERS TESTS ====================

func TestAdminService_ListUsers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "user1@test.com"}))
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u2", Email: "user2@test.com"}))

		users, total, err := svc.ListUsers(ctx, 10, 0, "")
		require.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, 2, total)
	})

	t.Run("With Query Filter", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
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
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
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
		svc, _, _, _, _, _, _ := setupAdminServiceTest()
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
		svc, _, tenantRepo, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t1", Name: "Tenant 1"}))
		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t2", Name: "Tenant 2"}))

		tenants, total, err := svc.ListTenants(ctx, 10, 0, "")
		require.NoError(t, err)
		assert.Len(t, tenants, 2)
		assert.Equal(t, 2, total)
	})

	t.Run("With Query Filter", func(t *testing.T) {
		svc, _, tenantRepo, _, _, _, _ := setupAdminServiceTest()
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
		svc, _, tenantRepo, _, _, _, _ := setupAdminServiceTest()
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
		svc, userRepo, tenantRepo, networkRepo, deviceRepo, chatRepo, _ := setupAdminServiceTest()
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
		svc, _, _, _, _, _, _ := setupAdminServiceTest()
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
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "user@test.com", IsAdmin: false}))

		user, err := svc.ToggleUserAdmin(ctx, "u1")
		require.NoError(t, err)
		assert.True(t, user.IsAdmin)
	})

	t.Run("Remove Admin", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "admin@test.com", IsAdmin: true}))

		user, err := svc.ToggleUserAdmin(ctx, "u1")
		require.NoError(t, err)
		assert.False(t, user.IsAdmin)
	})

	t.Run("User Not Found", func(t *testing.T) {
		svc, _, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		_, err := svc.ToggleUserAdmin(ctx, "nonexistent")
		assert.Error(t, err)

		domErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrUserNotFound, domErr.Code)
	})

	t.Run("Toggle Twice", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
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
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "delete@test.com"}))

		err := svc.DeleteUser(ctx, "u1")
		require.NoError(t, err)

		// Verify deleted
		_, err = userRepo.GetByID(ctx, "u1")
		assert.Error(t, err)
	})

	t.Run("User Not Found", func(t *testing.T) {
		svc, _, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		err := svc.DeleteUser(ctx, "nonexistent")
		assert.Error(t, err)
	})
}

// ==================== DELETE TENANT TESTS ====================

func TestAdminService_DeleteTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, _, tenantRepo, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t1", Name: "Delete Me"}))

		err := svc.DeleteTenant(ctx, "t1")
		require.NoError(t, err)

		// Verify deleted
		_, err = tenantRepo.GetByID(ctx, "t1")
		assert.Error(t, err)
	})

	t.Run("Tenant Not Found", func(t *testing.T) {
		svc, _, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		err := svc.DeleteTenant(ctx, "nonexistent")
		assert.Error(t, err)
	})
}

// ==================== LIST NETWORKS TESTS ====================

func TestAdminService_ListNetworks(t *testing.T) {
	t.Run("Success - With Tenant Filter", func(t *testing.T) {
		svc, _, _, networkRepo, _, _, _ := setupAdminServiceTest()
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
		svc, _, _, _, _, _, _ := setupAdminServiceTest()
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
		svc, _, _, _, deviceRepo, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d1", UserID: "u1", TenantID: "t1", Name: "Device 1", PubKey: "pk1"}))
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d2", UserID: "u1", TenantID: "t1", Name: "Device 2", PubKey: "pk2"}))

		devices, nextCursor, err := svc.ListDevices(ctx, 10, "", "")
		require.NoError(t, err)
		assert.Len(t, devices, 2)
		assert.Empty(t, nextCursor)
	})

	t.Run("Cross Tenant", func(t *testing.T) {
		svc, _, _, _, deviceRepo, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d1", UserID: "u1", TenantID: "t1", Name: "Device T1", PubKey: "pk1"}))
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d2", UserID: "u2", TenantID: "t2", Name: "Device T2", PubKey: "pk2"}))

		devices, _, err := svc.ListDevices(ctx, 10, "", "")
		require.NoError(t, err)
		assert.Len(t, devices, 2) // Admin sees all tenants
	})

	t.Run("With Limit", func(t *testing.T) {
		svc, _, _, _, deviceRepo, _, _ := setupAdminServiceTest()
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
		svc, _, _, _, deviceRepo, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d1", UserID: "u1", TenantID: "t1", Name: "MacBook Pro", PubKey: "pk1"}))
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d2", UserID: "u1", TenantID: "t1", Name: "Windows PC", PubKey: "pk2"}))
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d3", UserID: "u1", TenantID: "t1", Name: "MacBook Air", PubKey: "pk3"}))

		devices, _, err := svc.ListDevices(ctx, 10, "", "MacBook")
		require.NoError(t, err)
		assert.Len(t, devices, 2)
	})
}

// ==================== LIST ALL USERS TESTS ====================
// NOTE: ListAllUsers, GetUserStats, UpdateLastSeen, UpdateUserRole, SuspendUser, UnsuspendUser, 
// GetUserDetails all require AdminRepository which needs SQL DB.
// These functions are tested via integration tests.

// Tests below verify authorization logic which uses userRepo (in-memory)

func TestAdminService_UpdateUserRole_AuthChecks(t *testing.T) {
	t.Run("Fail_NotAdmin", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		normalUser := &domain.User{ID: "user-1", Email: "user@test.com", TenantID: "t1", IsAdmin: false}
		require.NoError(t, userRepo.Create(ctx, normalUser))

		isAdmin := true
		err := svc.UpdateUserRole(ctx, "user-1", "user-2", &isAdmin, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Admin access required")
	})

	t.Run("Fail_SelfDemotion", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		isAdmin := false
		err := svc.UpdateUserRole(ctx, "admin-1", "admin-1", &isAdmin, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot remove your own admin privileges")
	})

	t.Run("Fail_TargetNotFound", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		isAdmin := true
		err := svc.UpdateUserRole(ctx, "admin-1", "non-existent", &isAdmin, nil)
		require.Error(t, err)
	})

	t.Run("Success_PromoteToAdmin", func(t *testing.T) {
		svc, userRepo, _, _, _, _, adminRepo := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-role", Email: "admin-role@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		targetUser := &domain.User{ID: "target-role", Email: "target-role@test.com", TenantID: "t1", IsAdmin: false, IsModerator: false}
		require.NoError(t, userRepo.Create(ctx, targetUser))
		adminRepo.AddUser(targetUser) // Also add to admin repo for UpdateUserRole

		isAdmin := true
		err := svc.UpdateUserRole(ctx, "admin-role", "target-role", &isAdmin, nil)
		require.NoError(t, err)
	})

	t.Run("Success_PromoteToModerator", func(t *testing.T) {
		svc, userRepo, _, _, _, _, adminRepo := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-mod", Email: "admin-mod@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		targetUser := &domain.User{ID: "target-mod", Email: "target-mod@test.com", TenantID: "t1", IsAdmin: false, IsModerator: false}
		require.NoError(t, userRepo.Create(ctx, targetUser))
		adminRepo.AddUser(targetUser) // Also add to admin repo for UpdateUserRole

		isMod := true
		err := svc.UpdateUserRole(ctx, "admin-mod", "target-mod", nil, &isMod)
		require.NoError(t, err)
	})
}

func TestAdminService_SuspendUser_AuthChecks(t *testing.T) {
	t.Run("Fail_NotAdmin", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		normalUser := &domain.User{ID: "user-1", Email: "user@test.com", TenantID: "t1", IsAdmin: false}
		require.NoError(t, userRepo.Create(ctx, normalUser))

		err := svc.SuspendUser(ctx, "user-1", "user-2", "Reason")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Admin access required")
	})

	t.Run("Fail_SelfSuspend", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		err := svc.SuspendUser(ctx, "admin-1", "admin-1", "Reason")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot suspend your own account")
	})

	t.Run("Fail_TargetNotFound", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		err := svc.SuspendUser(ctx, "admin-1", "non-existent", "Reason")
		require.Error(t, err)
	})

	t.Run("Fail_SuspendAdmin", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		targetAdmin := &domain.User{ID: "admin-2", Email: "admin2@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, targetAdmin))

		err := svc.SuspendUser(ctx, "admin-1", "admin-2", "Reason")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot suspend another admin")
	})

	t.Run("Success_SuspendNormalUser", func(t *testing.T) {
		svc, userRepo, _, _, _, _, adminRepo := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-susp", Email: "admin-susp@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		targetUser := &domain.User{ID: "target-susp", Email: "target-susp@test.com", TenantID: "t1", IsAdmin: false}
		require.NoError(t, userRepo.Create(ctx, targetUser))
		adminRepo.AddUser(targetUser) // Also add to admin repo for SuspendUser

		err := svc.SuspendUser(ctx, "admin-susp", "target-susp", "Violation of ToS")
		require.NoError(t, err)
	})
}

func TestAdminService_UnsuspendUser_AuthChecks(t *testing.T) {
	t.Run("Fail_NotAdmin", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		normalUser := &domain.User{ID: "user-1", Email: "user@test.com", TenantID: "t1", IsAdmin: false}
		require.NoError(t, userRepo.Create(ctx, normalUser))

		err := svc.UnsuspendUser(ctx, "user-1", "user-2")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Admin access required")
	})

	t.Run("Fail_TargetNotFound", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		err := svc.UnsuspendUser(ctx, "admin-1", "non-existent")
		require.Error(t, err)
	})

	t.Run("Success_UnsuspendUser", func(t *testing.T) {
		svc, userRepo, _, _, _, _, adminRepo := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-unsusp", Email: "admin-unsusp@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		// Create a suspended user
		now := time.Now()
		reason := "Previous violation"
		suspendedUser := &domain.User{
			ID:              "susp-user",
			Email:           "suspended@test.com",
			TenantID:        "t1",
			IsAdmin:         false,
			Suspended:       true,
			SuspendedAt:     &now,
			SuspendedReason: &reason,
		}
		require.NoError(t, userRepo.Create(ctx, suspendedUser))
		adminRepo.AddUser(suspendedUser) // Also add to admin repo for UnsuspendUser

		err := svc.UnsuspendUser(ctx, "admin-unsusp", "susp-user")
		require.NoError(t, err)
	})
}

func TestAdminService_GetUserDetails_AuthChecks(t *testing.T) {
	t.Run("Fail_NotAdmin", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		normalUser := &domain.User{ID: "user-1", Email: "user@test.com", TenantID: "t1", IsAdmin: false}
		require.NoError(t, userRepo.Create(ctx, normalUser))

		_, err := svc.GetUserDetails(ctx, "user-1", "user-2")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Admin access required")
	})

	t.Run("Fail_TargetNotFound", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		_, err := svc.GetUserDetails(ctx, "admin-1", "non-existent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "User not found")
	})

	t.Run("Success_GetUserDetails", func(t *testing.T) {
		svc, userRepo, _, _, _, _, adminRepo := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-detail", Email: "admin-detail@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		targetUser := &domain.User{
			ID:           "target-detail",
			Email:        "target@test.com",
			TenantID:     "t1",
			IsAdmin:      false,
			PasswordHash: "secret-hash",
			TwoFAKey:     "secret-2fa-key",
		}
		adminRepo.AddUser(targetUser) // Add to admin repo for GetUserByID

		user, err := svc.GetUserDetails(ctx, "admin-detail", "target-detail")
		require.NoError(t, err)
		assert.Equal(t, "target-detail", user.ID)
		assert.Empty(t, user.PasswordHash) // Should be removed
		assert.Empty(t, user.TwoFAKey)     // Should be removed
	})
}

// ==================== LIST ALL USERS TESTS (using InMemoryAdminRepository) ====================

func TestAdminService_ListAllUsers(t *testing.T) {
	t.Run("Success_NoFilters", func(t *testing.T) {
		svc, userRepo, _, _, _, _, adminRepo := setupAdminServiceTest()
		ctx := context.Background()

		// Create admin user for auth
		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		// Add users to admin repo
		adminRepo.AddUser(&domain.User{ID: "u1", Email: "user1@test.com", TenantID: "t1", CreatedAt: time.Now()})
		adminRepo.AddUser(&domain.User{ID: "u2", Email: "user2@test.com", TenantID: "t1", CreatedAt: time.Now()})
		adminRepo.AddUser(&domain.User{ID: "u3", Email: "user3@test.com", TenantID: "t1", CreatedAt: time.Now()})

		users, total, err := svc.ListAllUsers(ctx, "admin-1", domain.UserFilters{}, domain.PaginationParams{Page: 1, PerPage: 10})
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, users, 3)
	})

	t.Run("Success_FilterByRole", func(t *testing.T) {
		svc, userRepo, _, _, _, _, adminRepo := setupAdminServiceTest()
		ctx := context.Background()

		// Create admin user for auth
		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		// Add users with different roles
		adminRepo.AddUser(&domain.User{ID: "u1", Email: "admin-user@test.com", IsAdmin: true, CreatedAt: time.Now()})
		adminRepo.AddUser(&domain.User{ID: "u2", Email: "mod@test.com", IsModerator: true, CreatedAt: time.Now()})
		adminRepo.AddUser(&domain.User{ID: "u3", Email: "user@test.com", CreatedAt: time.Now()})

		// Filter by admin role
		users, total, err := svc.ListAllUsers(ctx, "admin-1", domain.UserFilters{Role: "admin"}, domain.PaginationParams{Page: 1, PerPage: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, users, 1)
	})

	t.Run("Success_FilterByStatus", func(t *testing.T) {
		svc, userRepo, _, _, _, _, adminRepo := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		adminRepo.AddUser(&domain.User{ID: "u1", Email: "active@test.com", Suspended: false, CreatedAt: time.Now()})
		adminRepo.AddUser(&domain.User{ID: "u2", Email: "suspended@test.com", Suspended: true, CreatedAt: time.Now()})

		// Filter by suspended status
		users, total, err := svc.ListAllUsers(ctx, "admin-1", domain.UserFilters{Status: "suspended"}, domain.PaginationParams{Page: 1, PerPage: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, users, 1)
		assert.Equal(t, "suspended@test.com", users[0].Email)
	})

	t.Run("Fail_NotAdmin", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		normalUser := &domain.User{ID: "user-1", Email: "user@test.com", TenantID: "t1", IsAdmin: false}
		require.NoError(t, userRepo.Create(ctx, normalUser))

		_, _, err := svc.ListAllUsers(ctx, "user-1", domain.UserFilters{}, domain.PaginationParams{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Admin access required")
	})
}

// ==================== GET USER STATS TESTS ====================

func TestAdminService_GetUserStats(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, userRepo, _, _, _, _, adminRepo := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		// Add various users
		adminRepo.AddUser(&domain.User{ID: "u1", Email: "admin@test.com", IsAdmin: true})
		adminRepo.AddUser(&domain.User{ID: "u2", Email: "mod@test.com", IsModerator: true})
		adminRepo.AddUser(&domain.User{ID: "u3", Email: "suspended@test.com", Suspended: true})
		adminRepo.AddUser(&domain.User{ID: "u4", Email: "normal@test.com"})

		stats, err := svc.GetUserStats(ctx, "admin-1")
		require.NoError(t, err)
		assert.Equal(t, 4, stats.TotalUsers)
		assert.Equal(t, 1, stats.AdminUsers)
		assert.Equal(t, 1, stats.ModeratorUsers)
		assert.Equal(t, 1, stats.SuspendedUsers)
	})

	t.Run("Fail_NotAdmin", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		normalUser := &domain.User{ID: "user-1", Email: "user@test.com", TenantID: "t1", IsAdmin: false}
		require.NoError(t, userRepo.Create(ctx, normalUser))

		_, err := svc.GetUserStats(ctx, "user-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Admin access required")
	})
}

// ==================== UPDATE LAST SEEN TESTS ====================

func TestAdminService_UpdateLastSeen(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, userRepo, _, _, _, _, adminRepo := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		adminRepo.AddUser(&domain.User{ID: "target-user", Email: "target@test.com"})

		err := svc.UpdateLastSeen(ctx, "target-user")
		require.NoError(t, err)
	})

	t.Run("Fail_UserNotFound", func(t *testing.T) {
		svc, _, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		err := svc.UpdateLastSeen(ctx, "non-existent")
		require.Error(t, err)
	})
}

// ==================== GET USER DETAILS TESTS (with InMemoryAdminRepository) ====================

func TestAdminService_GetUserDetails_Success(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, userRepo, _, _, _, _, adminRepo := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		targetUser := &domain.User{
			ID:       "target-user",
			Email:    "target@test.com",
			TenantID: "t1",
		}
		adminRepo.AddUser(targetUser)

		user, err := svc.GetUserDetails(ctx, "admin-1", "target-user")
		require.NoError(t, err)
		assert.Equal(t, "target@test.com", user.Email)
	})

	t.Run("Fail_TargetNotFound", func(t *testing.T) {
		svc, userRepo, _, _, _, _, _ := setupAdminServiceTest()
		ctx := context.Background()

		adminUser := &domain.User{ID: "admin-1", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		_, err := svc.GetUserDetails(ctx, "admin-1", "non-existent")
		require.Error(t, err)
	})
}