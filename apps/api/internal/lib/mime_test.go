package lib

import "testing"

func TestClassifyType(t *testing.T) {
	cases := map[string]string{
		"image/jpeg":      "image",
		"image/png":       "image",
		"video/mp4":       "video",
		"application/pdf": "document",
		"text/plain":      "document",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "document",
		"application/zip":              "archive",
		"application/x-rar-compressed": "archive",
		"application/octet-stream":     "file",
		"":                             "file",
	}
	for mime, want := range cases {
		if got := ClassifyType(mime); got != want {
			t.Errorf("ClassifyType(%q) = %q, want %q", mime, got, want)
		}
	}
}
