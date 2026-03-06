# 검색 시스템 인덱싱 설계

## 대원칙

```
RDB  = Source of Truth (원본, 권한, 상태)
KV   = 파생상품 캐시 (비싼 외부 API 호출 결과)
ES   = 검색 캐시 (벡터 + BM25 인덱스)

진실은 RDB에만. 나머지는 전부 파생상품.
ES는 언제든 날려도 됨. KV에서 복구. KV도 날리면 RDB에서 재생성.
```

---

## 파일시스템 철학 (리눅스/유닉스)

파일 시스템 조작은 리눅스 철학을 따른다.

```
source.id  = inode  (파일의 실제 identity, 불변)
source.key = 경로   (포인터, 변경 가능)

mv = UPDATE source SET key = :new_key  (id는 그대로)
cp = 새 Source INSERT                  (새 id)
rm = deleted_at = NOW()
```

key에 UNIQUE 제약을 걸고 UPDATE로 처리하면
PostgreSQL row-level lock이 리눅스 `rename()` syscall의 atomic 동작을 보장한다.

```
파일 이동/소유권 이전 = key UPDATE
  → source.id 유지
  → SourceIndex 유지
  → Part, ES 전부 그대로
  → 재인덱싱 불필요

리버저닝 (내용 변경) = deleted_at + 새 Source INSERT
  → 새 source.id (새 파일)
  → 재인덱싱 필요

삭제 = deleted_at = NOW()
```

---

## 용어 정의

```
재시도      파이프라인 실패 후 같은 source_id로 다시 실행
재인덱싱    내용이 바뀐 새 문서. 새 source_id 발급. 사실상 새 파일 생성.
```

---

## 테이블 (4개)

### Source (파일)

```sql
Source
  id          PK
  key         VARCHAR UNIQUE  -- /users/{user_id}/{path}/{filename}
  raw_key     VARCHAR         -- S3 원본 경로. nullable.
  created_at  TIMESTAMP
  deleted_at  TIMESTAMP       -- null: 존재함, 값: 삭제됨
```

Source는 파일의 존재만 표현한다. 인덱싱 상태를 모른다.

#### 상태 표현

```
deleted_at IS NULL  → 존재함
deleted_at NOT NULL → 삭제됨
```

#### 설계 원칙

- key UNIQUE: 같은 경로에 파일이 두 개 존재할 수 없음 (리눅스와 동일)
- 폴더 테이블 없음 (key의 path가 폴더)
- owner 컬럼 없음 (key의 namespace가 owner)
- 원본 파일은 S3에 저장 (raw_key)

#### key 네임스페이스

```
/users/{user_id}/   → 개인 파일 (owner = user_id)
/root/              → 시스템 파일
/public/            → 전체 공개
```

### SourceIndex (인덱싱 완료 audit)

```sql
SourceIndex
  id           PK
  source_id    FK → Source.id  UNIQUE  -- Source당 1개 보장
  indexed_at   TIMESTAMP
  indexed_by   VARCHAR    -- 파이프라인 버전
  extractor    VARCHAR    -- 사용한 extractor
  llm_model    VARCHAR    -- 사용한 LLM
  created_at   TIMESTAMP
```

- 런타임에 사용하지 않는다. audit 전용.
- ES까지 완료된 시점에 INSERT.
- source_id UNIQUE: Source당 1개. 재인덱싱은 새 source_id이므로 중복 없음.
- source_id에 인덱스 필수 (EXISTS 서브쿼리 성능).

#### 검색 가능 조건

```sql
WHERE deleted_at IS NULL
AND EXISTS (
  SELECT 1 FROM source_index
  WHERE source_id = source.id
)
```

### Part (문서 조각)

```sql
Part
  id          PK
  source_id   FK → Source.id
  page        INT         -- 페이지 번호
  offset      INT         -- 페이지 내 순서
  data        JSONB       -- Part 내용 전체
  created_at  TIMESTAMP
```

