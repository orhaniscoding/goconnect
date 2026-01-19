package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannel_IsDeleted(t *testing.T) {
	t.Run("not deleted", func(t *testing.T) {
		channel := &Channel{}
		assert.False(t, channel.IsDeleted())
	})

	t.Run("deleted", func(t *testing.T) {
		now := timeNow()
		channel := &Channel{DeletedAt: &now}
		assert.True(t, channel.IsDeleted())
	})
}

func TestChannel_IsVoice(t *testing.T) {
	t.Run("voice channel", func(t *testing.T) {
		channel := &Channel{Type: ChannelTypeVoice}
		assert.True(t, channel.IsVoice())
		assert.False(t, channel.IsText())
	})

	t.Run("text channel", func(t *testing.T) {
		channel := &Channel{Type: ChannelTypeText}
		assert.True(t, channel.IsText())
		assert.False(t, channel.IsVoice())
	})
}

func TestChannel_GetParentID(t *testing.T) {
	tenantID := "tenant-123"
	sectionID := "section-123"
	networkID := "network-123"

	t.Run("tenant parent", func(t *testing.T) {
		channel := &Channel{TenantID: &tenantID}
		assert.Equal(t, tenantID, channel.GetParentID())
	})

	t.Run("section parent", func(t *testing.T) {
		channel := &Channel{SectionID: &sectionID}
		assert.Equal(t, sectionID, channel.GetParentID())
	})

	t.Run("network parent", func(t *testing.T) {
		channel := &Channel{NetworkID: &networkID}
		assert.Equal(t, networkID, channel.GetParentID())
	})

	t.Run("no parent", func(t *testing.T) {
		channel := &Channel{}
		assert.Equal(t, "", channel.GetParentID())
	})
}

func TestValidateChannelName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		wantError bool
	}{
		{
			name:      "valid name",
			input:     "general",
			expected:  "general",
			wantError: false,
		},
		{
			name:      "converts to lowercase",
			input:     "General",
			expected:  "general",
			wantError: false,
		},
		{
			name:      "spaces become hyphens",
			input:     "voice chat",
			expected:  "voice-chat",
			wantError: false,
		},
		{
			name:      "removes invalid chars",
			input:     "test@channel!",
			expected:  "testchannel",
			wantError: false,
		},
		{
			name:      "empty name",
			input:     "",
			wantError: true,
		},
		{
			name:      "only special chars",
			input:     "@#$%",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateChannelName(tt.input)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGenerateChannelID(t *testing.T) {
	id1 := GenerateChannelID()
	id2 := GenerateChannelID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "ch_")
}

func TestCreateChannelRequest_ApplyDefaults(t *testing.T) {
	req := &CreateChannelRequest{
		Name: "test-channel",
	}

	req.ApplyDefaults()

	assert.Equal(t, ChannelTypeText, req.Type)
	assert.NotNil(t, req.Bitrate)
	assert.Equal(t, 64000, *req.Bitrate)
	assert.NotNil(t, req.UserLimit)
	assert.Equal(t, 0, *req.UserLimit)
	assert.NotNil(t, req.Slowmode)
	assert.Equal(t, 0, *req.Slowmode)
}
