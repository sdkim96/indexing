package cu

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sdkim96/indexing/internal/mime"
	"github.com/sdkim96/indexing/internal/uri"
	"github.com/sdkim96/indexing/part"
)

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

var _ part.Part = (*CUPart)(nil)

type CUPart struct {
	role            Role
	data            Data
	page            int
	offset          int
	createdAt       int64
	sectionHeadings []string
	mimeType        mime.Type
}

func (p *CUPart) MimeType() mime.Type {
	return p.mimeType
}

func (p *CUPart) Text() string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("## [%s] p.%d\n", p.role, p.page))

	if len(p.sectionHeadings) > 0 {
		buf.WriteString(fmt.Sprintf("> %s\n", strings.Join(p.sectionHeadings, " > ")))
	}

	buf.WriteString("\n")
	buf.WriteString(p.data.GetText())

	return buf.String()
}

func (p *CUPart) Raw() []byte {
	b, _ := json.Marshal(struct {
		Role            Role     `json:"role"`
		Page            int      `json:"page"`
		Offset          int      `json:"offset"`
		CreatedAt       int64    `json:"created_at"`
		SectionHeadings []string `json:"section_headings,omitempty"`
		Data            any      `json:"data"`
	}{
		Role:            p.role,
		Page:            p.page,
		Offset:          p.offset,
		CreatedAt:       p.createdAt,
		SectionHeadings: p.sectionHeadings,
		Data:            p.data.Raw(),
	})
	return b
}

func (p *CUPart) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Role            Role      `json:"role"`
		MimeType        mime.Type `json:"mime_type"`
		Page            int       `json:"page"`
		Offset          int       `json:"offset"`
		CreatedAt       int64     `json:"created_at"`
		SectionHeadings []string  `json:"section_headings,omitempty"`
		Data            Data      `json:"data"`
	}{
		Role:            p.role,
		MimeType:        p.mimeType,
		Page:            p.page,
		Offset:          p.offset,
		CreatedAt:       p.createdAt,
		SectionHeadings: p.sectionHeadings,
		Data:            p.data,
	})
}

func (p *CUPart) UnmarshalJSON(data []byte) error {
	var v struct {
		Role            Role            `json:"role"`
		MimeType        mime.Type       `json:"mime_type"`
		Page            int             `json:"page"`
		Offset          int             `json:"offset"`
		CreatedAt       int64           `json:"created_at"`
		SectionHeadings []string        `json:"section_headings,omitempty"`
		Data            json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	p.role = v.Role
	p.mimeType = v.MimeType
	p.page = v.Page
	p.offset = v.Offset
	p.createdAt = v.CreatedAt
	p.sectionHeadings = v.SectionHeadings

	var header struct {
		Type DataType `json:"type"`
	}
	if err := json.Unmarshal(v.Data, &header); err != nil {
		return err
	}

	switch header.Type {
	case TextDataType:
		var d TextData
		if err := json.Unmarshal(v.Data, &d); err != nil {
			return err
		}
		p.data = d
	case ImageDataType:
		var d ImageData
		if err := json.Unmarshal(v.Data, &d); err != nil {
			return err
		}
		p.data = d
	case TableDataType:
		var d TableData
		if err := json.Unmarshal(v.Data, &d); err != nil {
			return err
		}
		p.data = d
	default:
		return fmt.Errorf("unknown data type: %q", header.Type)
	}

	return nil
}

func NewTextPart(role Role, page, offset int, text string, headings []string) part.Part {
	return &CUPart{
		role:            role,
		data:            TextData{Type: TextDataType, Text: text},
		page:            page,
		offset:          offset,
		createdAt:       time.Now().Unix(),
		sectionHeadings: headings,
		mimeType:        mime.MimeTxt,
	}
}

func NewImageURLPart(page, offset int, caption string, uri uri.URI, mimeType mime.Type, headings []string) part.Part {
	return &CUPart{
		role:            RoleImage,
		data:            ImageData{Type: ImageDataType, Text: caption, Image: Image{URI: uri}},
		page:            page,
		offset:          offset,
		createdAt:       time.Now().Unix(),
		sectionHeadings: headings,
		mimeType:        mimeType,
	}
}

func NewTablePart(page, offset int, text string, table map[string]any, headings []string) part.Part {
	return &CUPart{
		role:            RoleTable,
		data:            TableData{Type: TableDataType, Text: text, Table: table},
		page:            page,
		offset:          offset,
		createdAt:       time.Now().Unix(),
		sectionHeadings: headings,
		mimeType:        mime.MimeJSON,
	}
}
