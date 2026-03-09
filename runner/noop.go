package runner

import (
	"context"

	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
)

// NoopEnricher does nothing. Used when no Enricher is provided.
type NoopEnricher struct{}

func (n *NoopEnricher) Enrich(_ context.Context, _ []part.Part) ([]search.SearchDoc, error) {
	return nil, nil
}

// NoopSearchWriter does nothing. Used when no SearchWriter is provided.
type NoopSearchWriter struct{}

func (n *NoopSearchWriter) Write(_ context.Context, _ []search.SearchDoc) error {
	return nil
}
