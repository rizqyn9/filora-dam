package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"

	"github.com/rizqyn9/filora-dam/api/internal/auth"
	"github.com/rizqyn9/filora-dam/api/internal/config"
	"github.com/rizqyn9/filora-dam/api/internal/database"
	"github.com/rizqyn9/filora-dam/api/internal/database/db"
	"github.com/rizqyn9/filora-dam/api/internal/lib"
	"github.com/rizqyn9/filora-dam/api/internal/modules/account"
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

	// --- Encryption ---
	encryptFn := func(plaintext []byte) ([]byte, error) {
		if cfg.EncryptionKey == "" {
			// Development fallback: no encryption
			return plaintext, nil
		}
		return lib.Encrypt(plaintext, cfg.EncryptionKey)
	}

	// --- Compose modules ---

	// Account (Clerk user sync + auth resolver)
	accountRepo := account.NewRepository(queries)
	accountService := account.NewService(accountRepo)
	accountWebhook := account.NewWebhookHandler(accountService)

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
	storageService := storage.NewService(storageRepo, encryptFn)
	storageHandler := storage.NewHandler(storageService)

	// Archive worker (also implements JobCreator)
	archiveWorker := storage.NewWorker(queries, storageRepo, logger)

	// Assets
	assetRepo := asset.NewRepository(queries)
	assetService := asset.NewService(assetRepo, storageService, nil, archiveWorker)
	assetHandler := asset.NewHandler(assetService)

	// --- Server ---
	srv := server.New(logger)

	// Public routes (no auth)
	srv.App.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	account.RegisterRoutes(srv.App, accountWebhook)

	// Protected routes (auth required)
	api := srv.App.Group("/api/v1", auth.Middleware(accountService))

	space.RegisterRoutes(api, spaceHandler)
	folder.RegisterRoutes(api, folderHandler)
	tag.RegisterRoutes(api, tagHandler)
	session.RegisterRoutes(api, sessionHandler)
	storage.RegisterRoutes(api, storageHandler)
	asset.RegisterRoutes(api, assetHandler)

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
