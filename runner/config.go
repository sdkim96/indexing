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
	"github.com/sdkim96/indexing/analyze"
	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/indexing/enrich"
	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
)

// Config holds all implementations for running the indexing pipeline.
//
// Required fields:
//   - Provider: resolves a URI to an Input
//   - Analyzer: processes an Input into Parts
//   - PartWriter: persists the analyzed Parts
//
// Optional fields (noop implementations used if nil):
//   - Enricher: enriches Parts into SearchDocs
//   - SearchWriter: writes SearchDocs to a search engine
//   - Cache: caches expensive operations such as API calls
type Config struct {
	Provider     input.Provider
	Analyzer     analyze.Analyzer
	PartWriter   part.PartWriter
	Enricher     enrich.Enricher
	SearchWriter search.SearchWriter
	Cache        cache.Cache
}

type ConfigOpt func(*Config)

func WithProvider(p input.Provider) ConfigOpt {
	return func(cfg *Config) {
		cfg.Provider = p
	}
}

func WithAnalyzer(a analyze.Analyzer) ConfigOpt {
	return func(cfg *Config) {
		cfg.Analyzer = a
	}
}

func WithPartWriter(w part.PartWriter) ConfigOpt {
	return func(cfg *Config) {
		cfg.PartWriter = w
	}
}

func WithEnricher(e enrich.Enricher) ConfigOpt {
	return func(cfg *Config) {
		cfg.Enricher = e
	}
}

func WithSearchWriter(w search.SearchWriter) ConfigOpt {
	return func(cfg *Config) {
		cfg.SearchWriter = w
	}
}

func WithCache(c cache.Cache) ConfigOpt {
	return func(cfg *Config) {
		cfg.Cache = c
	}
}
