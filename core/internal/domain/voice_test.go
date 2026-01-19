package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Helper function for tests
func timeNow() time.Time {
	return time.Now()
}

func TestVoiceState_IsMuted(t *testing.T) {
	t.Run("not muted", func(t *testing.T) {
		state := &VoiceState{}
		assert.False(t, state.IsMuted())
	})

	t.Run("self muted", func(t *testing.T) {
		state := &VoiceState{SelfMute: true}
		assert.True(t, state.IsMuted())
	})

	t.Run("server muted", func(t *testing.T) {
		state := &VoiceState{ServerMute: true}
		assert.True(t, state.IsMuted())
	})

	t.Run("both muted", func(t *testing.T) {
		state := &VoiceState{SelfMute: true, ServerMute: true}
		assert.True(t, state.IsMuted())
	})
}

func TestVoiceState_IsDeafened(t *testing.T) {
	t.Run("not deafened", func(t *testing.T) {
		state := &VoiceState{}
		assert.False(t, state.IsDeafened())
	})

	t.Run("self deafened", func(t *testing.T) {
		state := &VoiceState{SelfDeaf: true}
		assert.True(t, state.IsDeafened())
	})

	t.Run("server deafened", func(t *testing.T) {
		state := &VoiceState{ServerDeaf: true}
		assert.True(t, state.IsDeafened())
	})
}

func TestUserPresence_IsOnline(t *testing.T) {
	tests := []struct {
		status   PresenceStatus
		expected bool
	}{
		{PresenceStatusOnline, true},
		{PresenceStatusIdle, true},
		{PresenceStatusDND, true},
		{PresenceStatusInvisible, false},
		{PresenceStatusOffline, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			presence := &UserPresence{Status: tt.status}
			assert.Equal(t, tt.expected, presence.IsOnline())
		})
	}
}

func TestUserPresence_GetEffectiveStatus(t *testing.T) {
	t.Run("invisible shows as offline", func(t *testing.T) {
		presence := &UserPresence{Status: PresenceStatusInvisible}
		assert.Equal(t, PresenceStatusOffline, presence.GetEffectiveStatus())
	})

	t.Run("online shows as online", func(t *testing.T) {
		presence := &UserPresence{Status: PresenceStatusOnline}
		assert.Equal(t, PresenceStatusOnline, presence.GetEffectiveStatus())
	})

	t.Run("dnd shows as dnd", func(t *testing.T) {
		presence := &UserPresence{Status: PresenceStatusDND}
		assert.Equal(t, PresenceStatusDND, presence.GetEffectiveStatus())
	})
}

func TestPresenceStatusConstants(t *testing.T) {
	assert.Equal(t, PresenceStatus("online"), PresenceStatusOnline)
	assert.Equal(t, PresenceStatus("idle"), PresenceStatusIdle)
	assert.Equal(t, PresenceStatus("dnd"), PresenceStatusDND)
	assert.Equal(t, PresenceStatus("invisible"), PresenceStatusInvisible)
	assert.Equal(t, PresenceStatus("offline"), PresenceStatusOffline)
}

func TestActivityTypeConstants(t *testing.T) {
	assert.Equal(t, ActivityType("playing"), ActivityTypePlaying)
	assert.Equal(t, ActivityType("listening"), ActivityTypeListening)
	assert.Equal(t, ActivityType("watching"), ActivityTypeWatching)
	assert.Equal(t, ActivityType("streaming"), ActivityTypeStreaming)
}
