package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test tenant invite
func mkTenantInvite(id, tenantID, code, createdBy string, maxUses int, expiresIn time.Duration) *domain.TenantInvite {
	now := time.Now()
	var expiresAt *time.Time
	if expiresIn > 0 {
		exp := now.Add(expiresIn)
		expiresAt = &exp
	}
	return &domain.TenantInvite{
		ID:        id,
		TenantID:  tenantID,
		Code:      code,
		MaxUses:   maxUses,
		UseCount:  0,
		ExpiresAt: expiresAt,
		CreatedBy: createdBy,
		CreatedAt: now,
	}
}

func TestNewInMemoryTenantInviteRepository(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.invites)
	assert.Equal(t, 0, len(repo.invites))
}

func TestTenantInviteRepository_Create_Success(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	invite := mkTenantInvite("", "tenant-1", "ABC123", "user-1", 10, 24*time.Hour)

	err := repo.Create(ctx, invite)

	require.NoError(t, err)
	assert.NotEmpty(t, invite.ID)
	assert.Equal(t, 1, len(repo.invites))
}

func TestTenantInviteRepository_Create_WithExistingID(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	invite := mkTenantInvite("invite-123", "tenant-1", "XYZ789", "user-1", 5, 24*time.Hour)

	err := repo.Create(ctx, invite)

	require.NoError(t, err)
	assert.Equal(t, "invite-123", invite.ID)
}

func TestTenantInviteRepository_Create_DuplicateCode(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	invite1 := mkTenantInvite("", "tenant-1", "SAME123", "user-1", 10, 24*time.Hour)
	invite2 := mkTenantInvite("", "tenant-2", "same123", "user-2", 5, 24*time.Hour) // case insensitive

	err1 := repo.Create(ctx, invite1)
	require.NoError(t, err1)

	err2 := repo.Create(ctx, invite2)

	require.Error(t, err2)
	domainErr, ok := err2.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrValidation, domainErr.Code)
}

func TestTenantInviteRepository_Create_UnlimitedUses(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	invite := mkTenantInvite("", "tenant-1", "UNLIMITED", "user-1", 0, 24*time.Hour)

	err := repo.Create(ctx, invite)

	require.NoError(t, err)
	assert.Equal(t, 0, invite.MaxUses)
}

func TestTenantInviteRepository_Create_NoExpiration(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	invite := mkTenantInvite("", "tenant-1", "NOEXP", "user-1", 10, 0)

	err := repo.Create(ctx, invite)

	require.NoError(t, err)
	assert.Nil(t, invite.ExpiresAt)
}

func TestTenantInviteRepository_GetByID_Success(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	invite := mkTenantInvite("invite-1", "tenant-1", "ABC123", "user-1", 10, 24*time.Hour)
	_ = repo.Create(ctx, invite)

	result, err := repo.GetByID(ctx, "invite-1")

	require.NoError(t, err)
	assert.Equal(t, "invite-1", result.ID)
	assert.Equal(t, "tenant-1", result.TenantID)
	assert.Equal(t, "ABC123", result.Code)
}

func TestTenantInviteRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()

	result, err := repo.GetByID(ctx, "non-existent")

	require.Error(t, err)
	assert.Nil(t, result)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantInviteRepository_GetByCode_Success(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	invite := mkTenantInvite("", "tenant-1", "JOIN123", "user-1", 10, 24*time.Hour)
	_ = repo.Create(ctx, invite)

	result, err := repo.GetByCode(ctx, "join123") // case insensitive

	require.NoError(t, err)
	assert.Equal(t, "JOIN123", result.Code)
}

func TestTenantInviteRepository_GetByCode_WithWhitespace(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	invite := mkTenantInvite("", "tenant-1", "TRIM123", "user-1", 10, 24*time.Hour)
	_ = repo.Create(ctx, invite)

	result, err := repo.GetByCode(ctx, "  TRIM123  ") // should trim

	require.NoError(t, err)
	assert.Equal(t, "TRIM123", result.Code)
}

