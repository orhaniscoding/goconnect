package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFriendship_Status(t *testing.T) {
	t.Run("pending", func(t *testing.T) {
		f := &Friendship{Status: FriendshipStatusPending}
		assert.True(t, f.IsPending())
		assert.False(t, f.IsAccepted())
		assert.False(t, f.IsBlocked())
	})

	t.Run("accepted", func(t *testing.T) {
		f := &Friendship{Status: FriendshipStatusAccepted}
		assert.False(t, f.IsPending())
		assert.True(t, f.IsAccepted())
		assert.False(t, f.IsBlocked())
	})

	t.Run("blocked", func(t *testing.T) {
		f := &Friendship{Status: FriendshipStatusBlocked}
		assert.False(t, f.IsPending())
		assert.False(t, f.IsAccepted())
		assert.True(t, f.IsBlocked())
	})
}

func TestGenerateFriendshipID(t *testing.T) {
	id1 := GenerateFriendshipID()
	id2 := GenerateFriendshipID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "fr_")
}

func TestDMChannel_IsGroupDM(t *testing.T) {
	t.Run("regular DM", func(t *testing.T) {
		dm := &DMChannel{Type: DMChannelTypeDM}
		assert.False(t, dm.IsGroupDM())
	})

	t.Run("group DM", func(t *testing.T) {
		dm := &DMChannel{Type: DMChannelTypeGroupDM}
		assert.True(t, dm.IsGroupDM())
	})
}

func TestDMMessage_IsDeleted(t *testing.T) {
	t.Run("not deleted", func(t *testing.T) {
		msg := &DMMessage{}
		assert.False(t, msg.IsDeleted())
	})

	t.Run("deleted", func(t *testing.T) {
		now := time.Now()
		msg := &DMMessage{DeletedAt: &now}
		assert.True(t, msg.IsDeleted())
	})
}

func TestGenerateDMChannelID(t *testing.T) {
	id1 := GenerateDMChannelID()
	id2 := GenerateDMChannelID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "dm_")
}

func TestGenerateDMMessageID(t *testing.T) {
	id1 := GenerateDMMessageID()
	id2 := GenerateDMMessageID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "dmsg_")
}

func TestFriendshipStatusConstants(t *testing.T) {
	assert.Equal(t, FriendshipStatus("pending"), FriendshipStatusPending)
	assert.Equal(t, FriendshipStatus("accepted"), FriendshipStatusAccepted)
	assert.Equal(t, FriendshipStatus("blocked"), FriendshipStatusBlocked)
}

func TestDMChannelTypeConstants(t *testing.T) {
	assert.Equal(t, DMChannelType("dm"), DMChannelTypeDM)
	assert.Equal(t, DMChannelType("group_dm"), DMChannelTypeGroupDM)
}
