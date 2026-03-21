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

package openai

import (
	"context"
	"encoding/json"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/sdkim96/indexing/cache"
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
func (e *OpenAIEnricher) Enrich(ctx context.Context, sourceID string, parts []part.Part, c cache.Cache) ([]search.SearchDoc, error) {
	oaiClient := openai.NewClient(
		option.WithAPIKey(e.apiKey),
	)

	semantifyReq := NewResponseAPIParam(
		WithSystemMessage(enrichPrompt),
		WithPartsAsUserMessage(parts),
		WithModel("gpt-5-nano"),
		WithResponseFormat[Document](),
		WithReasoningEffort("medium"),
	)
	val, err := c.GetOrSet(ctx, semantifyReq.FingerPrint("semantify"), func() ([]byte, error) {
		return Semantify(ctx, oaiClient, semantifyReq)
	})
	if err != nil {
		return nil, err
	}
	var d Document
	err = json.Unmarshal(val, &d)
	if err != nil {
		return nil, err
	}

	var docs []search.SearchDoc
	for _, chunk := range d.Chunks {

		embeddingReq := NewEmbeddingAPIParam(
			WithEmbeddingInput(chunk.Summary),
			WithEmbeddingModel("text-embedding-3-small"),
		)
		val, err := c.GetOrSet(ctx, embeddingReq.FingerPrint("embedding"), func() ([]byte, error) {
			return Embed(ctx, oaiClient, embeddingReq)
		})
		if err != nil {
			return nil, err
		}
		var embedding []float64
		err = json.Unmarshal(val, &embedding)
		if err != nil {
			return nil, err
		}
		docs = append(docs, OpenAISearchDoc{
			Title:     chunk.Topic,
			Embedding: Embedding{vector: embedding},
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

func Semantify(ctx context.Context, c openai.Client, req *ResponseAPIParam) ([]byte, error) {
	resp, err := c.Responses.New(ctx, req.ToRequestParam())
	if err != nil {
		return nil, err
	}
	return []byte(resp.OutputText()), nil
}

func Embed(ctx context.Context, c openai.Client, req *EmbeddingAPIParam) ([]byte, error) {
	resp, err := c.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(req.input),
		},
		Model: req.model,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(resp.Data[0].Embedding)
}
