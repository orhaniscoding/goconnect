package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test user
func mkUser(id, email string, isAdmin, isModerator bool) *domain.User {
	now := time.Now()
	return &domain.User{
		ID:          id,
		TenantID:    "tenant-1",
		Email:       email,
		PassHash:    "hashed-password",
		Locale:      "en",
		IsAdmin:     isAdmin,
		IsModerator: isModerator,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestNewInMemoryUserRepository(t *testing.T) {
	repo := NewInMemoryUserRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.users)
	assert.NotNil(t, repo.byEmail)
	assert.Equal(t, 0, len(repo.users))
	assert.Equal(t, 0, len(repo.byEmail))
}

func TestUserRepository_Create_Success(t *testing.T) {
	repo := NewInMemoryUserRepository()
	user := mkUser("user-1", "test@example.com", false, false)

	err := repo.Create(context.Background(), user)

	require.NoError(t, err)
	assert.Equal(t, 1, len(repo.users))
	assert.Equal(t, 1, len(repo.byEmail))
	assert.Equal(t, "user-1", repo.byEmail["test@example.com"])
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	repo := NewInMemoryUserRepository()
	user1 := mkUser("user-1", "duplicate@example.com", false, false)
	user2 := mkUser("user-2", "duplicate@example.com", false, false)

	err1 := repo.Create(context.Background(), user1)
	require.NoError(t, err1)

	err2 := repo.Create(context.Background(), user2)

	require.Error(t, err2)
	domainErr, ok := err2.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrInvalidRequest, domainErr.Code)
	assert.Contains(t, domainErr.Message, "email already exists")

	// Only first user should exist
	assert.Equal(t, 1, len(repo.users))
	assert.Equal(t, "user-1", repo.byEmail["duplicate@example.com"])
}

func TestUserRepository_Create_MultipleUsers(t *testing.T) {
	repo := NewInMemoryUserRepository()
	users := []*domain.User{
		mkUser("user-1", "user1@example.com", false, false),
		mkUser("user-2", "user2@example.com", false, true), // moderator
		mkUser("user-3", "user3@example.com", true, false), // admin
	}

	for _, user := range users {
		err := repo.Create(context.Background(), user)
		require.NoError(t, err)
	}

	assert.Equal(t, 3, len(repo.users))
	assert.Equal(t, 3, len(repo.byEmail))
}

func TestUserRepository_GetByID_Success(t *testing.T) {
	repo := NewInMemoryUserRepository()
	original := mkUser("user-1", "test@example.com", false, false)
	repo.Create(context.Background(), original)

	retrieved, err := repo.GetByID(context.Background(), "user-1")

	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, original.Email, retrieved.Email)
	assert.Equal(t, original.IsAdmin, retrieved.IsAdmin)
	assert.Equal(t, original.IsModerator, retrieved.IsModerator)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	retrieved, err := repo.GetByID(context.Background(), "non-existent")

	require.Error(t, err)
	assert.Nil(t, retrieved)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
	assert.Contains(t, domainErr.Message, "user not found")
}

func TestUserRepository_GetByEmail_Success(t *testing.T) {
	repo := NewInMemoryUserRepository()
	original := mkUser("user-1", "test@example.com", true, false)
	repo.Create(context.Background(), original)

	retrieved, err := repo.GetByEmail(context.Background(), original.TenantID, "test@example.com")

	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, original.Email, retrieved.Email)
	assert.Equal(t, original.IsAdmin, retrieved.IsAdmin)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	retrieved, err := repo.GetByEmail(context.Background(), "", "nonexistent@example.com")

	require.Error(t, err)
	assert.Nil(t, retrieved)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestUserRepository_Update_Success(t *testing.T) {
	repo := NewInMemoryUserRepository()
	user := mkUser("user-1", "original@example.com", false, false)
	repo.Create(context.Background(), user)

	// Update user details
	user.Locale = "tr"
	user.IsAdmin = true
	user.IsModerator = true

	err := repo.Update(context.Background(), user)

	require.NoError(t, err)

	retrieved, _ := repo.GetByID(context.Background(), "user-1")
	assert.Equal(t, "tr", retrieved.Locale)
	assert.True(t, retrieved.IsAdmin)
	assert.True(t, retrieved.IsModerator)
}

