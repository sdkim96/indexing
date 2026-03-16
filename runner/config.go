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
// Essentials:
//   - Provider
//   - Analyzer
//   - PartWriter
//
// Optionals:
//   - Enricher
//   - SearchWriter
//   - Cache
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
