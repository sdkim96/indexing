package analysis

import (
	"fmt"
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
		for _, section := range content.Sections {
			for _, elem := range section.Elements {
				switch {
				case strings.HasPrefix(elem, "/paragraphs/"):
					idx := parseIndex(elem)
					if idx < 0 || idx >= len(content.Paragraphs) {
						continue
					}
					p := content.Paragraphs[idx]
					page := findPageNumber(content.Pages, p.Span.Offset)
					parts = append(parts, part.NewTextPart(fileID, page, p.Span.Offset, p.Content))

				case strings.HasPrefix(elem, "/tables/"):
					idx := parseIndex(elem)
					if idx < 0 || idx >= len(content.Tables) {
						continue
					}
					t := content.Tables[idx]
					page := findPageNumber(content.Pages, t.Span.Offset)
					parts = append(parts, part.NewTablePart(fileID, page, t.Span.Offset, tableToText(t), tableToMap(t)))

				case strings.HasPrefix(elem, "/figures/"):
					idx := parseIndex(elem)
					if idx < 0 || idx >= len(content.Figures) {
						continue
					}
					f := content.Figures[idx]
					page := findPageNumber(content.Pages, f.Span.Offset)

					var textParts []string
					for _, el := range f.Elements {
						if strings.HasPrefix(el, "/paragraphs/") {
							pIdx := parseIndex(el)
							if pIdx >= 0 && pIdx < len(content.Paragraphs) {
								textParts = append(textParts, content.Paragraphs[pIdx].Content)
							}
						}
					}
					text := strings.Join(textParts, " ")
					s3Key := fmt.Sprintf("figures/%s/%s.png", fileID, f.ID)

					parts = append(parts, part.NewImagePart(fileID, page, f.Span.Offset, text, part.Image{Key: s3Key}))
					figCh <- FigureRequest{
						OpID:       result.ID,
						ContentIdx: contentIdx,
						FigureID:   f.ID,
						S3Key:      s3Key,
					}
				}
			}
		}
	}
	return parts
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
	var sb strings.Builder
	for _, cell := range t.Cells {
		sb.WriteString(cell.Content)
		sb.WriteString(" ")
	}
	return strings.TrimSpace(sb.String())
}

func tableToMap(t Table) map[string]any {
	return map[string]any{
		"rowCount":    t.RowCount,
		"columnCount": t.ColumnCount,
		"cells":       t.Cells,
	}
}