func TestTenantInviteRepository_GetByCode_NotFound(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()

	result, err := repo.GetByCode(ctx, "NOTEXIST")

	require.Error(t, err)
	assert.Nil(t, result)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantInviteRepository_Delete_Success(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	invite := mkTenantInvite("invite-1", "tenant-1", "DEL123", "user-1", 10, 24*time.Hour)
	_ = repo.Create(ctx, invite)

	err := repo.Delete(ctx, "invite-1")

	require.NoError(t, err)
	assert.Equal(t, 0, len(repo.invites))
}

func TestTenantInviteRepository_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "non-existent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantInviteRepository_ListByTenant_Success(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	_ = repo.Create(ctx, mkTenantInvite("", "tenant-1", "CODE1", "user-1", 10, 24*time.Hour))
	_ = repo.Create(ctx, mkTenantInvite("", "tenant-1", "CODE2", "user-1", 5, 24*time.Hour))
	_ = repo.Create(ctx, mkTenantInvite("", "tenant-2", "CODE3", "user-1", 10, 24*time.Hour))

	results, err := repo.ListByTenant(ctx, "tenant-1")

	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestTenantInviteRepository_ListByTenant_Empty(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()

	results, err := repo.ListByTenant(ctx, "empty-tenant")

	require.NoError(t, err)
	assert.Equal(t, 0, len(results))
}

func TestTenantInviteRepository_IncrementUseCount_Success(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	invite := mkTenantInvite("invite-1", "tenant-1", "INC123", "user-1", 10, 24*time.Hour)
	_ = repo.Create(ctx, invite)

	err := repo.IncrementUseCount(ctx, "invite-1")

	require.NoError(t, err)

	updated, _ := repo.GetByID(ctx, "invite-1")
	assert.Equal(t, 1, updated.UseCount)
}

func TestTenantInviteRepository_IncrementUseCount_Multiple(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	invite := mkTenantInvite("invite-1", "tenant-1", "MULTI", "user-1", 10, 24*time.Hour)
	_ = repo.Create(ctx, invite)

	for i := 0; i < 5; i++ {
		err := repo.IncrementUseCount(ctx, "invite-1")
		require.NoError(t, err)
	}

	updated, _ := repo.GetByID(ctx, "invite-1")
	assert.Equal(t, 5, updated.UseCount)
}

func TestTenantInviteRepository_IncrementUseCount_NotFound(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()

	err := repo.IncrementUseCount(ctx, "non-existent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantInviteRepository_Revoke_Success(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	invite := mkTenantInvite("invite-1", "tenant-1", "REVOKE", "user-1", 10, 24*time.Hour)
	_ = repo.Create(ctx, invite)

	err := repo.Revoke(ctx, "invite-1")

	require.NoError(t, err)

	updated, _ := repo.GetByID(ctx, "invite-1")
	assert.NotNil(t, updated.RevokedAt)
}

func TestTenantInviteRepository_Revoke_NotFound(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()

	err := repo.Revoke(ctx, "non-existent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantInviteRepository_DeleteExpired_Success(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()

	// Create expired invite directly (bypassing Create to set ExpiresAt in past)
	expired := &domain.TenantInvite{
		ID:        "inv-expired",
		TenantID:  "tenant-1",
		Code:      "EXPIRED",
		CreatedBy: "user-1",
		MaxUses:   10,
		CreatedAt: time.Now(),
	}
	pastTime := time.Now().Add(-1 * time.Hour)
	expired.ExpiresAt = &pastTime
	repo.invites[expired.ID] = expired

	// Create valid invite
	valid := mkTenantInvite("inv-valid", "tenant-1", "VALID", "user-1", 10, 24*time.Hour)
	_ = repo.Create(ctx, valid)

	// Create non-expiring invite
	noExp := mkTenantInvite("inv-noexp", "tenant-1", "NOEXP", "user-1", 10, 0)
	_ = repo.Create(ctx, noExp)

	count, err := repo.DeleteExpired(ctx)

	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 2, len(repo.invites))

	// Verify correct invites remain
	_, err = repo.GetByID(ctx, "inv-valid")
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, "inv-noexp")
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, "inv-expired")
	require.Error(t, err)
}

func TestTenantInviteRepository_DeleteExpired_NoneExpired(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()
	_ = repo.Create(ctx, mkTenantInvite("", "tenant-1", "VALID1", "user-1", 10, 24*time.Hour))
	_ = repo.Create(ctx, mkTenantInvite("", "tenant-1", "VALID2", "user-1", 10, 24*time.Hour))

	count, err := repo.DeleteExpired(ctx)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, 2, len(repo.invites))
}

func TestTenantInviteRepository_FullLifecycle(t *testing.T) {
	repo := NewInMemoryTenantInviteRepository()
	ctx := context.Background()

	// Create
	invite := mkTenantInvite("", "tenant-1", "LIFECYCLE", "user-1", 3, 24*time.Hour)
	err := repo.Create(ctx, invite)
	require.NoError(t, err)
	inviteID := invite.ID

	// Use invite 2 times
	_ = repo.IncrementUseCount(ctx, inviteID)
	_ = repo.IncrementUseCount(ctx, inviteID)

	// Verify use count
	updated, _ := repo.GetByID(ctx, inviteID)
	assert.Equal(t, 2, updated.UseCount)

	// Revoke
	_ = repo.Revoke(ctx, inviteID)

	// Verify revoked
	revoked, _ := repo.GetByID(ctx, inviteID)
	assert.NotNil(t, revoked.RevokedAt)

	// Delete
	err = repo.Delete(ctx, inviteID)
	require.NoError(t, err)

	// Verify deleted
	_, err = repo.GetByID(ctx, inviteID)
	require.Error(t, err)
}

func TestTenantInviteRepository_DeleteAllByTenant(t *testing.T) {
repo := NewInMemoryTenantInviteRepository()
ctx := context.Background()

// Create invites
_ = repo.Create(ctx, mkTenantInvite("", "tenant-1", "code1", "user-1", 10, time.Hour))
_ = repo.Create(ctx, mkTenantInvite("", "tenant-1", "code2", "user-2", 10, time.Hour))
_ = repo.Create(ctx, mkTenantInvite("", "tenant-2", "code3", "user-3", 10, time.Hour))

err := repo.DeleteAllByTenant(ctx, "tenant-1")

require.NoError(t, err)
results1, _ := repo.ListByTenant(ctx, "tenant-1")
assert.Empty(t, results1)
results2, _ := repo.ListByTenant(ctx, "tenant-2")
assert.Len(t, results2, 1)
}
