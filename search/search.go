package search

import (
	"context"

	"github.com/sdkim96/indexing/uri"
)

// SearchDoc represents a document to be indexed in a search engine.
type SearchDoc interface {
	Fields() map[string]any
}

// SearchWriter writes SearchDocs to a search engine.
type SearchWriter interface {
	Write(ctx context.Context, URI uri.URI, docs []SearchDoc) error
}
