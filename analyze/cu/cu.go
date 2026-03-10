package cu

import (
	"context"
	"fmt"
	"time"

	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/internal/storage"
	"github.com/sdkim96/indexing/part"
)

const MaxPollInterval = 60 * time.Second

// CU (Content Understanding) is responsible for analyzing raw data and producing structured parts.
type CU struct {
	blob             storage.Client
	http             *HTTPClient
	pollCallbackFunc func(status OperationStatus)
	pollInterval     time.Duration
}

func New(
	blob storage.Client,
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
			return ConvertToParts(ctx, result, cu.http, cu.blob)

		case StatusFailed, StatusCanceled:
			return nil, fmt.Errorf("analysis failed or canceled")
		}
	}
}
