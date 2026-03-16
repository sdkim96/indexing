package runner

import (
	"time"

	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
	"github.com/sdkim96/indexing/uri"
)

// IndexingContext carries data across pipeline stages.
type IndexingContext struct {
	// Input, PartWrite, SearchWrite Keys are used to identify data location in each stage.
	InputKey, PartWriteKey, SearchWriteKey uri.URI
	StartedAt                              time.Time

	// Populated during pipeline execution.
	Parts      []part.Part
	SearchDocs []search.SearchDoc
}

func NewICtx(inputKey, partWriteKey, searchWriteKey string) *IndexingContext {
	return &IndexingContext{
		InputKey:       uri.URI(inputKey),
		PartWriteKey:   uri.URI(partWriteKey),
		SearchWriteKey: uri.URI(searchWriteKey),
		StartedAt:      time.Now(),
	}
}
