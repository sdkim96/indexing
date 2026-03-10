package file

import (
	"io"

	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/internal/mime"
)

var _ input.Input = (*FileInput)(nil)

type FileInput struct {
	readCloser io.ReadCloser
	mimeType   mime.Type
	meta       map[string]any
}

func (f *FileInput) Read(p []byte) (n int, err error) {
	return f.readCloser.Read(p)
}

func (f *FileInput) Close() error {
	return f.readCloser.Close()
}

func (f *FileInput) MimeType() mime.Type {
	return f.mimeType
}

func (f *FileInput) Meta() map[string]any {
	return f.meta
}

func NewFileInput(readCloser io.ReadCloser, mimeType mime.Type, meta map[string]any) input.Input {
	return &FileInput{
		readCloser: readCloser,
		mimeType:   mimeType,
		meta:       meta,
	}
}
