package openai_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/sdkim96/indexing/analyze/cu"
	openaiEnrich "github.com/sdkim96/indexing/enrich/openai"
	"github.com/sdkim96/indexing/part"
)

// ── fixtures ──────────────────────────────────────────

func newParts() []part.Part {
	data, _ := os.ReadFile("testdata/result_cowboys.json")
	var cuParts []cu.CUPart
	_ = json.Unmarshal(data, &cuParts)

	var result []part.Part
	for _, p := range cuParts {
		result = append(result, part.Part(&p))
	}
	return result
}

func newSemanticDoc(t *testing.T) openaiEnrich.Document {
	t.Helper()
	data, err := os.ReadFile("testdata/semantify_result_cowboys.json")
	if err != nil {
		t.Skip("testdata/semantify_result_cowboys.json not found: run TestSemantify first")
	}

	var oaiResp map[string]any
	_ = json.Unmarshal(data, &oaiResp)

	output := oaiResp["output"].([]any)
	message := output[1].(map[string]any)
	content := message["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)

	var doc openaiEnrich.Document
	_ = json.Unmarshal([]byte(text), &doc)
	return doc
}

// ── helpers ───────────────────────────────────────────

func apiKey(t *testing.T) string {
	t.Helper()
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	return key
}

func newClient(t *testing.T) openai.Client {
	t.Helper()
	return openai.NewClient(option.WithAPIKey(apiKey(t)))
}

// ── Semantify ─────────────────────────────────────────

func TestSemantify(t *testing.T) {
	parts := newParts()
	doc, err := openaiEnrich.Semantify(context.Background(), newClient(t), parts)
	if err != nil {
		t.Fatalf("Semantify: %v", err)
	}
	if len(doc.Chunks) == 0 {
		t.Fatal("expected chunks, got 0")
	}
	for _, chunk := range doc.Chunks {
		if chunk.Topic == "" {
			t.Errorf("chunk has empty topic")
		}
		if chunk.Summary == "" {
			t.Errorf("chunk %q has empty summary", chunk.Topic)
		}
		if len(chunk.Keywords) == 0 {
			t.Errorf("chunk %q has empty keywords", chunk.Topic)
		}
		for _, idx := range chunk.Idxs {
			if idx < 0 || idx >= len(parts) {
				t.Errorf("idx %d out of range [0, %d)", idx, len(parts))
			}
		}
	}

	// testdata/semantify_result_cowboys.json 저장
	raw, _ := json.MarshalIndent(doc, "", "  ")
	_ = os.MkdirAll("testdata", 0755)
	if err := os.WriteFile("testdata/semantify_result_cowboys.json", raw, 0644); err != nil {
		t.Logf("warn: failed to save semantify_result_cowboys.json: %v", err)
	} else {
		t.Logf("saved testdata/semantify_result_cowboys.json")
	}

	t.Logf("chunks=%d", len(doc.Chunks))
	for _, c := range doc.Chunks {
		t.Logf("  %q → idxs=%v", c.Topic, c.Idxs)
		t.Logf("    summary: %s", c.Summary)
		t.Logf("    keywords: %v", c.Keywords)
	}
}

func TestSemantify_NoCoverage(t *testing.T) {
	parts := newParts()
	doc := newSemanticDoc(t) // API 호출 없이 캐시 사용

	covered := make(map[int]bool)
	for _, chunk := range doc.Chunks {
		for _, idx := range chunk.Idxs {
			covered[idx] = true
		}
	}
	for i := range parts {
		if !covered[i] {
			t.Errorf("part[%d] not covered in any chunk", i)
		}
	}
}

// ── Embed ─────────────────────────────────────────────

func TestEmbed(t *testing.T) {
	emb, err := openaiEnrich.Embed(context.Background(), newClient(t), "American cowboy history")
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}

	vector := emb.Vector()
	if len(vector) == 0 {
		t.Error("embedding vector is empty")
	}
	t.Logf("dim: %d", len(vector))
}

// ── Enrich (E2E) ──────────────────────────────────────

func TestOpenAIEnricher_Enrich(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	enricher := openaiEnrich.New(apiKey)
	parts := newParts()

	docs, err := enricher.Enrich(context.Background(), parts)
	if err != nil {
		t.Fatalf("Enrich failed: %v", err)
	}
	if len(docs) == 0 {
		t.Fatal("expected docs, got 0")
	}

	for i, doc := range docs {
		fields := doc.Fields()
		summary, _ := fields["summary"].(string)
		keywords, _ := fields["keywords"].([]string)
		vector, _ := fields["embedding"].([]float64)

		if summary == "" {
			t.Errorf("doc[%d] summary empty", i)
		}
		if len(keywords) == 0 {
			t.Errorf("doc[%d] keywords empty", i)
		}
		if len(vector) == 0 {
			t.Errorf("doc[%d] embedding empty", i)
		}
		t.Logf("doc[%d] topic=%q keywords=%v", i, fields["topic"], keywords)
	}
}
