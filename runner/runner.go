package runner

import (
	"context"
	"iter"
	"time"

	"github.com/sdkim96/indexing/analyze"
	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/indexing/enrich"
	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
)

// Event represents the completion of a pipeline stage.
type Event struct {
	Stage    string // "provide" | "analyze" | "part" | "enrich" | "search"
	ICtx     *IndexingContext
	Duration time.Duration
}

// Runner executes the indexing pipeline stages in order.
type Runner struct {
	provider     input.Provider
	analyzer     analyze.Analyzer
	partWriter   part.PartWriter
	enricher     enrich.Enricher
	searchWriter search.SearchWriter
	cache        cache.Cache
}

// Run executes the pipeline and yields Events via an iterator.
func (r *Runner) Run(ctx context.Context, ictx *IndexingContext) iter.Seq2[Event, error] {
	return func(yield func(Event, error) bool) {

		start := time.Now()
		input, err := r.provider.Provide(ctx, ictx.InputKey)
		if !yield(Event{"provide", ictx, time.Since(start)}, err) {
			return
		}
		defer input.Close()

		start = time.Now()
		parts, err := r.analyzer.Analyze(ctx, input, r.cache)
		if !yield(Event{"analyze", ictx, time.Since(start)}, err) {
			return
		}
		ictx.Parts = parts

		start = time.Now()
		err = r.partWriter.Write(ctx, ictx.PartWriteKey, ictx.Parts)
		if !yield(Event{"part", ictx, time.Since(start)}, err) {
			return
		}

		start = time.Now()
		docs, err := r.enricher.Enrich(ctx, ictx.Parts, r.cache)
		if !yield(Event{"enrich", ictx, time.Since(start)}, err) {
			return
		}
		ictx.SearchDocs = docs

		start = time.Now()
		err = r.searchWriter.Write(ctx, ictx.SearchWriteKey, ictx.SearchDocs)
		if !yield(Event{"search", ictx, time.Since(start)}, err) {
			return
		}
	}
}
