package analysis

import (
	"time"
)

// AnalysisResult holds the top-level result returned by the analyzer.
type AnalysisResult struct {
	AnalyzerID     *string           `json:"analyzerId,omitempty"`
	APIVersion     *string           `json:"apiVersion,omitempty"`
	CreatedAt      *time.Time        `json:"createdAt,omitempty"`
	StringEncoding *string           `json:"stringEncoding,omitempty"`
	Contents       []AnalysisContent `json:"contents"`
}

// AnalysisContentKind indicates the type of analyzed content.
type AnalysisContentKind string

const (
	// AnalysisContentKindDocument represents a document (e.g. PDF, image).
	AnalysisContentKindDocument AnalysisContentKind = "document"
	// AnalysisContentKindAudioVisual represents audio/visual content.
	AnalysisContentKindAudioVisual AnalysisContentKind = "audioVisual"
)

// OperationStatus represents the lifecycle state of an async operation.
type OperationStatus string

const (
	StatusNotStarted OperationStatus = "NotStarted"
	StatusRunning    OperationStatus = "Running"
	StatusSucceeded  OperationStatus = "Succeeded"
	StatusFailed     OperationStatus = "Failed"
	StatusCanceled   OperationStatus = "Canceled"
)

// Span represents a text range within the markdown output via offset and length.
type Span struct {
	Offset int `json:"offset"`
	Length int `json:"length"`
}

// Word represents a single recognized word on a page.
type Word struct {
	Content    string  `json:"content"`
	Span       Span    `json:"span"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"` // Bounding polygon in "D(page, x1,y1,...)" format
}

// Line represents a line of text recognized on a page.
type Line struct {
	Content string `json:"content"`
	Source  string `json:"source"`
	Span    Span   `json:"span"`
}

// Page represents a single page within a document, including its words and lines.
type Page struct {
	PageNumber int     `json:"pageNumber"`
	Angle      float64 `json:"angle"`  // Rotation angle in degrees
	Width      float64 `json:"width"`  // Page width in the unit specified by AnalysisContent.Unit
	Height     float64 `json:"height"` // Page height in the unit specified by AnalysisContent.Unit
	Spans      []Span  `json:"spans"`
	Words      []Word  `json:"words"`
	Lines      []Line  `json:"lines"`
}

// Paragraph represents a logical paragraph extracted from the document.
// Role is only set for special paragraphs such as titles, headings, headers, footers, and page numbers.
type Paragraph struct {
	Role    string `json:"role,omitempty"` // "title" | "sectionHeading" | "pageHeader" | "pageFooter" | "pageNumber"
	Content string `json:"content"`
	Source  string `json:"source"`
	Span    Span   `json:"span"`
}

// Section represents a logical section of the document.
// Elements contains JSON pointer-style references such as "/paragraphs/0", "/tables/1", "/figures/2".
type Section struct {
	Span     Span     `json:"span"`
	Elements []string `json:"elements"` // e.g. ["/paragraphs/0", "/tables/1", "/figures/2"]
}

// TableCell represents a single cell within a table.
type TableCell struct {
	Kind        string   `json:"kind"` // "columnHeader" | "content"
	RowIndex    int      `json:"rowIndex"`
	ColumnIndex int      `json:"columnIndex"`
	RowSpan     int      `json:"rowSpan"`
	ColumnSpan  int      `json:"columnSpan"`
	Content     string   `json:"content"`
	Source      string   `json:"source"`
	Span        Span     `json:"span"`
	Elements    []string `json:"elements,omitempty"` // References to paragraphs inside this cell
}

// TableCaption represents the caption associated with a table.
type TableCaption struct {
	Content  string   `json:"content"`
	Source   string   `json:"source"`
	Span     Span     `json:"span"`
	Elements []string `json:"elements,omitempty"`
}

// Table represents a table extracted from the document.
type Table struct {
	RowCount    int           `json:"rowCount"`
	ColumnCount int           `json:"columnCount"`
	Cells       []TableCell   `json:"cells"`
	Source      string        `json:"source"`
	Span        Span          `json:"span"`
	Caption     *TableCaption `json:"caption,omitempty"`
}

