package input

import "io"

type Input interface {
	io.ReadCloser
	MimeType() string
	Meta() map[string]any
}
