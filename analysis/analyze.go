package analysis

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sdkim96/indexing/job"
	"github.com/sdkim96/indexing/part"
)

const MaxPollInterval = 60 * time.Second

type Analysis struct {
	job job.Job
	Client
	pollCallback func(status OperationStatus)
	pollInterval time.Duration
}

func NewAnalysis(j job.Job, endpoint, apiKey string) *Analysis {
	return &Analysis{
		job:    j,
		Client: Client{endpoint: endpoint, apiKey: apiKey, httpClient: http.DefaultClient},
	}
}

type AnalysisOption func(*Analysis)

func WithHTTPClient(c *http.Client) AnalysisOption {
	return func(a *Analysis) {
		a.httpClient = c
	}
}

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

type AnalysisCompleted struct {
	Parts []part.Part `json:"parts"`
}

func (a *Analysis) Do(
	ctx context.Context,
	figCh chan<- FigureRequest,
	opts ...AnalysisOption,
) (AnalysisCompleted, error) {

	for _, opt := range opts {
		opt(a)
	}

	opLocation, err := a.Start(ctx, a.job.File())
	if err != nil {
		return AnalysisCompleted{}, err
	}
	interval := a.pollInterval
	if interval < 1 {
		interval = 1 * time.Second
	}

	for {
		result, err := a.GetResult(ctx, opLocation)
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
			return AnalysisCompleted{Parts: ConvertToParts(a.job.File().ID, result, figCh)}, nil
		case StatusFailed, StatusCanceled:
			return AnalysisCompleted{}, fmt.Errorf("analysis failed or canceled")
		}
	}
}
