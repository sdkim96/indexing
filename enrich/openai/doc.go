package openai

import "encoding/json"

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

type OpenAISearchDoc struct {
	Title string
	Embedding
	SummaryAndKeywords
	Meta map[string]any
}

func (d OpenAISearchDoc) Fields() map[string]any {
	fields := make(map[string]any)
	fields["embedding"] = d.Embedding.Vector()
	fields["summary"] = d.Summary.Text()
	fields["keywords"] = d.Keywords.Words()
	for k, v := range d.Meta {
		fields[k] = v
	}
	return fields
}

func (d OpenAISearchDoc) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Fields())
}

func (d *OpenAISearchDoc) UnmarshalJSON(data []byte) error {
	var raw struct {
		Title     string         `json:"title"`
		Embedding []float64      `json:"embedding"`
		Summary   string         `json:"summary"`
		Keywords  []string       `json:"keywords"`
		Meta      map[string]any `json:"meta"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	d.Title = raw.Title
	d.SummaryAndKeywords.Summary = Summary{text: raw.Summary}
	d.SummaryAndKeywords.Keywords = Keywords{words: raw.Keywords}
	d.Embedding = Embedding{vector: raw.Embedding}
	d.Meta = raw.Meta
	return nil
}
