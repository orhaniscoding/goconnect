package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// FRIENDSHIP (User-to-user relationships)
// ═══════════════════════════════════════════════════════════════════════════

// FriendshipStatus represents the status of a friendship
type FriendshipStatus string

const (
	FriendshipStatusPending  FriendshipStatus = "pending"
	FriendshipStatusAccepted FriendshipStatus = "accepted"
	FriendshipStatusBlocked  FriendshipStatus = "blocked"
)

// Friendship represents a relationship between two users
type Friendship struct {
	ID         string           `json:"id" db:"id"`
	UserID     string           `json:"user_id" db:"user_id"`
	FriendID   string           `json:"friend_id" db:"friend_id"`
	Status     FriendshipStatus `json:"status" db:"status"`
	CreatedAt  time.Time        `json:"created_at" db:"created_at"`
	AcceptedAt *time.Time       `json:"accepted_at,omitempty" db:"accepted_at"`

	// Enriched fields
	User   *User `json:"user,omitempty" db:"-"`
	Friend *User `json:"friend,omitempty" db:"-"`
}

// IsPending checks if the friendship request is pending
func (f *Friendship) IsPending() bool {
	return f.Status == FriendshipStatusPending
}

// IsAccepted checks if the friendship is accepted
func (f *Friendship) IsAccepted() bool {
	return f.Status == FriendshipStatusAccepted
}

// IsBlocked checks if the user is blocked
func (f *Friendship) IsBlocked() bool {
	return f.Status == FriendshipStatusBlocked
}

// SendFriendRequestRequest for sending a friend request
type SendFriendRequestRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// ListFriendsRequest for listing friends
type ListFriendsRequest struct {
	Status string `form:"status"` // pending, accepted, blocked
	Limit  int    `form:"limit,default=50"`
	Cursor string `form:"cursor"`
}

// GenerateFriendshipID generates a new friendship ID
func GenerateFriendshipID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("fr_%s", hex.EncodeToString(bytes))
}

// ═══════════════════════════════════════════════════════════════════════════
// DM CHANNEL (Direct messages)
// ═══════════════════════════════════════════════════════════════════════════

// DMChannelType represents the type of DM channel
type DMChannelType string

const (
	DMChannelTypeDM      DMChannelType = "dm"
	DMChannelTypeGroupDM DMChannelType = "group_dm"
)

// DMChannel represents a direct message conversation
type DMChannel struct {
	ID        string        `json:"id" db:"id"`
	Type      DMChannelType `json:"type" db:"type"`
	Name      *string       `json:"name,omitempty" db:"name"`
	Icon      *string       `json:"icon,omitempty" db:"icon"`
	OwnerID   *string       `json:"owner_id,omitempty" db:"owner_id"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`

	// Enriched fields
	Members      []DMChannelMember `json:"members,omitempty" db:"-"`
	LastMessage  *DMMessage        `json:"last_message,omitempty" db:"-"`
	UnreadCount  int               `json:"unread_count,omitempty" db:"-"`
}

// IsGroupDM checks if this is a group DM
func (d *DMChannel) IsGroupDM() bool {
	return d.Type == DMChannelTypeGroupDM
}

// DMChannelMember represents a member of a DM channel
type DMChannelMember struct {
	ChannelID string    `json:"channel_id" db:"channel_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Muted     bool      `json:"muted" db:"muted"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`

	// Enriched fields
	User *User `json:"user,omitempty" db:"-"`
}

// DMMessage represents a message in a DM channel
type DMMessage struct {
	ID              string     `json:"id" db:"id"`
	ChannelID       string     `json:"channel_id" db:"channel_id"`
	AuthorID        string     `json:"author_id" db:"author_id"`
	Content         string     `json:"content" db:"content"`
	Encrypted       bool       `json:"encrypted" db:"encrypted"`
	Attachments     []Attachment `json:"attachments" db:"-"`
	AttachmentsJSON string     `json:"-" db:"attachments"`
	ReplyToID       *string    `json:"reply_to_id,omitempty" db:"reply_to_id"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	EditedAt        *time.Time `json:"edited_at,omitempty" db:"edited_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Enriched fields
	Author  *User      `json:"author,omitempty" db:"-"`
	ReplyTo *DMMessage `json:"reply_to,omitempty" db:"-"`
}

// IsDeleted checks if the message is deleted
func (m *DMMessage) IsDeleted() bool {
	return m.DeletedAt != nil
}

// CreateDMChannelRequest for creating a 1:1 DM
type CreateDMChannelRequest struct {
	RecipientID string `json:"recipient_id" binding:"required"`
}

// CreateGroupDMRequest for creating a group DM
type CreateGroupDMRequest struct {
	Name       string   `json:"name,omitempty" binding:"max=100"`
	Recipients []string `json:"recipients" binding:"required,min=1,max=9"`
}

// SendDMMessageRequest for sending a DM
type SendDMMessageRequest struct {
	Content       string   `json:"content" binding:"required,min=1,max=4000"`
	ReplyToID     *string  `json:"reply_to_id,omitempty"`
	AttachmentIDs []string `json:"attachment_ids,omitempty"`
}

// ListDMChannelsRequest for listing DM channels
type ListDMChannelsRequest struct {
	Limit  int    `form:"limit,default=50"`
	Cursor string `form:"cursor"`
}

// GenerateDMChannelID generates a new DM channel ID
func GenerateDMChannelID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("dm_%s", hex.EncodeToString(bytes))
}

// GenerateDMMessageID generates a new DM message ID
func GenerateDMMessageID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("dmsg_%s", hex.EncodeToString(bytes))
}
