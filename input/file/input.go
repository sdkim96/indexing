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
	"io"

	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/mime"
)

var _ input.Input = (*FileInput)(nil)

type FileInput struct {
	readCloser io.ReadCloser
	mimeType   mime.Type
	meta       map[string]any
}

func NewFileInput(readCloser io.ReadCloser, mimeType mime.Type, meta map[string]any) input.Input {
	return &FileInput{
		readCloser: readCloser,
		mimeType:   mimeType,
		meta:       meta,
	}
}

func (f *FileInput) Read(p []byte) (n int, err error) {
	return f.readCloser.Read(p)
}

func (f *FileInput) Close() error {
	return f.readCloser.Close()
}

func (f *FileInput) MimeType() mime.Type {
	return f.mimeType
}

func (f *FileInput) Meta() map[string]any {
	return f.meta
}
