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
func newHasher(secret []byte) func(string) string {
    if len(secret) == 0 { return nil }
    var h hash.Hash
    return func(v string) string {
        h = hmac.New(sha256.New, secret)
        _, _ = h.Write([]byte(v))
        sum := h.Sum(nil)
        return base64.RawURLEncoding.EncodeToString(sum[:18])
    }
}
