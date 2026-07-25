package adapters

import (
	"context"
	"io"
)

// cloudinaryAdapter (serving layer). Concrete SDK calls land in phase 7.
type cloudinaryAdapter struct {
	creds CloudinaryCredentials
}

func (a *cloudinaryAdapter) Upload(ctx context.Context, in UploadInput) (*UploadResult, error) {
	return nil, ErrNotImplemented
}
func (a *cloudinaryAdapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, ErrNotImplemented
}
func (a *cloudinaryAdapter) Delete(ctx context.Context, key string) error { return ErrNotImplemented }
func (a *cloudinaryAdapter) Exists(ctx context.Context, key string) (bool, error) {
	return false, ErrNotImplemented
}
