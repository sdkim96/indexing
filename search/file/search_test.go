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

package file

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sdkim96/indexing/enrich/openai"
	"github.com/sdkim96/indexing/search"
	"github.com/sdkim96/indexing/urio"
)

const testDocURI = "file:///Users/sungdongkim/works/indexing/search/file/testdata/enrich_result_cowboys.json"
const testURI = "file:///Users/sungdongkim/works/indexing/search/file/testdata/search_docs.json"

func newDocs() []search.SearchDoc {
	path := urio.URI(testDocURI).Path()
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var raw []openai.OpenAISearchDoc
	if err := json.Unmarshal(data, &raw); err != nil {
		panic(err)
	}

	docs := make([]search.SearchDoc, len(raw))
	for i := range raw {
		docs[i] = &raw[i]
	}
	return docs
}

func TestSearchWriter_Write(t *testing.T) {
	uri := urio.URI(testURI)
	writer, err := NewFileSearchWriter(uri)
	if err != nil {
		t.Fatalf("failed to create FileSearchWriter: %v", err)
	}
	writer.Write(context.Background(), newDocs())
}
