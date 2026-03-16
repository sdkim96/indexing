package storage

import (
	"context"
	"io"

	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/uri"
)

type Meta struct {
	MimeType mime.Type
	FileName string
	Size     int64
}

// Client read/write with a storage backend (e.g. local filesystem, S3, GCS).
type Client interface {

	// Open the file at the given path for reading. The caller should close the returned ReadCloser.
	Open(ctx context.Context, path string) (io.ReadCloser, Meta, error)

	// Create a new file at the given path for writing. The caller should close the returned WriteCloser.
	Create(ctx context.Context, path string, mimeType mime.Type) (io.WriteCloser, uri.URI, error)
}
