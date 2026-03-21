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
	"fmt"
	"strings"
)

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
