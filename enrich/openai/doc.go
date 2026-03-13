package openai

type Embedding struct {
	vector []float64
}

func (e Embedding) Vector() []float64 {
	return e.vector
}

type Summary struct {
	text string
}

func (s Summary) Text() string {
	return s.text
}

type Keywords struct {
	words []string
}

func (k Keywords) Words() []string {
	return k.words
}

type SummaryAndKeywords struct {
	Summary  Summary
	Keywords Keywords
}

func (s SummaryAndKeywords) Fields() map[string]any {
	fields := make(map[string]any)
	fields["summary"] = s.Summary.Text()
	fields["keywords"] = s.Keywords.Words()
	return fields
}

type SearchDoc struct {
	Title string
	Embedding
	SummaryAndKeywords
	Meta map[string]any
}

func (d SearchDoc) Fields() map[string]any {
	fields := make(map[string]any)
	fields["embedding"] = d.Embedding.Vector()
	fields["summary"] = d.Summary.Text()
	fields["keywords"] = d.Keywords.Words()
	for k, v := range d.Meta {
		fields[k] = v
	}
	return fields
}