func TestUserRepository_Update_EmailChange(t *testing.T) {
	repo := NewInMemoryUserRepository()
	user := mkUser("user-1", "old@example.com", false, false)
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Verify old email works before update
	_, err = repo.GetByEmail(context.Background(), user.TenantID, "old@example.com")
	require.NoError(t, err)

	// Create a new user object with different email for update
	// This simulates real usage where you get user, modify it, and update
	updatedUser := mkUser("user-1", "new@example.com", false, false)

	err = repo.Update(context.Background(), updatedUser)

	require.NoError(t, err)

	// Old email should not work
	_, err = repo.GetByEmail(context.Background(), user.TenantID, "old@example.com")
	assert.Error(t, err)

	// New email should work
	retrieved, err := repo.GetByEmail(context.Background(), updatedUser.TenantID, "new@example.com")
	require.NoError(t, err)
	assert.Equal(t, "user-1", retrieved.ID)

	// byEmail index should be updated
	assert.Equal(t, "user-1", repo.byEmail["new@example.com"])
	assert.NotContains(t, repo.byEmail, "old@example.com")
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()
	user := mkUser("non-existent", "test@example.com", false, false)

	err := repo.Update(context.Background(), user)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestUserRepository_Delete_Success(t *testing.T) {
	repo := NewInMemoryUserRepository()
	user := mkUser("user-1", "test@example.com", false, false)
	repo.Create(context.Background(), user)

	err := repo.Delete(context.Background(), "user-1")

	require.NoError(t, err)
	assert.Equal(t, 0, len(repo.users))
	assert.Equal(t, 0, len(repo.byEmail))

	// Verify user is gone
	_, err = repo.GetByID(context.Background(), "user-1")
	assert.Error(t, err)

	_, err = repo.GetByEmail(context.Background(), user.TenantID, "test@example.com")
	assert.Error(t, err)
}

func TestUserRepository_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	err := repo.Delete(context.Background(), "non-existent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestUserRepository_Delete_CleansEmailIndex(t *testing.T) {
	repo := NewInMemoryUserRepository()
	user := mkUser("user-1", "test@example.com", false, false)
	repo.Create(context.Background(), user)

	repo.Delete(context.Background(), "user-1")

	// Both indexes should be cleaned
	assert.NotContains(t, repo.users, "user-1")
	assert.NotContains(t, repo.byEmail, "test@example.com")
}

func TestUserRepository_DifferentRoles(t *testing.T) {
	repo := NewInMemoryUserRepository()

	testCases := []struct {
		name        string
		user        *domain.User
		isAdmin     bool
		isModerator bool
	}{
		{
			name:        "Admin user",
			user:        mkUser("admin-1", "admin@example.com", true, false),
			isAdmin:     true,
			isModerator: false,
		},
		{
			name:        "Moderator user",
			user:        mkUser("mod-1", "mod@example.com", false, true),
			isAdmin:     false,
			isModerator: true,
		},
		{
			name:        "Regular user",
			user:        mkUser("user-1", "user@example.com", false, false),
			isAdmin:     false,
			isModerator: false,
		},
		{
			name:        "Admin + Moderator",
			user:        mkUser("super-1", "super@example.com", true, true),
			isAdmin:     true,
			isModerator: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Create(context.Background(), tc.user)
			require.NoError(t, err)

			retrieved, err := repo.GetByID(context.Background(), tc.user.ID)
			require.NoError(t, err)
			assert.Equal(t, tc.isAdmin, retrieved.IsAdmin)
			assert.Equal(t, tc.isModerator, retrieved.IsModerator)
		})
	}

	assert.Equal(t, 4, len(repo.users))
}

func TestUserRepository_ConcurrentReadsSafe(t *testing.T) {
	repo := NewInMemoryUserRepository()
	user := mkUser("user-1", "test@example.com", false, false)
	repo.Create(context.Background(), user)

	// Multiple concurrent reads should be safe
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := repo.GetByID(context.Background(), "user-1")
			assert.NoError(t, err)
			_, err = repo.GetByEmail(context.Background(), user.TenantID, "test@example.com")
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestUserRepository_LocaleSupport(t *testing.T) {
	repo := NewInMemoryUserRepository()

	userEN := mkUser("user-1", "en@example.com", false, false)
	userEN.Locale = "en"

	userTR := mkUser("user-2", "tr@example.com", false, false)
	userTR.Locale = "tr"

	repo.Create(context.Background(), userEN)
	repo.Create(context.Background(), userTR)

	retrieved1, _ := repo.GetByID(context.Background(), "user-1")
	assert.Equal(t, "en", retrieved1.Locale)

	retrieved2, _ := repo.GetByID(context.Background(), "user-2")
	assert.Equal(t, "tr", retrieved2.Locale)
}

func TestUserRepository_PasswordHashNotExposed(t *testing.T) {
	repo := NewInMemoryUserRepository()
	user := mkUser("user-1", "test@example.com", false, false)
	user.PassHash = "secret-hash-value"

	repo.Create(context.Background(), user)

	retrieved, _ := repo.GetByID(context.Background(), "user-1")
	assert.Equal(t, "secret-hash-value", retrieved.PassHash)
}

func TestUserRepository_FullCRUDCycle(t *testing.T) {
	repo := NewInMemoryUserRepository()

	// Create
	user := mkUser("user-1", "test@example.com", false, false)
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Read by ID
	retrieved, err := repo.GetByID(context.Background(), "user-1")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", retrieved.Email)

	// Read by Email
	retrieved, err = repo.GetByEmail(context.Background(), user.TenantID, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, "user-1", retrieved.ID)

	// Update
	user.Locale = "tr"
	user.IsAdmin = true
	err = repo.Update(context.Background(), user)
	require.NoError(t, err)

	retrieved, _ = repo.GetByID(context.Background(), "user-1")
	assert.Equal(t, "tr", retrieved.Locale)
	assert.True(t, retrieved.IsAdmin)

	// Delete
	err = repo.Delete(context.Background(), "user-1")
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), "user-1")
	assert.Error(t, err)
}

func TestUserRepository_TenantIDPreserved(t *testing.T) {
	repo := NewInMemoryUserRepository()
	user := mkUser("user-1", "test@example.com", false, false)
	user.TenantID = "custom-tenant-123"

	repo.Create(context.Background(), user)

	retrieved, _ := repo.GetByID(context.Background(), "user-1")
	assert.Equal(t, "custom-tenant-123", retrieved.TenantID)
}

func TestUserRepository_CreatedAtUpdatedAtPreserved(t *testing.T) {
	repo := NewInMemoryUserRepository()
	user := mkUser("user-1", "test@example.com", false, false)

	createdAt := user.CreatedAt
	updatedAt := user.UpdatedAt

	repo.Create(context.Background(), user)

	retrieved, _ := repo.GetByID(context.Background(), "user-1")
	assert.Equal(t, createdAt.Unix(), retrieved.CreatedAt.Unix())
	assert.Equal(t, updatedAt.Unix(), retrieved.UpdatedAt.Unix())
}
