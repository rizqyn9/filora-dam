package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/clerk"
	"github.com/rizqynugroho9/filora-dam/api/internal/config"
	"github.com/rizqynugroho9/filora-dam/api/internal/database"
	"github.com/rizqynugroho9/filora-dam/api/internal/middleware"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/account"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/rbac"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/session"
	"github.com/rizqynugroho9/filora-dam/api/internal/server"
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
	log.Info().Msg("database connected")

	// Modules
	accountRepo := account.NewRepository(db.Pool)
	accountSvc := account.NewService(accountRepo)
	accountHandler := account.NewHandler(accountSvc, cfg.ClerkWebhookSigningSecret)

	authorizer := auth.NewAuthorizer(db.Pool)

	rbacRepo := rbac.NewRepository(db.Pool)
	rbacSvc := rbac.NewService(rbacRepo)
	rbacHandler := rbac.NewHandler(rbacSvc, authorizer)

	sessionRepo := session.NewRepository(db.Pool)
	sessionSvc := session.NewService(sessionRepo, cfg.CLITokenTTLHours)
	sessionHandler := session.NewHandler(sessionSvc)

	// Auth: Clerk verifier is optional (nil rejects protected routes gracefully).
	var clerkVerifier middleware.ClerkVerifier
	if cfg.ClerkSecretKey != "" {
		clerkVerifier = clerk.NewVerifier(cfg.ClerkSecretKey)
	} else {
		log.Warn().Msg("CLERK_SECRET_KEY not set; protected routes will reject requests")
	}
	authMW := middleware.RequireAuth(middleware.AuthDeps{
		Clerk:    clerkVerifier,
		Sessions: sessionSvc,
		Accounts: accountSvc,
	})

	app := server.New(server.Deps{
		Config:  cfg,
		DB:      db,
		AuthMW:  authMW,
		Account: accountHandler,
		RBAC:    rbacHandler,
		Session: sessionHandler,
	})

	go func() {
		addr := ":" + cfg.Port
		log.Info().Str("addr", addr).Str("env", cfg.Env).Msg("server starting")
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
	log.Info().Msg("server stopped")
}
