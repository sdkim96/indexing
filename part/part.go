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

package part

import (
	"context"

	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/uri"
)

// Part represents a piece of content extracted from the input data. These are the basic units that will be enriched and indexed.
// Each Part should be self-contained and addressable, allowing for targeted enrichment and retrieval.
//
// We recommend keeping Parts to represent the logical atoms of the content.
//
// For example, if the input is a PDF document, each Part could represent a single semantically meaningful chunk, such as a sentence.
type Part interface {

	// MimeType returns the MIME type of the content in this Part, which can be used to determine how to process it during enrichment.
	MimeType() mime.Type

	// Text returns the textual content of the Part.
	// This is designed to be the primary element that LLMs consume.
	Text() string

	// Raw returns the raw byte content of the Part, which could be useful for marshalling or representing non-textual data.
	Raw() []byte
}

// PartWriter writes Parts to a destination identified by a URI, such as a database or file storage.
type PartWriter interface {

	// Write takes a list of Parts and writes them to the specified URI.
	Write(ctx context.Context, URI uri.URI, parts []Part) error
}
