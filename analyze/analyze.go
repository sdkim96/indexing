package analyze

import (
	"context"

	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/part"
)

// Analyzer analyzes raw data and produces Parts.
type Analyzer interface {
	Analyze(ctx context.Context, input input.Input, cache cache.Cache) ([]part.Part, error)
}