// FigureCaption represents the caption associated with a figure.
type FigureCaption struct {
	Content  string   `json:"content"`
	Source   string   `json:"source"`
	Span     Span     `json:"span"`
	Elements []string `json:"elements,omitempty"`
}

// Figure represents an image or chart extracted from the document.
type Figure struct {
	ID       string         `json:"id,omitempty"` // Referenced in markdown as "figures/<id>"
	Source   string         `json:"source"`
	Span     Span           `json:"span"`
	Elements []string       `json:"elements,omitempty"` // Paragraphs inside the figure (e.g. axis labels)
	Caption  *FigureCaption `json:"caption,omitempty"`
}

// AnalysisContent holds the analyzed result for a single input document or audio/visual segment.
type AnalysisContent struct {
	Kind       AnalysisContentKind     `json:"kind"`
	MimeType   string                  `json:"mimeType"`
	AnalyzerID *string                 `json:"analyzerId,omitempty"`
	Category   *string                 `json:"category,omitempty"`
	Path       *string                 `json:"path,omitempty"`
	Markdown   *string                 `json:"markdown,omitempty"` // Full document rendered as Markdown
	Fields     map[string]ContentField `json:"fields,omitempty"`   // Named fields extracted by a custom analyzer

	// Document-specific fields (kind == "document")
	StartPageNumber *int    `json:"startPageNumber,omitempty"`
	EndPageNumber   *int    `json:"endPageNumber,omitempty"`
	Unit            *string `json:"unit,omitempty"` // Measurement unit for coordinates, e.g. "inch" or "pixel"

	// Layout analysis fields (populated by prebuilt-layout and similar analyzers)
	Pages      []Page      `json:"pages,omitempty"`
	Paragraphs []Paragraph `json:"paragraphs,omitempty"`
	Sections   []Section   `json:"sections,omitempty"`
	Tables     []Table     `json:"tables,omitempty"`
	Figures    []Figure    `json:"figures,omitempty"`

	// Audio/visual-specific fields (kind == "audioVisual")
	StartTimeMs *int `json:"startTimeMs,omitempty"`
	EndTimeMs   *int `json:"endTimeMs,omitempty"`
	Width       *int `json:"width,omitempty"`
	Height      *int `json:"height,omitempty"`
}

// ContentField represents a single named field extracted by the analyzer.
type ContentField struct {
	Type       string        `json:"type"`                 // "string" | "date" | "time" | "number" | "integer" | "boolean" | "array" | "object" | "json"
	Confidence *float64      `json:"confidence,omitempty"` // Extraction confidence score (0.0 – 1.0)
	Source     *string       `json:"source,omitempty"`     // Bounding polygon reference
	Spans      []ContentSpan `json:"spans,omitempty"`      // Positions within the markdown output

	// Typed value — only one of these is populated based on Type
	ValueString  *string                 `json:"valueString,omitempty"`
	ValueNumber  *float64                `json:"valueNumber,omitempty"`
	ValueInteger *int64                  `json:"valueInteger,omitempty"`
	ValueBoolean *bool                   `json:"valueBoolean,omitempty"`
	ValueDate    *string                 `json:"valueDate,omitempty"` // ISO 8601: YYYY-MM-DD
	ValueTime    *string                 `json:"valueTime,omitempty"` // ISO 8601: hh:mm:ss
	ValueArray   []ContentField          `json:"valueArray,omitempty"`
	ValueObject  map[string]ContentField `json:"valueObject,omitempty"`
	ValueJson    interface{}             `json:"valueJson,omitempty"`
}

// ContentSpan represents the position of a field value within the markdown output.
// Kept separate from Span to preserve backward compatibility with ContentField.
type ContentSpan struct {
	Offset int `json:"offset"`
	Length int `json:"length"`
}

// Operation represents an asynchronous analysis job returned by the service.
type Operation struct {
	ID     string          `json:"id"`
	Status OperationStatus `json:"status"`
	Result *AnalysisResult `json:"result,omitempty"`
}

type FigureRequest struct {
	OpID       string
	ContentIdx int
	FigureID   string
	S3Key      string
}
