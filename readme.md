# indexing-repo 설계 문서

## 1. 개요

indexing-repo는 다양한 도메인의 소스 데이터를 인덱싱하는 파이프라인 라이브러리다.
파이프라인 로직과 인터페이스를 정의하고, 구현체는 외부에서 주입받는다.

**핵심 철학**
- indexing-repo는 인터페이스만 안다. 구현체는 모른다.
- 도메인이 구현체를 만들어 Config에 주입한다.
- `analyze` + `part` 는 필수. `enrich` / `search` / `cache` 는 선택.
- DDL은 domain-repo가 소유한다. indexing-repo에 DDL 없음.

---

## 2. 디렉토리 구조

```
indexing-repo/
  cmd/
    indexing/
      main.go              # 진입점. Config 조립 + Pipeline 실행

  runner/
    runner.go              # Pipeline 인터페이스 + Runner 구현체
    config.go              # Config 구조체
    wire.go                # 구현체 조립 → Runner 반환
    context.go             # IndexingContext (단계 간 데이터 전달)
    noop.go                # NoopEnricher, NoopSearchWriter

  input/
    input.go               # Input 인터페이스 (io.ReadCloser + MimeType + Meta)
    provider.go            # Provider 인터페이스 (sourceID → Input)
    file/
      file.go              # 파일 시스템 Provider 구현체
      input.go             # FileInput 구현체

  analyze/
    analyze.go             # Analyzer 인터페이스
    cu/
      cu.go                # Content Understanding 클라이언트 구현체
      http.go              # Azure AI Services HTTP 클라이언트
      models.go            # CU API 응답 구조체
      traverse.go          # sections 트리 순회 → []Part
      data.go              # Data 인터페이스 (TextData, ImageData, TableData)
      part.go              # CUPart 구현체 (패키지 내부 전용)
      cu_test.go           # 테스트

  part/
    part.go                # Part + PartWriter 인터페이스
    file/
      writer.go            # 파일 기반 PartWriter 구현체 (placeholder)

  enrich/
    enrich.go              # Enricher 인터페이스
    llm/
      llm.go               # LLM 기반 구현체 (placeholder)

  search/
    search.go              # SearchDoc + SearchWriter 인터페이스
    es/
      es.go                # Elasticsearch 구현체 (placeholder)

  cache/
    cache.go               # Cache + Cacheable 인터페이스

  internal/
    mime.go                # MIME 타입 추측 유틸리티
    blob/
      blob.go              # BlobMeta + Client 인터페이스 (blob 저장소)
      file.go              # 파일 시스템 blob Client 구현체
```

---

## 3. 인터페이스

### 3.1 필수 / 선택 구분

| 인터페이스 | 필수여부 | 역할 |
|---|---|---|
| `Provider` | 필수 | sourceID → Input 변환 |
| `Analyzer` | 필수 | Input → Part 분석 |
| `PartWriter` | 필수 | 분석된 Part 저장 |
| `Enricher` | 선택 | 검색용 가공 (없으면 noop) |
| `SearchWriter` | 선택 | 검색엔진 적재 (없으면 noop) |
| `Cache` | 선택 | 비싼 작업 결과 캐시 |

### 3.2 전체 시그니처

```go
// input/input.go
type Input interface {
    Read(p []byte) (n int, err error)
    Close() error
    MimeType() string
    Meta()     map[string]any
}

// input/provider.go
type Provider interface {
    Provide(ctx context.Context, sourceID string) (Input, error)
}

// part/part.go
type Part interface {
    MimeType() string
    Text()     string
    Raw()      any
}

type PartWriter interface {
    Write(ctx context.Context, sourceID string, parts []Part) error
}

// analyze/analyze.go
type Analyzer interface {
    Analyze(ctx context.Context, input input.Input) ([]part.Part, error)
}

// enrich/enrich.go
type Enricher interface {
    Enrich(ctx context.Context, parts []part.Part) ([]search.SearchDoc, error)
}

// search/search.go
type SearchDoc interface {
    SourceID() string
    PartIDs()  []string
    Fields()   map[string]any
}

type SearchWriter interface {
    Write(ctx context.Context, docs []SearchDoc) error
}

// cache/cache.go
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, bool, error)
    Set(ctx context.Context, key string, value []byte) error
}

type Cacheable interface {
    FingerPrint() string
}

// runner/runner.go
type Pipeline interface {
    Run(ctx context.Context, sourceID string) error
}
```

