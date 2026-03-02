package storage

import (
	"mime"
	"net/http"
	"path/filepath"
)

func detectContentType(filename string, content []byte) string {
	if mimeType := mime.TypeByExtension(filepath.Ext(filename)); mimeType != "" {
		return mimeType
	}

	if len(content) > 512 {
		content = content[:512]
	}
	return http.DetectContentType(content)
}
