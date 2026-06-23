package lib

import (
	"mime"
	"path/filepath"
	"strings"
)

// AssetType represents the type of asset
type AssetType string

const (
	TypeImage    AssetType = "image"
	TypeVideo    AssetType = "video"
	TypeDocument AssetType = "document"
	TypeArchive  AssetType = "archive"
	TypeFile     AssetType = "file"
)

// GetAssetType determines asset type from MIME type
func GetAssetType(mimeType string) AssetType {
	if strings.HasPrefix(mimeType, "image/") {
		return TypeImage
	}
	if strings.HasPrefix(mimeType, "video/") {
		return TypeVideo
	}

	// Document types
	documentTypes := map[string]bool{
		"application/pdf":        true,
		"application/msword":     true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
		"application/vnd.ms-excel":                                                  true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
		"application/vnd.ms-powerpoint":                                             true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
		"text/plain":             true,
		"text/csv":               true,
	}
	if documentTypes[mimeType] {
		return TypeDocument
	}

	// Archive types
	archiveTypes := map[string]bool{
		"application/zip":             true,
		"application/x-zip-compressed": true,
		"application/x-rar-compressed": true,
		"application/x-tar":            true,
		"application/gzip":             true,
		"application/x-7z-compressed":  true,
	}
	if archiveTypes[mimeType] {
		return TypeArchive
	}

	return TypeFile
}

// DetectMimeType detects MIME type from filename
func DetectMimeType(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "application/octet-stream"
	}

	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "application/octet-stream"
	}

	return mimeType
}
