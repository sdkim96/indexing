package enrich

import (
	"context"

	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
)

// Enricher processes Parts and produces SearchDocs for indexing.
type Enricher interface {
	Enrich(ctx context.Context, parts []part.Part, send chan<- cache.Cache) ([]search.SearchDoc, error)
}
