package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"

	"github.com/rizqyn9/filora-dam/api/internal/config"
	"github.com/rizqyn9/filora-dam/api/internal/database"
	"github.com/rizqyn9/filora-dam/api/internal/database/db"
	"github.com/rizqyn9/filora-dam/api/internal/modules/asset"
	"github.com/rizqyn9/filora-dam/api/internal/modules/folder"
	"github.com/rizqyn9/filora-dam/api/internal/modules/session"
	"github.com/rizqyn9/filora-dam/api/internal/modules/space"
	"github.com/rizqyn9/filora-dam/api/internal/modules/storage"
	"github.com/rizqyn9/filora-dam/api/internal/modules/tag"
	"github.com/rizqyn9/filora-dam/api/internal/server"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

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

	// --- Compose modules ---

	// Spaces
	spaceRepo := space.NewRepository(queries)
	spaceService := space.NewService(spaceRepo)
	spaceHandler := space.NewHandler(spaceService)

	// Folders
	folderRepo := folder.NewRepository(queries)
	folderService := folder.NewService(folderRepo)
	folderHandler := folder.NewHandler(folderService)

	// Tags
	tagRepo := tag.NewRepository(queries)
	tagService := tag.NewService(tagRepo)
	tagHandler := tag.NewHandler(tagService)

	// Sessions
	sessionRepo := session.NewRepository(queries)
	sessionService := session.NewService(sessionRepo)
	sessionHandler := session.NewHandler(sessionService)

	// Storage
	storageRepo := storage.NewRepository(queries)
	storageService := storage.NewService(storageRepo, encryptCredentials)
	storageHandler := storage.NewHandler(storageService)

	// Archive worker (also implements JobCreator interface)
	archiveWorker := storage.NewWorker(queries, storageRepo, logger)

	// Assets
	assetRepo := asset.NewRepository(queries)
	assetService := asset.NewService(assetRepo, storageService, nil, archiveWorker) // uploader=nil until adapter wired
	assetHandler := asset.NewHandler(assetService)

	// --- Server ---
	srv := server.New(logger)

	api := srv.App.Group("/api/v1")
	space.RegisterRoutes(api, spaceHandler)
	folder.RegisterRoutes(api, folderHandler)
	tag.RegisterRoutes(api, tagHandler)
	session.RegisterRoutes(api, sessionHandler)
	storage.RegisterRoutes(api, storageHandler)
	asset.RegisterRoutes(api, assetHandler)

	// Health check
	srv.App.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// --- Graceful shutdown ---
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Info().Msg("shutting down server")
		cancel()
		_ = srv.App.Shutdown()
	}()

	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info().Str("addr", addr).Msg("starting server")
	if err := srv.App.Listen(addr); err != nil {
		logger.Fatal().Err(err).Msg("server failed")
	}
}

// ponytail: placeholder encryption. Upgrade path: AES-GCM with key from env/vault.
func encryptCredentials(plaintext []byte) ([]byte, error) {
	// TODO: implement real encryption with ENCRYPTION_KEY from env
	return plaintext, nil
}
