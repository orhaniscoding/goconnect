package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// MESSAGE (Channel messages with rich features)
// ═══════════════════════════════════════════════════════════════════════════

// Message represents a message in a channel
type Message struct {
	ID              string     `json:"id" db:"id"`
	ChannelID       string     `json:"channel_id" db:"channel_id"`
	AuthorID        string     `json:"author_id" db:"author_id"`
	Content         string     `json:"content" db:"content"`
	ReplyToID       *string    `json:"reply_to_id,omitempty" db:"reply_to_id"`
	ThreadID        *string    `json:"thread_id,omitempty" db:"thread_id"`
	Attachments     []Attachment `json:"attachments" db:"-"`
	AttachmentsJSON string     `json:"-" db:"attachments"`
	Embeds          []Embed    `json:"embeds" db:"-"`
	EmbedsJSON      string     `json:"-" db:"embeds"`
	Mentions        []string   `json:"mentions" db:"mentions"`
	MentionRoles    []string   `json:"mention_roles" db:"mention_roles"`
	MentionEveryone bool       `json:"mention_everyone" db:"mention_everyone"`
	Pinned          bool       `json:"pinned" db:"pinned"`
	EditedAt        *time.Time `json:"edited_at,omitempty" db:"edited_at"`
	Encrypted       bool       `json:"encrypted" db:"encrypted"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Enriched fields
	Author    *User      `json:"author,omitempty" db:"-"`
	ReplyTo   *Message   `json:"reply_to,omitempty" db:"-"`
	Reactions []Reaction `json:"reactions,omitempty" db:"-"`
}

// IsDeleted checks if the message is deleted
func (m *Message) IsDeleted() bool {
	return m.DeletedAt != nil
}

// IsEdited checks if the message was edited
func (m *Message) IsEdited() bool {
	return m.EditedAt != nil
}

// IsReply checks if this is a reply to another message
func (m *Message) IsReply() bool {
	return m.ReplyToID != nil
}

// IsThreadReply checks if this is part of a thread
func (m *Message) IsThreadReply() bool {
	return m.ThreadID != nil
}

// Attachment represents a file attached to a message
type Attachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
	Width       *int   `json:"width,omitempty"`
	Height      *int   `json:"height,omitempty"`
}

// Embed represents a rich embed (link preview, etc.)
type Embed struct {
	Type        string  `json:"type"`
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	URL         string  `json:"url,omitempty"`
	Thumbnail   *string `json:"thumbnail,omitempty"`
	Color       *string `json:"color,omitempty"`
}

// Reaction represents an emoji reaction on a message
type Reaction struct {
	MessageID string    `json:"message_id" db:"message_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Emoji     string    `json:"emoji" db:"emoji"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Enriched/aggregated fields
	Count int   `json:"count,omitempty" db:"-"`
	Users []User `json:"users,omitempty" db:"-"`
}

// ReactionSummary is an aggregated view of reactions
type ReactionSummary struct {
	Emoji string `json:"emoji"`
	Count int    `json:"count"`
	Me    bool   `json:"me"` // Did the current user react with this emoji
}

// CreateMessageRequest for creating a new message
type CreateMessageRequest struct {
	Content         string   `json:"content" binding:"required,min=1,max=4000"`
	ReplyToID       *string  `json:"reply_to_id,omitempty"`
	AttachmentIDs   []string `json:"attachment_ids,omitempty"`
	MentionEveryone bool     `json:"mention_everyone,omitempty"`
}

// UpdateMessageRequest for editing a message
type UpdateMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=4000"`
}

// ListMessagesRequest for listing messages
type ListMessagesRequest struct {
	Before string `form:"before"`            // Get messages before this ID
	After  string `form:"after"`             // Get messages after this ID
	Limit  int    `form:"limit,default=50"`  // Max 100
}

// GenerateMessageID generates a new message ID
func GenerateMessageID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("msg_%s", hex.EncodeToString(bytes))
}

// GenerateAttachmentID generates a new attachment ID
func GenerateAttachmentID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("att_%s", hex.EncodeToString(bytes))
}
