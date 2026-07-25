// Command migrate runs SQL migration files against the database using pgx.
// No external tools (psql, migrate CLI) required — just Go and a DATABASE_URL.
//
// Usage:
//
//	go run cmd/migrate/main.go                              # runs all pending migrations
//	go run cmd/migrate/main.go internal/database/migrations/001_galleries_to_spaces.sql  # runs a specific file
//
// Reads DATABASE_URL from the environment (or .env file via godotenv).
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const migrationsDir = "internal/database/migrations"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal().Msg("DATABASE_URL is required (set in environment or .env file)")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to ping database")
	}
	log.Info().Msg("database connected")

	// Determine which files to run.
	files, err := resolveFiles(os.Args[1:])
	if err != nil {
		log.Fatal().Err(err).Msg("failed to resolve migration files")
	}

	if len(files) == 0 {
		log.Info().Msg("no migration files to run")
		return
	}

	for _, file := range files {
		log.Info().Str("file", file).Msg("executing migration")

		sql, err := os.ReadFile(file)
		if err != nil {
			log.Fatal().Err(err).Str("file", file).Msg("failed to read migration file")
		}

		start := time.Now()
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			log.Fatal().Err(err).Str("file", file).Msg("migration failed")
		}

		log.Info().
			Str("file", file).
			Dur("duration", time.Since(start)).
			Msg("migration completed")
	}

	log.Info().Int("count", len(files)).Msg("all migrations applied successfully")
}

// resolveFiles returns the list of SQL files to execute. If args are provided,
// they are used as explicit file paths. Otherwise, all .sql files in the
// migrations directory are collected and sorted alphabetically.
func resolveFiles(args []string) ([]string, error) {
	if len(args) > 0 {
		// Explicit files passed as arguments.
		for _, f := range args {
			if _, err := os.Stat(f); err != nil {
				return nil, fmt.Errorf("file not found: %s", f)
			}
		}
		return args, nil
	}

	// Auto-discover from migrations directory.
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".sql" {
			continue
		}
		files = append(files, filepath.Join(migrationsDir, e.Name()))
	}
	sort.Strings(files)
	return files, nil
}
