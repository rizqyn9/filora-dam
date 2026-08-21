package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

// Worker processes archive sync jobs in a loop.
type Worker struct {
	queries  *db.Queries
	repo     *Repository
	service  *Service
	registry *Registry
	logger   zerolog.Logger
}

func NewWorker(queries *db.Queries, repo *Repository, service *Service, registry *Registry, logger zerolog.Logger) *Worker {
	return &Worker{
		queries:  queries,
		repo:     repo,
		service:  service,
		registry: registry,
		logger:   logger,
	}
}

// Run polls for pending archive jobs and processes them.
func (w *Worker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info().Msg("archive worker stopping")
			return
		case <-ticker.C:
			w.processNext(ctx)
		}
	}
}

func (w *Worker) processNext(ctx context.Context) {
	job, err := w.queries.ClaimNextArchiveJob(ctx)
	if err != nil {
		return // no jobs available
	}

	w.logger.Info().Int64("job_id", job.ID).Str("asset_id", job.AssetID.String()).Msg("processing archive job")

	err = w.archiveAsset(ctx, job.AssetID)
	if err != nil {
		w.logger.Error().Err(err).Int64("job_id", job.ID).Msg("archive job failed")
		nextRetry := time.Now().Add(backoff(int(job.Attempts)))
		errMsg := err.Error()
		_ = w.queries.FailArchiveJob(ctx, db.FailArchiveJobParams{
			ID:          job.ID,
			LastError:   &errMsg,
			NextRetryAt: pgtype.Timestamptz{Time: nextRetry, Valid: true},
		})
		return
	}

	_ = w.queries.CompleteArchiveJob(ctx, job.ID)
	w.logger.Info().Int64("job_id", job.ID).Msg("archive job completed")
}

func (w *Worker) archiveAsset(ctx context.Context, assetID uuid.UUID) error {
	ctx, span := otel.Tracer("filora-api").Start(ctx, "archive.sync_asset")
	defer span.End()

	span.SetAttributes(attribute.String("asset.id", assetID.String()))

	// 1. Get the serving location
	servingLoc, err := w.repo.GetServingLocation(ctx, assetID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "get serving location failed")
		return fmt.Errorf("get serving location: %w", err)
	}
	if servingLoc.RemotePath == nil {
		err := fmt.Errorf("serving location has no remote path")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// 2. Get serving adapter and download the file
	servingAdapter, err := w.registry.Get(ctx, servingLoc.AccountID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "get serving adapter failed")
		return fmt.Errorf("get serving adapter: %w", err)
	}

	body, err := servingAdapter.Download(ctx, *servingLoc.RemotePath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "download from serving failed")
		return fmt.Errorf("download from serving: %w", err)
	}
	defer func() { _ = body.Close() }()

	// 3. Get the asset record for size info
	asset, err := w.queries.GetAssetByID(ctx, assetID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("get asset: %w", err)
	}

	// 4. Elect an archive account
	archiveAccount, err := w.service.ElectAccount(ctx, db.StorageLayerArchive, asset.SizeBytes)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "elect archive account failed")
		return fmt.Errorf("elect archive account: %w", err)
	}
	span.SetAttributes(
		attribute.Int64("archive.account_id", archiveAccount.ID),
		attribute.String("archive.provider", string(archiveAccount.Provider)),
	)

	// 5. Get archive adapter and upload
	archiveAdapter, err := w.registry.Get(ctx, archiveAccount.ID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("get archive adapter: %w", err)
	}

	key := fmt.Sprintf("archive/%s/%s", assetID.String(), asset.OriginalFilename)
	result, err := archiveAdapter.Upload(ctx, UploadInput{
		Key:         key,
		Body:        body,
		Size:        asset.SizeBytes,
		ContentType: asset.MimeType,
	})
	if err != nil {
		_, _ = w.repo.CreateLocation(ctx, db.CreateStorageLocationParams{
			AssetID:   assetID,
			AccountID: archiveAccount.ID,
			Layer:     db.StorageLayerArchive,
			Status:    db.LocationStatusFailed,
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "upload to archive failed")
		slog.ErrorContext(ctx, "archive upload failed", "error", err, "asset_id", assetID, "account_id", archiveAccount.ID)
		return fmt.Errorf("upload to archive: %w", err)
	}

	// 6. Record successful archive location
	_, err = w.repo.CreateLocation(ctx, db.CreateStorageLocationParams{
		AssetID:    assetID,
		AccountID:  archiveAccount.ID,
		Layer:      db.StorageLayerArchive,
		Status:     db.LocationStatusStored,
		RemotePath: &result.RemotePath,
		RemoteUrl:  &result.RemoteURL,
	})
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("record archive location: %w", err)
	}

	_ = w.repo.IncrementUsage(ctx, archiveAccount.ID, asset.SizeBytes)

	slog.InfoContext(ctx, "asset archived",
		"asset_id", assetID,
		"archive_account_id", archiveAccount.ID,
		"size_bytes", asset.SizeBytes,
	)

	return nil
}

// CreateArchiveJob implements asset.JobCreator.
func (w *Worker) CreateArchiveJob(ctx context.Context, assetID uuid.UUID) error {
	_, err := w.queries.CreateArchiveSyncJob(ctx, assetID)
	return err
}

// ponytail: exponential backoff, capped at 1 hour. Upgrade: configurable per-job policy.
func backoff(attempt int) time.Duration {
	d := time.Duration(1<<uint(attempt)) * time.Minute
	if d > time.Hour {
		d = time.Hour
	}
	return d
}
