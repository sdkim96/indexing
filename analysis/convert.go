package analysis

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/sdkim96/indexing/part"
)

func ConvertToParts(fileID string, result Operation, figCh chan<- FigureRequest) []part.Part {
	if result.Result == nil {
		return nil
	}

	var parts []part.Part

	for contentIdx, content := range result.Result.Contents {
		if len(content.Sections) == 0 {
			continue
		}

		ctx := &traverseCtx{
			ContentIdx:   contentIdx,
			Content:      content,
			Visited:      make(map[string]bool),
			CaptionElems: buildCaptionSet(content),
			FigCh:        figCh,
			FileID:       fileID,
			OpID:         result.ID,
		}
		parts = append(parts, ctx.traverse(0)...)
	}

	return parts
}

type traverseCtx struct {
	ContentIdx   int
	Content      AnalysisContent
	Visited      map[string]bool
	CaptionElems map[string]bool
	FigCh        chan<- FigureRequest
	FileID       string
	OpID         string
	headingStack []string
}

func (ctx *traverseCtx) traverse(sectionIdx int) []part.Part {
	var parts []part.Part

	depthBefore := len(ctx.headingStack)

	for _, elem := range ctx.Content.Sections[sectionIdx].Elements {
		if ctx.Visited[elem] {
			continue
		}
		ctx.Visited[elem] = true

		switch {
		case strings.HasPrefix(elem, "/sections/"):
			idx := parseIndex(elem)
			if idx < 0 || idx >= len(ctx.Content.Sections) {
				continue
			}
			parts = append(parts, ctx.traverse(idx)...)

		case strings.HasPrefix(elem, "/paragraphs/"):
			if ctx.CaptionElems[elem] {
				continue
			}
			idx := parseIndex(elem)
			if idx < 0 || idx >= len(ctx.Content.Paragraphs) {
				continue
			}
			p := ctx.Content.Paragraphs[idx]
			switch part.Role(p.Role) {
			case part.RolePageHeader, part.RolePageFooter, part.RolePageNumber:
				continue
			}
			role := part.Role(p.Role)
			if role == "" {
				role = part.RoleContent
			}

			if role == part.RoleSectionHeading {
				ctx.headingStack = append(ctx.headingStack, p.Content)
			}

			page := findPageNumber(ctx.Content.Pages, p.Span.Offset)
			pt := part.NewTextPart(ctx.FileID, role, page, p.Span.Offset, p.Content, slices.Clone(ctx.headingStack))
			parts = append(parts, pt)

		case strings.HasPrefix(elem, "/tables/"):
			idx := parseIndex(elem)
			if idx < 0 || idx >= len(ctx.Content.Tables) {
				continue
			}
			t := ctx.Content.Tables[idx]
			page := findPageNumber(ctx.Content.Pages, t.Span.Offset)
			pt := part.NewTablePart(ctx.FileID, page, t.Span.Offset, tableToText(t), tableToMap(t), slices.Clone(ctx.headingStack))
			parts = append(parts, pt)

		case strings.HasPrefix(elem, "/figures/"):
			idx := parseIndex(elem)
			if idx < 0 || idx >= len(ctx.Content.Figures) {
				continue
			}
			f := ctx.Content.Figures[idx]
			page := findPageNumber(ctx.Content.Pages, f.Span.Offset)
			text := ""
			if f.Caption != nil {
				text = f.Caption.Content
			}
			s3Key := fmt.Sprintf("figures/%s/%s.png", ctx.FileID, f.ID)
			pt := part.NewImagePart(ctx.FileID, page, f.Span.Offset, text, s3Key, slices.Clone(ctx.headingStack))
			parts = append(parts, pt)
			ctx.FigCh <- FigureRequest{OpID: ctx.OpID, ContentIdx: ctx.ContentIdx, FigureID: f.ID, S3Key: s3Key}
		}
	}

	ctx.headingStack = ctx.headingStack[:depthBefore]

	return parts
}

func buildCaptionSet(content AnalysisContent) map[string]bool {
	set := make(map[string]bool)
	for _, f := range content.Figures {
		if f.Caption != nil {
			for _, el := range f.Caption.Elements {
				set[el] = true
			}
		}
	}
	for _, t := range content.Tables {
		if t.Caption != nil {
			for _, el := range t.Caption.Elements {
				set[el] = true
			}
		}
	}
	return set
}

func findPageNumber(pages []Page, offset int) int {
	for _, page := range pages {
		for _, span := range page.Spans {
			if offset >= span.Offset && offset < span.Offset+span.Length {
				return page.PageNumber
			}
		}
	}
	return 1
}

func parseIndex(elem string) int {
	parts := strings.Split(elem, "/")
	if len(parts) < 3 {
		return -1
	}
	idx, err := strconv.Atoi(parts[2])
	if err != nil {
		return -1
	}
	return idx
}

func tableToText(t Table) string {
	if len(t.Cells) == 0 {
		return ""
	}

	maxRow, maxCol := 0, 0
	for _, cell := range t.Cells {
		if cell.RowIndex > maxRow {
			maxRow = cell.RowIndex
		}
		if cell.ColumnIndex > maxCol {
			maxCol = cell.ColumnIndex
		}
	}

	grid := make([][]string, maxRow+1)
	for i := range grid {
		grid[i] = make([]string, maxCol+1)
	}
	for _, cell := range t.Cells {
		grid[cell.RowIndex][cell.ColumnIndex] = cell.Content
	}

	colWidths := make([]int, maxCol+1)
	for _, row := range grid {
		for j, cell := range row {
			if len([]rune(cell)) > colWidths[j] {
				colWidths[j] = len([]rune(cell))
			}
		}
	}

	var sb strings.Builder
	for i, row := range grid {
		sb.WriteString("| ")
		for j, cell := range row {
			padding := colWidths[j] - len([]rune(cell))
			sb.WriteString(cell)
			sb.WriteString(strings.Repeat(" ", padding))
			sb.WriteString(" | ")
		}
		sb.WriteString("\n")

		if i == 0 {
			sb.WriteString("|")
			for _, w := range colWidths {
				sb.WriteString(strings.Repeat("-", w+2))
				sb.WriteString("|")
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func tableToMap(t Table) map[string]any {
	headers := make(map[int]string)
	for _, cell := range t.Cells {
		if cell.RowIndex == 0 {
			headers[cell.ColumnIndex] = cell.Content
		}
	}

	rowMap := make(map[int]map[string]any)
	for _, cell := range t.Cells {
		if cell.RowIndex == 0 {
			continue
		}
		if rowMap[cell.RowIndex] == nil {
			rowMap[cell.RowIndex] = make(map[string]any)
		}
		header := headers[cell.ColumnIndex]
		if header == "" {
			header = fmt.Sprintf("col_%d", cell.ColumnIndex)
		}
		rowMap[cell.RowIndex][header] = cell.Content
	}

	rows := make([]map[string]any, 0, len(rowMap))
	for i := 1; i <= t.RowCount; i++ {
		if row, ok := rowMap[i]; ok {
			rows = append(rows, row)
		}
	}

	return map[string]any{
		"table": rows,
	}
}
