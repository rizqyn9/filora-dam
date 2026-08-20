package asset

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

// StorageService is the interface this module needs from the storage module.
type StorageService interface {
	ElectAccount(ctx context.Context, layer db.StorageLayer, sizeBytes int64) (*db.StorageAccount, error)
}

// Uploader handles the physical upload and records a storage_locations row.
type Uploader interface {
	UploadAndRecord(ctx context.Context, assetID uuid.UUID, accountID int64, layer db.StorageLayer, key string, body io.Reader, size int64, contentType string) (*db.StorageLocation, error)
}

// JobCreator enqueues archive sync jobs.
type JobCreator interface {
	CreateArchiveJob(ctx context.Context, assetID uuid.UUID) error
}

// SpaceQuota checks and updates space storage usage.
type SpaceQuota interface {
	CheckQuota(ctx context.Context, spaceID uuid.UUID, additionalBytes int64) error
	IncrementUsage(ctx context.Context, spaceID uuid.UUID, bytes int64) error
	DecrementUsage(ctx context.Context, spaceID uuid.UUID, bytes int64) error
}

type Service struct {
	repo       *Repository
	storage    StorageService
	uploader   Uploader
	jobCreator JobCreator
	quota      SpaceQuota
}

func NewService(repo *Repository, storage StorageService, uploader Uploader, jobCreator JobCreator, quota SpaceQuota) *Service {
	return &Service{
		repo:       repo,
		storage:    storage,
		uploader:   uploader,
		jobCreator: jobCreator,
		quota:      quota,
	}
}

func (s *Service) GetAsset(ctx context.Context, id uuid.UUID) (*db.Asset, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("asset not found: %w", err)
		}
		return nil, err
	}
	return a, nil
}

func (s *Service) ListAssets(ctx context.Context, params ListAssetsParams) ([]db.Asset, error) {
	if params.Limit == 0 {
		params.Limit = 50
	}
	if params.FolderID != nil {
		return s.repo.ListByFolder(ctx, params.SpaceID, *params.FolderID, params.Limit, params.Offset)
	}
	return s.repo.ListBySpaceRoot(ctx, params.SpaceID, params.Limit, params.Offset)
}

// Upload handles the full upload flow:
// 1. Check space quota
// 2. Hash the file for dedup check
// 3. If hash exists → create reference only (no physical upload)
// 4. If new → elect account, upload + record location, create asset + reference
// 5. Increment space usage
// 6. Enqueue archive job
func (s *Service) Upload(ctx context.Context, userID int64, input UploadInput) (*db.Asset, error) {
	// Read body and compute hash
	hasher := sha256.New()
	content, err := io.ReadAll(io.TeeReader(input.Body, hasher))
	if err != nil {
		return nil, fmt.Errorf("read upload body: %w", err)
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))
	actualSize := int64(len(content))

	// Check space quota
	if err := s.quota.CheckQuota(ctx, input.SpaceID, actualSize); err != nil {
		return nil, fmt.Errorf("quota exceeded: %w", err)
	}

	// Dedup check (global)
	existing, err := s.repo.GetByChecksum(ctx, checksum)
	if err == nil && existing != nil {
		// Asset exists — just create a new reference (no physical upload needed)
		_, err = s.repo.CreateReference(ctx, existing.ID, input.SpaceID, input.FolderID)
		if err != nil {
			return nil, fmt.Errorf("create reference for existing asset: %w", err)
		}
		// Still increment space usage (the space now references this data)
		_ = s.quota.IncrementUsage(ctx, input.SpaceID, existing.SizeBytes)
		return existing, nil
	}

	// Elect a serving account
	account, err := s.storage.ElectAccount(ctx, db.StorageLayerServing, actualSize)
	if err != nil {
		return nil, fmt.Errorf("elect storage account: %w", err)
	}

	// Generate storage key
	assetID := uuid.New()
	key := fmt.Sprintf("%s/%s", assetID.String(), input.Filename)

	// Upload to provider AND record storage_locations row
	_, err = s.uploader.UploadAndRecord(ctx, assetID, account.ID, db.StorageLayerServing, key, bytes.NewReader(content), actualSize, input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("upload to storage: %w", err)
	}

	// Create asset record
	asset, err := s.repo.Create(ctx, db.CreateAssetParams{
		OriginalFilename: input.Filename,
		Name:             input.Filename,
		MimeType:         input.ContentType,
		SizeBytes:        actualSize,
		ChecksumSha256:   checksum,
		UploadedBy:       userID,
	})
	if err != nil {
		return nil, fmt.Errorf("create asset record: %w", err)
	}

	// Create reference
	_, err = s.repo.CreateReference(ctx, asset.ID, input.SpaceID, input.FolderID)
	if err != nil {
		return nil, fmt.Errorf("create asset reference: %w", err)
	}

	// Increment space usage
	_ = s.quota.IncrementUsage(ctx, input.SpaceID, actualSize)

	// Enqueue archive job — log error if it fails, don't silently swallow
	if err := s.jobCreator.CreateArchiveJob(ctx, asset.ID); err != nil {
		// Return the asset (upload succeeded) but wrap the error for the caller to log
		// The asset is usable, but not yet archived — this is a degraded state
		return asset, fmt.Errorf("asset uploaded but archive job failed: %w", err)
	}

	return asset, nil
}

func (s *Service) CreateReference(ctx context.Context, req CreateReferenceRequest) (*db.AssetReference, error) {
	return s.repo.CreateReference(ctx, req.AssetID, req.SpaceID, req.FolderID)
}

func (s *Service) DeleteReference(ctx context.Context, refID int64, spaceID uuid.UUID, sizeBytes int64) error {
	if err := s.repo.SoftDeleteReference(ctx, refID); err != nil {
		return err
	}
	// Decrement space usage when a reference is removed
	_ = s.quota.DecrementUsage(ctx, spaceID, sizeBytes)
	return nil
}

func (s *Service) RestoreReference(ctx context.Context, refID int64) error {
	return s.repo.RestoreReference(ctx, refID)
}

func (s *Service) RenameAsset(ctx context.Context, id uuid.UUID, name string) error {
	return s.repo.UpdateName(ctx, id, name)
}
