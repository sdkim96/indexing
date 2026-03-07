package analysis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sdkim96/indexing/job"
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

type Client struct {
	httpClient *http.Client
	endpoint   string
	apiKey     string
}

func NewClient(endpoint, apiKey string, client *http.Client) *Client {

	return &Client{
		httpClient: client,
		endpoint:   endpoint,
		apiKey:     apiKey,
	}
}

func (c *Client) Start(ctx context.Context, f *job.File) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", startURL(c.endpoint), bytes.NewReader(f.Bytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", f.MimeType)

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

func (c *Client) GetResult(ctx context.Context, opLocation string) (Operation, error) {
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

func (c *Client) GetFigure(
	ctx context.Context,
	fig FigureRequest,
) ([]byte, string, error) {
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
	return body, contentType, nil
}

func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.Header.Set(APIHeaderKey, c.apiKey)
	return c.httpClient.Do(req.WithContext(ctx))
}
