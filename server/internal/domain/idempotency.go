package domain

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// IdempotencyRecord represents an idempotency key record
type IdempotencyRecord struct {
	Key       string    `json:"key"`
	BodyHash  string    `json:"body_hash"`
	Response  []byte    `json:"response,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// IdempotencyTTL is the default TTL for idempotency keys (24 hours)
const IdempotencyTTL = 24 * time.Hour

// GenerateIdempotencyKey generates a random idempotency key for testing
func GenerateIdempotencyKey() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(999999999))
	return fmt.Sprintf("idem_%d_%d", time.Now().Unix(), n.Int64())
}

// NewIdempotencyRecord creates a new idempotency record
func NewIdempotencyRecord(key, bodyHash string) *IdempotencyRecord {
	now := time.Now()
	return &IdempotencyRecord{
		Key:       key,
		BodyHash:  bodyHash,
		CreatedAt: now,
		ExpiresAt: now.Add(IdempotencyTTL),
	}
}

// IsExpired checks if the idempotency record has expired
func (r *IdempotencyRecord) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}
