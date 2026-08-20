package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

// Uploader implements asset.Uploader by routing to the correct adapter via the registry.
type Uploader struct {
	registry *Registry
	repo     *Repository
}

func NewUploader(registry *Registry, repo *Repository) *Uploader {
	return &Uploader{registry: registry, repo: repo}
}

// Upload uploads a file to the given account and records the storage location.
func (u *Uploader) Upload(ctx context.Context, accountID int64, key string, body io.Reader, size int64, contentType string) (remotePath, remoteURL string, err error) {
	adapter, err := u.registry.Get(ctx, accountID)
	if err != nil {
		return "", "", fmt.Errorf("get adapter: %w", err)
	}

	result, err := adapter.Upload(ctx, UploadInput{
		Key:         key,
		Body:        body,
		Size:        size,
		ContentType: contentType,
	})
	if err != nil {
		return "", "", fmt.Errorf("upload via adapter: %w", err)
	}

	// Increment account usage
	_ = u.repo.IncrementUsage(ctx, accountID, size)

	return result.RemotePath, result.RemoteURL, nil
}

// UploadAndRecord uploads a file and creates a storage_locations record.
func (u *Uploader) UploadAndRecord(ctx context.Context, assetID uuid.UUID, accountID int64, layer db.StorageLayer, key string, body io.Reader, size int64, contentType string) (*db.StorageLocation, error) {
	remotePath, remoteURL, err := u.Upload(ctx, accountID, key, body, size, contentType)
	if err != nil {
		// Record failure
		loc, _ := u.repo.CreateLocation(ctx, db.CreateStorageLocationParams{
			AssetID:   assetID,
			AccountID: accountID,
			Layer:     layer,
			Status:    db.LocationStatusFailed,
		})
		return loc, err
	}

	loc, err := u.repo.CreateLocation(ctx, db.CreateStorageLocationParams{
		AssetID:    assetID,
		AccountID:  accountID,
		Layer:      layer,
		Status:     db.LocationStatusStored,
		RemotePath: &remotePath,
		RemoteUrl:  &remoteURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create storage location record: %w", err)
	}

	return loc, nil
}
