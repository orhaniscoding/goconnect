package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// CHANNEL (Text/Voice communication)
// ═══════════════════════════════════════════════════════════════════════════

// ChannelType defines the type of channel
type ChannelType string

const (
	ChannelTypeText         ChannelType = "text"
	ChannelTypeVoice        ChannelType = "voice"
	ChannelTypeAnnouncement ChannelType = "announcement"
)

// Channel represents a text or voice channel
type Channel struct {
	ID          string      `json:"id" db:"id"`
	TenantID    *string     `json:"tenant_id,omitempty" db:"tenant_id"`
	SectionID   *string     `json:"section_id,omitempty" db:"section_id"`
	NetworkID   *string     `json:"network_id,omitempty" db:"network_id"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description,omitempty" db:"description"`
	Type        ChannelType `json:"type" db:"type"`
	Position    int         `json:"position" db:"position"`

	// Voice channel settings
	Bitrate   int `json:"bitrate" db:"bitrate"`
	UserLimit int `json:"user_limit" db:"user_limit"`

	// Moderation
	Slowmode int  `json:"slowmode" db:"slowmode"`
	NSFW     bool `json:"nsfw" db:"nsfw"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Enriched fields
	Messages       []Message       `json:"messages,omitempty" db:"-"`
	VoiceStates    []VoiceState    `json:"voice_states,omitempty" db:"-"`
	MemberCount    int             `json:"member_count,omitempty" db:"-"`
}

// IsDeleted checks if the channel is soft-deleted
func (c *Channel) IsDeleted() bool {
	return c.DeletedAt != nil
}

// IsVoice checks if this is a voice channel
func (c *Channel) IsVoice() bool {
	return c.Type == ChannelTypeVoice
}

// IsText checks if this is a text channel
func (c *Channel) IsText() bool {
	return c.Type == ChannelTypeText
}

// GetParentID returns the parent ID (tenant, section, or network)
func (c *Channel) GetParentID() string {
	if c.TenantID != nil {
		return *c.TenantID
	}
	if c.SectionID != nil {
		return *c.SectionID
	}
	if c.NetworkID != nil {
		return *c.NetworkID
	}
	return ""
}

// CreateChannelRequest for creating a new channel
type CreateChannelRequest struct {
	Name        string      `json:"name" binding:"required,min=1,max=100"`
	Description string      `json:"description,omitempty" binding:"max=500"`
	Type        ChannelType `json:"type" binding:"required,oneof=text voice announcement"`
	Bitrate     *int        `json:"bitrate,omitempty" binding:"omitempty,min=8000,max=384000"`
	UserLimit   *int        `json:"user_limit,omitempty" binding:"omitempty,min=0,max=99"`
	Slowmode    *int        `json:"slowmode,omitempty" binding:"omitempty,min=0,max=21600"`
	NSFW        bool        `json:"nsfw,omitempty"`
}

// ApplyDefaults sets default values
func (r *CreateChannelRequest) ApplyDefaults() {
	if r.Type == "" {
		r.Type = ChannelTypeText
	}
	if r.Bitrate == nil {
		defaultBitrate := 64000
		r.Bitrate = &defaultBitrate
	}
	if r.UserLimit == nil {
		defaultLimit := 0
		r.UserLimit = &defaultLimit
	}
	if r.Slowmode == nil {
		defaultSlowmode := 0
		r.Slowmode = &defaultSlowmode
	}
}

// UpdateChannelRequest for updating a channel
type UpdateChannelRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
	Position    *int    `json:"position,omitempty" binding:"omitempty,min=0"`
	Bitrate     *int    `json:"bitrate,omitempty" binding:"omitempty,min=8000,max=384000"`
	UserLimit   *int    `json:"user_limit,omitempty" binding:"omitempty,min=0,max=99"`
	Slowmode    *int    `json:"slowmode,omitempty" binding:"omitempty,min=0,max=21600"`
	NSFW        *bool   `json:"nsfw,omitempty"`
}

// GenerateChannelID generates a new channel ID
func GenerateChannelID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("ch_%s", hex.EncodeToString(bytes))
}

// ValidateChannelName validates and sanitizes a channel name
func ValidateChannelName(name string) (string, error) {
	name = strings.TrimSpace(name)

	// Collapse consecutive spaces and convert to lowercase with hyphens (Discord style)
	spaceRegex := regexp.MustCompile(`\s+`)
	name = spaceRegex.ReplaceAllString(name, "-")
	name = strings.ToLower(name)

	// Remove invalid characters
	validChars := regexp.MustCompile(`[^a-z0-9-_]`)
	name = validChars.ReplaceAllString(name, "")

	if len(name) < 1 {
		return "", NewError(ErrValidation, "channel name is required", map[string]any{
			"field": "name",
			"min":   1,
		})
	}

	if len(name) > 100 {
		return "", NewError(ErrValidation, "channel name must be at most 100 characters", map[string]any{
			"field": "name",
			"max":   100,
		})
	}

	return name, nil
}

// ListChannelsRequest for listing channels
type ListChannelsRequest struct {
	Type   string `form:"type"`             // Filter by type
	Limit  int    `form:"limit,default=50"` // Max 100
	Cursor string `form:"cursor"`
}
