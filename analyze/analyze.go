package analyze

import (
	"context"

	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/part"
)

// Analyzer analyzes raw data and produces Parts.
type Analyzer interface {
	Analyze(ctx context.Context, input input.Input) ([]part.Part, error)
}
