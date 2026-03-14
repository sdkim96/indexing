package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

type CacheWriter interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte) error
}

type Cache struct {
	Key   string
	Value []byte
}

type Cacheable interface {
	FingerPrint() string
}

func Sha256(chunks ...[]byte) string {
	h := sha256.New()

	for _, c := range chunks {
		h.Write(c)
	}

	return hex.EncodeToString(h.Sum(nil))
}
