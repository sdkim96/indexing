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
	// Stage is the name of the completed stage.
	// Possible values: "provide" | "analyze" | "part" | "enrich" | "search"
	Stage string

	// Duration is the time taken to complete the stage.
	Duration time.Duration
}

// Runner executes the indexing pipeline stages in order.
//
// Each component (provider, analyzer, partWriter, enricher, searchWriter)
// is configured at construction time and already knows its source or destination.
// The Runner itself does not manage URIs or resource locations —
// those concerns belong to each component.
//
// Use New to construct a Runner with the required components.
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
// The sourceID is an identifier that the Runner passes to each stage,
// which can be used for logging or as a key for storage.
//
// Stages are executed in order:
//  1. provide  — reads the source into an Input
//  2. analyze  — processes the Input into Parts
//  3. part     — persists the Parts to storage
//  4. enrich   — enriches Parts into SearchDocs
//  5. search   — writes SearchDocs to the search index
//
// If a stage returns an error, it is yielded alongside the Event.
// The caller decides whether to abort by returning false from the range body.
//
// Example:
//
//	for event, err := range r.Run(ctx, sourceID) {
//	    if err != nil {
//	        log.Fatalf("stage %s failed: %v", event.Stage, err)
//	    }
//	    log.Printf("stage %s done in %s", event.Stage, event.Duration)
//	}
func (r *Runner) Run(ctx context.Context, sourceID string) iter.Seq2[Event, error] {
	return func(yield func(Event, error) bool) {

		start := time.Now()
		inp, err := r.provider.Provide(ctx, sourceID)
		if !yield(Event{"provide", time.Since(start)}, err) {
			return
		}
		defer inp.Close()

		start = time.Now()
		parts, err := r.analyzer.Analyze(ctx, sourceID, inp, r.cache)
		if !yield(Event{"analyze", time.Since(start)}, err) {
			return
		}

		start = time.Now()
		err = r.partWriter.Write(ctx, sourceID, parts)
		if !yield(Event{"part", time.Since(start)}, err) {
			return
		}

		start = time.Now()
		docs, err := r.enricher.Enrich(ctx, sourceID, parts, r.cache)
		if !yield(Event{"enrich", time.Since(start)}, err) {
			return
		}

		start = time.Now()
		err = r.searchWriter.Write(ctx, sourceID, docs)
		if !yield(Event{"search", time.Since(start)}, err) {
			return
		}
	}
}
