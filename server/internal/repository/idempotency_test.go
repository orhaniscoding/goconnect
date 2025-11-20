package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test idempotency record
func mkIdempotencyRecord(key, bodyHash string, ttl time.Duration) *domain.IdempotencyRecord {
	now := time.Now()
	return &domain.IdempotencyRecord{
		Key:       key,
		BodyHash:  bodyHash,
		Response:  []byte(`{"status":"success"}`),
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}
}

func TestNewInMemoryIdempotencyRepository(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.records)
	assert.Equal(t, 0, len(repo.records))
}

func TestIdempotencyRepository_Set_Success(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()
	record := mkIdempotencyRecord("key-1", "hash-1", 24*time.Hour)

	err := repo.Set(ctx, record)

	require.NoError(t, err)
	assert.Equal(t, 1, len(repo.records))
}

func TestIdempotencyRepository_Get_Success(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()
	original := mkIdempotencyRecord("key-1", "hash-1", 24*time.Hour)
	repo.Set(ctx, original)

	retrieved, err := repo.Get(ctx, "key-1")

	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, original.Key, retrieved.Key)
	assert.Equal(t, original.BodyHash, retrieved.BodyHash)
	assert.Equal(t, original.Response, retrieved.Response)
}

func TestIdempotencyRepository_Get_NotFound(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()

	retrieved, err := repo.Get(ctx, "non-existent")

	require.Error(t, err)
	assert.Nil(t, retrieved)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
	assert.Contains(t, domainErr.Message, "not found")
}

func TestIdempotencyRepository_Get_Expired(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()

	// Create record that expires in 1 millisecond
	record := mkIdempotencyRecord("expired-key", "hash-1", 1*time.Millisecond)
	repo.Set(ctx, record)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	retrieved, err := repo.Get(ctx, "expired-key")

	require.Error(t, err)
	assert.Nil(t, retrieved)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
	assert.Contains(t, domainErr.Message, "expired")
}

func TestIdempotencyRepository_Set_Update(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()

	// Set initial record
	record1 := mkIdempotencyRecord("key-1", "hash-1", 24*time.Hour)
	record1.Response = []byte(`{"status":"first"}`)
	repo.Set(ctx, record1)

	// Update with new record
	record2 := mkIdempotencyRecord("key-1", "hash-2", 24*time.Hour)
	record2.Response = []byte(`{"status":"second"}`)
	repo.Set(ctx, record2)

	retrieved, err := repo.Get(ctx, "key-1")
	require.NoError(t, err)
	assert.Equal(t, "hash-2", retrieved.BodyHash)
	assert.Equal(t, []byte(`{"status":"second"}`), retrieved.Response)
}

func TestIdempotencyRepository_Delete_Success(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()
	record := mkIdempotencyRecord("key-1", "hash-1", 24*time.Hour)
	repo.Set(ctx, record)

	err := repo.Delete(ctx, "key-1")

	require.NoError(t, err)
	assert.Equal(t, 0, len(repo.records))

	_, err = repo.Get(ctx, "key-1")
	assert.Error(t, err)
}

func TestIdempotencyRepository_Delete_NonExistent(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "non-existent")

	// Delete should succeed even if key doesn't exist
	require.NoError(t, err)
}

func TestIdempotencyRepository_Cleanup_RemovesExpired(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()

	// Add expired record
	expiredRecord := mkIdempotencyRecord("expired-1", "hash-1", -1*time.Hour)
	repo.Set(ctx, expiredRecord)

	// Add valid record
	validRecord := mkIdempotencyRecord("valid-1", "hash-2", 24*time.Hour)
	repo.Set(ctx, validRecord)

	assert.Equal(t, 2, len(repo.records))

	// Run cleanup
	err := repo.Cleanup(ctx)

	require.NoError(t, err)
	assert.Equal(t, 1, len(repo.records))

	// Expired should be gone
	_, err = repo.Get(ctx, "expired-1")
	assert.Error(t, err)

	// Valid should still exist
	_, err = repo.Get(ctx, "valid-1")
	assert.NoError(t, err)
}

