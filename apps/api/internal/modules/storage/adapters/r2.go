package adapters

import (
	"context"
	"io"
)

// r2Adapter (S3-compatible). Concrete SDK calls land in phase 7/8.
type r2Adapter struct {
	creds R2Credentials
}

func (a *r2Adapter) Upload(ctx context.Context, in UploadInput) (*UploadResult, error) {
	return nil, ErrNotImplemented
}
func (a *r2Adapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, ErrNotImplemented
}
func (a *r2Adapter) Delete(ctx context.Context, key string) error { return ErrNotImplemented }
func (a *r2Adapter) Exists(ctx context.Context, key string) (bool, error) {
	return false, ErrNotImplemented
}
