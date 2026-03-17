# 🏃 Indexing Runner

Indexing Runner is a kit that provides a canonical way to form an indexing pipeline, serving several interfaces and useful implementations. You can build your own implementations, or use prebuilt tools such as **Azure Content Understanding Analyzer** or **OpenAI Enricher** to orchestrate this pipeline.

---

## 🔗 URI-First Design

Every I/O boundary in this pipeline is addressed by a **URI**. There are no raw file paths or opaque strings — every source and destination is expressed as a typed URI.

```
file://cowboys.pdf       → input source on local filesystem
file://parts.json        → part write destination
file://search.json       → search write destination
```

The URI scheme determines which `storage.Client` implementation handles the I/O:

| Scheme | Implementation |
|---|---|
| `file://` | `storage.FileSystemClient` |
| `https://` | Azure Blob Storage (planned) |

This makes the pipeline **location-agnostic** — swapping from local filesystem to cloud storage only requires changing the URI scheme and injecting a different `storage.Client`. The pipeline code itself does not change.

```go
// URI parsing
type URI string

func (u URI) Scheme() string  // "file", "https", ...
func (u URI) Path() string    // "/absolute/path" or "relative/path"
func (u URI) Validate() error // checks scheme and path
```

---

## 🏛️ Architecture

The main flow of this pipeline is:

1. 📥 Reading bytes from an unknown source
2. 🔍 Analyzing the bytes and generating parts
3. 💾 Storing the parts
4. ✨ Enriching the parts, optimizing the data form
5. 🔎 Indexing the data to a search engine
6. ⚡ Caching expensive computations or I/O calls

```
Input → Analyze → Part Storage → Enrich → Search Indexing
```

The library defines interfaces for each stage. Domain teams inject their own implementations via the builder pattern. `analyze` + `part` are **required**; `enrich`, `search`, and `cache` are **optional**.

---

## 📁 Project Structure

```
cmd/indexing/          Entry point — assembles Config and runs pipeline
runner/                Pipeline orchestration (Runner, Config, Context, Wire)
input/                 Input interface + Provider interface
  file/                Filesystem Provider implementation
analyze/               Analyzer interface
  cu/                  Azure Content Understanding implementation
part/                  Part + PartWriter interfaces
  file/                Filesystem PartWriter implementation
enrich/                Enricher interface
  openai/              OpenAI-based enrichment (semantification + embeddings)
search/                SearchDoc + SearchWriter interfaces
  file/                Filesystem JSON SearchWriter
  es/                  Elasticsearch (placeholder)
cache/                 Cache + Hasher interfaces
  file/                Filesystem cache implementation
storage/               Storage abstraction (filesystem, Azure Blob)
uri/                   URI parsing ({scheme}://{path})
mime/                  MIME type utilities
```

---

## 🔌 Interfaces

| Interface | Required | Role |
|---|---|---|
| `Provider` | ✅ Yes | URI → Input |
| `Analyzer` | ✅ Yes | Input → []Part |
| `PartWriter` | ✅ Yes | Persists analyzed Parts |
| `Enricher` | ⬜ No | Parts → []SearchDoc (noop if absent) |
| `SearchWriter` | ⬜ No | Writes SearchDocs to search engine (noop if absent) |
| `Cache` | ⬜ No | Caches expensive API calls |

```go
// 📥 input
type Input interface {
    io.ReadCloser
    MimeType() mime.Type
    Meta()     map[string]any
}

type Provider interface {
    Provide(ctx context.Context, URI uri.URI) (Input, error)
}

// 🔍 analyze
type Analyzer interface {
    Analyze(ctx context.Context, input input.Input, cache cache.Cache) ([]part.Part, error)
}

// 🧩 part
type Part interface {
    MimeType() mime.Type
    Text()     string
    Raw()      []byte
}

type PartWriter interface {
    Write(ctx context.Context, URI uri.URI, parts []Part) error
}

// ✨ enrich
type Enricher interface {
    Enrich(ctx context.Context, parts []part.Part, cache cache.Cache) ([]search.SearchDoc, error)
}

// 🔎 search
type SearchDoc interface {
    Fields() map[string]any
}

type SearchWriter interface {
    Write(ctx context.Context, URI uri.URI, docs []SearchDoc) error
}

// ⚡ cache
type Cache interface {
    GetOrSet(ctx context.Context, key string, fn func() ([]byte, error)) ([]byte, error)
}
```

---

## ⚙️ Pipeline Execution

The `Runner` executes stages in order and yields `Event`s via `iter.Seq2[Event, error]`. The caller decides whether to continue or abort on error.

```go
type Event struct {
    Stage    string           // "provide" | "analyze" | "part" | "enrich" | "search"
    ICtx     *IndexingContext
    Duration time.Duration
}
```

---

## 🚀 Usage

```go
// 1️⃣ Create storage client
fsClient, _ := storage.NewFileSystemClient("testdata")

// 2️⃣ Assemble pipeline
r, _ := runner.New(
    runner.WithProvider(fileinput.New(fsClient)),
    runner.WithAnalyzer(cu.New(fsClient, cu.NewClient(endpoint, apiKey, http.DefaultClient))),
    runner.WithPartWriter(partfile.New(fsClient)),
    runner.WithEnricher(openai.New(oaiApiKey)),
    runner.WithSearchWriter(searchfile.New(fsClient)),
    runner.WithCache(filecache.New(fsClient)),
)

// 3️⃣ Create indexing context with URIs
ictx := runner.NewICtx(
    "file://cowboys.pdf",   // input source
    "file://parts.json",    // part write destination
    "file://search.json",   // search write destination
)

// 4️⃣ Run pipeline
for event, err := range r.Run(ctx, ictx) {
    if err != nil {
        log.Fatalf("failed at %s: %v", event.Stage, err)
    }
    log.Printf("stage %s done in %s", event.Stage, event.Duration)
}
```

---

## 🔁 Idempotency

The Runner does not guarantee idempotency — it provides the **structure** for it. Each implementation is responsible for its own guarantees.

| Stage | Strategy |
|---|---|
| 🔍 Analyze | Cache hit skips Azure CU API call |
| 💾 Part | Overwrite on write |
| ✨ Enrich | Cache hit skips OpenAI API calls |
| 🔎 Search | Upsert semantics in search engine |

On re-execution, cached stages are skipped automatically if the `Cache` implementation is provided.

---

## 📦 Dependencies

- **Go 1.23+** — required for `iter.Seq2`
- `github.com/openai/openai-go` — OpenAI client
- `github.com/google/uuid` — UUID generation
- `github.com/invopop/jsonschema` — JSON schema for structured output

---

## 🧭 Dependency Direction

```
runner/    →  input, analyze, part, enrich, search, cache
analyze/   →  input, part, cache
enrich/    →  part, search, cache
input/     →  uri, mime
part/      →  uri, mime
search/    →  uri
cache/     →  (independent)
storage/   →  (independent)
uri/       →  (independent)
mime/      →  (independent)
```