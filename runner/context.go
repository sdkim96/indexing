package runner

import (
	"time"

	"github.com/sdkim96/indexing/internal/uri"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
)

// IndexingContext carries data across pipeline stages.
type IndexingContext struct {
	// 실행 메타
	InputKey       uri.URI
	PartWriteKey   uri.URI
	SearchWriteKey uri.URI
	StartedAt      time.Time
	CacheChanSize  int

	// 단계별로 채워짐
	Parts      []part.Part        // Analyzer 이후 존재
	SearchDocs []search.SearchDoc // Enricher 이후 존재
}
