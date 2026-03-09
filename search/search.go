package search

import "context"

// SearchDoc represents a document to be indexed in a search engine.
type SearchDoc interface {
	SourceID() string
	PartIDs() []string
	Fields() map[string]any
}

// SearchWriter writes SearchDocs to a search engine.
type SearchWriter interface {
	Write(ctx context.Context, docs []SearchDoc) error
}
