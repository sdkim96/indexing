package part

import (
	"context"
)

type Part interface {
	MimeType() string
	Text() string
	Raw() any
}

// PartWriter persists Parts.
type PartWriter interface {
	Write(ctx context.Context, sourceID string, parts []Part) error
}
