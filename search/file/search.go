package file

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sdkim96/indexing/internal/mime"
	"github.com/sdkim96/indexing/internal/storage"
	"github.com/sdkim96/indexing/internal/uri"
	"github.com/sdkim96/indexing/search"
)

type FileSearchWriter struct {
	client *storage.FileSystemClient
}

func New(client *storage.FileSystemClient) *FileSearchWriter {
	return &FileSearchWriter{client: client}
}

var _ search.SearchWriter = (*FileSearchWriter)(nil)

func (w *FileSearchWriter) Write(ctx context.Context, URI uri.URI, docs []search.SearchDoc) error {
	if err := URI.Validate(); err != nil {
		return err
	}
	if scheme := URI.Scheme(); scheme != "file" {
		return fmt.Errorf("The scheme must be file://. Check your URI: %s", string(URI))
	}
	fp, _, err := w.client.Create(ctx, URI.Path(), mime.MimeJSON)
	if err != nil {
		return err
	}
	defer fp.Close()
	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return err
	}
	if _, err := fp.Write(data); err != nil {
		return err
	}
	return nil
}
