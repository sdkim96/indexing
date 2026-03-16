package file

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sdkim96/indexing/enrich/openai"
	"github.com/sdkim96/indexing/search"
	"github.com/sdkim96/indexing/storage"
	"github.com/sdkim96/indexing/uri"
)

const testDocURI = "file:///Users/sungdongkim/works/indexing/search/file/testdata/enrich_result_cowboys.json"
const testURI = "file:///Users/sungdongkim/works/indexing/search/file/testdata/search_docs.json"

func newTestClient() *storage.FileSystemClient {
	client, err := storage.NewFileSystemClient("")
	if err != nil {
		panic(err)
	}
	return client
}

func newDocs() []search.SearchDoc {
	path := uri.URI(testDocURI).Path()
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var raw []openai.OpenAISearchDoc
	if err := json.Unmarshal(data, &raw); err != nil {
		panic(err)
	}

	docs := make([]search.SearchDoc, len(raw))
	for i := range raw {
		docs[i] = &raw[i]
	}
	return docs
}

func TestSearchWriter_Write(t *testing.T) {
	client := newTestClient()
	writer := New(client)

	writer.Write(context.Background(), testURI, newDocs())
}
