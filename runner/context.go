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

	// Parts are populated during pipeline execution.
	Parts []part.Part

	// SearchDocs are populated during pipeline execution.
	SearchDocs []search.SearchDoc
}

func NewICtx(opts ...IContextOpt) *IndexingContext {
	ctx := &IndexingContext{
		StartedAt: time.Now(),
	}
	for _, opt := range opts {
		opt(ctx)
	}
	return ctx
}

type IContextOpt func(*IndexingContext)

func WithInputKey(key string) IContextOpt {
	return func(ctx *IndexingContext) {
		ctx.InputKey = uri.URI(key)
	}
}

func WithPartWriteKey(key string) IContextOpt {
	return func(ctx *IndexingContext) {
		ctx.PartWriteKey = uri.URI(key)
	}
}

func WithSearchWriteKey(key string) IContextOpt {
	return func(ctx *IndexingContext) {
		ctx.SearchWriteKey = uri.URI(key)
	}
}
