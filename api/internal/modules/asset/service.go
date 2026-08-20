package asset

import (
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

// Uploader handles the physical upload to a storage provider.
type Uploader interface {
	Upload(ctx context.Context, accountID int64, key string, body io.Reader, size int64, contentType string) (remotePath, remoteURL string, err error)
}

// JobCreator enqueues archive sync jobs.
type JobCreator interface {
	CreateArchiveJob(ctx context.Context, assetID uuid.UUID) error
}

type Service struct {
	repo       *Repository
	storage    StorageService
	uploader   Uploader
	jobCreator JobCreator
}

func NewService(repo *Repository, storage StorageService, uploader Uploader, jobCreator JobCreator) *Service {
	return &Service{
		repo:       repo,
		storage:    storage,
		uploader:   uploader,
		jobCreator: jobCreator,
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
// 1. Hash the file for dedup check
// 2. If hash exists → create reference only (no upload)
// 3. If new → elect account, upload, create asset + reference, enqueue archive job
func (s *Service) Upload(ctx context.Context, userID int64, spaceID uuid.UUID, folderID *uuid.UUID, filename, contentType string, size int64, body io.Reader) (*db.Asset, error) {
	// Read body and compute hash
	hasher := sha256.New()
	content, err := io.ReadAll(io.TeeReader(body, hasher))
	if err != nil {
		return nil, fmt.Errorf("read upload body: %w", err)
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))
	actualSize := int64(len(content))

	// Dedup check (global)
	existing, err := s.repo.GetByChecksum(ctx, checksum)
	if err == nil && existing != nil {
		// Asset exists — just create a new reference
		_, err = s.repo.CreateReference(ctx, existing.ID, spaceID, folderID)
		if err != nil {
			return nil, fmt.Errorf("create reference for existing asset: %w", err)
		}
		return existing, nil
	}

	// Elect a serving account
	account, err := s.storage.ElectAccount(ctx, db.StorageLayerServing, actualSize)
	if err != nil {
		return nil, fmt.Errorf("elect storage account: %w", err)
	}

	// Generate storage key
	assetID := uuid.New()
	key := fmt.Sprintf("%s/%s", assetID.String(), filename)

	// Upload to provider
	_, _, err = s.uploader.Upload(ctx, account.ID, key, byteReader(content), actualSize, contentType)
	if err != nil {
		return nil, fmt.Errorf("upload to storage: %w", err)
	}

	// Create asset record
	asset, err := s.repo.Create(ctx, db.CreateAssetParams{
		OriginalFilename: filename,
		Name:             filename,
		MimeType:         contentType,
		SizeBytes:        actualSize,
		ChecksumSha256:   checksum,
		UploadedBy:       userID,
	})
	if err != nil {
		return nil, fmt.Errorf("create asset record: %w", err)
	}

	// Create reference
	_, err = s.repo.CreateReference(ctx, asset.ID, spaceID, folderID)
	if err != nil {
		return nil, fmt.Errorf("create asset reference: %w", err)
	}

	// Enqueue archive job (async, best-effort)
	_ = s.jobCreator.CreateArchiveJob(ctx, asset.ID)

	return asset, nil
}

func (s *Service) CreateReference(ctx context.Context, req CreateReferenceRequest) (*db.AssetReference, error) {
	return s.repo.CreateReference(ctx, req.AssetID, req.SpaceID, req.FolderID)
}

func (s *Service) DeleteReference(ctx context.Context, refID int64) error {
	return s.repo.SoftDeleteReference(ctx, refID)
}

func (s *Service) RestoreReference(ctx context.Context, refID int64) error {
	return s.repo.RestoreReference(ctx, refID)
}

func (s *Service) RenameAsset(ctx context.Context, id uuid.UUID, name string) error {
	return s.repo.UpdateName(ctx, id, name)
}

// byteReader wraps a byte slice as an io.Reader.
type byteReaderWrapper struct {
	data []byte
	pos  int
}

func byteReader(b []byte) io.Reader {
	return &byteReaderWrapper{data: b}
}

func (r *byteReaderWrapper) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
