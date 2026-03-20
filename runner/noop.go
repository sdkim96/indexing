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

	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
)

// NoopEnricher does nothing. Used when no Enricher is provided.
type NoopEnricher struct{}

func (n *NoopEnricher) Enrich(_ context.Context, _ []part.Part, _ cache.Cache) ([]search.SearchDoc, error) {
	return nil, nil
}

// NoopSearchWriter does nothing. Used when no SearchWriter is provided.
type NoopSearchWriter struct{}

func (n *NoopSearchWriter) Write(_ context.Context, _ []search.SearchDoc) error {
	return nil
}

type NoopCache struct{}

func (n *NoopCache) GetOrSet(ctx context.Context, key string, fn func() ([]byte, error)) ([]byte, error) {
	return fn()
}
