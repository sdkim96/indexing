package input

import (
	"context"
	"errors"

	"github.com/sdkim96/indexing/internal/uri"
)

// Provider is responsible for providing an Input based on a URI.
// The process that provider do
// is by reading the data from the source identified by the URI
// and returning an Input that can be used to read the data.
type Provider interface {
	Provide(ctx context.Context, key uri.URI) (Input, error)
}

var ErrUnsupportedSourceScheme error = errors.New("unsupported source key. check the scheme.")
var ErrInvalidSourceKey error = errors.New("invalid source key. must follow {scheme}://{path}")
