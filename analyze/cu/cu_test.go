package cu_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sdkim96/indexing/analyze/cu"
	"github.com/sdkim96/indexing/input"
	input_file "github.com/sdkim96/indexing/input/file"
	"github.com/sdkim96/indexing/internal/storage"
)

func newTestCU(t *testing.T, url, apiKey string, opts ...cu.CUOptions) *cu.CU {
	t.Helper()
	storageClient, err := storage.NewFileSystemClient("testdata/blob_storage")
	if err != nil {
		return nil
	}

	httpClient := cu.NewClient(url, apiKey, http.DefaultClient)
	return cu.New(storageClient, httpClient, opts...)
}

func newFileInput(t *testing.T, path string) input.Input {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	return input_file.NewFileInput(f, "application/pdf", map[string]any{
		"filename": path,
	})
}

func TestAnalyze_Success(t *testing.T) {
	var pollCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST":
			w.Header().Set("Operation-Location", "http://"+r.Host+"/operations/123")
			w.WriteHeader(http.StatusAccepted)

		case r.Method == "GET" && r.URL.Path == "/operations/123":
			cnt := pollCount.Add(1)
			var op cu.Operation
			if cnt < 2 {
				op = cu.Operation{Status: cu.StatusRunning}
			} else {
				op = cu.Operation{
					Status: cu.StatusSucceeded,
					Result: &cu.AnalysisResult{
						Contents: []cu.AnalysisContent{
							{Kind: cu.AnalysisContentKindDocument, MimeType: "application/pdf"},
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

	analyzer := newTestCU(t, srv.URL, "test-api-key",
		cu.WithPollInterval(10*time.Millisecond),
	)
	inp := newFileInput(t, "testdata/test_with_image.pdf")
	defer inp.Close()

	_, err := analyzer.Analyze(context.Background(), inp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pollCount.Load() < 2 {
		t.Fatalf("expected at least 2 polls, got %d", pollCount.Load())
	}
}

func TestAnalyze_Failed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST":
			w.Header().Set("Operation-Location", "http://"+r.Host+"/operations/456")
			w.WriteHeader(http.StatusAccepted)
		case r.Method == "GET":
			op := cu.Operation{Status: cu.StatusFailed}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(op)
		}
	}))
	defer srv.Close()

	analyzer := newTestCU(t, srv.URL, "test-api-key",
		cu.WithPollInterval(10*time.Millisecond),
	)
	inp := input_file.NewFileInput(nopCloser{}, "application/pdf", nil)

	_, err := analyzer.Analyze(context.Background(), inp)
	if err == nil {
		t.Fatal("expected error for failed analysis")
	}
}

func TestAnalyze_ContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST":
			w.Header().Set("Operation-Location", "http://"+r.Host+"/operations/789")
			w.WriteHeader(http.StatusAccepted)
		case r.Method == "GET":
			op := cu.Operation{Status: cu.StatusRunning}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(op)
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	analyzer := newTestCU(t, srv.URL, "test-api-key",
		cu.WithPollInterval(10*time.Millisecond),
	)
	inp := input_file.NewFileInput(nopCloser{}, "application/pdf", nil)

	_, err := analyzer.Analyze(ctx, inp)
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestAnalyze_Integration(t *testing.T) {
	endpoint := os.Getenv("AZURE_AI_SERVICES_ENDPOINT")
	apiKey := os.Getenv("AZURE_AI_FOUNDARY_API_KEY")
	if endpoint == "" || apiKey == "" {
		t.Skip("integration env not set")
	}

	analyzer := newTestCU(t, endpoint, apiKey,
		cu.WithPollCallback(func(status cu.OperationStatus) {
			t.Logf("poll status: %s", status)
		}),
	)
	inp := newFileInput(t, "testdata/report.pdf")
	defer inp.Close()

	parts, err := analyzer.Analyze(context.Background(), inp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b, _ := json.MarshalIndent(parts, "", "  ")
	os.WriteFile("testdata/result_report.json", b, 0644)
	t.Logf("got %d parts", len(parts))
}

func TestGetFigure(t *testing.T) {
	endpoint := os.Getenv("AZURE_AI_SERVICES_ENDPOINT")
	apiKey := os.Getenv("AZURE_AI_FOUNDARY_API_KEY")
	if endpoint == "" || apiKey == "" {
		t.Skip("integration env not set")
	}

	client := cu.NewClient(endpoint, apiKey, http.DefaultClient)
	data, contentType, err := client.GetFigure(
		context.Background(),
		cu.FigureRequest{
			OpID:       "02494fe5-3589-4c42-9aa1-f1c03de0152a",
			ContentIdx: 0,
			FigureID:   "1.1",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("got figure with content type: %s, size: %d bytes", contentType, len(data))
}

func TestConvertToParts(t *testing.T) {
	endpoint := os.Getenv("AZURE_AI_SERVICES_ENDPOINT")
	apiKey := os.Getenv("AZURE_AI_FOUNDARY_API_KEY")
	if endpoint == "" || apiKey == "" {
		t.Skip("integration env not set")
	}

	var op cu.Operation
	data, _ := os.ReadFile("testdata/cowboys.json")
	json.Unmarshal(data, &op)

	httpClient := cu.NewClient(endpoint, apiKey, http.DefaultClient)
	storageClient, err := storage.NewFileSystemClient("testdata/blob_storage")
	if err != nil {
		t.Fatalf("failed to create storage client: %v", err)
	}

	parts, err := cu.ConvertToParts(context.Background(), op, httpClient, storageClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b, _ := json.MarshalIndent(parts, "", "  ")
	os.WriteFile("testdata/cowboys_converted_parts.json", b, 0644)
	t.Logf("converted parts: %d", len(parts))
}

// nopCloser는 빈 io.ReadCloser 테스트용
type nopCloser struct{}

func (nopCloser) Read(p []byte) (int, error) { return 0, nil }
func (nopCloser) Close() error               { return nil }
