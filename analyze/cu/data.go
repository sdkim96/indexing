package cu

type DataType string

const (
	TextDataType  DataType = "text"
	ImageDataType DataType = "image"
	TableDataType DataType = "table"
)

type Data interface {
	GetText() string
	Raw() any
	GetType() DataType
}

type TextData struct {
	Type DataType `json:"type"`
	Text string   `json:"text"`
}

func (t TextData) GetType() DataType { return TextDataType }
func (t TextData) GetText() string   { return t.Text }
func (t TextData) Raw() any          { return t.Text }

type ImageData struct {
	Type  DataType `json:"type"`
	Text  string   `json:"text"`
	Image Image    `json:"image"`
}

func (i ImageData) GetType() DataType { return ImageDataType }
func (i ImageData) GetText() string   { return i.Text }
func (i ImageData) Raw() any          { return i.Image }

type TableData struct {
	Type  DataType       `json:"type"`
	Text  string         `json:"text"`
	Table map[string]any `json:"table"`
}

func (t TableData) GetType() DataType { return TableDataType }
func (t TableData) GetText() string   { return t.Text }
func (t TableData) Raw() any          { return t.Table }

type Image struct {
	Key string `json:"key"`
}
