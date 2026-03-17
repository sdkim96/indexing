// Copyright 2026 Sungdong Kim
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// Event represents the completion of a single pipeline stage.
// It is yielded by Runner.Run after each stage completes, regardless of success or failure.
type Event struct {
	Stage    string // "provide" | "analyze" | "part" | "enrich" | "search"
	ICtx     *IndexingContext
	Duration time.Duration
}

// Runner executes the indexing pipeline stages in order.
// It is assembled via New and configured with a Config.
// Each stage is executed sequentially; the caller controls error handling
// by deciding whether to continue or abort via the iter.Seq2 iterator.
type Runner struct {
	provider     input.Provider
	analyzer     analyze.Analyzer
	partWriter   part.PartWriter
	enricher     enrich.Enricher
	searchWriter search.SearchWriter
	cache        cache.Cache
}

// Run executes the indexing pipeline and yields an Event after each stage completes.
// It returns an iter.Seq2[Event, error] that the caller ranges over.
//
// Stages are executed in order:
//  1. provide  — resolves the input URI to an Input
//  2. analyze  — processes the Input into Parts
//  3. part     — persists the Parts to storage
//  4. enrich   — enriches Parts into SearchDocs
//  5. search   — writes SearchDocs to the search engine
//
// If a stage returns an error, it is yielded alongside the Event.
// The caller decides whether to abort by returning false from the range body.
//
// Example:
//
//	for event, err := range r.Run(ctx, ictx) {
//	    if err != nil {
//	        log.Fatalf("failed at %s: %v", event.Stage, err)
//	    }
//	    log.Printf("stage %s done in %s", event.Stage, event.Duration)
//	}
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
