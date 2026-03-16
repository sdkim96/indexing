package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

// Cache provides methods to read and write cache data.
type Cache interface {
	GetOrSet(ctx context.Context, key string, fn func() ([]byte, error)) ([]byte, error)
}

// Hasher provides a method to generate a fingerprint for caching purposes.
// The FingerPrint method should return a unique string that represents the content of the item being cached.
// This allows the caching mechanism to identify when the same content is being processed, enabling cache hits and improving performance.
type Hasher interface {
	FingerPrint(prefix string) string
}

// Sha256 provides the canonical way to generate FingerPrints for cache items.
// It takes one or more byte slices as input and
// returns a SHA-256 hash of the combined input as a hexadecimal string.
func Sha256(chunks ...[]byte) string {
	h := sha256.New()

	for _, c := range chunks {
		h.Write(c)
	}

	return hex.EncodeToString(h.Sum(nil))
}
