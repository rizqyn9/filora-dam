package adapters

import (
	"context"
	"io"
)

// imagekitAdapter (serving layer). Concrete SDK calls land in phase 7.
type imagekitAdapter struct {
	creds ImageKitCredentials
}

func (a *imagekitAdapter) Upload(ctx context.Context, in UploadInput) (*UploadResult, error) {
	return nil, ErrNotImplemented
}
func (a *imagekitAdapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, ErrNotImplemented
}
func (a *imagekitAdapter) Delete(ctx context.Context, key string) error { return ErrNotImplemented }
func (a *imagekitAdapter) Exists(ctx context.Context, key string) (bool, error) {
	return false, ErrNotImplemented
}
