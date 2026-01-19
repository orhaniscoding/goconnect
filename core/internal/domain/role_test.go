package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRole_HasHigherPosition(t *testing.T) {
	role1 := &Role{Position: 10}
	role2 := &Role{Position: 5}
	role3 := &Role{Position: 10}

	assert.True(t, role1.HasHigherPosition(role2))
	assert.False(t, role2.HasHigherPosition(role1))
	assert.False(t, role1.HasHigherPosition(role3))
}

func TestRole_CanManage(t *testing.T) {
	t.Run("admin can manage any role", func(t *testing.T) {
		admin := &Role{IsAdmin: true, Position: 1}
		highRole := &Role{Position: 100}

		assert.True(t, admin.CanManage(highRole))
	})

	t.Run("higher position can manage lower", func(t *testing.T) {
		highRole := &Role{Position: 10}
		lowRole := &Role{Position: 5}

		assert.True(t, highRole.CanManage(lowRole))
		assert.False(t, lowRole.CanManage(highRole))
	})

	t.Run("same position cannot manage", func(t *testing.T) {
		role1 := &Role{Position: 10}
		role2 := &Role{Position: 10}

		assert.False(t, role1.CanManage(role2))
	})
}

func TestValidateRoleName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		wantError bool
	}{
		{
			name:      "valid name",
			input:     "Admin",
			expected:  "Admin",
			wantError: false,
		},
		{
			name:      "trims whitespace",
			input:     "  Moderator  ",
			expected:  "Moderator",
			wantError: false,
		},
		{
			name:      "empty name",
			input:     "",
			wantError: true,
		},
		{
			name:      "too long",
			input:     string(make([]byte, 101)),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateRoleName(tt.input)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidateHexColor(t *testing.T) {
	tests := []struct {
		name      string
		color     string
		wantError bool
	}{
		{"valid color", "#FF5733", false},
		{"valid lowercase", "#ff5733", false},
		{"empty color", "", false},
		{"missing hash", "FF5733", true},
		{"too short", "#FFF", true},
		{"too long", "#FF5733AA", true},
		{"invalid chars", "#GGGGGG", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHexColor(tt.color)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateRoleID(t *testing.T) {
	id1 := GenerateRoleID()
	id2 := GenerateRoleID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "role_")
}

func TestChannelPermissionOverride_IsRoleOverride(t *testing.T) {
	roleID := "role-123"
	userID := "user-123"

	t.Run("role override", func(t *testing.T) {
		override := &ChannelPermissionOverride{RoleID: &roleID}
		assert.True(t, override.IsRoleOverride())
		assert.False(t, override.IsUserOverride())
	})

	t.Run("user override", func(t *testing.T) {
		override := &ChannelPermissionOverride{UserID: &userID}
		assert.False(t, override.IsRoleOverride())
		assert.True(t, override.IsUserOverride())
	})
}

func TestPermissionConstants(t *testing.T) {
	// Verify permission constants are properly defined
	assert.Equal(t, "server.manage", PermServerManage)
	assert.Equal(t, "channel.send_messages", PermChannelSendMessages)
	assert.Equal(t, "voice.connect", PermVoiceConnect)
	assert.Equal(t, "network.connect", PermNetworkConnect)
}
