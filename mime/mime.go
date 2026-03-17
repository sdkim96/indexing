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

package mime

import "strings"

// Type represents the MIME type of a file or data.
// It is a string that follows the format "type/subtype", such as "application/pdf" or "text/plain".
type Type string

const (
	MimePDF     Type = "application/pdf"
	MimeDocx    Type = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	MimeDoc     Type = "application/msword"
	MimeXlsx    Type = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	MimePptx    Type = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	MimeTxt     Type = "text/plain"
	MimeMd      Type = "text/markdown"
	MimeHTML    Type = "text/html"
	MimeCSV     Type = "text/csv"
	MimeJSON    Type = "application/json"
	MimeXML     Type = "application/xml"
	MimePNG     Type = "image/png"
	MimeJPG     Type = "image/jpeg"
	MimeGIF     Type = "image/gif"
	MimeWebP    Type = "image/webp"
	MimeMP4     Type = "video/mp4"
	MimeMP3     Type = "audio/mpeg"
	MimeZIP     Type = "application/zip"
	MimeUnknown Type = "application/octet-stream"
)

var mimeMapping = map[string]Type{
	".pdf":  MimePDF,
	".docx": MimeDocx,
	".doc":  MimeDoc,
	".xlsx": MimeXlsx,
	".pptx": MimePptx,
	".txt":  MimeTxt,
	".md":   MimeMd,
	".html": MimeHTML,
	".csv":  MimeCSV,
	".json": MimeJSON,
	".xml":  MimeXML,
	".png":  MimePNG,
	".jpg":  MimeJPG,
	".jpeg": MimeJPG,
	".gif":  MimeGIF,
	".webp": MimeWebP,
	".mp4":  MimeMP4,
	".mp3":  MimeMP3,
	".zip":  MimeZIP,
}

func TextTypes() []Type {
	return []Type{MimeTxt, MimeMd, MimeHTML, MimeCSV, MimeJSON, MimeXML}
}
func ImageTypes() []Type {
	return []Type{MimePNG, MimeJPG, MimeGIF, MimeWebP}
}

func GuessMimeType(path string) Type {
	for ext, mime := range mimeMapping {
		if strings.HasSuffix(path, ext) {
			return mime
		}
	}
	return MimeUnknown
}

func GuessExtension(mimeType Type) string {
	for ext, mime := range mimeMapping {
		if mime == mimeType {
			return ext
		}
	}
	return ""
}
