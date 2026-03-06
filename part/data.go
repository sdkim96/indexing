package part

type DataType string

const (
	TextDataType  DataType = "text"
	ImageDataType DataType = "image"
	TableDataType DataType = "table"
)

// Data is an interface that represents the data contained in a Part. It has a method GetType() that returns the type of data.
type Data interface {
	GetType() DataType
}

// TextData is a struct that represents text data in a Part. It includes the type of data and the text content.
type TextData struct {
	Type DataType `json:"type"`
	Text string   `json:"text"`
}

func (t TextData) GetType() DataType {
	return TextDataType
}

// ImageData is a struct that represents image data in a Part. It includes the type of data, a text description, and the image metadata.
type ImageData struct {
	Type  DataType `json:"type"`
	Text  string   `json:"text"`
	Image Image    `json:"image"`
}

func (i ImageData) GetType() DataType {
	return ImageDataType
}

// TableData is a struct that represents table data in a Part. It includes the type of data, a text description, and the table content as a map.
type TableData struct {
	Type  DataType       `json:"type"`
	Text  string         `json:"text"`
	Table map[string]any `json:"table"`
}

func (t TableData) GetType() DataType {
	return TableDataType
}

// Image is a struct that represents the metadata of an image, including its key, MIME type, and size in bytes.
type Image struct {
	Key      string `json:"key"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}
