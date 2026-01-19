package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSection_IsDeleted(t *testing.T) {
	t.Run("not deleted", func(t *testing.T) {
		section := &Section{}
		assert.False(t, section.IsDeleted())
	})

	t.Run("deleted", func(t *testing.T) {
		now := timeNow()
		section := &Section{DeletedAt: &now}
		assert.True(t, section.IsDeleted())
	})
}

func TestValidateSectionName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		wantError bool
	}{
		{
			name:      "valid name",
			input:     "General",
			expected:  "General",
			wantError: false,
		},
		{
			name:      "valid name with spaces",
			input:     "Voice Channels",
			expected:  "Voice Channels",
			wantError: false,
		},
		{
			name:      "trims whitespace",
			input:     "  Trimmed  ",
			expected:  "Trimmed",
			wantError: false,
		},
		{
			name:      "collapses multiple spaces",
			input:     "Multiple   Spaces",
			expected:  "Multiple Spaces",
			wantError: false,
		},
		{
			name:      "empty name",
			input:     "",
			wantError: true,
		},
		{
			name:      "only whitespace",
			input:     "   ",
			wantError: true,
		},
		{
			name:      "too long name",
			input:     string(make([]byte, 101)),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateSectionName(tt.input)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGenerateSectionID(t *testing.T) {
	id1 := GenerateSectionID()
	id2 := GenerateSectionID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "sec_")
}

func TestCreateSectionRequest_ApplyDefaults(t *testing.T) {
	req := &CreateSectionRequest{
		Name: "Test Section",
	}

	req.ApplyDefaults()

	assert.Equal(t, SectionVisibilityVisible, req.Visibility)
}
