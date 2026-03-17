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

	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/storage"
	"github.com/sdkim96/indexing/uri"
)

type FileProvider struct {
	client *storage.FileSystemClient
}

func New(client *storage.FileSystemClient) FileProvider {
	return FileProvider{client: client}
}

var _ input.Provider = (*FileProvider)(nil)

type FileMeta struct {
	name string
	size int64
}

// FileProvider assumes that the form of key must be "file://{path}" and provides FileInput.
func (p FileProvider) Provide(ctx context.Context, URI uri.URI) (input.Input, error) {
	if err := URI.Validate(); err != nil {
		return nil, input.ErrInvalidSourceKey
	}
	if URI.Scheme() != "file" {
		return nil, input.ErrUnsupportedSourceScheme
	}

	f, meta, err := p.client.Open(ctx, URI.Path())
	if err != nil {
		return nil, err
	}
	var m map[string]any = map[string]any{
		"name": meta.FileName,
		"size": meta.Size,
	}
	return NewFileInput(f, mime.GuessMimeType(URI.Path()), m), nil
}
