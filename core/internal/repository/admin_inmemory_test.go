package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryAdminRepository_ListAllUsers(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryAdminRepository()

	// Add test users
	username1 := "admin1"
	username2 := "mod1"
	username3 := "user1"
	users := []*domain.User{
		{ID: "u1", Email: "admin@test.com", Username: &username1, TenantID: "t1", IsAdmin: true, IsModerator: false, Suspended: false},
		{ID: "u2", Email: "mod@test.com", Username: &username2, TenantID: "t1", IsAdmin: false, IsModerator: true, Suspended: false},
		{ID: "u3", Email: "user@test.com", Username: &username3, TenantID: "t2", IsAdmin: false, IsModerator: false, Suspended: true},
	}
	for _, u := range users {
		repo.AddUser(u)
	}

	t.Run("List all without filters", func(t *testing.T) {
		list, total, err := repo.ListAllUsers(ctx, domain.UserFilters{}, domain.PaginationParams{})
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, list, 3)
	})

	t.Run("Filter by role admin", func(t *testing.T) {
		list, total, err := repo.ListAllUsers(ctx, domain.UserFilters{Role: "admin"}, domain.PaginationParams{})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, list, 1)
		assert.Equal(t, "u1", list[0].ID)
	})

	t.Run("Filter by role moderator", func(t *testing.T) {
		list, total, err := repo.ListAllUsers(ctx, domain.UserFilters{Role: "moderator"}, domain.PaginationParams{})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, list, 1)
		assert.Equal(t, "u2", list[0].ID)
	})

	t.Run("Filter by role user", func(t *testing.T) {
		list, total, err := repo.ListAllUsers(ctx, domain.UserFilters{Role: "user"}, domain.PaginationParams{})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, list, 1)
		assert.Equal(t, "u3", list[0].ID)
	})

	t.Run("Filter by status active", func(t *testing.T) {
		list, total, err := repo.ListAllUsers(ctx, domain.UserFilters{Status: "active"}, domain.PaginationParams{})
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, list, 2)
	})

	t.Run("Filter by status suspended", func(t *testing.T) {
		list, total, err := repo.ListAllUsers(ctx, domain.UserFilters{Status: "suspended"}, domain.PaginationParams{})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, list, 1)
		assert.Equal(t, "u3", list[0].ID)
	})

	t.Run("Filter by tenant", func(t *testing.T) {
		list, total, err := repo.ListAllUsers(ctx, domain.UserFilters{TenantID: "t1"}, domain.PaginationParams{})
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, list, 2)
	})

	t.Run("Search by email", func(t *testing.T) {
		list, total, err := repo.ListAllUsers(ctx, domain.UserFilters{Search: "admin@"}, domain.PaginationParams{})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, list, 1)
		assert.Equal(t, "u1", list[0].ID)
	})

	t.Run("Search by username", func(t *testing.T) {
		list, total, err := repo.ListAllUsers(ctx, domain.UserFilters{Search: "mod1"}, domain.PaginationParams{})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, list, 1)
		assert.Equal(t, "u2", list[0].ID)
	})

	t.Run("Pagination", func(t *testing.T) {
		list, total, err := repo.ListAllUsers(ctx, domain.UserFilters{}, domain.PaginationParams{Page: 1, PerPage: 2})
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, list, 2)
	})

	t.Run("Pagination beyond results", func(t *testing.T) {
		list, total, err := repo.ListAllUsers(ctx, domain.UserFilters{}, domain.PaginationParams{Page: 10, PerPage: 2})
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, list, 0)
	})
}

func TestInMemoryAdminRepository_GetUserStats(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryAdminRepository()

	// Add test users
	username1 := "admin"
	users := []*domain.User{
		{ID: "u1", Email: "admin@test.com", Username: &username1, IsAdmin: true, IsModerator: false, Suspended: false},
		{ID: "u2", Email: "mod@test.com", IsAdmin: false, IsModerator: true, Suspended: false},
		{ID: "u3", Email: "user@test.com", IsAdmin: false, IsModerator: false, Suspended: true},
		{ID: "u4", Email: "user2@test.com", IsAdmin: false, IsModerator: false, Suspended: false},
	}
	for _, u := range users {
		repo.AddUser(u)
	}

	stats, err := repo.GetUserStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, stats.TotalUsers)
	assert.Equal(t, 1, stats.AdminUsers)
	assert.Equal(t, 1, stats.ModeratorUsers)
	assert.Equal(t, 1, stats.SuspendedUsers)
}

