package file

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/storage"
	"github.com/sdkim96/indexing/uri"
)

type FilePartWriter struct {
	client *storage.FileSystemClient
}

func New(client *storage.FileSystemClient) FilePartWriter {
	return FilePartWriter{client: client}
}

var _ part.PartWriter = (*FilePartWriter)(nil)

func (w FilePartWriter) Write(ctx context.Context, URI uri.URI, parts []part.Part) error {
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

	data, err := json.Marshal(parts)
	if err != nil {
		return err
	}
	if _, err := fp.Write(data); err != nil {
		return err
	}

	return nil
}
