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

package file

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/sdkim96/indexing/cache"
)

// FileCache is a file-based cache that stores values as files in a directory.
// Each key maps to a file within the directory.
type FileCache struct {
	dir         string
	hitCallback func(key string)
}

// New creates a new FileCache that stores cache files in the given directory.
func New(dir string, opts ...FileCacheOptions) *FileCache {
	c := &FileCache{dir: dir}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type FileCacheOptions func(*FileCache)

// WithCacheHitCallback sets a callback that is called when a cache hit occurs.
func WithCacheHitCallback(callback func(key string)) FileCacheOptions {
	return func(c *FileCache) {
		c.hitCallback = callback
	}
}

var _ cache.Cache = (*FileCache)(nil)

// GetOrSet retrieves the cached value for the given key.
// If the key does not exist, it calls fn to generate the value,
// stores it in the cache, and returns it.
// If storing fails, the generated value is still returned without error.
func (c *FileCache) GetOrSet(ctx context.Context, key string, fn func() ([]byte, error)) ([]byte, error) {
	path := filepath.Join(c.dir, key)

	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err == nil {
		defer f.Close()
		if c.hitCallback != nil {
			c.hitCallback(key)
		}
		return io.ReadAll(f)
	}

	data, err := fn()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return data, nil
	}
	w, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return data, nil
	}
	defer w.Close()
	w.Write(data)

	return data, nil
}
