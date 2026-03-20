# 🏃 Indexing

A Go library that provides a canonical way to form an indexing pipeline, serving composable interfaces and useful implementations. Build your own implementations, or use prebuilt tools such as **Azure Content Understanding Analyzer** or **OpenAI Enricher** to orchestrate this pipeline.

---

## 🏛️ Architecture

The main flow of this pipeline is:

1. 📥 Reading bytes from a source
2. 🔍 Analyzing the bytes and generating parts
3. 💾 Storing the parts
4. ✨ Enriching the parts, optimizing the data form
5. 🔎 Indexing the data to a search engine
6. ⚡ Caching expensive computations or I/O calls

```
Input → Analyze → Part Storage → Enrich → Search Indexing
```

The library defines interfaces for each stage. Domain teams inject their own implementations via the builder pattern. `provider` + `analyzer` + `partWriter` are **required**; `enricher`, `searchWriter`, and `cache` are **optional**.

Each component is **self-contained** — it receives its URI (or destination) at construction time and manages its own I/O. The Runner does not pass URIs between stages.

---

## 🔗 URI

Every I/O boundary is addressed by a **URI** (`urio.URI`). There are no raw file paths or opaque strings — every source and destination is expressed as a typed URI.

```
file://cowboys.pdf       → input source on local filesystem
file://parts.json        → part write destination
file://search.json       → search write destination
```

The URI scheme determines which implementation handles the I/O:

| Scheme | Example |
|---|---|
| `file://` | Local filesystem |
| `blob://` | Azure Blob Storage |

```go
type URI string

func (u URI) Scheme() string  // "file", "blob", ...
func (u URI) Path() string    // path after "://"
func (u URI) Validate() error // checks scheme and path
```

---

## 📁 Project Structure

```
examples/              Usage examples
runner/                Pipeline orchestration (Runner, Config, Wire)
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
cache/                 Cache + Hasher interfaces
  file/                Filesystem cache implementation
urio/                  URI parsing ({scheme}://{path})
mime/                  MIME type utilities
```

---

## 🔌 Interfaces

| Interface | Required | Role |
|---|---|---|
| `Provider` | ✅ Yes | Reads source into Input |
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
    Provide(ctx context.Context) (Input, error)
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
    Write(ctx context.Context, parts []Part) error
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
    Write(ctx context.Context, docs []SearchDoc) error
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
    Stage    string          // "provide" | "analyze" | "part" | "enrich" | "search"
    Duration time.Duration
}
```

---

## 🚀 Usage

```go
// 1️⃣ Create components with their URIs
provider, _ := fileinput.NewFileProvider(urio.URI("file://testdata/cowboys.pdf"))
partWriter, _ := partfile.NewFilePartWriter(urio.URI("file://testdata/parts.json"))
searchWriter, _ := searchfile.NewFileSearchWriter(urio.URI("file://testdata/search.json"))
cache := filecache.New("testdata/cache")
analyzer := cu.New(
    cu.NewClient(endpoint, apiKey, http.DefaultClient),
    func(ctx context.Context, name string) (urio.WriteCloser, error) {
        return cu.NewFileFigWriter(urio.URI("file://testdata/" + name))
    },
)

// 2️⃣ Assemble pipeline
r, _ := runner.New(
    runner.WithProvider(provider),
    runner.WithAnalyzer(analyzer),
    runner.WithPartWriter(partWriter),
    runner.WithEnricher(openai.New(oaiApiKey)),
    runner.WithSearchWriter(searchWriter),
    runner.WithCache(cache),
)

// 3️⃣ Run pipeline
for event, err := range r.Run(ctx) {
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

- **Go 1.25+** — required for `iter.Seq2`
- `github.com/openai/openai-go` — OpenAI client
- `github.com/google/uuid` — UUID generation
- `github.com/invopop/jsonschema` — JSON schema for structured output

---

## 🧭 Dependency Direction

```
runner/    →  input, analyze, part, enrich, search, cache
analyze/   →  input, part, cache, urio, mime
enrich/    →  part, search, cache
input/     →  urio, mime
part/      →  mime
search/    →  (independent)
cache/     →  (independent)
urio/      →  (independent)
mime/      →  (independent)
```
