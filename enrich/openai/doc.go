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

type SearchDoc struct {
	Embedding
	Summary
	Keywords
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
