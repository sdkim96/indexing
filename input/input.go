package input

import (
	"io"

	"github.com/sdkim96/indexing/internal/mime"
)

type Input interface {
	io.ReadCloser
	MimeType() mime.Type
	Meta() map[string]any
}
