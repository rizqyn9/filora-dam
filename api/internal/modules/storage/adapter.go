package storage

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
	RemoteURL  string // public URL (serving layer only)
}

// Adapter is the port that all storage provider implementations must satisfy.
// Business logic never touches provider SDKs directly — only through this interface.
type Adapter interface {
	Upload(ctx context.Context, input UploadInput) (*UploadResult, error)
	Delete(ctx context.Context, key string) error
}
