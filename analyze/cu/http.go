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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sdkim96/indexing/mime"
)

const APIHeaderKey = "Ocp-Apim-Subscription-Key"
const StartEndpointPath = "/contentunderstanding/analyzers/prebuilt-layout:analyzeBinary?api-version=2025-11-01"
const GetFigureEndpointPath = "/contentunderstanding/analyzerResults/%s/files/contents/%d/figures/%s?api-version=2025-11-01"

func startURL(endpoint string) string {
	return endpoint + StartEndpointPath
}

func figureURL(endpoint, opID string, contentIdx int, figureID string) string {
	return fmt.Sprintf(endpoint+GetFigureEndpointPath, opID, contentIdx, figureID)
}

type FigureRequest struct {
	OpID       string
	ContentIdx int
	FigureID   string
	S3Key      string
}

type HTTPClient struct {
	client   *http.Client
	endpoint string
	apiKey   string
}

func NewClient(endpoint, apiKey string, client *http.Client) *HTTPClient {

	return &HTTPClient{
		client:   client,
		endpoint: endpoint,
		apiKey:   apiKey,
	}
}

func (c *HTTPClient) Start(ctx context.Context, b *Blob) (string, error) {

	req, err := http.NewRequestWithContext(ctx, "POST", startURL(c.endpoint), bytes.NewReader(b.Data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", b.Mime)

	resp, err := c.do(ctx, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	opLocation := resp.Header.Get("Operation-Location")
	if opLocation == "" {
		return "", fmt.Errorf("Operation-Location header not found")
	}

	return opLocation, nil
}

func (c *HTTPClient) GetResult(ctx context.Context, opLocation string) (Operation, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", opLocation, nil)
	if err != nil {
		return Operation{}, err
	}
	resp, err := c.do(ctx, req)
	if err != nil {
		return Operation{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Operation{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return Operation{}, err
	}

	var op Operation
	err = json.Unmarshal(body, &op)
	if err != nil {
		return Operation{}, err
	}

	return op, nil
}

func (c *HTTPClient) GetFigure(
	ctx context.Context,
	fig FigureRequest,
) ([]byte, mime.Type, error) {
	url := figureURL(c.endpoint, fig.OpID, fig.ContentIdx, fig.FigureID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := resp.Header.Get("Content-Type")
	return body, mime.Type(contentType), nil
}

func (c *HTTPClient) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.Header.Set(APIHeaderKey, c.apiKey)
	return c.client.Do(req.WithContext(ctx))
}
