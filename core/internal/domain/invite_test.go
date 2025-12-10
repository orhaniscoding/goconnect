package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== InviteToken.IsValid Tests ====================

func TestInviteToken_IsValid(t *testing.T) {
	t.Run("Valid When Not Revoked And Not Expired", func(t *testing.T) {
		token := &InviteToken{
			ID:        "inv123",
			Token:     "abc123",
			ExpiresAt: time.Now().Add(time.Hour),
			UsesMax:   0, // Unlimited
			UsesLeft:  0,
		}
		assert.True(t, token.IsValid())
	})

	t.Run("Invalid When Revoked", func(t *testing.T) {
		revokedAt := time.Now()
		token := &InviteToken{
			ID:        "inv123",
			Token:     "abc123",
			ExpiresAt: time.Now().Add(time.Hour),
			RevokedAt: &revokedAt,
		}
		assert.False(t, token.IsValid())
	})

	t.Run("Invalid When Expired", func(t *testing.T) {
		token := &InviteToken{
			ID:        "inv123",
			Token:     "abc123",
			ExpiresAt: time.Now().Add(-time.Hour), // Expired
		}
		assert.False(t, token.IsValid())
	})

	t.Run("Invalid When Uses Exhausted", func(t *testing.T) {
		token := &InviteToken{
			ID:        "inv123",
			Token:     "abc123",
			ExpiresAt: time.Now().Add(time.Hour),
			UsesMax:   5,
			UsesLeft:  0,
		}
		assert.False(t, token.IsValid())
	})

	t.Run("Valid When Uses Remaining", func(t *testing.T) {
		token := &InviteToken{
			ID:        "inv123",
			Token:     "abc123",
			ExpiresAt: time.Now().Add(time.Hour),
			UsesMax:   5,
			UsesLeft:  3,
		}
		assert.True(t, token.IsValid())
	})
}

// ==================== InviteToken.DecrementUse Tests ====================

func TestInviteToken_DecrementUse(t *testing.T) {
	t.Run("Decrements Uses Left", func(t *testing.T) {
		token := &InviteToken{
			ID:        "inv123",
			Token:     "abc123",
			ExpiresAt: time.Now().Add(time.Hour),
			UsesMax:   5,
			UsesLeft:  5,
		}

		err := token.DecrementUse()
		require.NoError(t, err)
		assert.Equal(t, 4, token.UsesLeft)
	})

	t.Run("Does Not Decrement For Unlimited", func(t *testing.T) {
		token := &InviteToken{
			ID:        "inv123",
			Token:     "abc123",
			ExpiresAt: time.Now().Add(time.Hour),
			UsesMax:   0, // Unlimited
			UsesLeft:  0,
		}

		err := token.DecrementUse()
		require.NoError(t, err)
		assert.Equal(t, 0, token.UsesLeft) // Still 0
	})

	t.Run("Returns Error When Invalid", func(t *testing.T) {
		token := &InviteToken{
			ID:        "inv123",
			Token:     "abc123",
			ExpiresAt: time.Now().Add(-time.Hour), // Expired
		}

		err := token.DecrementUse()
		require.Error(t, err)
		domErr, ok := err.(*Error)
		require.True(t, ok)
		assert.Equal(t, ErrInviteTokenExpired, domErr.Code)
	})
}

// ==================== GenerateInviteToken Tests ====================

func TestGenerateInviteToken(t *testing.T) {
	t.Run("Generates Token", func(t *testing.T) {
		token, err := GenerateInviteToken()
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Len(t, token, 32) // 16 bytes = 32 hex chars
	})

	t.Run("Generates Unique Tokens", func(t *testing.T) {
		token1, err := GenerateInviteToken()
		require.NoError(t, err)

		token2, err := GenerateInviteToken()
		require.NoError(t, err)

		assert.NotEqual(t, token1, token2)
	})
}

// ==================== GenerateInviteID Tests ====================

func TestGenerateInviteID(t *testing.T) {
	t.Run("Generates ID With Prefix", func(t *testing.T) {
		id := GenerateInviteID()
		assert.Contains(t, id, "inv_")
	})

	t.Run("Generates Unique IDs", func(t *testing.T) {
		id1 := GenerateInviteID()
		id2 := GenerateInviteID()
		assert.NotEqual(t, id1, id2)
	})
}

// ==================== Struct Tests ====================

func TestInviteToken(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		now := time.Now()
		token := InviteToken{
			ID:        "inv123",
			NetworkID: "net123",
			TenantID:  "tenant123",
			Token:     "abc123xyz",
			CreatedBy: "user123",
			ExpiresAt: now.Add(24 * time.Hour),
			UsesMax:   10,
			UsesLeft:  5,
			CreatedAt: now,
		}

		assert.Equal(t, "inv123", token.ID)
		assert.Equal(t, "net123", token.NetworkID)
		assert.Equal(t, 10, token.UsesMax)
	})
}

func TestCreateInviteRequest(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		req := CreateInviteRequest{
			ExpiresIn: 3600, // 1 hour
			UsesMax:   5,
		}

		assert.Equal(t, 3600, req.ExpiresIn)
		assert.Equal(t, 5, req.UsesMax)
	})
}

func TestInviteTokenResponse(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		resp := InviteTokenResponse{
			ID:        "inv123",
			NetworkID: "net123",
			Token:     "abc123",
			InviteURL: "https://example.com/join/abc123",
			ExpiresAt: time.Now().Add(time.Hour),
			UsesMax:   10,
			UsesLeft:  5,
			CreatedAt: time.Now(),
			IsActive:  true,
		}

		assert.Equal(t, "inv123", resp.ID)
		assert.True(t, resp.IsActive)
	})
}
