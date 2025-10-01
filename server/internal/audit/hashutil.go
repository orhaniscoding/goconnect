package audit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"hash"
)

// newHasher returns a closure that deterministically hashes input strings using
// HMAC-SHA256(secret, value) truncated to the first 18 bytes (144 bits) and
// base64-url encodes without padding. Returns nil if secret empty.
func newHasher(secret []byte) func(string) string { // legacy single-secret helper (kept for backward compat)
	if len(secret) == 0 { return nil }
	var h hash.Hash
	return func(v string) string {
		h = hmac.New(sha256.New, secret)
		_, _ = h.Write([]byte(v))
		sum := h.Sum(nil)
		return base64.RawURLEncoding.EncodeToString(sum[:18])
	}
}

// multiSecretHasher returns a hasher using the first secret as active while still retaining
// the full slice for potential future verification use (e.g., comparing historical values).
// If secrets empty or first is empty returns nil.
func multiSecretHasher(secrets [][]byte) (hashFunc func(string) string, activeSecret []byte, all [][]byte) {
	if len(secrets) == 0 || len(secrets[0]) == 0 { return nil, nil, nil }
	active := secrets[0]
	var h hash.Hash
	return func(v string) string {
		h = hmac.New(sha256.New, active)
		_, _ = h.Write([]byte(v))
		sum := h.Sum(nil)
		return base64.RawURLEncoding.EncodeToString(sum[:18])
	}, active, secrets
}

// tryHashAll returns the expected hash if any of the provided secrets match the already-hashed value.
// This is a utility for future verification routines that may need to reconcile historical hashes
// when secrets rotate. For now it's unused but included to anchor rotation design.
func tryHashAll(raw string, secrets [][]byte) []string {
	out := make([]string, 0, len(secrets))
	for _, s := range secrets {
		if len(s) == 0 { continue }
		h := hmac.New(sha256.New, s)
		_, _ = h.Write([]byte(raw))
		sum := h.Sum(nil)
		out = append(out, base64.RawURLEncoding.EncodeToString(sum[:18]))
	}
	return out
}
