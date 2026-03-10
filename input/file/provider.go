package file

import (
	"context"
	"os"

	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/internal/mime"
	"github.com/sdkim96/indexing/internal/uri"
)

type FileProvider struct {
}

var _ input.Provider = (*FileProvider)(nil)

type FileMeta struct {
	name string
	size int64
}

// FileProvider assumes that the form of key must be "file://{path}" and provides FileInput.
func (p *FileProvider) Provide(ctx context.Context, key uri.URI) (input.Input, error) {
	if err := key.Validate(); err != nil {
		return nil, input.ErrInvalidSourceKey
	}
	if key.Scheme() != "file" {
		return nil, input.ErrUnsupportedSourceScheme
	}

	file, err := os.Open(key.Path())
	if err != nil {
		return nil, err
	}

	var meta map[string]any
	if info, err := file.Stat(); err == nil {
		meta = map[string]any{
			"name": info.Name(),
			"size": info.Size(),
		}
	} else {
		meta = map[string]any{}
	}

	return NewFileInput(file, mime.GuessMimeType(key.Path()), meta), nil
}
