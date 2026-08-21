package asset

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
	"github.com/rizqyn9/filora-dam/api/internal/telemetry"
)

var tracer = otel.Tracer("filora-api")

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

// AssetRepository is the interface the service needs from its data layer.
type AssetRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*db.Asset, error)
	GetByChecksum(ctx context.Context, hash string) (*db.Asset, error)
	Create(ctx context.Context, params db.CreateAssetParams) (*db.Asset, error)
	CreateReference(ctx context.Context, assetID, spaceID uuid.UUID, folderID *uuid.UUID) (*db.AssetReference, error)
	SoftDeleteReference(ctx context.Context, id int64) error
	RestoreReference(ctx context.Context, id int64) error
	ListByFolder(ctx context.Context, spaceID, folderID uuid.UUID, limit, offset int32) ([]db.Asset, error)
	ListBySpaceRoot(ctx context.Context, spaceID uuid.UUID, limit, offset int32) ([]db.Asset, error)
	UpdateName(ctx context.Context, id uuid.UUID, name string) error
}

type Service struct {
	repo       AssetRepository
	storage    StorageService
	uploader   Uploader
	jobCreator JobCreator
	quota      SpaceQuota
	metrics    *telemetry.Metrics
}

func NewService(repo AssetRepository, storage StorageService, uploader Uploader, jobCreator JobCreator, quota SpaceQuota, metrics *telemetry.Metrics) *Service {
	return &Service{
		repo:       repo,
		storage:    storage,
		uploader:   uploader,
		jobCreator: jobCreator,
		quota:      quota,
		metrics:    metrics,
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

// Upload handles the full upload flow with observability.
func (s *Service) Upload(ctx context.Context, userID int64, input UploadInput) (*db.Asset, error) {
	ctx, span := tracer.Start(ctx, "asset.upload")
	defer span.End()

	start := time.Now()
	span.SetAttributes(
		attribute.String("asset.filename", input.Filename),
		attribute.String("asset.content_type", input.ContentType),
		attribute.Int64("asset.size_bytes", input.Size),
		attribute.String("space.id", input.SpaceID.String()),
	)

	// Read body and compute hash
	hasher := sha256.New()
	content, err := io.ReadAll(io.TeeReader(input.Body, hasher))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "read body failed")
		slog.ErrorContext(ctx, "upload read body failed", "error", err)
		return nil, fmt.Errorf("read upload body: %w", err)
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))
	actualSize := int64(len(content))

	// Check space quota
	if err := s.quota.CheckQuota(ctx, input.SpaceID, actualSize); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "quota exceeded")
		slog.WarnContext(ctx, "upload quota exceeded", "space_id", input.SpaceID, "size", actualSize)
		return nil, fmt.Errorf("quota exceeded: %w", err)
	}

	// Dedup check (global)
	existing, err := s.repo.GetByChecksum(ctx, checksum)
	if err == nil && existing != nil {
		span.SetAttributes(attribute.Bool("dedup.hit", true))

		_, err = s.repo.CreateReference(ctx, existing.ID, input.SpaceID, input.FolderID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "create reference failed")
			return nil, fmt.Errorf("create reference for existing asset: %w", err)
		}

		_ = s.quota.IncrementUsage(ctx, input.SpaceID, existing.SizeBytes)

		slog.InfoContext(ctx, "asset dedup hit",
			"asset_id", existing.ID,
			"checksum", checksum,
			"space_id", input.SpaceID,
		)

		s.metrics.UploadsTotal.Add(ctx, 1, metric.WithAttributes(
			attribute.String("mime_type", input.ContentType),
			attribute.Bool("dedup_hit", true),
		))
		return existing, nil
	}
	span.SetAttributes(attribute.Bool("dedup.hit", false))

	// Elect a serving account
	account, err := s.storage.ElectAccount(ctx, db.StorageLayerServing, actualSize)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "elect account failed")
		slog.ErrorContext(ctx, "elect storage account failed", "error", err, "layer", "serving")
		return nil, fmt.Errorf("elect storage account: %w", err)
	}
	span.SetAttributes(
		attribute.Int64("storage.account_id", account.ID),
		attribute.String("storage.provider", string(account.Provider)),
	)

	// Generate storage key
	assetID := uuid.New()
	key := fmt.Sprintf("%s/%s", assetID.String(), input.Filename)

	// Upload to provider AND record storage_locations row
	_, err = s.uploader.UploadAndRecord(ctx, assetID, account.ID, db.StorageLayerServing, key, bytes.NewReader(content), actualSize, input.ContentType)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "upload to storage failed")
		slog.ErrorContext(ctx, "upload to storage failed", "error", err, "account_id", account.ID)
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
		span.RecordError(err)
		span.SetStatus(codes.Error, "create asset record failed")
		return nil, fmt.Errorf("create asset record: %w", err)
	}

	// Create reference
	_, err = s.repo.CreateReference(ctx, asset.ID, input.SpaceID, input.FolderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "create reference failed")
		return nil, fmt.Errorf("create asset reference: %w", err)
	}

	// Increment space usage
	_ = s.quota.IncrementUsage(ctx, input.SpaceID, actualSize)

	// Enqueue archive job
	if err := s.jobCreator.CreateArchiveJob(ctx, asset.ID); err != nil {
		span.AddEvent("archive_job_failed")
		slog.WarnContext(ctx, "archive job creation failed", "error", err, "asset_id", asset.ID)
		return asset, fmt.Errorf("asset uploaded but archive job failed: %w", err)
	}

	// --- Metrics + logs ---
	duration := float64(time.Since(start).Milliseconds())
	s.metrics.UploadsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", string(account.Provider)),
		attribute.String("mime_type", input.ContentType),
		attribute.Bool("dedup_hit", false),
	))
	s.metrics.UploadsBytes.Add(ctx, actualSize, metric.WithAttributes(
		attribute.String("provider", string(account.Provider)),
	))
	s.metrics.UploadsDuration.Record(ctx, duration, metric.WithAttributes(
		attribute.String("provider", string(account.Provider)),
	))

	slog.InfoContext(ctx, "asset uploaded",
		"asset_id", asset.ID,
		"size_bytes", actualSize,
		"provider", account.Provider,
		"account_id", account.ID,
		"duration_ms", duration,
	)

	return asset, nil
}

func (s *Service) CreateReference(ctx context.Context, req CreateReferenceRequest) (*db.AssetReference, error) {
	return s.repo.CreateReference(ctx, req.AssetID, req.SpaceID, req.FolderID)
}

func (s *Service) DeleteReference(ctx context.Context, refID int64, spaceID uuid.UUID, sizeBytes int64) error {
	if err := s.repo.SoftDeleteReference(ctx, refID); err != nil {
		return err
	}
	_ = s.quota.DecrementUsage(ctx, spaceID, sizeBytes)
	return nil
}

func (s *Service) RestoreReference(ctx context.Context, refID int64) error {
	return s.repo.RestoreReference(ctx, refID)
}

func (s *Service) RenameAsset(ctx context.Context, id uuid.UUID, name string) error {
	return s.repo.UpdateName(ctx, id, name)
}
