package file

import (
	"context"
	"testing"

	"github.com/sdkim96/indexing/internal/mime"
	"github.com/sdkim96/indexing/internal/storage"
	"github.com/sdkim96/indexing/internal/uri"
	"github.com/sdkim96/indexing/part"
)

type MockPart struct {
	mimeType mime.Type
	text     string
	raw      []byte
}

func NewMockPart(mimeType mime.Type, text string, raw []byte) *MockPart {
	return &MockPart{
		mimeType: mimeType,
		text:     text,
		raw:      raw,
	}
}

func (p *MockPart) MimeType() mime.Type {
	return p.mimeType
}

func (p *MockPart) Text() string {
	return p.text
}

func (p *MockPart) Raw() []byte {
	return p.raw
}

func (p *MockPart) MarshalJSON() ([]byte, error) {
	return []byte(`{"mimeType":"` + string(p.mimeType) + `","text":"` + p.text + `"}`), nil
}

func newWriter() *FilePartWriter {
	client, _ := storage.NewFileSystemClient("testdata")
	return New(client)
}

func TestFileWriter(t *testing.T) {
	writer := newWriter()
	path := uri.URI("file://testdata/test.json")

	var parts []part.Part
	parts = append(parts, NewMockPart(mime.MimeTxt, "Hello, World!", []byte("Hello, World!")))
	parts = append(parts, NewMockPart(mime.MimeTxt, "Goodbye, World!", []byte("Goodbye, World!")))

	writer.Write(context.Background(), path, parts)
}
