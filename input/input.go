package input

import (
	"io"

	"github.com/sdkim96/indexing/mime"
)

// Input represents the data to be indexed.
// It provides methods to read the data.
type Input interface {
	io.ReadCloser
	MimeType() mime.Type
	Meta() map[string]any
}
