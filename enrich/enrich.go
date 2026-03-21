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

package enrich

import (
	"context"

	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
)

// Enricher processes Parts and produces SearchDocs for indexing.
// Implementations are responsible for taking the analyzed Parts and enriching them with additional information,
// such as generating embeddings, extracting metadata, or performing any necessary transformations before indexing.
//
// The cache parameter is optional — pass nil or a cache.NoopCache if caching
// is not needed. When provided, cache should only be used for expensive
// operations such as external API calls, not for memoizing Enrich itself.
type Enricher interface {

	// Enrich takes a list of document parts, processes them, and returns a list of SearchDoc objects
	// that contain the enriched information for each topic identified in the document.
	// The sourceID is a unique identifier of the source (e.g. a file, a provided data stream).
	Enrich(ctx context.Context, sourceID string, parts []part.Part, cache cache.Cache) ([]search.SearchDoc, error)
}
