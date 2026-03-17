// Copyright 2026 Sungdong Kim
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

// Cache provides methods to read and write cache data.
type Cache interface {

	// GetOrSet retrieves the cached value for the given key.
	//
	// If the key does not exist, it calls the provided function to generate the value,
	// stores it in the cache, and then returns it.
	GetOrSet(ctx context.Context, key string, fn func() ([]byte, error)) ([]byte, error)
}

// Hasher provides a method to generate a fingerprint for caching purposes.
type Hasher interface {

	// FingerPrint generates a unique string representation of the item for caching purposes.
	// You can optionally provide a prefix to namespace the fingerprint, which can be useful to avoid collisions between different types of cache items.
	FingerPrint(prefix string) string
}

// Sha256 provides the canonical way to generate FingerPrints for cache items.
//
// It takes one or more byte slices as input and
// returns a SHA-256 hash of the combined input as a hexadecimal string.
func Sha256(chunks ...[]byte) string {
	h := sha256.New()

	for _, c := range chunks {
		h.Write(c)
	}

	return hex.EncodeToString(h.Sum(nil))
}
