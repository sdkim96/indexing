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
	"fmt"
	"os"

	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/urio"
)

type FileProvider struct {
	uri urio.URI
}

func NewFileProvider(uri urio.URI) (FileProvider, error) {
	if uri.Scheme() != "file" {
		return FileProvider{}, fmt.Errorf("unsupported URI scheme: %s. expected 'file'", uri.Scheme())
	}
	return FileProvider{uri: uri}, nil
}

var _ input.Provider = (*FileProvider)(nil)

// Provide opens the file at the URI given at construction time and returns a FileInput.
func (p FileProvider) Provide(ctx context.Context) (input.Input, error) {
	f, err := os.Open(p.uri.Path())
	if err != nil {
		return nil, fmt.Errorf("failed to read search docs from file: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	var m map[string]any = map[string]any{
		"name": info.Name(),
		"size": info.Size(),
	}
	return NewFileInput(f, mime.GuessMimeType(p.uri.Path()), m), nil

}
