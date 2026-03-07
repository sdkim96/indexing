package part

import "github.com/sdkim96/indexing/cache"

type Part struct {
	ID        string `json:"id"`
	FileID    string `json:"source_id"`
	Data      Data   `json:"data"`
	Page      int    `json:"page"`
	Offset    int    `json:"offset"`
	CreatedAt int64  `json:"created_at"`
}

func NewTextPart(fileID string, page, offset int, text string) Part {
	return Part{
		ID: "part-text-" + cache.Sha256(
			[]byte(text),
		),
		FileID: fileID,
		Data:   TextData{Type: TextDataType, Text: text},
		Page:   page,
		Offset: offset,
	}
}
func NewImagePart(fileID string, page, offset int, text string, img Image) Part {
	return Part{
		ID: "part-image-" + cache.Sha256(
			[]byte(text),
			[]byte(img.Key),
		),
		FileID: fileID,
		Data:   ImageData{Type: ImageDataType, Text: text, Image: img},
		Page:   page,
		Offset: offset,
	}
}
func NewTablePart(fileID string, page, offset int, text string, table map[string]any) Part {
	return Part{
		ID: "part-table-" + cache.Sha256(
			[]byte(text),
		),
		FileID: fileID,
		Data:   TableData{Type: TableDataType, Text: text, Table: table},
		Page:   page,
		Offset: offset,
	}
}
