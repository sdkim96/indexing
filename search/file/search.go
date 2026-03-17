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

	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/search"
	"github.com/sdkim96/indexing/storage"
	"github.com/sdkim96/indexing/uri"
)

type FileSearchWriter struct {
	client *storage.FileSystemClient
}

func New(client *storage.FileSystemClient) FileSearchWriter {
	return FileSearchWriter{client: client}
}

var _ search.SearchWriter = (*FileSearchWriter)(nil)

func (w FileSearchWriter) Write(ctx context.Context, URI uri.URI, docs []search.SearchDoc) error {
	if err := URI.Validate(); err != nil {
		return err
	}
	if scheme := URI.Scheme(); scheme != "file" {
		return fmt.Errorf("The scheme must be file://. Check your URI: %s", string(URI))
	}
	fp, _, err := w.client.Create(ctx, URI.Path(), mime.MimeJSON)
	if err != nil {
		return err
	}
	defer fp.Close()
	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return err
	}
	if _, err := fp.Write(data); err != nil {
		return err
	}
	return nil
}
