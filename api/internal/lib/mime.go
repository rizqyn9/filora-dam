package lib

import (
	"net/http"
	"strings"
)

// DetectContentType sniffs the MIME type from the first bytes of content.
func DetectContentType(head []byte) string {
	return http.DetectContentType(head)
}

// ClassifyType maps a MIME type to Filora's asset type
// (image | video | document | archive | file).
func ClassifyType(mime string) string {
	m := strings.ToLower(mime)
	switch {
	case strings.HasPrefix(m, "image/"):
		return "image"
	case strings.HasPrefix(m, "video/"):
		return "video"
	case m == "application/pdf",
		strings.HasPrefix(m, "text/"),
		strings.Contains(m, "msword"),
		strings.Contains(m, "officedocument"),
		strings.Contains(m, "ms-excel"),
		strings.Contains(m, "ms-powerpoint"),
		strings.Contains(m, "opendocument"):
		return "document"
	case strings.Contains(m, "zip"),
		strings.Contains(m, "x-rar"),
		strings.Contains(m, "x-tar"),
		strings.Contains(m, "gzip"),
		strings.Contains(m, "x-7z"):
		return "archive"
	default:
		return "file"
	}
}
