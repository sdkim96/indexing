package openai_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sdkim96/indexing/analyze/cu"
	openaiEnrich "github.com/sdkim96/indexing/enrich/openai"
	"github.com/sdkim96/indexing/part"
)

func newParts() []part.Part {
	data, _ := os.ReadFile("testdata/result_cowboys.json")
	var parts []cu.CUPart
	_ = json.Unmarshal(data, &parts)

	var result []part.Part
	for _, p := range parts {
		result = append(result, part.Part(&p))
	}
	return result
}
func TestOpenAIEnricher_Enrich(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	enricher := openaiEnrich.New(apiKey)

	parts := newParts()

	ctx := context.Background()
	docs, err := enricher.Enrich(ctx, parts)
	if err != nil {
		t.Fatalf("Enrich failed: %v", err)
	}

	t.Logf("docs: %+v", docs)
	_ = docs
}