- (page, offset) 튜플이 문서 내 위치를 완전히 표현
- 전역 순서가 필요하면 런타임에 `ORDER BY page, offset`으로 계산
- RDB에서 Part를 검색하지 않음 (source_id + page + offset으로 꺼내기만)
- 검색은 ES가 함. RDB는 저장과 조회만.
- deleted_at 없음 (Source 수명을 따라감)
- 전 필드 immutable (동기화 문제 원천 차단)

#### data JSONB 구조

모든 Part에 `type`과 `text`가 공통으로 존재한다.
`text`는 해당 Part의 텍스트 표현이며, summary 생성 시 type 분기 없이 `data.text`만 읽으면 된다.
`text`는 같은 입력에서 항상 같은 결과가 나와야 한다 (deterministic).
미디어 Part는 type명과 동일한 키에 원본 정보를 담는다.
위치 정보 (page, offset)는 컬럼으로 관리하므로 data에 중복 저장하지 않는다.

```json
{ "type": "text", "text": "1분기 매출은 100억으로 전년 대비 15% 증가했다." }
```

```json
{
  "type": "image",
  "text": "매출 추이 그래프. 1분기 100억, 2분기 120억. 우상향 추세.",
  "image": { "key": "https://myaccount.blob.core.windows.net/parts/img_001.png", "mime_type": "image/png", "size": 204800 }
}
```

```json
{
  "type": "table",
  "text": "| Name | Corp | Remark |\n|------|------|--------|\n| Foo  |      |        |\n| Bar  | Microsoft | Dummy |",
  "table": {
    "data": {
      "rowCount": 3,
      "columnCount": 3,
      "cells": [{"kind": "columnHeader", "rowIndex": 0, "columnIndex": 0, "content": "Name"}, "..."],
      "caption": "Table 1: This is a dummy table"
    }
  }
}
```

#### 규칙

```
text       항상 존재. 모든 Part의 텍스트 표현. deterministic.
{type}     미디어/구조화 데이터면 존재 (image, table). 원본 정보.

summary 만들 때 → data.text만 읽으면 됨. type 분기 불필요.
렌더링할 때     → data.image.key로 Blob 접근 / data.table.data로 구조 접근.
LLM에 전달할 때 → data를 그대로 전달. 중간 가공 불필요.
```

타입별 저장 방식:

```
text   → text만. 추가 키 없음.
image  → text(caption) + image.key(Blob URL) + image.mime_type + image.size
table  → text(markdown) + table.data(CU 원본 JSON)
```

#### Page boundary invariant (I1)

```
Part는 절대 페이지를 넘지 않는다.
cross-page 문맥은 grouping이 담당한다.
```

### source_read_permission (ACL)

```sql
source_read_permission
  source_id   FK → Source.id  ┐
  user_id     INT             ┘ 복합 PK
```

- owner는 key로 판단 → ACL row 불필요
- 공유할 때만 row 추가
- 소유권 이전 시 ACL도 함께 정리 필요

---

## Content Understanding (Azure)

OCR, 레이아웃 분석, 요소 분류를 단일 API 호출로 처리한다.

```
원본 파일
  ↓
Azure Content Understanding (prebuilt-layout)
  ↓
sections 트리 (paragraph / table / figure 순서 보존)
  ↓
list[Part]
```

split 정책, Extractor 인터페이스, 별도 chunker가 불필요하다.
Content Understanding이 이미 semantic unit으로 쪼개서 순서까지 정해서 준다.

### 응답 구조 핵심

```
sections[].elements  → 문서 내 순서 (트리 구조, 재귀 순회)
paragraphs[]         → 텍스트 단위. role로 필터링.
tables[]             → markdown에 <table>로 포함됨.
figures[]            → id로 바이너리 별도 호출 (무료).
```

### paragraph role 처리

```
SKIP_ROLES = { pageHeader, pageFooter, pageNumber }
→ Part로 만들지 않음

sectionHeading, title, 일반 paragraph
→ Part (type=text)
```

### sections 순회 → list[Part]

