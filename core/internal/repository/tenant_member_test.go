package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test tenant member
func mkTenantMember(id, tenantID, userID string, role domain.TenantRole, nickname string) *domain.TenantMember {
	return &domain.TenantMember{
		ID:        id,
		TenantID:  tenantID,
		UserID:    userID,
		Role:      role,
		Nickname:  nickname,
		JoinedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestNewInMemoryTenantMemberRepository(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.members)
	assert.Equal(t, 0, len(repo.members))
}

func TestTenantMemberRepository_Create_Success(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("", "tenant-1", "user-1", domain.TenantRoleMember, "JohnDoe")

	err := repo.Create(ctx, member)

	require.NoError(t, err)
	assert.NotEmpty(t, member.ID)
	assert.Equal(t, 1, len(repo.members))
}

func TestTenantMemberRepository_Create_WithExistingID(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("member-123", "tenant-1", "user-1", domain.TenantRoleMember, "JohnDoe")

	err := repo.Create(ctx, member)

	require.NoError(t, err)
	assert.Equal(t, "member-123", member.ID)
}

func TestTenantMemberRepository_Create_DuplicateMembership(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member1 := mkTenantMember("", "tenant-1", "user-1", domain.TenantRoleMember, "JohnDoe")
	member2 := mkTenantMember("", "tenant-1", "user-1", domain.TenantRoleAdmin, "Johnny")

	err1 := repo.Create(ctx, member1)
	require.NoError(t, err1)

	err2 := repo.Create(ctx, member2)

	require.Error(t, err2)
	domainErr, ok := err2.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrValidation, domainErr.Code)
}

func TestTenantMemberRepository_Create_SameUserDifferentTenants(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member1 := mkTenantMember("", "tenant-1", "user-1", domain.TenantRoleMember, "")
	member2 := mkTenantMember("", "tenant-2", "user-1", domain.TenantRoleMember, "")

	err1 := repo.Create(ctx, member1)
	require.NoError(t, err1)

	err2 := repo.Create(ctx, member2)

	require.NoError(t, err2)
	assert.Equal(t, 2, len(repo.members))
}

func TestTenantMemberRepository_GetByID_Success(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("member-1", "tenant-1", "user-1", domain.TenantRoleAdmin, "Admin")
	_ = repo.Create(ctx, member)

	result, err := repo.GetByID(ctx, "member-1")

	require.NoError(t, err)
	assert.Equal(t, "member-1", result.ID)
	assert.Equal(t, "tenant-1", result.TenantID)
	assert.Equal(t, "user-1", result.UserID)
	assert.Equal(t, domain.TenantRoleAdmin, result.Role)
}

func TestTenantMemberRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	result, err := repo.GetByID(ctx, "non-existent")

	require.Error(t, err)
	assert.Nil(t, result)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantMemberRepository_GetByUserAndTenant_Success(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("", "tenant-1", "user-1", domain.TenantRoleVIP, "VIPUser")
	_ = repo.Create(ctx, member)

	result, err := repo.GetByUserAndTenant(ctx, "user-1", "tenant-1")

	require.NoError(t, err)
	assert.Equal(t, "user-1", result.UserID)
	assert.Equal(t, "tenant-1", result.TenantID)
	assert.Equal(t, domain.TenantRoleVIP, result.Role)
}

func TestTenantMemberRepository_GetByUserAndTenant_NotFound(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	result, err := repo.GetByUserAndTenant(ctx, "user-1", "tenant-1")

	require.Error(t, err)
	assert.Nil(t, result)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantMemberRepository_Update_Success(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "OldNick")
	_ = repo.Create(ctx, member)

	member.Nickname = "NewNick"
	member.Role = domain.TenantRoleAdmin
	err := repo.Update(ctx, member)

	require.NoError(t, err)

	updated, _ := repo.GetByID(ctx, "member-1")
	assert.Equal(t, "NewNick", updated.Nickname)
	assert.Equal(t, domain.TenantRoleAdmin, updated.Role)
}

func TestTenantMemberRepository_Update_NotFound(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("non-existent", "tenant-1", "user-1", domain.TenantRoleMember, "")

	err := repo.Update(ctx, member)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantMemberRepository_Delete_Success(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "")
	_ = repo.Create(ctx, member)

	err := repo.Delete(ctx, "member-1")

	require.NoError(t, err)
	assert.Equal(t, 0, len(repo.members))
}

func TestTenantMemberRepository_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "non-existent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantMemberRepository_ListByTenant_Success(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	_ = repo.Create(ctx, mkTenantMember("", "tenant-1", "user-1", domain.TenantRoleMember, ""))
	_ = repo.Create(ctx, mkTenantMember("", "tenant-1", "user-2", domain.TenantRoleAdmin, ""))
	_ = repo.Create(ctx, mkTenantMember("", "tenant-2", "user-3", domain.TenantRoleMember, ""))

	results, cursor, err := repo.ListByTenant(ctx, "tenant-1", "", 10, "")

	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.Empty(t, cursor)
}

