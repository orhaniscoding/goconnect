package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test invite token
func mkInviteToken(id, networkID, tenantID, token, createdBy string, expiresIn time.Duration, usesMax int) *domain.InviteToken {
	now := time.Now()
	return &domain.InviteToken{
		ID:        id,
		NetworkID: networkID,
		TenantID:  tenantID,
		Token:     token,
		CreatedBy: createdBy,
		ExpiresAt: now.Add(expiresIn),
		UsesMax:   usesMax,
		UsesLeft:  usesMax,
		CreatedAt: now,
		RevokedAt: nil,
	}
}

func TestNewInMemoryInviteTokenRepository(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.byID)
	assert.NotNil(t, repo.byToken)
	assert.NotNil(t, repo.byNetwork)
	assert.Equal(t, 0, len(repo.byID))
}

func TestInviteTokenRepository_Create_Success(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()
	token := mkInviteToken("inv-1", "net-1", "tenant-1", "abc123", "user-1", 24*time.Hour, 10)

	err := repo.Create(ctx, token)

	require.NoError(t, err)
	assert.Equal(t, 1, len(repo.byID))
	assert.Equal(t, 1, len(repo.byToken))
	assert.Equal(t, 1, len(repo.byNetwork["net-1"]))
}

func TestInviteTokenRepository_Create_DuplicateID(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()
	token1 := mkInviteToken("inv-1", "net-1", "tenant-1", "abc123", "user-1", 24*time.Hour, 10)
	token2 := mkInviteToken("inv-1", "net-2", "tenant-1", "xyz789", "user-1", 24*time.Hour, 5)

	err1 := repo.Create(ctx, token1)
	require.NoError(t, err1)

	err2 := repo.Create(ctx, token2)

	require.Error(t, err2)
	domainErr, ok := err2.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrConflict, domainErr.Code)
}

func TestInviteTokenRepository_Create_DuplicateToken(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()
	token1 := mkInviteToken("inv-1", "net-1", "tenant-1", "same-token", "user-1", 24*time.Hour, 10)
	token2 := mkInviteToken("inv-2", "net-1", "tenant-1", "same-token", "user-1", 24*time.Hour, 5)

	err1 := repo.Create(ctx, token1)
	require.NoError(t, err1)

	err2 := repo.Create(ctx, token2)

	require.Error(t, err2)
	domainErr, ok := err2.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrConflict, domainErr.Code)
}

func TestInviteTokenRepository_GetByID_Success(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()
	token := mkInviteToken("inv-1", "net-1", "tenant-1", "abc123", "user-1", 24*time.Hour, 10)
	_ = repo.Create(ctx, token)

	result, err := repo.GetByID(ctx, "inv-1")

	require.NoError(t, err)
	assert.Equal(t, "inv-1", result.ID)
	assert.Equal(t, "abc123", result.Token)
}

func TestInviteTokenRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()

	result, err := repo.GetByID(ctx, "nonexistent")

	require.Error(t, err)
	assert.Nil(t, result)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrInviteTokenNotFound, domainErr.Code)
}

func TestInviteTokenRepository_GetByToken_Success(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()
	token := mkInviteToken("inv-1", "net-1", "tenant-1", "abc123", "user-1", 24*time.Hour, 10)
	_ = repo.Create(ctx, token)

	result, err := repo.GetByToken(ctx, "abc123")

	require.NoError(t, err)
	assert.Equal(t, "inv-1", result.ID)
	assert.Equal(t, "net-1", result.NetworkID)
}

func TestInviteTokenRepository_GetByToken_NotFound(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()

	result, err := repo.GetByToken(ctx, "nonexistent")

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestInviteTokenRepository_ListByNetwork_Success(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()

	// Create tokens for different networks
	token1 := mkInviteToken("inv-1", "net-1", "tenant-1", "token1", "user-1", 24*time.Hour, 10)
	token2 := mkInviteToken("inv-2", "net-1", "tenant-1", "token2", "user-1", 24*time.Hour, 5)
	token3 := mkInviteToken("inv-3", "net-2", "tenant-1", "token3", "user-1", 24*time.Hour, 3)

	_ = repo.Create(ctx, token1)
	_ = repo.Create(ctx, token2)
	_ = repo.Create(ctx, token3)

	result, err := repo.ListByNetwork(ctx, "net-1")

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestInviteTokenRepository_ListByNetwork_ExcludesExpired(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()

	// Create an expired token
	expiredToken := mkInviteToken("inv-expired", "net-1", "tenant-1", "expired", "user-1", -1*time.Hour, 10)
	validToken := mkInviteToken("inv-valid", "net-1", "tenant-1", "valid", "user-1", 24*time.Hour, 10)

	_ = repo.Create(ctx, expiredToken)
	_ = repo.Create(ctx, validToken)

	result, err := repo.ListByNetwork(ctx, "net-1")

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "inv-valid", result[0].ID)
}

func TestInviteTokenRepository_ListByNetwork_ExcludesRevoked(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()

	// Create and revoke a token
	revokedToken := mkInviteToken("inv-revoked", "net-1", "tenant-1", "revoked", "user-1", 24*time.Hour, 10)
	validToken := mkInviteToken("inv-valid", "net-1", "tenant-1", "valid", "user-1", 24*time.Hour, 10)

	_ = repo.Create(ctx, revokedToken)
	_ = repo.Create(ctx, validToken)
	_ = repo.Revoke(ctx, "inv-revoked")

	result, err := repo.ListByNetwork(ctx, "net-1")

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "inv-valid", result[0].ID)
}