```python
SKIP_ROLES = {"pageHeader", "pageFooter", "pageNumber"}

def traverse(element_ref, response) → list[Part]:
    if "/paragraphs/" in element_ref:
        p = resolve(element_ref)
        if p.role in SKIP_ROLES: return []
        return [Part(type="text", text=p.content)]

    if "/tables/" in element_ref:
        t = resolve(element_ref)
        return [Part(type="text", text=markdown_of(t))]  # markdown에 포함됨

    if "/figures/" in element_ref:
        f = resolve(element_ref)
        bytes = GET /figures/{f.id}          # 무료
        s3_key = S3.upload(bytes)
        caption = f.caption.content if f.caption else ""
        return [Part(type="image", text=caption, image={ key: s3_key })]

    if "/sections/" in element_ref:
        s = resolve(element_ref)
        return flatten([traverse(e) for e in s.elements])
```

### KV 캐시

```
key:   "raw:{sha256(raw_file)}:{sha256(cu_params)}"
value: Content Understanding 응답 JSON 전체 (figure 바이너리 제외)
```

---

## Summary 전략

```
Part (RDB) → Summary (KV 캐시 + ES)
원본          파생상품
```

---

## KV 캐시 스키마

### 캐시의 목적

```
캐시의 목적은 비싼 외부 API 호출을 빠르게 해결하는 것이다.

캐시 대상:   Extractor 호출, LLM 호출 (과금)
캐시 비대상: split, 그룹핑 (내부 연산)
```

### 캐시의 보장 범위

```
✓ 파이프라인 재시도 → idempotent 보장
✗ 재인덱싱 (새 source_id) → idempotent 보장 안 함
```

### Content Understanding 캐시 (2단계 내)

```
key:   "raw:{sha256(raw_file)}:{sha256(cu_params)}"
value: Content Understanding 응답 JSON (figure 바이너리 제외)
```

- `sha256(raw_file)`: 파일 내용 기반. mv해도 캐시 유지.
- `sha256(cu_params)`: analyzer 버전, 옵션 등 모든 파라미터를 hash로 흡수.

### LLM 캐시 (4단계 내)

```
key:   "llm:{sha256(prompt + params)}"
value: LLM 응답
```

- 그룹핑이든 Summary 생성이든 동일한 패턴으로 캐시.
- 재인덱싱 시에도 안 바뀐 그룹은 히트.

---

## ES 구조

```json
{
  "summary_id": "hash_value",
  "source_id": 10,
  "part_ids": [1, 2, 3],
  "summary_vector": [...],
  "summary": "2024년 1분기 매출 분석과 추이 그래프",
  "keywords": ["매출", "1분기", "그래프"]
}
```

```
summary_vector   의미 기반 검색
keywords         정확한 키워드 매칭 부스팅
summary          BM25 키워드 검색
```

---

## 인덱싱 파이프라인

### 설계 원칙

```
Source      = 파일의 존재. 인덱싱을 모름.
SourceIndex = 인덱싱 완료 audit. 런타임에 안 씀.
Part        = 인덱싱의 원자 단위 (immutable)
KV          = 비싼 외부 API 호출 결과 캐시
```

파이프라인 재시도 = 같은 source_id 재사용. KV 캐시 히트.
재인덱싱 = 새 source_id. 새 파일 생성과 동일.

### 전체 흐름

```
[1단계] Source 생성
[2단계] Part 생성  (Content Understanding → sections 순회)
[3단계] Summary + Keyword 생성 (LLM)
[4단계] ES 적재 → SourceIndex INSERT
```

---

### 1단계: Source 생성

```
S3 업로드 → Source INSERT
```

재시도면 기존 Source 재사용 (같은 source_id).

---

### 2단계: Part 생성

**invariant**: Part는 페이지를 넘지 않는다 (I1).

