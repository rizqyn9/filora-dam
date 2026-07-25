package adapters

import (
	"context"
	"io"
)

// gcsAdapter (archive layer). Concrete SDK calls land in phase 8.
type gcsAdapter struct {
	creds GCSCredentials
}

func (a *gcsAdapter) Upload(ctx context.Context, in UploadInput) (*UploadResult, error) {
	return nil, ErrNotImplemented
}
func (a *gcsAdapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, ErrNotImplemented
}
func (a *gcsAdapter) Delete(ctx context.Context, key string) error { return ErrNotImplemented }
func (a *gcsAdapter) Exists(ctx context.Context, key string) (bool, error) {
	return false, ErrNotImplemented
}
