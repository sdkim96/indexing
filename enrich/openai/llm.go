package openai

import (
	"context"
	"log"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/sdkim96/indexing/enrich"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
)

type OpenAIEnricher struct {
	apiKey string
}

var _ enrich.Enricher = (*OpenAIEnricher)(nil)

const (
	embeddingEndpoint = "https://api.openai.com/v1/embeddings"
	llmEndpoint       = "https://api.openai.com/v1/llm"
)

func (l *OpenAIEnricher) Enrich(ctx context.Context, parts []part.Part) ([]search.SearchDoc, error) {
	oaiClient := openai.NewClient(
		option.WithAPIKey(l.apiKey),
	)

	resp, err := oaiClient.Responses.New(ctx, NewResponseAPIParam(
		WithSystemMessage(""),
		WithUserMessage(""),
		WithModel("gpt-4o-mini"),
	).ToRequestParam())

	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Println(resp.OutputText())
}
