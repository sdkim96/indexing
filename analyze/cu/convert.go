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
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/sdkim96/indexing/part"
)

func ConvertToParts(
	ctx context.Context,
	sourceID string,
	result Operation,
	http *HTTPClient,
	figWriter FigureWriter,
) ([]part.Part, error) {
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
			opID:         result.ID,
		}
		p, figs := tCtx.traverse(0)
		parts = append(parts, p...)
		figureRequests = append(figureRequests, figs...)
	}

	for _, req := range figureRequests {
		uri, mimeType, err := uploadFigure(ctx, sourceID, req, http, figWriter)
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

type traverseCtx struct {
	contentIdx   int
	content      AnalysisContent
	visited      map[string]bool
	captionElems map[string]bool
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
	elParts := strings.Split(elem, "/")
	if len(elParts) < 3 {
		return -1
	}
	idx, err := strconv.Atoi(elParts[2])
	if err != nil {
		return -1
	}
	return idx
}
