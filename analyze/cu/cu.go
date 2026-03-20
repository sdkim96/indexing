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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sdkim96/indexing/analyze"
	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/urio"
)

const MaxPollInterval = 60 * time.Second

type CU struct {
	http             *HTTPClient
	figWriter        func(ctx context.Context, name string, mimeType mime.Type) (urio.WriteCloser, error)
	pollCallbackFunc func(status OperationStatus)
	pollInterval     time.Duration
}

var _ analyze.Analyzer = (*CU)(nil)

func New(
	http *HTTPClient,
	figWriter func(ctx context.Context, name string, mimeType mime.Type) (urio.WriteCloser, error),
	opts ...CUOptions,
) *CU {
	cu := &CU{
		figWriter: figWriter,
		http:      http,
	}
	for _, opt := range opts {
		opt(cu)
	}
	return cu
}

type CUOptions func(*CU)

func WithPollCallback(callback func(status OperationStatus)) CUOptions {
	return func(cu *CU) {
		cu.pollCallbackFunc = callback
	}
}

func WithPollInterval(interval time.Duration) CUOptions {
	return func(cu *CU) {
		cu.pollInterval = interval
	}
}

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

func (w FileFigWriter) Write(p []byte) (n int, err error) {
	return w.f.Write(p)
}
func (w FileFigWriter) Close() error {
	return w.f.Close()
}
func (w FileFigWriter) URI() urio.URI {
	return w.uri
}

var _ urio.WriteCloser = (*FileFigWriter)(nil)

type Blob struct {
	Data []byte
	Mime string
}

func (b *Blob) FingerPrint(prefix string) string {
	key := cache.Sha256(b.Data, []byte(b.Mime))
	if prefix != "" {
		key = prefix + ":" + key
	}
	return key
}

var _ cache.Hasher = (*Blob)(nil)

func (cu *CU) Analyze(ctx context.Context, inp input.Input, c cache.Cache) ([]part.Part, error) {
	data, err := io.ReadAll(inp)
	if err != nil {
		return nil, err
	}
	b := &Blob{
		Data: data,
		Mime: string(inp.MimeType()),
	}

	val, err := c.GetOrSet(ctx, b.FingerPrint("cu"), func() ([]byte, error) {
		opLocation, err := cu.http.Start(ctx, b)
		if err != nil {
			return nil, err
		}
		op, err := cu.poll(ctx, opLocation)
		if err != nil {
			return nil, err
		}
		return json.Marshal(op)
	})
	if err != nil {
		return nil, err
	}

	var op Operation
	if err := json.Unmarshal(val, &op); err != nil {
		return nil, err
	}

	return ConvertToParts(ctx, op, cu.http, cu.figWriter)
}

func (cu *CU) poll(ctx context.Context, opLocation string) (*Operation, error) {
	interval := cu.pollInterval
	if interval < 1 {
		interval = 1 * time.Second
	}

	for {
		op, err := cu.http.GetResult(ctx, opLocation)
		if err != nil {
			return nil, err
		}
		if cu.pollCallbackFunc != nil {
			cu.pollCallbackFunc(op.Status)
		}

		switch op.Status {
		case StatusNotStarted, StatusRunning:
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(interval):
			}
			interval *= 2
			if interval > MaxPollInterval {
				interval = MaxPollInterval
			}

		case StatusSucceeded:
			return &op, nil

		case StatusFailed, StatusCanceled:
			return nil, fmt.Errorf("analysis failed or canceled")
		}
	}
}
