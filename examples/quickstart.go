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

package examples

import (
	"context"
	"fmt"
	"net/http"
	"os"

	cu "github.com/sdkim96/indexing/analyze/cu"
	cache "github.com/sdkim96/indexing/cache/file"
	oai "github.com/sdkim96/indexing/enrich/openai"
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
		runner.WithEnricher(oai.New(oaiApiKey)),
		runner.WithSearchWriter(search.New(storage)),
		runner.WithCache(cache.New(storage, cache.WithCacheHitCallback(func(key string) { fmt.Println("Cache Hit!!! Key:", key) }))),
	)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	ictx := runner.NewICtx(
		runner.WithInputKey("file://cowboys.txt"),
		runner.WithPartWriteKey("file://parts.json"),
		runner.WithSearchWriteKey("file://search.json"),
	)
	for event, err := range r.Run(ctx, ictx) {
		if err != nil {
			panic(err)
		}
		println("Stage:", event.Stage, "Duration:", event.Duration.String())
	}
}
