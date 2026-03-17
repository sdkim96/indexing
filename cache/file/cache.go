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

	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/storage"
)

type FileCache struct {
	client      *storage.FileSystemClient
	hitCallback func(key string)
}

func New(client *storage.FileSystemClient, opts ...FileCacheOptions) FileCache {
	c := FileCache{client: client}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

type FileCacheOptions func(*FileCache)

func WithCacheHitCallback(callback func(key string)) FileCacheOptions {
	return func(c *FileCache) {
		c.hitCallback = callback
	}
}

var _ cache.Cache = (*FileCache)(nil)

func (c FileCache) GetOrSet(ctx context.Context, key string, fn func() ([]byte, error)) ([]byte, error) {

	f, _, err := c.client.Open(ctx, key)
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

	w, _, err := c.client.Create(ctx, key, mime.MimeJSON)
	if err != nil {
		return data, nil
	}
	defer w.Close()
	w.Write(data)

	return data, nil
}
