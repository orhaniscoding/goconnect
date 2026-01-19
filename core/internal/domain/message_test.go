package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMessage_IsDeleted(t *testing.T) {
	t.Run("not deleted", func(t *testing.T) {
		msg := &Message{}
		assert.False(t, msg.IsDeleted())
	})

	t.Run("deleted", func(t *testing.T) {
		now := time.Now()
		msg := &Message{DeletedAt: &now}
		assert.True(t, msg.IsDeleted())
	})
}

func TestMessage_IsEdited(t *testing.T) {
	t.Run("not edited", func(t *testing.T) {
		msg := &Message{}
		assert.False(t, msg.IsEdited())
	})

	t.Run("edited", func(t *testing.T) {
		now := time.Now()
		msg := &Message{EditedAt: &now}
		assert.True(t, msg.IsEdited())
	})
}

func TestMessage_IsReply(t *testing.T) {
	t.Run("not a reply", func(t *testing.T) {
		msg := &Message{}
		assert.False(t, msg.IsReply())
	})

	t.Run("is a reply", func(t *testing.T) {
		replyToID := "msg-123"
		msg := &Message{ReplyToID: &replyToID}
		assert.True(t, msg.IsReply())
	})
}

func TestMessage_IsThreadReply(t *testing.T) {
	t.Run("not a thread reply", func(t *testing.T) {
		msg := &Message{}
		assert.False(t, msg.IsThreadReply())
	})

	t.Run("is a thread reply", func(t *testing.T) {
		threadID := "msg-thread"
		msg := &Message{ThreadID: &threadID}
		assert.True(t, msg.IsThreadReply())
	})
}

func TestGenerateMessageID(t *testing.T) {
	id1 := GenerateMessageID()
	id2 := GenerateMessageID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "msg_")
}

func TestGenerateAttachmentID(t *testing.T) {
	id1 := GenerateAttachmentID()
	id2 := GenerateAttachmentID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "att_")
}

func TestAttachment(t *testing.T) {
	att := Attachment{
		ID:          "att-123",
		Filename:    "image.png",
		Size:        1024,
		ContentType: "image/png",
		URL:         "https://cdn.example.com/image.png",
	}

	assert.Equal(t, "att-123", att.ID)
	assert.Equal(t, "image.png", att.Filename)
	assert.Equal(t, int64(1024), att.Size)
	assert.Equal(t, "image/png", att.ContentType)
}

func TestEmbed(t *testing.T) {
	embed := Embed{
		Type:        "link",
		Title:       "Example Site",
		Description: "A description",
		URL:         "https://example.com",
	}

	assert.Equal(t, "link", embed.Type)
	assert.Equal(t, "Example Site", embed.Title)
}

func TestReactionSummary(t *testing.T) {
	summary := ReactionSummary{
		Emoji: "👍",
		Count: 5,
		Me:    true,
	}

	assert.Equal(t, "👍", summary.Emoji)
	assert.Equal(t, 5, summary.Count)
	assert.True(t, summary.Me)
}
