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

// Pipeline is the top-level interface for running an indexing pipeline.
type Pipeline interface {
	Run(ctx context.Context, sourceID string) error
}

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

		// 0단계: sourceID → Input 변환
		start := time.Now()
		input, err := r.provider.Provide(ctx, ictx.SourceID)
		if !yield(Event{"provide", ictx, time.Since(start)}, err) {
			return
		}
		defer input.Close()

		// 1단계: 읽고 쪼갠다
		start = time.Now()
		parts, err := r.analyzer.Analyze(ctx, input)
		if !yield(Event{"analyze", ictx, time.Since(start)}, err) {
			return
		}
		ictx.Parts = parts

		// 2단계: 저장한다
		start = time.Now()
		err = r.partWriter.Write(ctx, ictx.SourceID, ictx.Parts)
		if !yield(Event{"part", ictx, time.Since(start)}, err) {
			return
		}

		// 3단계: 가공한다
		start = time.Now()
		docs, err := r.enricher.Enrich(ctx, ictx.Parts)
		if !yield(Event{"enrich", ictx, time.Since(start)}, err) {
			return
		}
		ictx.SearchDocs = docs

		// 4단계: 적재한다
		start = time.Now()
		err = r.searchWriter.Write(ctx, ictx.SearchDocs)
		if !yield(Event{"search", ictx, time.Since(start)}, err) {
			return
		}
	}
}
