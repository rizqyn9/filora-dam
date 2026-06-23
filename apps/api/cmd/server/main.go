package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/rizqynugroho9/filora-dam/api/internal/config"
	"github.com/rizqynugroho9/filora-dam/api/internal/database"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
	"github.com/rizqynugroho9/filora-dam/api/internal/middleware"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/account"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/asset"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✓ Database connected")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Filora DAM API",
		ErrorHandler: customErrorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	// Initialize JWT manager
	jwtManager := lib.NewJWTManager(cfg.JWTSecret)

	// Initialize auth middleware
	authMiddleware := middleware.AuthMiddleware(jwtManager)

	// Initialize modules
	accountRepo := account.NewRepository(db.Pool)
	accountService := account.NewService(accountRepo, jwtManager)
	accountHandler := account.NewHandler(accountService)

	storageRepo := storage.NewRepository(db.Pool)
	storageService := storage.NewService(storageRepo)
	storageHandler := storage.NewHandler(storageService)

	assetRepo := asset.NewRepository(db.Pool)
	assetService := asset.NewService(assetRepo)
	assetHandler := asset.NewHandler(assetService)

	// Routes
	app.Get("/", func(c fiber.Ctx) error {
		return lib.Success(c, fiber.Map{
			"name":    "Filora DAM API",
			"version": "0.1.0",
			"status":  "healthy",
		})
	})

	app.Get("/health", func(c fiber.Ctx) error {
		return lib.Success(c, fiber.Map{
			"status": "ok",
		})
	})

	// Register module routes
	accountHandler.RegisterRoutes(app)
	storageHandler.RegisterRoutes(app, authMiddleware)
	assetHandler.RegisterRoutes(app, authMiddleware)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Filora DAM API starting on http://localhost:%s", cfg.Port)
	log.Printf("   Environment: %s", cfg.Environment)

	// Graceful shutdown
	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func customErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return lib.Error(c, code, "ERROR", message)
}
