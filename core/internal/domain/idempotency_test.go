package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== IdempotencyTTL Tests ====================

func TestIdempotencyTTL(t *testing.T) {
	t.Run("TTL Is 24 Hours", func(t *testing.T) {
		assert.Equal(t, 24*time.Hour, IdempotencyTTL)
	})
}

// ==================== GenerateIdempotencyKey Tests ====================

func TestGenerateIdempotencyKey(t *testing.T) {
	t.Run("Generates Unique Keys", func(t *testing.T) {
		key1 := GenerateIdempotencyKey()
		key2 := GenerateIdempotencyKey()

		assert.NotEmpty(t, key1)
		assert.NotEmpty(t, key2)
		assert.NotEqual(t, key1, key2)
	})

	t.Run("Key Has Correct Prefix", func(t *testing.T) {
		key := GenerateIdempotencyKey()
		assert.Contains(t, key, "idem_")
	})
}

// ==================== NewIdempotencyRecord Tests ====================

func TestNewIdempotencyRecord(t *testing.T) {
	t.Run("Creates Record With Correct Fields", func(t *testing.T) {
		key := "test_key"
		bodyHash := "abc123hash"

		before := time.Now()
		record := NewIdempotencyRecord(key, bodyHash)
		after := time.Now()

		require.NotNil(t, record)
		assert.Equal(t, key, record.Key)
		assert.Equal(t, bodyHash, record.BodyHash)
		assert.True(t, record.CreatedAt.After(before.Add(-time.Second)))
		assert.True(t, record.CreatedAt.Before(after.Add(time.Second)))
		assert.Equal(t, record.CreatedAt.Add(IdempotencyTTL), record.ExpiresAt)
	})

	t.Run("Response Is Nil By Default", func(t *testing.T) {
		record := NewIdempotencyRecord("key", "hash")
		assert.Nil(t, record.Response)
	})
}

// ==================== IdempotencyRecord.IsExpired Tests ====================

func TestIdempotencyRecord_IsExpired(t *testing.T) {
	t.Run("Not Expired When Fresh", func(t *testing.T) {
		record := NewIdempotencyRecord("key", "hash")
		assert.False(t, record.IsExpired())
	})

	t.Run("Expired When Past ExpiresAt", func(t *testing.T) {
		record := &IdempotencyRecord{
			Key:       "key",
			BodyHash:  "hash",
			CreatedAt: time.Now().Add(-25 * time.Hour),
			ExpiresAt: time.Now().Add(-time.Hour), // Expired 1 hour ago
		}
		assert.True(t, record.IsExpired())
	})

	t.Run("Not Expired When ExpiresAt Is In Future", func(t *testing.T) {
		record := &IdempotencyRecord{
			Key:       "key",
			BodyHash:  "hash",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Hour),
		}
		assert.False(t, record.IsExpired())
	})
}

// ==================== IdempotencyRecord Struct Tests ====================

func TestIdempotencyRecord(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		now := time.Now()
		record := IdempotencyRecord{
			Key:       "key123",
			BodyHash:  "sha256:abc123",
			Response:  []byte(`{"status":"ok"}`),
			CreatedAt: now,
			ExpiresAt: now.Add(24 * time.Hour),
		}

		assert.Equal(t, "key123", record.Key)
		assert.Equal(t, "sha256:abc123", record.BodyHash)
		assert.NotNil(t, record.Response)
	})
}
