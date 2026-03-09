package blob

import (
	"context"
	"io"
)

type BlobMeta struct {
	MimeType string
	FileName string
	Size     int64
}

type Client interface {
	Open(ctx context.Context, path string) (io.ReadCloser, BlobMeta, error)
	Create(ctx context.Context, path string, mimeType string) (io.WriteCloser, error)
}
