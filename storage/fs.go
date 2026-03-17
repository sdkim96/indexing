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
	"os"
	"path/filepath"

	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/uri"
)

var _ Client = (*FileSystemClient)(nil)

type FileSystemClient struct {
	root string
}

func (c FileSystemClient) Root() string {
	return c.root
}

func NewFileSystemClient(root string) (*FileSystemClient, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &FileSystemClient{root: abs}, nil
}

func (c *FileSystemClient) Open(ctx context.Context, path string) (io.ReadCloser, Meta, error) {
	fullPath := c.resolve(path)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, Meta{}, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, Meta{}, err
	}
	return f, Meta{
		MimeType: mime.GuessMimeType(fullPath),
		FileName: filepath.Base(fullPath),
		Size:     info.Size(),
	}, nil
}

func (c *FileSystemClient) Create(ctx context.Context, path string, mimeType mime.Type) (io.WriteCloser, uri.URI, error) {
	fullPath := c.resolve(path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, uri.URI(""), err
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, uri.URI(""), err
	}
	return f, uri.URI("file://" + fullPath), nil
}

func (c *FileSystemClient) resolve(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(c.root, path)
}
