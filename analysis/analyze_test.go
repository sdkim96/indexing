package analysis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sdkim96/indexing/analysis"
	"github.com/sdkim96/indexing/job"
)

func TestDo_Success(t *testing.T) {
	var pollCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/contentunderstanding/analyzers/prebuilt-invoice:analyze":
			w.Header().Set("Operation-Location", "http://"+r.Host+"/operations/123")
			w.WriteHeader(http.StatusAccepted)

		case r.Method == "GET" && r.URL.Path == "/operations/123":
			cnt := pollCount.Add(1)
			var op analysis.Operation
			if cnt < 2 {
				op = analysis.Operation{Status: analysis.StatusRunning}
			} else {
				op = analysis.Operation{
					Status: analysis.StatusSucceeded,
					Result: &analysis.AnalysisResult{
						Contents: []analysis.AnalysisContent{
							{Kind: analysis.AnalysisContentKindDocument, MimeType: "application/pdf"},
						},
					},
				}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(op)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	pdfBytes, err := os.ReadFile("testdata/test_with_image.pdf")
	if err != nil {
		t.Fatalf("failed to read test pdf: %v", err)
	}
	f := job.File{Name: "test_with_image.pdf", Bytes: pdfBytes, MimeType: "application/pdf"}
	j := job.NewJob("job-1", "key-1", f)

	a := analysis.NewAnalysis(j, srv.URL, "test-api-key")
	_, err = a.Do(
		context.Background(),
		nil,
		analysis.WithPollInterval(10*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pollCount.Load() < 2 {
		t.Fatalf("expected at least 2 polls, got %d", pollCount.Load())
	}
}

func TestDo_Failed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST":
			w.Header().Set("Operation-Location", "http://"+r.Host+"/operations/456")
			w.WriteHeader(http.StatusAccepted)

		case r.Method == "GET":
			op := analysis.Operation{Status: analysis.StatusFailed}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(op)
		}
	}))
	defer srv.Close()

	f := job.File{Name: "test.pdf", Bytes: []byte("fake"), MimeType: "application/pdf"}
	j := job.NewJob("job-2", "key-2", f)

	a := analysis.NewAnalysis(j, srv.URL, "test-api-key")
	_, err := a.Do(
		context.Background(),
		nil,
		analysis.WithPollInterval(10*time.Millisecond),
	)
	if err == nil {
		t.Fatal("expected error for failed analysis")
	}
}

func TestDo_ContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST":
			w.Header().Set("Operation-Location", "http://"+r.Host+"/operations/789")
			w.WriteHeader(http.StatusAccepted)

		case r.Method == "GET":
			op := analysis.Operation{Status: analysis.StatusRunning}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(op)
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	f := job.File{Name: "test.pdf", Bytes: []byte("fake"), MimeType: "application/pdf"}
	j := job.NewJob("job-3", "key-3", f)

	a := analysis.NewAnalysis(j, srv.URL, "test-api-key")
	_, err := a.Do(
		ctx,
		nil,
		analysis.WithPollInterval(10*time.Millisecond),
	)
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestDo_Integration(t *testing.T) {
	endpoint := os.Getenv("AZURE_AI_SERVICES_ENDPOINT")
	apiKey := os.Getenv("AZURE_AI_FOUNDARY_API_KEY")
	if endpoint == "" || apiKey == "" {
		t.Skip("AZURE_AI_SERVICES_ENDPOINT and AZURE_AI_FOUNDARY_API_KEY not set, skipping integration test")
	}
	pdfBytes, err := os.ReadFile("testdata/cowboys.pdf")
	if err != nil {
		t.Fatalf("failed to read test pdf: %v", err)
	}
	f := job.File{Name: "cowboys.pdf", Bytes: pdfBytes, MimeType: "application/pdf"}
	j := job.NewJob("integration-1", "key-1", f)

	ch := make(chan analysis.FigureRequest, 5)

	ctx := context.Background()
	a := analysis.NewAnalysis(j, endpoint, apiKey)
	wg := sync.WaitGroup{}
	listen := sync.WaitGroup{}

	result, err := a.Do(
		ctx,
		ch,
		analysis.WithPollCallback(func(status analysis.OperationStatus) {
			t.Logf("poll status: %s", status)
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	close(ch)
	for req := range ch {
		wg.Add(1)
		listen.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("Starting GetFigure %s\n", req.FigureID)
			_, contentType, err := a.Client.GetFigure(ctx, req)
			fmt.Printf("Got figure %s with content type: %s, err: %v\n", req.FigureID, contentType, err) // 호출 후
		}()
	}

	wg.Wait()

	b, _ := json.MarshalIndent(result.Parts, "", "  ")
	os.WriteFile("testdata/result_cowboys.json", b, 0644)
	t.Logf("got %d parts", len(result.Parts))

}

func TestGetFigure(t *testing.T) {
	endpoint := os.Getenv("AZURE_AI_SERVICES_ENDPOINT")
	apiKey := os.Getenv("AZURE_AI_FOUNDARY_API_KEY")
	if endpoint == "" || apiKey == "" {
		t.Skip("AZURE_AI_SERVICES_ENDPOINT and AZURE_AI_FOUNDARY_API_KEY not set, skipping integration test")
	}
	client := analysis.NewClient(endpoint, apiKey, http.DefaultClient)

	bytes, contentType, err := client.GetFigure(
		context.Background(),
		analysis.FigureRequest{
			OpID:       "02494fe5-3589-4c42-9aa1-f1c03de0152a",
			ContentIdx: 0,
			FigureID:   "1.1",
			S3Key:      "",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("got figure with content type: %s, size: %d bytes", contentType, len(bytes))
}

func TestConvertToParts(t *testing.T) {
	fileID := "fake-file-id"
	ch := make(chan analysis.FigureRequest, 5)

	var op analysis.Operation
	data, _ := os.ReadFile("testdata/cowboys.json")

	json.Unmarshal(data, &op)
	parts := analysis.ConvertToParts(fileID, op, ch)
	close(ch)

	b, _ := json.MarshalIndent(parts, "", "  ")
	os.WriteFile("testdata/cowboys_converted_parts.json", b, 0644)
	t.Logf("converted parts: %s", string(b))

	for sig := range ch {
		t.Logf("%v", sig)
	}
}
