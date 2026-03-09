package runner

import (
	"github.com/sdkim96/indexing/analyze"
	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/indexing/enrich"
	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
)

// Config holds all dependencies for building a Pipeline.
type Config struct {
	Provider     input.Provider      // 필수
	Analyzer     analyze.Analyzer    // 필수
	PartWriter   part.PartWriter     // 필수
	Enricher     enrich.Enricher     // 선택 (nil → noop)
	SearchWriter search.SearchWriter // 선택 (nil → noop)
	CacheWriter  cache.Cache         // 선택 (nil → noop)
}
