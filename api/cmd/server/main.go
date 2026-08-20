package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// --- Crypto ---
	encryptFn := func(plaintext []byte) ([]byte, error) {
		if cfg.EncryptionKey == "" {
			return plaintext, nil
		}
		return lib.Encrypt(plaintext, cfg.EncryptionKey)
	}
	decryptFn := func(ciphertext []byte) ([]byte, error) {
		if cfg.EncryptionKey == "" {
			return ciphertext, nil
		}
		return lib.Decrypt(ciphertext, cfg.EncryptionKey)
	}

	// --- Compose modules ---

	// Account
	accountRepo := account.NewRepository(queries)
	accountService := account.NewService(accountRepo)
	accountWebhook := account.NewWebhookHandler(accountService, cfg.ClerkWebhookSecret)

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
	storageRegistry := storage.NewRegistry(storageRepo, decryptFn)
	storageUploader := storage.NewUploader(storageRegistry, storageRepo)

	// Archive worker
	archiveWorker := storage.NewWorker(queries, storageRepo, storageService, storageRegistry, logger)

	// Assets
	assetRepo := asset.NewRepository(queries)
	assetService := asset.NewService(assetRepo, storageService, storageUploader, archiveWorker, spaceService)
	assetHandler := asset.NewHandler(assetService, spaceService, logger)

	// --- Server ---
	srv := server.New(logger)

	// Public routes
	srv.App.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	account.RegisterRoutes(srv.App, accountWebhook)

	// Protected routes
	api := srv.App.Group("/api/v1", auth.Middleware(accountService))
	space.RegisterRoutes(api, spaceHandler)
	folder.RegisterRoutes(api, folderHandler)
	tag.RegisterRoutes(api, tagHandler)
	session.RegisterRoutes(api, sessionHandler)
	storage.RegisterRoutes(api, storageHandler)
	asset.RegisterRoutes(api, assetHandler)

	// --- Archive worker (background) ---
	go archiveWorker.Run(ctx, 10*time.Second)

	// --- Graceful shutdown ---
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Info().Msg("shutting down")
		cancel()
		_ = srv.App.Shutdown()
	}()

	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info().Str("addr", addr).Msg("starting server")
	if err := srv.App.Listen(addr); err != nil {
		logger.Fatal().Err(err).Msg("server failed")
	}
}