---

## 4. IndexingContext

파이프라인 실행 전체 생명주기를 들고 다니는 컨텍스트. 단계 간 데이터 전달 매개체다.

```go
// runner/context.go
type IndexingContext struct {
    // 실행 메타
    SourceID     string
    StartedAt    time.Time
    AttemptCount int    // 재시도 횟수 (로깅용)
    TraceID      string // 분산 추적용

    // 단계별로 채워짐
    Parts      []part.Part        // Analyzer 이후 존재
    SearchDocs []search.SearchDoc // Enricher 이후 존재
}
```

`Parts`와 `SearchDocs`는 모두 인터페이스 슬라이스다. Runner는 구현체를 모른다.

---

## 5. 파이프라인 단계

| 단계 | 작업 | 사용 인터페이스 | 필수 |
|---|---|---|---|
| 1단계 | 소스 제공 | `Provider` | 필수 |
| 2단계 | 분석 (OCR / Split) | `Analyzer` | 필수 |
| 3단계 | Part 저장 | `PartWriter` | 필수 |
| 4단계 | 검색용 가공 | `Enricher` | 선택 |
| 5단계 | 검색엔진 적재 | `SearchWriter` | 선택 |

---

## 6. Runner 구현

Runner는 순서만 안다. 각 단계가 뭘 하는지 모른다.
`iter.Seq2[Event, error]`를 반환하여 호출자가 단계별 결과를 스트리밍으로 받는다.

```go
// runner/runner.go
type Event struct {
    Stage    string           // "provide" | "analyze" | "part" | "enrich" | "search"
    ICtx     *IndexingContext
    Duration time.Duration
}

func (r *Runner) Run(ctx context.Context, ictx *IndexingContext) iter.Seq2[Event, error] {
    return func(yield func(Event, error) bool) {

        // 1단계: 소스 제공
        start := time.Now()
        inp, err := r.provider.Provide(ctx, ictx.SourceID)
        if !yield(Event{"provide", ictx, time.Since(start)}, err) { return }

        // 2단계: 분석한다
        start = time.Now()
        parts, err := r.analyzer.Analyze(ctx, inp)
        if !yield(Event{"analyze", ictx, time.Since(start)}, err) { return }
        ictx.Parts = parts

        // 3단계: 저장한다
        start = time.Now()
        err = r.partWriter.Write(ctx, ictx.SourceID, ictx.Parts)
        if !yield(Event{"part", ictx, time.Since(start)}, err) { return }

        // 4단계: 가공한다
        start = time.Now()
        docs, err := r.enricher.Enrich(ctx, ictx.Parts)
        if !yield(Event{"enrich", ictx, time.Since(start)}, err) { return }
        ictx.SearchDocs = docs

        // 5단계: 적재한다
        start = time.Now()
        err = r.searchWriter.Write(ctx, ictx.SearchDocs)
        if !yield(Event{"search", ictx, time.Since(start)}, err) { return }
    }
}
```

호출하는 쪽:

```go
ictx := &runner.IndexingContext{
    SourceID:  sourceID,
    StartedAt: time.Now(),
    TraceID:   traceID,
}

for event, err := range r.Run(ctx, ictx) {
    if err != nil {
        log.Printf("failed at %s: %v", event.Stage, err)
        return err
    }
    log.Printf("stage %s done in %s", event.Stage, event.Duration)
}
```

에러가 나도 어느 단계인지 바로 알 수 있고, 중단 여부는 호출자가 결정한다.

---

## 7. Idempotent 조건

