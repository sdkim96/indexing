package job

import (
	"strconv"

	"github.com/sdkim96/indexing/cache"
)

type File struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Bytes    []byte `json:"bytes"`
	MimeType string `json:"mime_type"`
}

func (f *File) Size() int {
	return len(f.Bytes)
}

func (f *File) FingerPrint() string {
	return cache.Sha256(f.Bytes, []byte(f.MimeType), []byte(strconv.Itoa(f.Size())))
}