func TestInviteTokenRepository_UseToken_Success(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()
	token := mkInviteToken("inv-1", "net-1", "tenant-1", "abc123", "user-1", 24*time.Hour, 10)
	_ = repo.Create(ctx, token)

	result, err := repo.UseToken(ctx, "abc123")

	require.NoError(t, err)
	assert.Equal(t, 9, result.UsesLeft)
}

func TestInviteTokenRepository_UseToken_Unlimited(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()
	token := mkInviteToken("inv-1", "net-1", "tenant-1", "abc123", "user-1", 24*time.Hour, 0) // unlimited
	_ = repo.Create(ctx, token)

	result, err := repo.UseToken(ctx, "abc123")

	require.NoError(t, err)
	assert.Equal(t, 0, result.UsesLeft) // remains 0 for unlimited
}

func TestInviteTokenRepository_UseToken_NotFound(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()

	result, err := repo.UseToken(ctx, "nonexistent")

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestInviteTokenRepository_UseToken_Expired(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()
	token := mkInviteToken("inv-1", "net-1", "tenant-1", "abc123", "user-1", -1*time.Hour, 10) // expired
	_ = repo.Create(ctx, token)

	result, err := repo.UseToken(ctx, "abc123")

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestInviteTokenRepository_UseToken_NoUsesLeft(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()
	token := mkInviteToken("inv-1", "net-1", "tenant-1", "abc123", "user-1", 24*time.Hour, 1)
	_ = repo.Create(ctx, token)

	// Use the only available use
	_, _ = repo.UseToken(ctx, "abc123")

	// Try to use again
	result, err := repo.UseToken(ctx, "abc123")

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestInviteTokenRepository_Revoke_Success(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()
	token := mkInviteToken("inv-1", "net-1", "tenant-1", "abc123", "user-1", 24*time.Hour, 10)
	_ = repo.Create(ctx, token)

	err := repo.Revoke(ctx, "inv-1")

	require.NoError(t, err)

	// Verify it's revoked
	result, _ := repo.GetByID(ctx, "inv-1")
	assert.NotNil(t, result.RevokedAt)
}

func TestInviteTokenRepository_Revoke_NotFound(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()

	err := repo.Revoke(ctx, "nonexistent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrInviteTokenNotFound, domainErr.Code)
}

func TestInviteTokenRepository_DeleteExpired_Success(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()

	// Create expired and valid tokens
	expiredToken1 := mkInviteToken("inv-exp-1", "net-1", "tenant-1", "expired1", "user-1", -2*time.Hour, 10)
	expiredToken2 := mkInviteToken("inv-exp-2", "net-1", "tenant-1", "expired2", "user-1", -1*time.Hour, 10)
	validToken := mkInviteToken("inv-valid", "net-1", "tenant-1", "valid", "user-1", 24*time.Hour, 10)

	_ = repo.Create(ctx, expiredToken1)
	_ = repo.Create(ctx, expiredToken2)
	_ = repo.Create(ctx, validToken)

	deleted, err := repo.DeleteExpired(ctx)

	require.NoError(t, err)
	assert.Equal(t, 2, deleted)
	assert.Equal(t, 1, len(repo.byID))
}

func TestInviteTokenRepository_DeleteExpired_NoExpired(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()

	validToken := mkInviteToken("inv-valid", "net-1", "tenant-1", "valid", "user-1", 24*time.Hour, 10)
	_ = repo.Create(ctx, validToken)

	deleted, err := repo.DeleteExpired(ctx)

	require.NoError(t, err)
	assert.Equal(t, 0, deleted)
	assert.Equal(t, 1, len(repo.byID))
}

func TestInviteTokenRepository_Concurrency(t *testing.T) {
	repo := NewInMemoryInviteTokenRepository()
	ctx := context.Background()
	token := mkInviteToken("inv-1", "net-1", "tenant-1", "abc123", "user-1", 24*time.Hour, 100)
	_ = repo.Create(ctx, token)

	// Run concurrent UseToken operations
	done := make(chan bool, 50)
	for i := 0; i < 50; i++ {
		go func() {
			_, _ = repo.UseToken(ctx, "abc123")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 50; i++ {
		<-done
	}

	// Verify the final count
	result, _ := repo.GetByID(ctx, "inv-1")
	assert.Equal(t, 50, result.UsesLeft)
}
