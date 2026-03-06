package analysis

import (
	"context"

	"net/http"
)

type Client struct {
	httpClient *http.Client
	Endpoint   string
	APIKey     string
}

func NewClient(endpoint, apiKey string) Client {
	return Client{
		httpClient: http.DefaultClient,
		Endpoint:   endpoint,
		APIKey:     apiKey,
	}
}

func (c Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.Header.Set("Ocp-Apim-Subscription-Key", c.APIKey)
	return c.httpClient.Do(req.WithContext(ctx))

}
