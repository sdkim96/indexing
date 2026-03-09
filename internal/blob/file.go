package blob

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/sdkim96/indexing/internal"
)

var _ Client = (*FileBlobClient)(nil)

type FileBlobClient struct{}

func (c *FileBlobClient) Open(ctx context.Context, path string) (io.ReadCloser, BlobMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, BlobMeta{}, err
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, BlobMeta{}, err
	}

	return f, BlobMeta{
		MimeType: internal.GuessMimeType(path),
		FileName: filepath.Base(path),
		Size:     info.Size(),
	}, nil
}

func (c *FileBlobClient) Create(ctx context.Context, path string, mimeType string) (io.WriteCloser, error) {
	return os.Create(path)
}
