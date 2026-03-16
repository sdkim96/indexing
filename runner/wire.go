package runner

import (
	"errors"

	"github.com/sdkim96/indexing/analyze"
	"github.com/sdkim96/indexing/enrich"
	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
)

// New validates the Config and assembles a Runner.
func New(opts ...ConfigOpt) (*Runner, error) {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.Provider == nil {
		return nil, errors.New("source provider is required")
	}
	if cfg.Analyzer == nil {
		return nil, errors.New("analyzer is required")
	}
	if cfg.PartWriter == nil {
		return nil, errors.New("part writer is required")
	}

	enricher := cfg.Enricher
	if enricher == nil {
		enricher = &NoopEnricher{}
	}

	sw := cfg.SearchWriter
	if sw == nil {
		sw = &NoopSearchWriter{}
	}

	cache := cfg.Cache
	if cache == nil {
		cache = &NoopCache{}
	}

	return &Runner{
		provider:     cfg.Provider,
		analyzer:     cfg.Analyzer,
		partWriter:   cfg.PartWriter,
		enricher:     enricher,
		searchWriter: sw,
		cache:        cache,
	}, nil
}

// ensure interfaces are satisfied at compile time
var (
	_ input.Provider      = (input.Provider)(nil)
	_ analyze.Analyzer    = (analyze.Analyzer)(nil)
	_ part.PartWriter     = (part.PartWriter)(nil)
	_ enrich.Enricher     = (enrich.Enricher)(nil)
	_ search.SearchWriter = (search.SearchWriter)(nil)
)
