package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/sdkim96/indexing/internal/mime"
	"github.com/sdkim96/indexing/internal/uri"
)

var _ Client = (*FileSystemClient)(nil)

type FileSystemClient struct {
	root string
}

func NewFileSystemClient(root string) (*FileSystemClient, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &FileSystemClient{root: abs}, nil
}

func (c *FileSystemClient) Open(ctx context.Context, path string) (io.ReadCloser, Meta, error) {
	fullPath := filepath.Join(c.root, path)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, Meta{}, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, Meta{}, err
	}
	return f, Meta{
		MimeType: mime.GuessMimeType(fullPath),
		FileName: filepath.Base(fullPath),
		Size:     info.Size(),
	}, nil
}

func (c *FileSystemClient) Create(ctx context.Context, path string, mimeType mime.Type) (io.WriteCloser, uri.URI, error) {
	fullPath := filepath.Join(c.root, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, uri.URI(""), err
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, uri.URI(""), err
	}
	return f, uri.URI("file://" + fullPath), nil
}
