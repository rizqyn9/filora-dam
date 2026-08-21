package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	iauth "github.com/rizqyn9/filora-dam/api/internal/auth"
	"github.com/rizqyn9/filora-dam/api/internal/config"
	"github.com/rizqyn9/filora-dam/api/internal/database"
	"github.com/rizqyn9/filora-dam/api/internal/database/db"
	"github.com/rizqyn9/filora-dam/api/internal/lib"
	"github.com/rizqyn9/filora-dam/api/internal/modules/asset"
	authmod "github.com/rizqyn9/filora-dam/api/internal/modules/auth"
	"github.com/rizqyn9/filora-dam/api/internal/modules/folder"
	"github.com/rizqyn9/filora-dam/api/internal/modules/space"
	"github.com/rizqyn9/filora-dam/api/internal/modules/storage"
	"github.com/rizqyn9/filora-dam/api/internal/modules/tag"
	"github.com/rizqyn9/filora-dam/api/internal/server"
	"github.com/rizqyn9/filora-dam/api/internal/telemetry"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- OpenTelemetry ---
	otelShutdown, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName:    "filora-api",
		ServiceVersion: "1.0.0",
		Environment:    cfg.Environment,
		Endpoint:       cfg.AxiomEndpoint,
		Token:          cfg.AxiomToken,
		Dataset:        cfg.AxiomDataset,
		MetricsDataset: cfg.AxiomMetricsDataset,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init otel")
	}
	defer func() { _ = otelShutdown(context.Background()) }()

	// Route slog → OTel logs (sent to Axiom alongside traces + metrics)
	if cfg.AxiomToken != "" {
		telemetry.SetupSlog()
	}

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

	// --- Auth ---
	tokenCache := iauth.NewTokenCache()
	appMetrics := telemetry.NewMetrics()
	authService := authmod.NewService(queries, tokenCache, appMetrics)
	authHandler := authmod.NewHandler(authService)

	// --- Modules ---
	spaceRepo := space.NewRepository(queries)
	spaceService := space.NewService(spaceRepo)
	spaceHandler := space.NewHandler(spaceService)

	folderRepo := folder.NewRepository(queries)
	folderService := folder.NewService(folderRepo)
	folderHandler := folder.NewHandler(folderService)

	tagRepo := tag.NewRepository(queries)
	tagService := tag.NewService(tagRepo)
	tagHandler := tag.NewHandler(tagService)

	storageRepo := storage.NewRepository(queries)
	storageService := storage.NewService(storageRepo, encryptFn)
	storageHandler := storage.NewHandler(storageService)
	storageRegistry := storage.NewRegistry(storageRepo, decryptFn)
	storageUploader := storage.NewUploader(storageRegistry, storageRepo)

	archiveWorker := storage.NewWorker(queries, storageRepo, storageService, storageRegistry, logger)

	assetRepo := asset.NewRepository(queries)
	assetService := asset.NewService(assetRepo, storageService, storageUploader, archiveWorker, spaceService, appMetrics)
	assetHandler := asset.NewHandler(assetService, spaceService, logger)

	// --- Server ---
	srv := server.New(logger)

	// OTel tracing middleware (first in chain)
	srv.App.Use(telemetry.FiberMiddleware())

	// Public routes (no auth)
	srv.App.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	authmod.RegisterRoutes(srv.App, authHandler)

	// Protected routes
	authMiddleware := iauth.Middleware(tokenCache, authService, authmod.IdleTTL)
	api := srv.App.Group("/api/v1", authMiddleware)

	authmod.RegisterProtectedRoutes(api, authHandler)
	space.RegisterRoutes(api, spaceHandler)
	folder.RegisterRoutes(api, folderHandler)
	tag.RegisterRoutes(api, tagHandler)
	storage.RegisterRoutes(api, storageHandler)
	asset.RegisterRoutes(api, assetHandler)

	// --- Archive worker ---
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
