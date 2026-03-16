package part

import (
	"context"

	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/uri"
)

type Part interface {
	MimeType() mime.Type
	Text() string
	Raw() []byte
}

// PartWriter persists Parts.
type PartWriter interface {
	Write(ctx context.Context, URI uri.URI, parts []Part) error
}
