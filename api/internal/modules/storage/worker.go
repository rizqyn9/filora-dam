package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rizqyn9/filora-dam/api/internal/database/db"
	"github.com/rs/zerolog"
)

// Worker processes archive sync jobs in a loop.
type Worker struct {
	queries *db.Queries
	repo    *Repository
	logger  zerolog.Logger
}

func NewWorker(queries *db.Queries, repo *Repository, logger zerolog.Logger) *Worker {
	return &Worker{queries: queries, repo: repo, logger: logger}
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
		// No jobs available — not an error
		return
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
	// TODO: implement actual archive copy
	// 1. Get serving location for asset
	// 2. Elect archive account
	// 3. Download from serving → upload to archive
	// 4. Create storage_location record for archive layer
	_ = assetID
	return fmt.Errorf("archive not implemented")
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
