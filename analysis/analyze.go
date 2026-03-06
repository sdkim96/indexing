package analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/sdkim96/indexing/job"
	"github.com/sdkim96/indexing/part"
)

const MaxPollInterval = 60 * time.Second

type AnalysisCompleted struct {
	Parts []part.Part `json:"parts"`
}

type Analysis struct {
	job          job.Job
	client       Client
	pollCallback func(status OperationStatus)
	pollInterval time.Duration
}

type AnalysisOption func(*Analysis)

func WithPollCallback(callback func(status OperationStatus)) AnalysisOption {
	return func(a *Analysis) {
		a.pollCallback = callback
	}
}

func WithPollInterval(interval time.Duration) AnalysisOption {
	return func(a *Analysis) {
		a.pollInterval = interval
	}
}

func (a *Analysis) Do(
	ctx context.Context,
	opt ...AnalysisOption,
) (AnalysisCompleted, error) {

	for _, o := range opt {
		o(a)
	}

	opLocation, err := a.client.start(ctx, a.job)
	if err != nil {
		return AnalysisCompleted{}, err
	}
	interval := a.pollInterval
	if interval < 1 {
		interval = 1 * time.Second
	}

	for {
		result, err := a.client.getResult(ctx, opLocation)
		if err != nil {
			return AnalysisCompleted{}, err
		}
		if a.pollCallback != nil {
			a.pollCallback(result.Status)
		}
		switch result.Status {
		case StatusNotStarted, StatusRunning:

			select {
			case <-ctx.Done():
				return AnalysisCompleted{}, ctx.Err()
			case <-time.After(interval):
			}
			interval *= 2
			if interval > MaxPollInterval {
				interval = MaxPollInterval
			}

		case StatusSucceeded:
			return AnalysisCompleted{Parts: convertToParts(result.Result)}, nil
		case StatusFailed, StatusCanceled:
			return AnalysisCompleted{}, fmt.Errorf("analysis failed or canceled")
		}
	}
}