```
CU 캐시 확인: "raw:{sha256(raw_file)}:{sha256(cu_params)}"
  히트 → KV에서 응답 JSON 꺼냄
  미스 → Content Understanding(prebuilt-layout) 호출 → KV 저장

→ sections 트리 재귀 순회:
    paragraph  → Part INSERT (type=text)    [SKIP_ROLES 제외]
    table      → Part INSERT (type=text)    [markdown 그대로]
    figure     → GET /figures/{id} → S3 → Part INSERT (type=image)
    section    → 재귀

Part INSERT (ON CONFLICT DO NOTHING)
```

---

### 3단계: Summary + Keyword 생성

```
LLM 캐시 확인: "llm:{sha256(prompt + params)}"
  히트 → 스킵
  미스 → 각 그룹의 data.text → LLM 호출 → { summary, keywords } → KV 저장
```

---

### 4단계: ES 적재

```
list[{ summary, keywords }] → ES INSERT
→ SourceIndex INSERT (indexed_at, extractor, llm_model 등)
```

SourceIndex INSERT가 파이프라인 완료의 유일한 신호.

---

### 재시도

```sql
-- 재시도 대상 (SourceIndex 없는 Source)
SELECT id FROM source
WHERE deleted_at IS NULL
AND NOT EXISTS (
  SELECT 1 FROM source_index WHERE source_id = source.id
)
```

```
같은 source_id로 재실행
Extractor 캐시 히트 → 재호출 없음
LLM 캐시 히트       → 재호출 없음
Part ON CONFLICT    → 중복 INSERT 스킵
```

---

## 검색 + 생성 (RAG) 흐름

### Searchable IDs

```sql
-- 본인 소유
SELECT id FROM source
WHERE key LIKE '/users/111/%'
AND deleted_at IS NULL
AND EXISTS (SELECT 1 FROM source_index WHERE source_id = source.id)

UNION

-- 공유받은 파일
SELECT source_id FROM source_read_permission
WHERE user_id = 111

UNION

-- 전체 공개
SELECT id FROM source
WHERE key LIKE '/public/%'
AND deleted_at IS NULL
AND EXISTS (SELECT 1 FROM source_index WHERE source_id = source.id)
```

### 전체 흐름

```
① Searchable IDs   RDB에서 권한 기반 source_id 목록 조회.
② Retrieval        ES에서 source_id 필터 + 하이브리드 검색. part_ids 반환.
③ Fetch            RDB에서 part_ids에 해당하는 Part.data 조회.
④ Generation       Part.data를 그대로 LLM에 전달하여 답변 생성.
```

---

## CRUD 전략

### Create

```
1단계  S3 업로드 → Source INSERT                                       [트랜잭션]

2단계  CU 캐시 확인: "raw:{sha256(raw_file)}:{sha256(cu_params)}"
         히트 → KV에서 응답 JSON 꺼냄
         미스 → Content Understanding 호출 → KV 저장                  [KV]
       sections 트리 순회 → Part INSERT (ON CONFLICT DO NOTHING)
       figure는 GET /figures/{id} → S3 업로드                          [트랜잭션/요소]

3단계  LLM 캐시 확인: "llm:{sha256(prompt + params)}"
         히트 → 스킵
         미스 → LLM 호출 → { summary, keywords } → KV 저장            [KV]

4단계  ES 색인 → SourceIndex INSERT                                    [ES + RDB 트랜잭션]
```

### Move / 소유권 이전 (mv)

```
UPDATE source SET key = :new_key WHERE id = :id
```

- source.id 유지 → SourceIndex 유지 → 재인덱싱 불필요
- Part, ES 전부 그대로
- key UNIQUE + PostgreSQL row-level lock → atomic (리눅스 rename() 동일)
- 소유권 이전 시 source_read_permission도 함께 정리

### Update / 재인덱싱 (cp + rm)

```
1. 기존 Source: deleted_at = NOW()
2. 새 Source INSERT (같은 key, 새 id)
3. 파이프라인 실행 (새 source_id)
4. 기존 ES 데이터는 배치가 나중에 삭제
```

내용이 바뀐 새 문서. 재시도와 명확히 구별된다.

### Delete (rm)