func TestTenantMemberRepository_ListByTenant_FilterByRole(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	_ = repo.Create(ctx, mkTenantMember("", "tenant-1", "user-1", domain.TenantRoleMember, ""))
	_ = repo.Create(ctx, mkTenantMember("", "tenant-1", "user-2", domain.TenantRoleAdmin, ""))
	_ = repo.Create(ctx, mkTenantMember("", "tenant-1", "user-3", domain.TenantRoleMember, ""))

	results, _, err := repo.ListByTenant(ctx, "tenant-1", string(domain.TenantRoleMember), 10, "")

	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
	for _, r := range results {
		assert.Equal(t, domain.TenantRoleMember, r.Role)
	}
}

func TestTenantMemberRepository_ListByTenant_Pagination(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		member := mkTenantMember("", "tenant-1", "user-"+string(rune('a'+i)), domain.TenantRoleMember, "")
		_ = repo.Create(ctx, member)
	}

	results, _, err := repo.ListByTenant(ctx, "tenant-1", "", 3, "")

	require.NoError(t, err)
	assert.Equal(t, 3, len(results))
	// Note: In-memory implementation doesn't use cursor
}

func TestTenantMemberRepository_ListByTenant_EmptyTenant(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	results, cursor, err := repo.ListByTenant(ctx, "empty-tenant", "", 10, "")

	require.NoError(t, err)
	assert.Equal(t, 0, len(results))
	assert.Empty(t, cursor)
}

func TestTenantMemberRepository_ListByUser_Success(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	_ = repo.Create(ctx, mkTenantMember("", "tenant-1", "user-1", domain.TenantRoleMember, ""))
	_ = repo.Create(ctx, mkTenantMember("", "tenant-2", "user-1", domain.TenantRoleAdmin, ""))
	_ = repo.Create(ctx, mkTenantMember("", "tenant-3", "user-2", domain.TenantRoleMember, ""))

	results, err := repo.ListByUser(ctx, "user-1")

	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestTenantMemberRepository_ListByUser_Empty(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	results, err := repo.ListByUser(ctx, "user-without-memberships")

	require.NoError(t, err)
	assert.Equal(t, 0, len(results))
}

func TestTenantMemberRepository_CountByTenant(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	_ = repo.Create(ctx, mkTenantMember("", "tenant-1", "user-1", domain.TenantRoleMember, ""))
	_ = repo.Create(ctx, mkTenantMember("", "tenant-1", "user-2", domain.TenantRoleAdmin, ""))
	_ = repo.Create(ctx, mkTenantMember("", "tenant-2", "user-3", domain.TenantRoleMember, ""))

	count, err := repo.CountByTenant(ctx, "tenant-1")

	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestTenantMemberRepository_CountByTenant_Empty(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	count, err := repo.CountByTenant(ctx, "empty-tenant")

	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestTenantMemberRepository_UpdateRole_Success(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "")
	_ = repo.Create(ctx, member)

	err := repo.UpdateRole(ctx, "member-1", domain.TenantRoleAdmin)

	require.NoError(t, err)

	updated, _ := repo.GetByID(ctx, "member-1")
	assert.Equal(t, domain.TenantRoleAdmin, updated.Role)
}

func TestTenantMemberRepository_UpdateRole_NotFound(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	err := repo.UpdateRole(ctx, "non-existent", domain.TenantRoleAdmin)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantMemberRepository_GetUserRole_Success(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("", "tenant-1", "user-1", domain.TenantRoleModerator, "")
	_ = repo.Create(ctx, member)

	role, err := repo.GetUserRole(ctx, "user-1", "tenant-1")

	require.NoError(t, err)
	assert.Equal(t, domain.TenantRoleModerator, role)
}

func TestTenantMemberRepository_GetUserRole_NotFound(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	role, err := repo.GetUserRole(ctx, "user-1", "tenant-1")

	require.Error(t, err)
	assert.Equal(t, domain.TenantRole(""), role)
}

func TestTenantMemberRepository_HasRole_Owner(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("", "tenant-1", "user-1", domain.TenantRoleOwner, "")
	_ = repo.Create(ctx, member)

	tests := []struct {
		requiredRole domain.TenantRole
		expected     bool
	}{
		{domain.TenantRoleOwner, true},
		{domain.TenantRoleAdmin, true},
		{domain.TenantRoleModerator, true},
		{domain.TenantRoleVIP, true},
		{domain.TenantRoleMember, true},
	}

	for _, tt := range tests {
		hasRole, err := repo.HasRole(ctx, "user-1", "tenant-1", tt.requiredRole)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, hasRole, "Owner should have %s access", tt.requiredRole)
	}
}

func TestTenantMemberRepository_HasRole_Member(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("", "tenant-1", "user-1", domain.TenantRoleMember, "")
	_ = repo.Create(ctx, member)

	tests := []struct {
		requiredRole domain.TenantRole
		expected     bool
	}{
		{domain.TenantRoleOwner, false},
		{domain.TenantRoleAdmin, false},
		{domain.TenantRoleModerator, false},
		{domain.TenantRoleVIP, false},
		{domain.TenantRoleMember, true},
	}

	for _, tt := range tests {
		hasRole, err := repo.HasRole(ctx, "user-1", "tenant-1", tt.requiredRole)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, hasRole, "Member should %shave %s access", map[bool]string{true: "", false: "not "}[tt.expected], tt.requiredRole)
	}
}

func TestTenantMemberRepository_HasRole_NotMember(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	hasRole, err := repo.HasRole(ctx, "user-1", "tenant-1", domain.TenantRoleMember)

	require.NoError(t, err) // In-memory implementation returns false without error
	assert.False(t, hasRole)
}

func TestTenantMemberRepository_AllRoleHierarchy(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	roleHierarchy := []domain.TenantRole{
		domain.TenantRoleOwner,
		domain.TenantRoleAdmin,
		domain.TenantRoleModerator,
		domain.TenantRoleVIP,
		domain.TenantRoleMember,
	}

	for i, userRole := range roleHierarchy {
		member := mkTenantMember("", "tenant-"+string(rune('1'+i)), "user-"+string(rune('1'+i)), userRole, "")
		_ = repo.Create(ctx, member)

		for j, requiredRole := range roleHierarchy {
			hasRole, err := repo.HasRole(ctx, member.UserID, member.TenantID, requiredRole)
			require.NoError(t, err)

			// User should have access to their role and all lower roles
			expected := i <= j
			assert.Equal(t, expected, hasRole, "%s should %shave %s access",
				userRole, map[bool]string{true: "", false: "not "}[expected], requiredRole)
		}
	}
}

func TestTenantMemberRepository_Ban_Success(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "")
	_ = repo.Create(ctx, member)

	err := repo.Ban(ctx, "member-1", "admin-1")

	require.NoError(t, err)
	result, _ := repo.GetByID(ctx, "member-1")
	assert.NotNil(t, result.BannedAt)
	assert.Equal(t, "admin-1", result.BannedBy)
}

