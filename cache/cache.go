package cache

import (
	"crypto/sha256"
	"encoding/hex"
)

type Cache interface {
	Set(key string, value any) error
	Get(key string) (any, error)
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