```
1. Source: deleted_at = NOW()
2. 배치가 주기적으로 ES에서 해당 source_id 삭제
```

---

## 동기화 전략

```
RDB → KV → ES   단방향만 존재
ES → RDB         절대 없음
KV → RDB         절대 없음
```

### ES 복구 (ES 날아갔을 때)

```
1. RDB에서 Part 전부 읽음
2. LLM 캐시 키 계산 → KV에서 summary 꺼냄 (LLM 호출 0번)
3. ES 재색인
```

### KV도 날아갔을 때

```
1. RDB에서 Part.data.text 읽음
2. LLM 호출 → summary 재생성 → KV 저장
3. ES 재색인
```

---

## 최소 상태 원칙

```
Source         deleted_at (파일 존재 여부만)
SourceIndex    audit 전용. 런타임에 안 씀. EXISTS로 인덱싱 완료 판별.
Part           위치는 (page, offset). data는 immutable.
ACL            있으면 row, 없으면 없는 거.
ES             상태 컬럼 없음.
```

없는 것이 설계:

- 폴더 테이블 없음
- owner 컬럼 없음
- status enum 없음
- indexed_at 컬럼 없음 (SourceIndex EXISTS로 대체)
- version 컬럼 없음
- Chunk 테이블 없음
- Summary 테이블 없음
- IndexingJob 테이블 없음
- ES → RDB 동기화 없음
- 전역 order 컬럼 없음

---

## 인덱싱 불변 원칙 (Invariants)

```
I1. Page boundary는 ingestion invariant
    Part는 절대 페이지를 넘지 않는다. cross-page 문맥은 grouping이 담당.

I2. Part 위치는 (page, offset) 튜플로 표현
    전역 순서는 런타임에 ORDER BY page, offset으로 계산.

I3. Content Understanding 캐시 anchor는 파일 내용 (sha256(raw_file))
    파일 이동/소유권 이전에도 캐시가 유지된다.

I4. Part.data.text는 항상 존재하고 deterministic
    같은 입력에서 같은 text가 나와야 캐시가 안정적.

I5. KV key space는 "파라미터 + 입력"만으로 결정
    외부 상태에 의존하지 않는다.

I6. KV는 파이프라인 재시도에 대한 idempotent에 집중한다
    재인덱싱(새 source_id)에 대한 idempotent는 보장하지 않는다.

I7. 캐시 대상은 비싼 외부 API 호출만이다
    Extractor (과금), LLM (과금). 내부 연산은 캐시하지 않는다.

I8. SourceIndex INSERT가 파이프라인 완료의 유일한 신호
    ES까지 완료된 시점에만 INSERT. 런타임은 EXISTS로 판별.

I9. 파일시스템 조작은 리눅스 철학을 따른다
    source.id = inode (불변)
    source.key = 경로 (변경 가능, UNIQUE)
    mv = key UPDATE / cp+rm = 재인덱싱 / rm = deleted_at
```

---

## 아키텍처 요약

```
RDB (Source + SourceIndex + Part + ACL)
  Source:      파일 존재 (deleted_at만)
  SourceIndex: 인덱싱 완료 audit (런타임 미사용)

KV 캐시 (비싼 외부 API 호출 결과만)
  "raw:{sha256(raw_file)}:{sha256(cu_params)}"   → Content Understanding 응답 JSON
  "llm:{sha256(prompt + params)}"                → LLM 응답

ES (summary_vector + keywords + BM25)

검색 가능 여부  → deleted_at IS NULL AND EXISTS (source_index)
검색 수행       → ES 하이브리드 검색 (summary 기반)
원본 조회       → RDB Part.data (ORDER BY page, offset)
미디어 렌더링   → S3 (data.image.key, data.table.key)
LLM 생성       → Part.data를 그대로 전달

파일 이동/소유권 → key UPDATE (id 불변, 재인덱싱 없음)
파이프라인 재시도 → 같은 source_id, KV 캐시 히트
재인덱싱         → 새 source_id, 이전 source deleted_at
```