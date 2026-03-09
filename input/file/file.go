package file

import (
	"context"
	"os"

	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/internal"
)

type FileProvider struct {
}

var _ input.Provider = (*FileProvider)(nil)

type FileMeta struct {
	name string
	size int64
}

// FileProvider assumes sourceID is a file path and provides an Input that reads from the file.
func (p *FileProvider) Provide(ctx context.Context, sourceID string) (input.Input, error) {
	file, err := os.Open(sourceID)
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

	return NewFileInput(file, internal.GuessMimeType(sourceID), meta), nil
}
