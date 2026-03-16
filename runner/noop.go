package runner

import (
	"context"

	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
	"github.com/sdkim96/indexing/uri"
)

// NoopEnricher does nothing. Used when no Enricher is provided.
type NoopEnricher struct{}

func (n *NoopEnricher) Enrich(_ context.Context, _ []part.Part, _ cache.Cache) ([]search.SearchDoc, error) {
	return nil, nil
}

// NoopSearchWriter does nothing. Used when no SearchWriter is provided.
type NoopSearchWriter struct{}

func (n *NoopSearchWriter) Write(_ context.Context, _ uri.URI, _ []search.SearchDoc) error {
	return nil
}

type NoopCache struct{}

func (n *NoopCache) GetOrSet(ctx context.Context, key string, fn func() ([]byte, error)) ([]byte, error) {
	return fn()
}
