// Copyright 2026 Sungdong Kim
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cu

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/indexing/urio"
)

type figureRequest struct {
	opID       string
	contentIdx int
	figureID   string
	fileID     string
	page       int
	offset     int
	caption    string
	headings   []string
}

func ConvertToParts(ctx context.Context, result Operation, http *HTTPClient, figWriter func(ctx context.Context, name string) (urio.WriteCloser, error)) ([]part.Part, error) {
	if result.Result == nil {
		return nil, fmt.Errorf("operation %q result is nil: check if the operation completed successfully", result.ID)
	}

	var parts []part.Part
	var figureRequests []figureRequest

	for contentIdx, content := range result.Result.Contents {
		if len(content.Sections) == 0 {
			continue
		}

		tCtx := &traverseCtx{
			contentIdx:   contentIdx,
			content:      content,
			visited:      make(map[string]bool),
			captionElems: buildCaptionSet(content),
			fileID:       uuid.New().String(),
			opID:         result.ID,
		}
		p, figs := tCtx.traverse(0)
		parts = append(parts, p...)
		figureRequests = append(figureRequests, figs...)
	}

	for _, req := range figureRequests {
		uri, mimeType, err := uploadFigure(ctx, req, http, figWriter)
		if err != nil {
			pt := NewTextPart(RoleImage, req.page, req.offset, req.caption, req.headings)
			parts = append(parts, pt)
			continue
		}
		pt := NewImageURLPart(req.page, req.offset, req.caption, uri, mimeType, req.headings)
		parts = append(parts, pt)
	}

	return parts, nil
}

func uploadFigure(ctx context.Context, req figureRequest, http *HTTPClient, figWriter func(ctx context.Context, name string) (urio.WriteCloser, error)) (uri urio.URI, mimeType mime.Type, err error) {
	data, contentType, err := http.GetFigure(ctx, FigureRequest{
		OpID:       req.opID,
		ContentIdx: req.contentIdx,
		FigureID:   req.figureID,
	})
	if err != nil {
		return "", "", err
	}

	ext := mime.GuessExtension(contentType)
	storageKey := fmt.Sprintf("figures/%s/%s%s", req.fileID, req.figureID, ext)
	w, err := figWriter(ctx, storageKey)
	if err != nil {
		return "", "", err
	}
	defer w.Close()

	if _, err = io.Copy(w, bytes.NewReader(data)); err != nil {
		return "", "", err
	}

	return w.URI(), contentType, nil
}

type traverseCtx struct {
	contentIdx   int
	content      AnalysisContent
	visited      map[string]bool
	captionElems map[string]bool
	fileID       string
	opID         string
	headingStack []string
}

func (ctx *traverseCtx) traverse(sectionIdx int) ([]part.Part, []figureRequest) {
	var parts []part.Part
	var figs []figureRequest

	depthBefore := len(ctx.headingStack)

	for _, elem := range ctx.content.Sections[sectionIdx].Elements {
		if ctx.visited[elem] {
			continue
		}
		ctx.visited[elem] = true

		switch {
		case strings.HasPrefix(elem, "/sections/"):
			idx := parseIndex(elem)
			if idx < 0 || idx >= len(ctx.content.Sections) {
				continue
			}
			p, f := ctx.traverse(idx)
			parts = append(parts, p...)
			figs = append(figs, f...)

		case strings.HasPrefix(elem, "/paragraphs/"):
			if ctx.captionElems[elem] {
				continue
			}
			idx := parseIndex(elem)
			if idx < 0 || idx >= len(ctx.content.Paragraphs) {
				continue
			}
			p := ctx.content.Paragraphs[idx]
			switch Role(p.Role) {
			case RolePageHeader, RolePageFooter, RolePageNumber:
				continue
			}
			role := Role(p.Role)
			if role == "" {
				role = RoleContent
			}
			if role == RoleSectionHeading {
				ctx.headingStack = append(ctx.headingStack, p.Content)
			}
			page := findPageNumber(ctx.content.Pages, p.Span.Offset)
			pt := NewTextPart(role, page, p.Span.Offset, p.Content, slices.Clone(ctx.headingStack))
			parts = append(parts, pt)

		case strings.HasPrefix(elem, "/tables/"):
			idx := parseIndex(elem)
			if idx < 0 || idx >= len(ctx.content.Tables) {
				continue
			}
			t := ctx.content.Tables[idx]
			page := findPageNumber(ctx.content.Pages, t.Span.Offset)
			pt := NewTablePart(page, t.Span.Offset, tableToText(t), tableToMap(t), slices.Clone(ctx.headingStack))
			parts = append(parts, pt)

		case strings.HasPrefix(elem, "/figures/"):
			idx := parseIndex(elem)
			if idx < 0 || idx >= len(ctx.content.Figures) {
				continue
			}
			f := ctx.content.Figures[idx]
			page := findPageNumber(ctx.content.Pages, f.Span.Offset)
			caption := ""
			if f.Caption != nil {
				caption = f.Caption.Content
			}
			figs = append(figs, figureRequest{
				opID:       ctx.opID,
				contentIdx: ctx.contentIdx,
				figureID:   f.ID,
				fileID:     ctx.fileID,
				page:       page,
				offset:     f.Span.Offset,
				caption:    caption,
				headings:   slices.Clone(ctx.headingStack),
			})
		}
	}

	ctx.headingStack = ctx.headingStack[:depthBefore]

	return parts, figs
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
