package analysis

import (
	"context"

	"github.com/sdkim96/indexing/job"
)

// AnalysisInput represents an input for analysis, which can be either a URL or raw data.
type AnalysisInput struct {
	URL          *string `json:"url,omitempty"`
	Data         []byte  `json:"data,omitempty"` // base64 encoded
	Name         *string `json:"name,omitempty"`
	MimeType     *string `json:"mimeType,omitempty"`
	ContentRange *string `json:"range,omitempty"`
}

// StartRequest represents a request to begin analysis with multiple inputs.
type StartRequest struct {
	Inputs []AnalysisInput `json:"inputs"`
}

func (c Client) start(ctx context.Context, j job.Job) (string, error) {

	return "", nil
}
