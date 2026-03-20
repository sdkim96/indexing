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

import "github.com/sdkim96/indexing/urio"

type DataType string

const (
	TextDataType  DataType = "text"
	ImageDataType DataType = "image"
	TableDataType DataType = "table"
)

type Data interface {
	GetText() string
	Raw() any
	GetType() DataType
}

type TextData struct {
	Type DataType `json:"type"`
	Text string   `json:"text"`
}

func (t TextData) GetType() DataType { return TextDataType }
func (t TextData) GetText() string   { return t.Text }
func (t TextData) Raw() any          { return t.Text }

type ImageData struct {
	Type  DataType `json:"type"`
	Text  string   `json:"text"`
	Image Image    `json:"image"`
}

func (i ImageData) GetType() DataType { return ImageDataType }
func (i ImageData) GetText() string   { return i.Text }
func (i ImageData) Raw() any          { return i.Image }

type TableData struct {
	Type  DataType       `json:"type"`
	Text  string         `json:"text"`
	Table map[string]any `json:"table"`
}

func (t TableData) GetType() DataType { return TableDataType }
func (t TableData) GetText() string   { return t.Text }
func (t TableData) Raw() any          { return t.Table }

type Image struct {
	URI urio.URI `json:"uri,omitempty"`
}
