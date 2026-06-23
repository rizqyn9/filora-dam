package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/storage/adapters"
)

var (
	ErrProviderNotFound   = errors.New("provider not found")
	ErrInvalidProviderType = errors.New("invalid provider type")
	ErrNoActiveProvider   = errors.New("no active storage provider available")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GetProviderByID retrieves a provider by ID
func (s *Service) GetProviderByID(ctx context.Context, id string) (*Provider, error) {
	provider, err := s.repo.GetProviderByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProviderNotFound
		}
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}
	return provider, nil
}

// ListProviders lists all providers for a user
func (s *Service) ListProviders(ctx context.Context, userID string, activeOnly bool) ([]*Provider, error) {
	var providers []*Provider
	var err error

	if activeOnly {
		providers, err = s.repo.ListActiveProviders(ctx, userID)
	} else {
		providers, err = s.repo.ListAllProviders(ctx, userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	return providers, nil
}

// CreateProvider creates a new storage provider
func (s *Service) CreateProvider(ctx context.Context, userID string, req *CreateProviderRequest) (*Provider, error) {
	// Validate provider type
	if !isValidProviderType(req.Type) {
		return nil, ErrInvalidProviderType
	}

	// Validate credentials by creating adapter (stub check)
	if err := s.validateProviderCredentials(req.Type, req.Credentials); err != nil {
		return nil, err
	}

	provider, err := s.repo.CreateProvider(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, nil
}

// DeactivateProvider deactivates a storage provider
func (s *Service) DeactivateProvider(ctx context.Context, providerID string) error {
	if err := s.repo.DeactivateProvider(ctx, providerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProviderNotFound
		}
		return fmt.Errorf("failed to deactivate provider: %w", err)
	}
	return nil
}

// SelectProvider selects an available provider for upload (simple round-robin for now)
func (s *Service) SelectProvider(ctx context.Context, userID string) (*Provider, error) {
	providers, err := s.repo.ListActiveProviders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	if len(providers) == 0 {
		return nil, ErrNoActiveProvider
	}

	// Simple strategy: return first active provider
	// TODO: Implement better selection strategy in Phase 7
	return providers[0], nil
}

// CreateAdapter creates a storage adapter for a provider
func (s *Service) CreateAdapter(provider *Provider) (adapters.StorageAdapter, error) {
	config := &adapters.AdapterConfig{
		Credentials: provider.Credentials,
	}

	switch provider.Type {
	case "cloudinary":
		return adapters.NewCloudinaryAdapter(config)
	case "imagekit":
		return adapters.NewImageKitAdapter(config)
	case "r2":
		return adapters.NewR2Adapter(config)
	default:
		return nil, ErrInvalidProviderType
	}
}

// validateProviderCredentials validates provider credentials by attempting to create an adapter
func (s *Service) validateProviderCredentials(providerType string, credentials map[string]interface{}) error {
	config := &adapters.AdapterConfig{
		Credentials: credentials,
	}

	switch providerType {
	case "cloudinary":
		_, err := adapters.NewCloudinaryAdapter(config)
		return err
	case "imagekit":
		_, err := adapters.NewImageKitAdapter(config)
		return err
	case "r2":
		_, err := adapters.NewR2Adapter(config)
		return err
	default:
		return ErrInvalidProviderType
	}
}

func isValidProviderType(providerType string) bool {
	validTypes := map[string]bool{
		"cloudinary": true,
		"imagekit":   true,
		"r2":         true,
	}
	return validTypes[providerType]
}
