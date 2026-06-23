package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Port        string `validate:"required"`
	Environment string `validate:"required,oneof=development production test"`
	DatabaseURL string `validate:"required,url"`
	JWTSecret   string `validate:"required,min=32"`

	// Storage providers
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string

	ImageKitPublicKey   string
	ImageKitPrivateKey  string
	ImageKitURLEndpoint string

	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
	R2Endpoint        string
}

func Load() (*Config, error) {
	// Load .env file if exists (development)
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "3000"),
		Environment: getEnv("ENV", "development"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),

		CloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET", ""),

		ImageKitPublicKey:   getEnv("IMAGEKIT_PUBLIC_KEY", ""),
		ImageKitPrivateKey:  getEnv("IMAGEKIT_PRIVATE_KEY", ""),
		ImageKitURLEndpoint: getEnv("IMAGEKIT_URL_ENDPOINT", ""),

		R2AccountID:       getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:      getEnv("R2_BUCKET_NAME", ""),
		R2Endpoint:        getEnv("R2_ENDPOINT", ""),
	}

	// Validate configuration
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
