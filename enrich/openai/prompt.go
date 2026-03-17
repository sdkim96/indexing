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

package openai

const enrichPrompt = `
You are a document analysis assistant.
You will be given an ordered list of document chunks extracted from a single document.
Each chunk is numbered and has a role (title, sectionHeading, content, table) and text content.

Analyze all chunks in order and group them by topic.

Rules:
- EVERY chunk index must appear in exactly one group. Do not skip any index.
- Write all summaries and topics in the same language as the document (Korean if Korean, English if English).
- Keywords must include: key concepts, proper nouns, AND any specific numbers/statistics (e.g. "272조", "45~60%", "LTV:CAC 3x").

For each group:
- topic: the main subject
- idxs: indexes of relevant chunks (no duplicates, no omissions)
- summary: 2-3 sentences optimized for search queries. Use different words than the original.
- keywords: 5-10 key terms including synonyms, related concepts, both Korean and English variants, and key statistics.

Return a JSON object matching the Document schema.
`

// Chunk represents a group of related document parts, identified by their indexes and a common topic.
type Chunk struct {
	Topic    string   `json:"topic"    jsonschema:"description=The topic of the chunk."`
	Idxs     []int    `json:"idxs"     jsonschema:"description=The indexes of the chunk relevant in the original document."`
	Summary  string   `json:"summary"  jsonschema:"description=2-3 sentence summary optimized for search. Use different words than the original text."`
	Keywords []string `json:"keywords" jsonschema:"description=5-10 key terms including synonyms and related concepts for search."`
}

// Document is the response format for linking step of enrichment, containing a list of chunks grouped by topic.
type Document struct {
	Chunks []Chunk `json:"chunks" jsonschema:"description=The chunks of the document. Divided by the topic."`
}
