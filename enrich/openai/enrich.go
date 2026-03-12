package openai

import (
	"context"
	"encoding/json"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/sdkim96/indexing/enrich"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
)

type OpenAIEnricher struct {
	apiKey string
}

func New(apiKey string) *OpenAIEnricher {
	return &OpenAIEnricher{apiKey: apiKey}
}

var _ enrich.Enricher = (*OpenAIEnricher)(nil)

func (e *OpenAIEnricher) Enrich(ctx context.Context, parts []part.Part) ([]search.SearchDoc, error) {
	oaiClient := openai.NewClient(
		option.WithAPIKey(e.apiKey),
	)

	resp, err := oaiClient.Responses.New(ctx, NewResponseAPIParam(
		WithSystemMessage(partsLinkingPrompt),
		WithPartsAsUserMessage(parts),
		WithModel("gpt-5-nano"),
		WithResponseFormat[Document](),
	).ToRequestParam())

	if err != nil {
		return nil, err
	}

	raw := []byte(resp.RawJSON())
	var result Document
	err = json.Unmarshal(raw, &result)
	if err != nil {
		return nil, err
	}
	os.WriteFile("output.json", raw, 0644)
	return nil, nil
}
