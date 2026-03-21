# Release v0.3.0

## Summary

v0.3.0 adds a `sourceID` parameter across all pipeline interfaces and refactors the `cu` package for cleaner file organization. The `sourceID` lets each stage identify the source being processed, useful for logging, storage keying, and figure naming. The `google/uuid` dependency is removed.

---

## Breaking Changes

### Interface changes

All pipeline interfaces now receive a `sourceID string` parameter:

| Interface | v0.2.1 | v0.3.0 |
|---|---|---|
| `Provider` | `Provide(ctx)` | `Provide(ctx, sourceID)` |
| `Analyzer` | `Analyze(ctx, input, cache)` | `Analyze(ctx, sourceID, input, cache)` |
| `PartWriter` | `Write(ctx, parts)` | `Write(ctx, sourceID, parts)` |
| `Enricher` | `Enrich(ctx, parts, cache)` | `Enrich(ctx, sourceID, parts, cache)` |
| `SearchWriter` | `Write(ctx, docs)` | `Write(ctx, sourceID, docs)` |
| `Runner` | `Run(ctx)` | `Run(ctx, sourceID)` |

### `cu.FileFigWriter` moved

`cu.NewFileFigWriter` moved to sub-package `cu/filefigure`:

```diff
- import cu "github.com/sdkim96/indexing/analyze/cu"
- cu.NewFileFigWriter(uri)

+ import cuFileFigure "github.com/sdkim96/indexing/analyze/cu/filefigure"
+ cuFileFigure.NewFileFigWriter(uri)
```

### `cu.FigureWriter` type alias

The anonymous `func(ctx, name, mimeType) (WriteCloser, error)` is now a named type `cu.FigureWriter`.

---

## What's New

- **`sourceID` across the pipeline** — every stage receives a source identifier, replacing the previous `uuid`-based figure naming with caller-controlled IDs
- **`cu.FigureWriter` type** — named function type for figure writer callbacks
- **`cu` package reorganization**:
  - `data.go` merged into `part.go` (Data types live with CUPart)
  - `FigureRequest` moved from `models.go` to `http.go`
  - `FileFigWriter` extracted to `cu/filefigure` sub-package
  - `traverse.go` replaced by `convert.go`, `figure.go`, `table.go`

---

## Removed

- `github.com/google/uuid` dependency — figure paths now use the caller-provided `sourceID` instead of generated UUIDs
- `cu/traverse.go` — split into `convert.go`, `figure.go`, `table.go`

---

## Stats

- **22 files changed**, 100 insertions, 473 deletions
- **2 files deleted** (`cu/data.go`, `cu/traverse.go`)
- **4 files added** (`cu/convert.go`, `cu/figure.go`, `cu/table.go`, `cu/filefigure/writer.go`)
