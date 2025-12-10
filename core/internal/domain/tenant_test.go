package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTenantRoleHierarchy(t *testing.T) {
	tests := []struct {
		role     TenantRole
		expected int
	}{
		{TenantRoleOwner, 100},
		{TenantRoleAdmin, 80},
		{TenantRoleModerator, 60},
		{TenantRoleVIP, 40},
		{TenantRoleMember, 20},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			result := TenantRoleHierarchy(tt.role)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTenantRole_HasPermission(t *testing.T) {
	tests := []struct {
		name     string
		role     TenantRole
		required TenantRole
		expected bool
	}{
		{"Owner has admin permission", TenantRoleOwner, TenantRoleAdmin, true},
		{"Owner has owner permission", TenantRoleOwner, TenantRoleOwner, true},
		{"Admin has moderator permission", TenantRoleAdmin, TenantRoleModerator, true},
		{"Admin does not have owner permission", TenantRoleAdmin, TenantRoleOwner, false},
		{"Moderator has member permission", TenantRoleModerator, TenantRoleMember, true},
		{"Member does not have moderator permission", TenantRoleMember, TenantRoleModerator, false},
		{"VIP has member permission", TenantRoleVIP, TenantRoleMember, true},
		{"Member has member permission", TenantRoleMember, TenantRoleMember, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.HasPermission(tt.required)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTenantMember_IsBanned(t *testing.T) {
	now := time.Now()

	t.Run("Not banned", func(t *testing.T) {
		member := &TenantMember{
			ID:       "tm_123",
			TenantID: "t1",
			UserID:   "u1",
			BannedAt: nil,
		}
		assert.False(t, member.IsBanned())
	})

	t.Run("Banned", func(t *testing.T) {
		member := &TenantMember{
			ID:       "tm_123",
			TenantID: "t1",
			UserID:   "u1",
			BannedAt: &now,
			BannedBy: "admin1",
		}
		assert.True(t, member.IsBanned())
	})
}

func TestGenerateTenantInviteCode(t *testing.T) {
	t.Run("Generates valid code", func(t *testing.T) {
		code, err := GenerateTenantInviteCode()
		assert.NoError(t, err)
		assert.Len(t, code, 8)

		// Check that code only contains valid characters
		const validChars = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
		for _, c := range code {
			assert.True(t, strings.ContainsRune(validChars, c), "Invalid character: %c", c)
		}
	})

	t.Run("Generates unique codes", func(t *testing.T) {
		codes := make(map[string]bool)
		for i := 0; i < 100; i++ {
			code, err := GenerateTenantInviteCode()
			assert.NoError(t, err)
			assert.False(t, codes[code], "Duplicate code generated: %s", code)
			codes[code] = true
		}
	})
}

func TestGenerateTenantMemberID(t *testing.T) {
	id := GenerateTenantMemberID()
	assert.True(t, strings.HasPrefix(id, "tm_"))
	assert.Len(t, id, 19) // "tm_" + 16 hex chars
}

func TestGenerateTenantInviteID(t *testing.T) {
	id := GenerateTenantInviteID()
	assert.True(t, strings.HasPrefix(id, "ti_"))
	assert.Len(t, id, 19) // "ti_" + 16 hex chars
}

func TestGenerateAnnouncementID(t *testing.T) {
	id := GenerateAnnouncementID()
	assert.True(t, strings.HasPrefix(id, "ann_"))
	assert.Len(t, id, 20) // "ann_" + 16 hex chars
}

func TestGenerateChatMessageID(t *testing.T) {
	id := GenerateChatMessageID()
	assert.True(t, strings.HasPrefix(id, "msg_"))
	assert.Len(t, id, 20) // "msg_" + 16 hex chars
}

func TestTenantInvite_IsValid(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-time.Hour)
	futureTime := now.Add(time.Hour)

	tests := []struct {
		name     string
		invite   *TenantInvite
		expected bool
	}{
		{
			name: "Valid invite - no restrictions",
			invite: &TenantInvite{
				ID:        "ti_123",
				MaxUses:   0,
				UseCount:  0,
				ExpiresAt: nil,
				RevokedAt: nil,
			},
			expected: true,
		},
		{
			name: "Valid invite - not yet expired",
			invite: &TenantInvite{
				ID:        "ti_124",
				MaxUses:   10,
				UseCount:  5,
				ExpiresAt: &futureTime,
				RevokedAt: nil,
			},
			expected: true,
		},
		{
			name: "Invalid - revoked",
			invite: &TenantInvite{
				ID:        "ti_125",
				RevokedAt: &now,
			},
			expected: false,
		},
		{
			name: "Invalid - expired",
			invite: &TenantInvite{
				ID:        "ti_126",
				ExpiresAt: &pastTime,
			},
			expected: false,
		},
		{
			name: "Invalid - max uses reached",
			invite: &TenantInvite{
				ID:       "ti_127",
				MaxUses:  5,
				UseCount: 5,
			},
			expected: false,
		},
		{
			name: "Invalid - exceeded max uses",
			invite: &TenantInvite{
				ID:       "ti_128",
				MaxUses:  5,
				UseCount: 10,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.invite.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTenantExtended_CanJoin(t *testing.T) {
	tests := []struct {
		name     string
		tenant   *TenantExtended
		expected bool
	}{
		{
			name: "Unlimited members",
			tenant: &TenantExtended{
				Tenant:      Tenant{MaxMembers: 0},
				MemberCount: 1000,
			},
			expected: true,
		},
		{
			name: "Has space",
			tenant: &TenantExtended{
				Tenant:      Tenant{MaxMembers: 100},
				MemberCount: 50,
			},
			expected: true,
		},
		{
			name: "At capacity",
			tenant: &TenantExtended{
				Tenant:      Tenant{MaxMembers: 100},
				MemberCount: 100,
			},
			expected: false,
		},
		{
			name: "Over capacity",
			tenant: &TenantExtended{
				Tenant:      Tenant{MaxMembers: 100},
				MemberCount: 150,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tenant.CanJoin()
			assert.Equal(t, tt.expected, result)
		})
	}
}
