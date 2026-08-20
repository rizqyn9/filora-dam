// Package adapters contains storage provider implementations.
// The Adapter interface and value types are defined here to avoid import cycles.
package adapters

import (
	"context"
	"io"
)

// UploadInput is the input for uploading a file to a storage provider.
type UploadInput struct {
	Key         string
	Body        io.Reader
	Size        int64
	ContentType string
}

// UploadResult is returned after a successful upload.
type UploadResult struct {
	RemotePath string
	RemoteURL  string
}

// Adapter is the port that all storage provider implementations must satisfy.
type Adapter interface {
	Upload(ctx context.Context, input UploadInput) (*UploadResult, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}
