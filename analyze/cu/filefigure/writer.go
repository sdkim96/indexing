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

package cu

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sdkim96/indexing/urio"
)

type FileFigWriter struct {
	f   *os.File
	uri urio.URI
}

func NewFileFigWriter(uri urio.URI) (*FileFigWriter, error) {
	path := uri.Path()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directories for %s: %w", path, err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for writing %s: %w", path, err)
	}

	return &FileFigWriter{
		f:   f,
		uri: uri,
	}, nil
}

var _ urio.WriteCloser = (*FileFigWriter)(nil)

func (w FileFigWriter) Write(p []byte) (n int, err error) {
	return w.f.Write(p)
}
func (w FileFigWriter) Close() error {
	return w.f.Close()
}
func (w FileFigWriter) URI() urio.URI {
	return w.uri
}
