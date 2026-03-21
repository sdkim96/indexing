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

package analyze

import (
	"context"

	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/part"
)

// Analyzer processes an Input and produces a list of Parts.
// Implementations are responsible for parsing the raw input data
// and splitting it into smaller, addressable Parts for enrichment and indexing.
//
// The cache parameter is optional — pass nil or a cache.NoopCache if caching
// is not needed. When provided, cache should only be used for expensive
// operations such as external API calls, not for memoizing Analyze itself.
type Analyzer interface {

	// Analyze reads from input, processes the data, and returns the resulting Parts.
	// The sourceID is an identifier that the Analyzer can use for logging or as a key for storage.
	Analyze(ctx context.Context, sourceID string, input input.Input, cache cache.Cache) ([]part.Part, error)
}
