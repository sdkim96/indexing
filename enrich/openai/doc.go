package openai

type Embedding struct {
	vector []float64
}
type Summary struct {
	text string
}
type Keywords struct {
	words []string
}

type Chunk struct {
	Topic string `json:"topic" jsonschema:"description=The topic of the chunk."`
	Idxs  []int  `json:"idxs" jsonschema:"description=The indexes of the chunk relevant in the original document."`
}

type Document struct {
	Chunks []Chunk `json:"chunks" jsonschema:"description=The chunks of the document. Divided by the topic."`
}

type SummaryAndKeywords struct {
	Summary  Summary
	Keywords Keywords
}

type SearchDoc struct {
	Embedding
	SummaryAndKeywords
	Meta map[string]any
}

func (d SearchDoc) Fields() map[string]any {
	fields := make(map[string]any)
	fields["embedding"] = d.Embedding.vector
	fields["summary"] = d.Summary.text
	fields["keywords"] = d.Keywords.words
	for k, v := range d.Meta {
		fields[k] = v
	}
	return fields
}