Runner는 idempotent를 보장하지 않는다. **"idempotent할 수 있는 구조를 제공"**할 뿐이다.
실제 보장은 각 구현체의 책임이다.

### 구현체가 지켜야 할 조건

| 구현체 | 조건 |
|---|---|
| `Analyzer` + `Cache` | 같은 sourceID + 파라미터면 CU/LLM 호출 스킵 |
| `PartWriter` | 중복 insert 무시 (ON CONFLICT DO NOTHING 또는 덮어쓰기) |
| `Enricher` + `Cache` | 같은 parts 입력이면 LLM 호출 스킵 |
| `SearchWriter` | upsert. 중복 적재 안됨 |

### 재실행 흐름

```
AttemptCount 증가 후 Run() 재호출

1단계 provide   → Provider.Provide() 재실행
2단계 analyze   → Cache.Get() hit → CU 호출 스킵
3단계 part      → ON CONFLICT DO NOTHING → 중복 저장 안됨
4단계 enrich    → Cache.Get() hit → LLM 호출 스킵
5단계 search    → upsert → 중복 적재 안됨
```

구현체가 조건을 지키면 처음부터 재실행해도 비용 없이 idempotent가 보장된다.

---

## 8. 조립 (wire.go)

domain-repo는 Config만 채워서 넘기면 된다. 구현체 조립은 wire.go가 전담한다.

```go
// runner/wire.go
func New(cfg Config) (*Runner, error) {
    // 필수 검증
    if cfg.Provider == nil {
        return nil, errors.New("provider is required")
    }
    if cfg.Analyzer == nil {
        return nil, errors.New("analyzer is required")
    }
    if cfg.PartWriter == nil {
        return nil, errors.New("part writer is required")
    }

    // 선택 → noop 폴백
    enricher := cfg.Enricher
    if enricher == nil { enricher = &NoopEnricher{} }

    searchWriter := cfg.SearchWriter
    if searchWriter == nil { searchWriter = &NoopSearchWriter{} }

    return &Runner{
        provider:     cfg.Provider,
        analyzer:     cfg.Analyzer,
        partWriter:   cfg.PartWriter,
        enricher:     enricher,
        searchWriter: searchWriter,
        cache:        cfg.CacheWriter,
    }, nil
}

// 기본 조립 예시 (cmd/indexing/main.go)
r, _ := runner.New(runner.Config{
    Provider:     fileinput.FileProvider{Root: "./data"},
    Analyzer:     cu.New(blobClient, httpClient),
    PartWriter:   partfile.New("./output"),
    Enricher:     llm.New(cfg.LLM),
    SearchWriter: es.New(cfg.ES),
    CacheWriter:  cachefile.New("./cache"),
})
```

---

## 9. 의존 방향

```
runner/   →  input, analyze, part, enrich, search, cache
input/    →  (독립)
analyze/  →  input, part  (Input, Part 인터페이스 참조)
cu/       →  input, part, cache, internal/blob  (Analyzer 구현체)
enrich/   →  part, search  (Part, SearchDoc 인터페이스 참조)
search/   →  (독립)
cache/    →  (독립)
internal/ →  (독립. blob은 internal 참조)

domain-repo   →  indexing-repo (인터페이스 + 타입 참조)
indexing-repo →  domain-repo   완전히 모름
```

**핵심 규칙**
- indexing-repo는 domain-repo를 import하지 않는다.
- `CUPart` 구현체는 `analyze/cu` 패키지 내부에서만 사용된다.
- DDL이 없다. Part 저장 방식은 PartWriter 구현체가 결정한다.
- 기본 구현체(`file/`)로 DB 없이 즉시 파이프라인을 실행할 수 있다.

---

## 10. 도메인별 확장

각 도메인은 필요한 인터페이스만 구현하여 주입한다.

| 도메인 | PartWriter | Enricher | SearchWriter | Cache |
|---|---|---|---|---|
| File | RDB | LLM (요약+키워드) | ES | file |
| Conversation | RDB | — noop | Graph DB | — noop |
| Memory | RDB | Vector 생성 | Vector DB | file |
