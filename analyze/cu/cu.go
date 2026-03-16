package cu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/sdkim96/indexing/analyze"
	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/storage"
)

const MaxPollInterval = 60 * time.Second

type CU struct {
	storage          storage.Client
	http             *HTTPClient
	pollCallbackFunc func(status OperationStatus)
	pollInterval     time.Duration
}

var _ analyze.Analyzer = (*CU)(nil)

func New(
	storage storage.Client,
	http *HTTPClient,
	opts ...CUOptions,
) *CU {
	cu := &CU{
		storage: storage,
		http:    http,
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

	return ConvertToParts(ctx, op, cu.http, cu.storage)
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
