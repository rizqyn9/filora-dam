package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rizqynugroho9/filora-dam/api/internal/config"
	"github.com/rizqynugroho9/filora-dam/api/internal/database"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/storage"
)

const (
	pollInterval = 10 * time.Second
	maxBackoff   = 5 * time.Minute
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}
	if cfg.IsProduction() {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	db, err := database.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	storageSvc := storage.NewService(storage.NewRepository(db.Pool))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info().Dur("poll", pollInterval).Msg("archive worker started")

	for {
		if ctx.Err() != nil {
			log.Info().Msg("archive worker stopped")
			return
		}

		processed := processOne(ctx, storageSvc)
		if processed {
			continue // drain the queue as fast as it fills
		}

		select {
		case <-ctx.Done():
			log.Info().Msg("archive worker stopped")
			return
		case <-time.After(pollInterval):
		}
	}
}

// processOne claims and handles a single job. Returns true if a job was claimed.
func processOne(ctx context.Context, svc *storage.Service) bool {
	job, err := svc.ClaimJob(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to claim archive job")
		return false
	}
	if job == nil {
		return false
	}

	if err := svc.ReplicateToArchive(ctx, job.AssetID); err != nil {
		if job.Attempts >= job.MaxAttempts {
			_ = svc.FailJob(ctx, job.ID, err)
			log.Error().Int64("job", job.ID).Str("asset", job.AssetID.String()).Err(err).
				Msg("archive job failed permanently")
			return true
		}
		backoff := time.Duration(1<<uint(job.Attempts)) * time.Second
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
		_ = svc.RetryJob(ctx, job.ID, err, backoff)
		log.Warn().Int64("job", job.ID).Dur("retry_in", backoff).Err(err).
			Msg("archive job will retry")
		return true
	}

	_ = svc.CompleteJob(ctx, job.ID)
	log.Info().Int64("job", job.ID).Str("asset", job.AssetID.String()).
		Msg("archive job completed")
	return true
}
