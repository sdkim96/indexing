# 🚀 Release v0.2.1

## Summary

v0.2.1 is a major simplification of the pipeline architecture. Components are now **self-contained** — each receives its URI at construction time and manages its own I/O. The Runner no longer brokers URIs or intermediate state between stages. This results in a cleaner API with **270 additions and 454 deletions** across 22 files.

---

## 🩹 Patch from v0.2.0

- `cu.figWriter` callback now receives `mime.Type` as a third parameter, allowing MIME-aware figure storage
- Affected: `cu.New`, `ConvertToParts`, `uploadFigure`, `examples/quickstart.go`

---

## ⚠️ Breaking Changes (from v0.1.0)

### Removed packages

- `storage/` — `storage.Client`, `FileSystemClient`, and Azure placeholder removed entirely. Components now handle their own I/O directly.
- `uri/` — renamed to `urio/`
- `search/es/` — Elasticsearch placeholder removed
- `runner/context.go` — `IndexingContext`, `NewICtx`, and all `IContextOpt` helpers removed

### Interface changes

| Interface | v0.1.0 | v0.2.1 |
|---|---|---|
| `Provider` | `Provide(ctx, URI) (Input, error)` | `Provide(ctx) (Input, error)` |
| `PartWriter` | `Write(ctx, URI, []Part) error` | `Write(ctx, []Part) error` |
| `SearchWriter` | `Write(ctx, URI, []SearchDoc) error` | `Write(ctx, []SearchDoc) error` |
| `Runner` | `Run(ctx, *IndexingContext)` | `Run(ctx)` |

### Struct changes

| Struct | v0.1.0 | v0.2.1 |
|---|---|---|
| `Event` | `Stage`, `ICtx`, `Duration` | `Stage`, `Duration` |

### Constructor changes

- `cu.New` — now takes `(*HTTPClient, figWriterFunc, ...CUOptions)` instead of `(storage.Client, *CUClient)`. The `figWriterFunc` now receives `mime.Type` so implementations can use the content type when writing figures.
- `file.NewFileProvider` — now takes `(urio.URI)` instead of `(storage.Client)`
- `file.NewFilePartWriter` — now takes `(urio.URI)` instead of `(storage.Client)`
- `file.NewFileSearchWriter` — now takes `(urio.URI)` instead of `(storage.Client)`
- `file.New` (cache) — now takes `(string, ...FileCacheOptions)` instead of `(storage.Client)`

---

## ✨ What's New

- **`urio` package** — renamed from `uri` with new `WriteCloser` interface for URI-aware writers
- **Self-contained components** — each component holds its URI internally; no URI passing at call time
- **`cu.FileFigWriter`** — new `urio.WriteCloser` implementation for writing figures to local files
- **`figWriter` receives `mime.Type`** — the figure writer callback now receives the content type from the API, enabling MIME-aware storage
- **`cache.FileCacheOptions`** — new `WithCacheHitCallback` option for cache hit observability

---

## 🗑️ What's Removed

- `storage.Client` interface and all implementations (`FileSystemClient`, Azure stub)
- `runner.IndexingContext` and the entire context-passing pattern
- `search/es` Elasticsearch placeholder
- URI parameters from all stage interfaces

---

## 📦 Migration Guide

```diff
- // v0.1.0
- fsClient, _ := storage.NewFileSystemClient("testdata")
- provider := fileinput.New(fsClient)
- r, _ := runner.New(...)
- ictx := runner.NewICtx(
-     runner.WithInputKey("file://cowboys.pdf"),
-     runner.WithPartWriteKey("file://parts.json"),
-     runner.WithSearchWriteKey("file://search.json"),
- )
- for event, err := range r.Run(ctx, ictx) { ... }

+ // v0.2.1
+ provider, _ := fileinput.NewFileProvider(urio.URI("file://testdata/cowboys.pdf"))
+ partWriter, _ := partfile.NewFilePartWriter(urio.URI("file://testdata/parts.json"))
+ searchWriter, _ := searchfile.NewFileSearchWriter(urio.URI("file://testdata/search.json"))
+ r, _ := runner.New(...)
+ for event, err := range r.Run(ctx) { ... }
```

---

## 📊 Stats

- **22 files changed**, 270 insertions, 454 deletions
- **5 files deleted** (`storage/azure.go`, `storage/fs.go`, `storage/storage.go`, `runner/context.go`, `search/es/es.go`)
- **1 file renamed** (`uri/uri.go` → `urio/urio.go`)
