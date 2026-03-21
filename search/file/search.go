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
	"fmt"
	"os"

	"github.com/sdkim96/indexing/search"
	"github.com/sdkim96/indexing/urio"
)

type FileSearchWriter struct {
	uri urio.URI
}

func NewFileSearchWriter(uri urio.URI) (FileSearchWriter, error) {
	if uri.Scheme() != "file" {
		return FileSearchWriter{}, fmt.Errorf("unsupported URI scheme: %s. expected 'file'", uri.Scheme())
	}
	return FileSearchWriter{uri: uri}, nil
}

var _ search.SearchWriter = (*FileSearchWriter)(nil)

func (w FileSearchWriter) Write(ctx context.Context, sourceID string, docs []search.SearchDoc) error {
	f, err := os.OpenFile(w.uri.Path(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to write search docs to file: %w", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(docs)
	if err != nil {
		return fmt.Errorf("failed to encode search docs to JSON: %w", err)
	}
	return nil
}