func TestInMemoryAdminRepository_UpdateUserRole(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryAdminRepository()

	// Add test user
	repo.AddUser(&domain.User{ID: "u1", Email: "test@test.com", IsAdmin: false, IsModerator: false})

	t.Run("Update to admin", func(t *testing.T) {
		isAdmin := true
		err := repo.UpdateUserRole(ctx, "u1", &isAdmin, nil)
		require.NoError(t, err)

		user, _ := repo.GetUserByID(ctx, "u1")
		assert.True(t, user.IsAdmin)
	})

	t.Run("Update to moderator", func(t *testing.T) {
		isMod := true
		err := repo.UpdateUserRole(ctx, "u1", nil, &isMod)
		require.NoError(t, err)

		user, _ := repo.GetUserByID(ctx, "u1")
		assert.True(t, user.IsModerator)
	})

	t.Run("User not found", func(t *testing.T) {
		isAdmin := true
		err := repo.UpdateUserRole(ctx, "non-existent", &isAdmin, nil)
		assert.Error(t, err)
	})
}

func TestInMemoryAdminRepository_SuspendUser(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryAdminRepository()

	repo.AddUser(&domain.User{ID: "u1", Email: "test@test.com", Suspended: false})

	t.Run("Suspend user successfully", func(t *testing.T) {
		err := repo.SuspendUser(ctx, "u1", "Policy violation", "admin1")
		require.NoError(t, err)

		user, _ := repo.GetUserByID(ctx, "u1")
		assert.True(t, user.Suspended)
		assert.NotNil(t, user.SuspendedAt)
		assert.Equal(t, "Policy violation", *user.SuspendedReason)
		assert.Equal(t, "admin1", *user.SuspendedBy)
	})

	t.Run("User not found", func(t *testing.T) {
		err := repo.SuspendUser(ctx, "non-existent", "reason", "admin")
		assert.Error(t, err)
	})
}

func TestInMemoryAdminRepository_UnsuspendUser(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryAdminRepository()

	now := time.Now()
	reason := "test"
	suspendedBy := "admin"
	repo.AddUser(&domain.User{
		ID:              "u1",
		Email:           "test@test.com",
		Suspended:       true,
		SuspendedAt:     &now,
		SuspendedReason: &reason,
		SuspendedBy:     &suspendedBy,
	})

	t.Run("Unsuspend user successfully", func(t *testing.T) {
		err := repo.UnsuspendUser(ctx, "u1")
		require.NoError(t, err)

		user, _ := repo.GetUserByID(ctx, "u1")
		assert.False(t, user.Suspended)
		assert.Nil(t, user.SuspendedAt)
		assert.Nil(t, user.SuspendedReason)
		assert.Nil(t, user.SuspendedBy)
	})

	t.Run("User not found", func(t *testing.T) {
		err := repo.UnsuspendUser(ctx, "non-existent")
		assert.Error(t, err)
	})
}

func TestInMemoryAdminRepository_GetUserByID(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryAdminRepository()

	repo.AddUser(&domain.User{ID: "u1", Email: "test@test.com"})

	t.Run("Get existing user", func(t *testing.T) {
		user, err := repo.GetUserByID(ctx, "u1")
		require.NoError(t, err)
		assert.Equal(t, "u1", user.ID)
		assert.Equal(t, "test@test.com", user.Email)
	})

	t.Run("User not found", func(t *testing.T) {
		_, err := repo.GetUserByID(ctx, "non-existent")
		assert.Error(t, err)
	})
}

func TestInMemoryAdminRepository_UpdateLastSeen(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryAdminRepository()

	repo.AddUser(&domain.User{ID: "u1", Email: "test@test.com"})

	t.Run("Update last seen successfully", func(t *testing.T) {
		err := repo.UpdateLastSeen(ctx, "u1")
		require.NoError(t, err)

		// Verify via ListAllUsers
		list, _, _ := repo.ListAllUsers(ctx, domain.UserFilters{}, domain.PaginationParams{})
		for _, item := range list {
			if item.ID == "u1" {
				assert.NotNil(t, item.LastSeen)
			}
		}
	})

	t.Run("User not found", func(t *testing.T) {
		err := repo.UpdateLastSeen(ctx, "non-existent")
		assert.Error(t, err)
	})
}
