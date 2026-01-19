package domain

import (
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// VOICE STATE (Who is in voice channels)
// ═══════════════════════════════════════════════════════════════════════════

// VoiceState represents a user's current voice connection
type VoiceState struct {
	UserID      string    `json:"user_id" db:"user_id"`
	ChannelID   string    `json:"channel_id" db:"channel_id"`
	SelfMute    bool      `json:"self_mute" db:"self_mute"`
	SelfDeaf    bool      `json:"self_deaf" db:"self_deaf"`
	ServerMute  bool      `json:"server_mute" db:"server_mute"`
	ServerDeaf  bool      `json:"server_deaf" db:"server_deaf"`
	ConnectedAt time.Time `json:"connected_at" db:"connected_at"`

	// Enriched fields
	User    *User    `json:"user,omitempty" db:"-"`
	Channel *Channel `json:"channel,omitempty" db:"-"`
}

// IsMuted checks if the user is muted (self or server)
func (v *VoiceState) IsMuted() bool {
	return v.SelfMute || v.ServerMute
}

// IsDeafened checks if the user is deafened (self or server)
func (v *VoiceState) IsDeafened() bool {
	return v.SelfDeaf || v.ServerDeaf
}

// UpdateVoiceStateRequest for updating voice state
type UpdateVoiceStateRequest struct {
	SelfMute *bool `json:"self_mute,omitempty"`
	SelfDeaf *bool `json:"self_deaf,omitempty"`
}

// ModerateVoiceStateRequest for moderator actions
type ModerateVoiceStateRequest struct {
	ServerMute *bool   `json:"server_mute,omitempty"`
	ServerDeaf *bool   `json:"server_deaf,omitempty"`
	MoveToChannel *string `json:"move_to_channel,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════════════
// USER PRESENCE (Online status)
// ═══════════════════════════════════════════════════════════════════════════

// PresenceStatus represents online status
type PresenceStatus string

const (
	PresenceStatusOnline    PresenceStatus = "online"
	PresenceStatusIdle      PresenceStatus = "idle"
	PresenceStatusDND       PresenceStatus = "dnd"       // Do Not Disturb
	PresenceStatusInvisible PresenceStatus = "invisible"
	PresenceStatusOffline   PresenceStatus = "offline"
)

// ActivityType represents what the user is doing
type ActivityType string

const (
	ActivityTypePlaying   ActivityType = "playing"
	ActivityTypeListening ActivityType = "listening"
	ActivityTypeWatching  ActivityType = "watching"
	ActivityTypeStreaming ActivityType = "streaming"
)

// UserPresence represents a user's online presence
type UserPresence struct {
	UserID        string         `json:"user_id" db:"user_id"`
	Status        PresenceStatus `json:"status" db:"status"`
	CustomStatus  *string        `json:"custom_status,omitempty" db:"custom_status"`
	ActivityType  *ActivityType  `json:"activity_type,omitempty" db:"activity_type"`
	ActivityName  *string        `json:"activity_name,omitempty" db:"activity_name"`
	LastSeen      time.Time      `json:"last_seen" db:"last_seen"`
	DesktopStatus *PresenceStatus `json:"desktop_status,omitempty" db:"desktop_status"`
	MobileStatus  *PresenceStatus `json:"mobile_status,omitempty" db:"mobile_status"`
	WebStatus     *PresenceStatus `json:"web_status,omitempty" db:"web_status"`

	// Enriched fields
	User *User `json:"user,omitempty" db:"-"`
}

// IsOnline checks if the user is online (not offline or invisible to others)
func (p *UserPresence) IsOnline() bool {
	return p.Status == PresenceStatusOnline || p.Status == PresenceStatusIdle || p.Status == PresenceStatusDND
}

// GetEffectiveStatus returns the visible status (invisible shows as offline)
func (p *UserPresence) GetEffectiveStatus() PresenceStatus {
	if p.Status == PresenceStatusInvisible {
		return PresenceStatusOffline
	}
	return p.Status
}

// UpdatePresenceRequest for updating presence
type UpdatePresenceRequest struct {
	Status       *PresenceStatus `json:"status,omitempty" binding:"omitempty,oneof=online idle dnd invisible"`
	CustomStatus *string         `json:"custom_status,omitempty" binding:"omitempty,max=128"`
	ActivityType *ActivityType   `json:"activity_type,omitempty" binding:"omitempty,oneof=playing listening watching streaming"`
	ActivityName *string         `json:"activity_name,omitempty" binding:"omitempty,max=128"`
}