func TestIdempotencyRepository_Cleanup_MultipleExpired(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()

	// Add multiple expired records
	for i := 1; i <= 5; i++ {
		record := mkIdempotencyRecord("expired-"+string(rune('a'+i)), "hash", -1*time.Hour)
		repo.Set(ctx, record)
	}

	// Add valid records
	for i := 1; i <= 3; i++ {
		record := mkIdempotencyRecord("valid-"+string(rune('a'+i)), "hash", 24*time.Hour)
		repo.Set(ctx, record)
	}

	assert.Equal(t, 8, len(repo.records))

	err := repo.Cleanup(ctx)

	require.NoError(t, err)
	assert.Equal(t, 3, len(repo.records))
}

func TestIdempotencyRepository_MultipleKeys(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()

	// Create multiple records
	records := []*domain.IdempotencyRecord{
		mkIdempotencyRecord("key-1", "hash-1", 24*time.Hour),
		mkIdempotencyRecord("key-2", "hash-2", 24*time.Hour),
		mkIdempotencyRecord("key-3", "hash-3", 24*time.Hour),
	}

	for _, record := range records {
		err := repo.Set(ctx, record)
		require.NoError(t, err)
	}

	assert.Equal(t, 3, len(repo.records))

	// Verify all can be retrieved
	for i, record := range records {
		retrieved, err := repo.Get(ctx, record.Key)
		require.NoError(t, err)
		assert.Equal(t, records[i].BodyHash, retrieved.BodyHash)
	}
}

func TestIdempotencyRepository_ResponseStorage(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()

	record := mkIdempotencyRecord("key-1", "hash-1", 24*time.Hour)
	record.Response = []byte(`{"status":"success","data":{"id":123}}`)

	repo.Set(ctx, record)

	retrieved, err := repo.Get(ctx, "key-1")
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"status":"success","data":{"id":123}}`), retrieved.Response)
}

func TestIdempotencyRepository_EmptyResponse(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()

	record := mkIdempotencyRecord("key-1", "hash-1", 24*time.Hour)
	record.Response = nil

	repo.Set(ctx, record)

	retrieved, err := repo.Get(ctx, "key-1")
	require.NoError(t, err)
	assert.Nil(t, retrieved.Response)
}

func TestIdempotencyRepository_ConcurrentAccess(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()

	// Set a record
	record := mkIdempotencyRecord("concurrent-key", "hash-1", 24*time.Hour)
	repo.Set(ctx, record)

	// Concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := repo.Get(ctx, "concurrent-key")
			assert.NoError(t, err)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestIdempotencyRepository_FullCycle(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()

	// Set
	record := mkIdempotencyRecord("cycle-key", "hash-1", 24*time.Hour)
	err := repo.Set(ctx, record)
	require.NoError(t, err)

	// Get
	retrieved, err := repo.Get(ctx, "cycle-key")
	require.NoError(t, err)
	assert.Equal(t, "hash-1", retrieved.BodyHash)

	// Update
	record.BodyHash = "hash-updated"
	err = repo.Set(ctx, record)
	require.NoError(t, err)

	retrieved, _ = repo.Get(ctx, "cycle-key")
	assert.Equal(t, "hash-updated", retrieved.BodyHash)

	// Delete
	err = repo.Delete(ctx, "cycle-key")
	require.NoError(t, err)

	_, err = repo.Get(ctx, "cycle-key")
	assert.Error(t, err)
}

func TestIdempotencyRepository_TTLValidation(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()

	// Short TTL record
	shortTTL := mkIdempotencyRecord("short-ttl", "hash-1", 50*time.Millisecond)
	repo.Set(ctx, shortTTL)

	// Should be accessible immediately
	_, err := repo.Get(ctx, "short-ttl")
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	_, err = repo.Get(ctx, "short-ttl")
	assert.Error(t, err)
}

func TestIdempotencyRepository_TimestampPreservation(t *testing.T) {
	repo := NewInMemoryIdempotencyRepository()
	ctx := context.Background()

	record := mkIdempotencyRecord("timestamp-key", "hash-1", 24*time.Hour)
	createdAt := record.CreatedAt
	expiresAt := record.ExpiresAt

	repo.Set(ctx, record)

	retrieved, err := repo.Get(ctx, "timestamp-key")
	require.NoError(t, err)

	assert.Equal(t, createdAt.Unix(), retrieved.CreatedAt.Unix())
	assert.Equal(t, expiresAt.Unix(), retrieved.ExpiresAt.Unix())
}
