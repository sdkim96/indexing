package cu

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/internal/blob"
	"github.com/sdkim96/indexing/part"
)

const MaxPollInterval = 60 * time.Second

// CU (Content Understanding) is responsible for analyzing raw data and producing structured parts.
type CU struct {
	blob             blob.Client
	http             *HTTPClient
	pollCallbackFunc func(status OperationStatus)
	pollInterval     time.Duration
}

func New(
	blob blob.Client,
	http *HTTPClient,
	opts ...CUOptions,
) *CU {
	cu := &CU{
		blob: blob,
		http: http,
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

func (cu *CU) Analyze(ctx context.Context, inp input.Input) ([]part.Part, error) {

	opLocation, err := cu.http.Start(ctx, inp)
	if err != nil {
		return nil, err
	}

	interval := cu.pollInterval
	if interval < 1 {
		interval = 1 * time.Second
	}

	for {
		result, err := cu.http.GetResult(ctx, opLocation)
		if err != nil {
			return nil, err
		}
		if cu.pollCallbackFunc != nil {
			cu.pollCallbackFunc(result.Status)
		}

		switch result.Status {
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
			figCh := make(chan FigureRequest)

			var parts []part.Part
			go func() {
				parts = ConvertToParts(result, figCh)
				close(figCh)
			}()

			for req := range figCh {
				data, mimeType, err := cu.http.GetFigure(ctx, req)
				if err != nil {
					continue
				}
				w, err := cu.blob.Create(ctx, req.FigureID, mimeType)
				if err != nil {
					continue
				}
				io.Copy(w, bytes.NewReader(data))
				w.Close()
			}
			return parts, nil

		case StatusFailed, StatusCanceled:
			return nil, fmt.Errorf("analysis failed or canceled")
		}
	}
}
