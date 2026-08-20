package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/rizqyn9/filora-dam/api/internal/config"
	"github.com/rizqyn9/filora-dam/api/internal/database"
	"github.com/rizqyn9/filora-dam/api/internal/database/db"
	"github.com/rizqyn9/filora-dam/api/internal/lib"
	"github.com/rizqyn9/filora-dam/api/internal/modules/storage"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "worker").Logger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	queries := db.New(pool)

	decryptFn := func(ciphertext []byte) ([]byte, error) {
		if cfg.EncryptionKey == "" {
			return ciphertext, nil
		}
		return lib.Decrypt(ciphertext, cfg.EncryptionKey)
	}
	encryptFn := func(plaintext []byte) ([]byte, error) {
		if cfg.EncryptionKey == "" {
			return plaintext, nil
		}
		return lib.Encrypt(plaintext, cfg.EncryptionKey)
	}

	storageRepo := storage.NewRepository(queries)
	storageService := storage.NewService(storageRepo, encryptFn)
	storageRegistry := storage.NewRegistry(storageRepo, decryptFn)
	worker := storage.NewWorker(queries, storageRepo, storageService, storageRegistry, logger)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Info().Msg("shutting down worker")
		cancel()
	}()

	logger.Info().Msg("starting archive worker")
	worker.Run(ctx, 10*time.Second)
}
