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

	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/urio"
)

type figureRequest struct {
	opID       string
	contentIdx int
	figureID   string
	page       int
	offset     int
	caption    string
	headings   []string
}

func uploadFigure(
	ctx context.Context,
	sourceID string,
	req figureRequest,
	http *HTTPClient,
	figWriter FigureWriter,
) (uri urio.URI, mimeType mime.Type, err error) {
	data, contentType, err := http.GetFigure(ctx, FigureRequest{
		OpID:       req.opID,
		ContentIdx: req.contentIdx,
		FigureID:   req.figureID,
	})
	if err != nil {
		return "", "", err
	}

	ext := mime.GuessExtension(contentType)
	w, err := figWriter(ctx, fmt.Sprintf("figures/%s/%s%s", sourceID, req.figureID, ext), contentType)
	if err != nil {
		return "", "", err
	}
	defer w.Close()

	if _, err = io.Copy(w, bytes.NewReader(data)); err != nil {
		return "", "", err
	}

	return w.URI(), contentType, nil
}
