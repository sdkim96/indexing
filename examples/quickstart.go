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

package main

import (
	"context"
	"net/http"
	"os"

	cu "github.com/sdkim96/indexing/analyze/cu"
	cuFileFigure "github.com/sdkim96/indexing/analyze/cu/filefigure"
	cache "github.com/sdkim96/indexing/cache/file"
	oai "github.com/sdkim96/indexing/enrich/openai"
	input "github.com/sdkim96/indexing/input/file"
	"github.com/sdkim96/indexing/mime"
	part "github.com/sdkim96/indexing/part/file"
	"github.com/sdkim96/indexing/runner"
	search "github.com/sdkim96/indexing/search/file"
	"github.com/sdkim96/indexing/urio"
)

func main() {

	cuFoundaryAPIKey := os.Getenv("AZURE_AI_FOUNDARY_API_KEY")
	cuAIServiceEndpoint := os.Getenv("AZURE_AI_SERVICES_ENDPOINT")
	oaiApiKey := os.Getenv("OPENAI_API_KEY")

	provider, err := input.NewFileProvider(urio.URI("file://testdata/cowboys.pdf"))
	if err != nil {
		panic(err)
	}
	partWriter, err := part.NewFilePartWriter(urio.URI("file://testdata/parts.json"))
	if err != nil {
		panic(err)
	}
	searchWriter, err := search.NewFileSearchWriter(urio.URI("file://testdata/search.json"))
	if err != nil {
		panic(err)
	}
	cache := cache.New("testdata/cache")
	cuAnalyzer := cu.New(
		cu.NewClient(cuAIServiceEndpoint, cuFoundaryAPIKey, http.DefaultClient),
		func(ctx context.Context, name string, mimeType mime.Type) (urio.WriteCloser, error) {
			return cuFileFigure.NewFileFigWriter(urio.URI("file://testdata/" + name))
		},
	)

	r, err := runner.New(
		runner.WithProvider(provider),
		runner.WithAnalyzer(cuAnalyzer),
		runner.WithPartWriter(partWriter),
		runner.WithEnricher(oai.New(oaiApiKey)),
		runner.WithSearchWriter(searchWriter),
		runner.WithCache(cache),
	)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	for event, err := range r.Run(ctx, "cowboys.pdf") {
		if err != nil {
			panic(err)
		}
		println("Stage:", event.Stage, "Duration:", event.Duration.String())
	}
}
