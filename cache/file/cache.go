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
