package input

import (
	"context"
	"errors"

	"github.com/sdkim96/indexing/uri"
)

// Provider reads the data from the source identified by the URI
// and returns an Input that can be used to read the data.
type Provider interface {
	Provide(ctx context.Context, URI uri.URI) (Input, error)
}

var ErrUnsupportedSourceScheme error = errors.New("unsupported source key. check the scheme.")
var ErrInvalidSourceKey error = errors.New("invalid source key. must follow {scheme}://{path}")
