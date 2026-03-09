package internal

import "strings"

var mimeMapping = map[string]string{
	".pdf":  "application/pdf",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".txt":  "text/plain",
}

func GuessMimeType(path string) string {
	for ext, mime := range mimeMapping {
		if strings.HasSuffix(path, ext) {
			return mime
		}
	}
	return "application/octet-stream"
}
