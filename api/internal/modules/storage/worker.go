package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/rizqynugroho9/filora-dam/api/internal/modules/storage/adapters"
)

// Job is a claimed archive-replication job.
type Job struct {
	ID          int64
	AssetID     uuid.UUID
	Attempts    int32
	MaxAttempts int32
}

// ClaimJob atomically claims the next due archive job, or returns nil if none.
func (s *Service) ClaimJob(ctx context.Context) (*Job, error) {
	row, ok, err := s.repo.ClaimArchiveJob(ctx)
	if err != nil || !ok {
		return nil, err
	}
	return &Job{ID: row.ID, AssetID: row.AssetID, Attempts: row.Attempts, MaxAttempts: row.MaxAttempts}, nil
}

func (s *Service) CompleteJob(ctx context.Context, id int64) error {
	return s.repo.MarkJobResult(ctx, id, "completed", nil, nil)
}

// RetryJob reschedules a job after a backoff delay.
func (s *Service) RetryJob(ctx context.Context, id int64, cause error, backoff time.Duration) error {
	msg := cause.Error()
	next := time.Now().Add(backoff)
	return s.repo.MarkJobResult(ctx, id, "pending", &msg, &next)
}

// FailJob marks a job permanently failed.
func (s *Service) FailJob(ctx context.Context, id int64, cause error) error {
	msg := cause.Error()
	return s.repo.MarkJobResult(ctx, id, "failed", &msg, nil)
}

// ReplicateToArchive copies an asset's serving copy into an active archive
// account and records an archive storage_location.
func (s *Service) ReplicateToArchive(ctx context.Context, assetID uuid.UUID) error {
	src, ok, err := s.repo.GetArchiveSource(ctx, assetID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no serving source for asset %s", assetID)
	}

	srcProvider, err := s.repo.GetRaw(ctx, src.ProviderID)
	if err != nil {
		return err
	}
	srcAdapter, err := adapters.NewAdapter(srcProvider.Type, srcProvider.Credentials)
	if err != nil {
		return err
	}
	reader, err := srcAdapter.Download(ctx, src.ProviderKey)
	if err != nil {
		return fmt.Errorf("download serving source: %w", err)
	}
	defer reader.Close()

	accounts, err := s.repo.ListActiveByLayer(ctx, "archive")
	if err != nil {
		return err
	}
	if len(accounts) == 0 {
		return fmt.Errorf("no active archive storage account configured")
	}
	// Account election is a backlog concern; use the first active account.
	arch := accounts[0]

	archAdapter, err := adapters.NewAdapter(arch.Type, arch.Credentials)
	if err != nil {
		return err
	}
	res, err := archAdapter.Upload(ctx, adapters.UploadInput{
		Key:         src.ProviderKey,
		ContentType: src.MimeType,
		Size:        src.Size,
		Reader:      reader,
	})
	if err != nil {
		return fmt.Errorf("upload to archive: %w", err)
	}

	var urlPtr *string
	if res.URL != "" {
		urlPtr = &res.URL
	}
	if err := s.repo.CreateLocation(ctx, assetID, arch.ID, "archive", res.Key, urlPtr, "stored"); err != nil {
		return err
	}
	_ = s.repo.AddUsed(ctx, arch.ID, src.Size)
	return nil
}
