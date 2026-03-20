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

package search

import (
	"context"
)

// SearchDoc represents a document to be indexed in the search engine. It contains the fields that will be stored and searchable.
type SearchDoc interface {

	// Fields returns a map of field names to their corresponding values for this SearchDoc.
	// The fields should include all necessary information for indexing and searching, such as title, content, metadata, etc.
	Fields() map[string]any
}

// SearchWriter writes SearchDocs to a search engine.
type SearchWriter interface {

	// Write takes a list of SearchDocs and writes them to the search engine.
	Write(ctx context.Context, docs []SearchDoc) error
}
