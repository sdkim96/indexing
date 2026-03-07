package part

import "github.com/sdkim96/indexing/cache"

type Role string

const (
	RoleContent        Role = "content"
	RoleImage          Role = "image"
	RoleTable          Role = "table"
	RoleTitle          Role = "title"
	RoleSectionHeading Role = "sectionHeading"
	RolePageHeader     Role = "pageHeader"
	RolePageFooter     Role = "pageFooter"
	RolePageNumber     Role = "pageNumber"
	RoleFootnote       Role = "footnote"
	RoleFormulaBlock   Role = "formulaBlock"
)

type Part struct {
	ID              string   `json:"id"`
	FileID          string   `json:"source_id"`
	Role            Role     `json:"role"`
	Data            Data     `json:"data"`
	Page            int      `json:"page"`
	Offset          int      `json:"offset"`
	CreatedAt       int64    `json:"created_at"`
	SectionHeadings []string `json:"section_headings,omitempty"`
}

func NewTextPart(fileID string, role Role, page, offset int, text string, headings []string) Part {
	return Part{
		ID:              "part-text-" + cache.Sha256([]byte(text)),
		FileID:          fileID,
		Role:            role,
		Data:            TextData{Type: TextDataType, Text: text},
		Page:            page,
		Offset:          offset,
		SectionHeadings: headings,
	}
}

func NewImagePart(fileID string, page, offset int, caption string, key string, headings []string) Part {
	return Part{
		ID:              "part-image-" + cache.Sha256([]byte(caption), []byte(key)),
		FileID:          fileID,
		Role:            RoleImage,
		Data:            ImageData{Type: ImageDataType, Text: caption, Image: Image{Key: key}},
		Page:            page,
		Offset:          offset,
		SectionHeadings: headings,
	}
}

func NewTablePart(fileID string, page, offset int, text string, table map[string]any, headings []string) Part {
	return Part{
		ID:              "part-table-" + cache.Sha256([]byte(text)),
		FileID:          fileID,
		Role:            RoleTable,
		Data:            TableData{Type: TableDataType, Text: text, Table: table},
		Page:            page,
		Offset:          offset,
		SectionHeadings: headings,
	}
}
