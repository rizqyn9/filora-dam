package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port          string
	DatabaseURL   string
	Environment   string
	EncryptionKey string // hex-encoded 32-byte AES key (optional in dev)

	// Axiom OTel
	AxiomEndpoint       string // default: api.axiom.co
	AxiomToken          string // Axiom API token
	AxiomDataset        string // dataset for traces + logs
	AxiomMetricsDataset string // dataset for metrics (separate per Axiom docs)
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                envOrDefault("PORT", "8080"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		Environment:         envOrDefault("ENVIRONMENT", "development"),
		EncryptionKey:       os.Getenv("ENCRYPTION_KEY"),
		AxiomEndpoint:       envOrDefault("AXIOM_ENDPOINT", "api.axiom.co"),
		AxiomToken:          os.Getenv("AXIOM_TOKEN"),
		AxiomDataset:        envOrDefault("AXIOM_DATASET", "filora"),
		AxiomMetricsDataset: envOrDefault("AXIOM_METRICS_DATASET", "filora-metrics"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
