package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

// Config holds all runtime configuration, loaded from the environment.
type Config struct {
	Port string `validate:"required"`
	Env  string `validate:"required,oneof=development production test"`

	DatabaseURL string `validate:"required"`

	// Clerk (web auth). Optional until the identity phase is wired.
	ClerkSecretKey            string
	ClerkWebhookSigningSecret string

	// CLI opaque token lifetime, in hours (0 = never expires).
	CLITokenTTLHours int
}

// IsProduction reports whether the app runs in the production environment.
func (c *Config) IsProduction() bool { return c.Env == "production" }

// Load reads configuration from the environment (and a .env file if present)
// and validates it.
func Load() (*Config, error) {
	// Best-effort: load .env in development. Missing file is not an error.
	_ = godotenv.Load()

	cfg := &Config{
		Port:                      getEnv("PORT", "3000"),
		Env:                       getEnv("ENV", "development"),
		DatabaseURL:               getEnv("DATABASE_URL", ""),
		ClerkSecretKey:            getEnv("CLERK_SECRET_KEY", ""),
		ClerkWebhookSigningSecret: getEnv("CLERK_WEBHOOK_SIGNING_SECRET", ""),
		CLITokenTTLHours:          getEnvInt("CLI_TOKEN_TTL_HOURS", 720),
	}

	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
