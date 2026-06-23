package adapters

import (
	"context"
	"errors"
	"io"
)

var (
	ErrNotImplemented = errors.New("adapter not implemented")
	ErrInvalidConfig  = errors.New("invalid adapter configuration")
	ErrUploadFailed   = errors.New("upload failed")
	ErrDownloadFailed = errors.New("download failed")
	ErrDeleteFailed   = errors.New("delete failed")
)

// UploadInput represents file upload input
type UploadInput struct {
	File     io.Reader
	Filename string
	MimeType string
	Size     int64
}

// UploadResult represents upload result from provider
type UploadResult struct {
	Key      string                 `json:"key"`
	URL      string                 `json:"url"`
	Size     int64                  `json:"size"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// StorageAdapter defines the interface that all storage providers must implement
type StorageAdapter interface {
	// Upload uploads a file to the storage provider
	Upload(ctx context.Context, input *UploadInput) (*UploadResult, error)

	// Download retrieves a file from the storage provider
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes a file from the storage provider
	Delete(ctx context.Context, key string) error

	// Exists checks if a file exists in the storage provider
	Exists(ctx context.Context, key string) (bool, error)

	// GetURL returns the public URL for a file
	GetURL(ctx context.Context, key string) (string, error)
}

// AdapterConfig represents common configuration for adapters
type AdapterConfig struct {
	Credentials map[string]interface{}
}
