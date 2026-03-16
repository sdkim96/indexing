package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/sdkim96/indexing/analyze/cu"
	cache "github.com/sdkim96/indexing/cache/file"
	enrich "github.com/sdkim96/indexing/enrich/openai"
	input "github.com/sdkim96/indexing/input/file"
	part "github.com/sdkim96/indexing/part/file"
	"github.com/sdkim96/indexing/runner"
	search "github.com/sdkim96/indexing/search/file"
	"github.com/sdkim96/indexing/storage"
)

func main() {
	storage, err := storage.NewFileSystemClient("testdata")
	if err != nil {
		panic(err)
	}

	cuFoundaryAPIKey := os.Getenv("AZURE_AI_FOUNDARY_API_KEY")
	cuAIServiceEndpoint := os.Getenv("AZURE_AI_SERVICES_ENDPOINT")
	oaiApiKey := os.Getenv("OPENAI_API_KEY")

	r, err := runner.New(
		runner.WithProvider(input.New(storage)),
		runner.WithAnalyzer(cu.New(storage, cu.NewClient(cuAIServiceEndpoint, cuFoundaryAPIKey, http.DefaultClient))),
		runner.WithPartWriter(part.New(storage)),
		runner.WithEnricher(enrich.New(oaiApiKey)),
		runner.WithSearchWriter(search.New(storage)),
		runner.WithCache(cache.New(storage, cache.WithCacheHitCallback(func(key string) { fmt.Println("Cache Hit!!! Key:", key) }))),
	)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	ictx := runner.NewICtx(
		"file://cowboys.pdf",
		"file://parts.json",
		"file://search.json",
	)
	for event, err := range r.Run(ctx, ictx) {
		if err != nil {
			panic(err)
		}
		println("Stage:", event.Stage, "Duration:", event.Duration.String())
	}
}