func TestTenantMemberRepository_Ban_NotFound(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	err := repo.Ban(ctx, "nonexistent", "admin-1")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantMemberRepository_Unban_Success(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "")
	_ = repo.Create(ctx, member)
	_ = repo.Ban(ctx, "member-1", "admin-1")

	err := repo.Unban(ctx, "member-1")

	require.NoError(t, err)
	result, _ := repo.GetByID(ctx, "member-1")
	assert.Nil(t, result.BannedAt)
	assert.Empty(t, result.BannedBy)
}

func TestTenantMemberRepository_Unban_NotFound(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	err := repo.Unban(ctx, "nonexistent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantMemberRepository_IsBanned_True(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "")
	_ = repo.Create(ctx, member)
	_ = repo.Ban(ctx, "member-1", "admin-1")

	isBanned, err := repo.IsBanned(ctx, "user-1", "tenant-1")

	require.NoError(t, err)
	assert.True(t, isBanned)
}

func TestTenantMemberRepository_IsBanned_False(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member := mkTenantMember("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "")
	_ = repo.Create(ctx, member)

	isBanned, err := repo.IsBanned(ctx, "user-1", "tenant-1")

	require.NoError(t, err)
	assert.False(t, isBanned)
}

func TestTenantMemberRepository_IsBanned_NotMember(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	isBanned, err := repo.IsBanned(ctx, "user-1", "tenant-1")

	require.NoError(t, err)
	assert.False(t, isBanned)
}

func TestTenantMemberRepository_ListBannedByTenant(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member1 := mkTenantMember("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "")
	member2 := mkTenantMember("member-2", "tenant-1", "user-2", domain.TenantRoleMember, "")
	member3 := mkTenantMember("member-3", "tenant-1", "user-3", domain.TenantRoleMember, "")
	_ = repo.Create(ctx, member1)
	_ = repo.Create(ctx, member2)
	_ = repo.Create(ctx, member3)
	_ = repo.Ban(ctx, "member-1", "admin")
	_ = repo.Ban(ctx, "member-3", "admin")

	banned, err := repo.ListBannedByTenant(ctx, "tenant-1")

	require.NoError(t, err)
	assert.Len(t, banned, 2)
}

func TestTenantMemberRepository_ListBannedByTenant_Empty(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()

	banned, err := repo.ListBannedByTenant(ctx, "tenant-1")

	require.NoError(t, err)
	assert.Empty(t, banned)
}

func TestTenantMemberRepository_DeleteAllByTenant(t *testing.T) {
	repo := NewInMemoryTenantMemberRepository()
	ctx := context.Background()
	member1 := mkTenantMember("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "")
	member2 := mkTenantMember("member-2", "tenant-1", "user-2", domain.TenantRoleMember, "")
	member3 := mkTenantMember("member-3", "tenant-2", "user-3", domain.TenantRoleMember, "")
	_ = repo.Create(ctx, member1)
	_ = repo.Create(ctx, member2)
	_ = repo.Create(ctx, member3)

	err := repo.DeleteAllByTenant(ctx, "tenant-1")

	require.NoError(t, err)
	members, _, _ := repo.ListByTenant(ctx, "tenant-1", "", 100, "")
	assert.Empty(t, members)
	members2, _, _ := repo.ListByTenant(ctx, "tenant-2", "", 100, "")
	assert.Len(t, members2, 1)
}
