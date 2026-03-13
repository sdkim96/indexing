package openai

import (
	"context"
	"encoding/json"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/sdkim96/indexing/enrich"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/search"
)

var _ enrich.Enricher = (*OpenAIEnricher)(nil)

// OpenAIEnricher is an implementation of the Enricher interface that uses OpenAI's API to enrich document parts.
type OpenAIEnricher struct {
	apiKey string
}

func New(apiKey string) *OpenAIEnricher {
	return &OpenAIEnricher{apiKey: apiKey}
}

// Enrich takes a list of document parts, processes them using OpenAI's API to group related parts together,
// summarize the content, extract keywords, and generate embeddings. It returns a list of SearchDoc objects
// that contain the enriched information for each topic identified in the document.
func (e *OpenAIEnricher) Enrich(ctx context.Context, parts []part.Part) ([]search.SearchDoc, error) {
	oaiClient := openai.NewClient(
		option.WithAPIKey(e.apiKey),
	)

	doc, err := Semantify(ctx, oaiClient, parts)
	if err != nil {
		return nil, err
	}

	var docs []search.SearchDoc
	for _, chunk := range doc.Chunks {
		embedding, err := Embed(ctx, oaiClient, chunk.Summary)
		if err != nil {
			return nil, err
		}
		docs = append(docs, SearchDoc{
			Title:     chunk.Topic,
			Embedding: embedding,
			SummaryAndKeywords: SummaryAndKeywords{
				Summary:  Summary{text: chunk.Summary},
				Keywords: Keywords{words: chunk.Keywords},
			},
			Meta: map[string]any{
				"topic": chunk.Topic,
				"idxs":  chunk.Idxs,
			},
		})
	}
	return docs, nil
}

func Semantify(ctx context.Context, c openai.Client, parts []part.Part) (Document, error) {
	resp, err := c.Responses.New(ctx, NewResponseAPIParam(
		WithSystemMessage(enrichPrompt),
		WithPartsAsUserMessage(parts),
		WithModel("gpt-5-nano"),
		WithResponseFormat[Document](),
	).ToRequestParam())
	if err != nil {
		return Document{}, err
	}

	var result Document
	if err := json.Unmarshal([]byte(resp.RawJSON()), &result); err != nil {
		return Document{}, err
	}
	return result, nil
}

func Embed(ctx context.Context, c openai.Client, text string) (Embedding, error) {
	resp, err := c.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(text),
		},
		Model: "text-embedding-3-small",
	})
	if err != nil {
		return Embedding{}, err
	}
	return Embedding{vector: resp.Data[0].Embedding}, nil
}
