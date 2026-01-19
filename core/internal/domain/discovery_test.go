package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerDiscovery_IsDiscoverable(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		d := &ServerDiscovery{Enabled: true}
		assert.True(t, d.IsDiscoverable())
	})

	t.Run("disabled", func(t *testing.T) {
		d := &ServerDiscovery{Enabled: false}
		assert.False(t, d.IsDiscoverable())
	})
}

func TestAllDiscoveryCategories(t *testing.T) {
	categories := AllDiscoveryCategories()

	assert.NotEmpty(t, categories)
	assert.Contains(t, categories, DiscoveryCategoryGaming)
	assert.Contains(t, categories, DiscoveryCategoryEducation)
	assert.Contains(t, categories, DiscoveryCategoryMusic)
	assert.Contains(t, categories, DiscoveryCategoryTech)
	assert.Contains(t, categories, DiscoveryCategoryOther)
}

func TestValidateVanityCode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		wantError bool
	}{
		{
			name:      "valid code",
			input:     "minecraft-tr",
			expected:  "minecraft-tr",
			wantError: false,
		},
		{
			name:      "converts to lowercase",
			input:     "MINECRAFT-TR",
			expected:  "minecraft-tr",
			wantError: false,
		},
		{
			name:      "valid alphanumeric",
			input:     "server123",
			expected:  "server123",
			wantError: false,
		},
		{
			name:      "too short",
			input:     "abc",
			wantError: true,
		},
		{
			name:      "too long",
			input:     "this-vanity-code-is-way-too-long-for-the-limit",
			wantError: true,
		},
		{
			name:      "starts with hyphen",
			input:     "-invalid",
			wantError: true,
		},
		{
			name:      "ends with hyphen",
			input:     "invalid-",
			wantError: true,
		},
		{
			name:      "consecutive hyphens",
			input:     "invalid--code",
			wantError: true,
		},
		{
			name:      "empty",
			input:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateVanityCode(tt.input)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDiscoveryCategoryConstants(t *testing.T) {
	assert.Equal(t, DiscoveryCategory("gaming"), DiscoveryCategoryGaming)
	assert.Equal(t, DiscoveryCategory("education"), DiscoveryCategoryEducation)
	assert.Equal(t, DiscoveryCategory("music"), DiscoveryCategoryMusic)
	assert.Equal(t, DiscoveryCategory("tech"), DiscoveryCategoryTech)
	assert.Equal(t, DiscoveryCategory("other"), DiscoveryCategoryOther)
}

func TestDiscoveryServerResponse(t *testing.T) {
	resp := DiscoveryServerResponse{
		ID:          "tenant-123",
		Name:        "Gaming Server",
		MemberCount: 1000,
		OnlineCount: 150,
		Featured:    true,
		Verified:    true,
	}

	assert.Equal(t, "tenant-123", resp.ID)
	assert.Equal(t, "Gaming Server", resp.Name)
	assert.Equal(t, 1000, resp.MemberCount)
	assert.True(t, resp.Featured)
	assert.True(t, resp.Verified)
}
