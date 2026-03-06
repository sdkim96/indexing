package analysis

import (
	"context"
	"time"
)

// AnalysisResult - 분석 결과
type AnalysisResult struct {
	AnalyzerID     *string           `json:"analyzerId,omitempty"`
	APIVersion     *string           `json:"apiVersion,omitempty"`
	CreatedAt      *time.Time        `json:"createdAt,omitempty"`
	StringEncoding *string           `json:"stringEncoding,omitempty"`
	Contents       []AnalysisContent `json:"contents"`
}

// AnalysisContentKind - 콘텐츠 종류
type AnalysisContentKind string

const (
	AnalysisContentKindDocument    AnalysisContentKind = "document"
	AnalysisContentKindAudioVisual AnalysisContentKind = "audioVisual"
)

type OperationStatus string

const (
	StatusNotStarted OperationStatus = "NotStarted"
	StatusRunning    OperationStatus = "Running"
	StatusSucceeded  OperationStatus = "Succeeded"
	StatusFailed     OperationStatus = "Failed"
	StatusCanceled   OperationStatus = "Canceled"
)

// AnalysisContent - 분석된 콘텐츠 (공통 필드)
type AnalysisContent struct {
	Kind       AnalysisContentKind     `json:"kind"`
	MimeType   string                  `json:"mimeType"`
	AnalyzerID *string                 `json:"analyzerId,omitempty"`
	Category   *string                 `json:"category,omitempty"`
	Path       *string                 `json:"path,omitempty"`
	Markdown   *string                 `json:"markdown,omitempty"`
	Fields     map[string]ContentField `json:"fields,omitempty"`

	// DocumentContent 전용 필드 (kind == "document")
	StartPageNumber *int    `json:"startPageNumber,omitempty"`
	EndPageNumber   *int    `json:"endPageNumber,omitempty"`
	Unit            *string `json:"unit,omitempty"`

	// AudioVisualContent 전용 필드 (kind == "audioVisual")
	StartTimeMs *int `json:"startTimeMs,omitempty"`
	EndTimeMs   *int `json:"endTimeMs,omitempty"`
	Width       *int `json:"width,omitempty"`
	Height      *int `json:"height,omitempty"`
}

// ContentField - 추출된 필드
type ContentField struct {
	Type       string        `json:"type"` // "string", "date", "time", "number", "integer", "boolean", "array", "object", "json"
	Confidence *float64      `json:"confidence,omitempty"`
	Source     *string       `json:"source,omitempty"`
	Spans      []ContentSpan `json:"spans,omitempty"`

	// 타입별 값
	ValueString  *string                 `json:"valueString,omitempty"`
	ValueNumber  *float64                `json:"valueNumber,omitempty"`
	ValueInteger *int64                  `json:"valueInteger,omitempty"`
	ValueBoolean *bool                   `json:"valueBoolean,omitempty"`
	ValueDate    *string                 `json:"valueDate,omitempty"` // ISO 8601 YYYY-MM-DD
	ValueTime    *string                 `json:"valueTime,omitempty"` // ISO 8601 hh:mm:ss
	ValueArray   []ContentField          `json:"valueArray,omitempty"`
	ValueObject  map[string]ContentField `json:"valueObject,omitempty"`
	ValueJson    interface{}             `json:"valueJson,omitempty"`
}

// ContentSpan - 마크다운 내 위치
type ContentSpan struct {
	Offset int `json:"offset"`
	Length int `json:"length"`
}

// OperationStatus - 폴링용 작업 상태
type Operation struct {
	ID     string          `json:"id"`
	Status OperationStatus `json:"status"` // "NotStarted", "Running", "Succeeded", "Failed", "Canceled"
	Result *AnalysisResult `json:"result,omitempty"`
}

func (c Client) getResult(ctx context.Context, opLocation string) (Operation, error) {
	return Operation{}, nil
}
