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

package storage

import (
	"context"
	"io"

	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/uri"
)

// Meta contains the metadata of a file in the specified storage backend,
// such as MIME type, file name, and size. This information is returned by the Client when opening a file for reading.
type Meta struct {
	MimeType mime.Type
	FileName string
	Size     int64
}

// Client read/write with a storage backend (e.g. local filesystem, S3, GCS).
type Client interface {

	// Open the file at the given path for reading. The caller should close the returned ReadCloser.
	Open(ctx context.Context, path string) (io.ReadCloser, Meta, error)

	// Create a new file at the given path for writing. The caller should close the returned WriteCloser.
	Create(ctx context.Context, path string, mimeType mime.Type) (io.WriteCloser, uri.URI, error)
}
