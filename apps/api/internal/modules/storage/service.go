package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/jackc/pgx/v5"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/asset"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/storage/adapters"
)

var (
	ErrProviderNotFound   = errors.New("provider not found")
	ErrInvalidProviderType = errors.New("invalid provider type")
	ErrNoActiveProvider   = errors.New("no active storage provider available")
)

type Service struct {
	repo        *Repository
	assetRepo   *asset.Repository
	assetService *asset.Service
}

func NewService(repo *Repository, assetRepo *asset.Repository, assetService *asset.Service) *Service {
	return &Service{
		repo:         repo,
		assetRepo:    assetRepo,
		assetService: assetService,
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

// SelectProvider selects an available provider using quota-aware selection
func (s *Service) SelectProvider(ctx context.Context, userID string) (*Provider, error) {
	providers, err := s.repo.ListActiveProviders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	if len(providers) == 0 {
		return nil, ErrNoActiveProvider
	}

	// Quota-aware selection: prefer provider with most available space
	var selected *Provider
	var maxAvailable int64 = -1

	for _, p := range providers {
		// If no quota set, treat as unlimited
		available := int64(9223372036854775807) // Max int64
		if p.Quota != nil && *p.Quota > 0 {
			available = *p.Quota - p.Used
			if available <= 0 {
				continue // Skip providers with no space
			}
		}

		if available > maxAvailable {
			maxAvailable = available
			selected = p
		}
	}

	if selected == nil {
		return nil, ErrNoActiveProvider
	}

	return selected, nil
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

// UploadFile handles the complete file upload workflow
func (s *Service) UploadFile(ctx context.Context, userID string, file io.Reader, filename string, size int64) (*asset.Asset, error) {
	// Read file into memory for hashing and uploading
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Calculate hash
	hash := lib.HashBytes(fileBytes)

	// Detect MIME type
	mimeType := lib.DetectMimeType(filename)
	assetType := string(lib.GetAssetType(mimeType))

	// Check for duplicate
	existingAsset, err := s.assetRepo.GetByHash(ctx, hash, userID)
	if err == nil && existingAsset != nil {
		// Asset with same hash already exists, return it
		return existingAsset, nil
	}

	// Select storage provider
	provider, err := s.SelectProvider(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create adapter
	adapter, err := s.CreateAdapter(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// Upload to provider
	uploadInput := &adapters.UploadInput{
		File:     bytes.NewReader(fileBytes),
		Filename: filename,
		MimeType: mimeType,
		Size:     size,
	}

	uploadResult, err := adapter.Upload(ctx, uploadInput)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to provider: %w", err)
	}

	// Create asset
	assetReq := &asset.CreateAssetRequest{
		Name:     filename,
		Type:     assetType,
		MimeType: mimeType,
		Size:     uploadResult.Size,
		Hash:     hash,
		Tags:     []string{},
		Metadata: uploadResult.Metadata,
	}

	newAsset, err := s.assetService.CreateAsset(ctx, userID, assetReq)
	if err != nil {
		// Try to cleanup uploaded file
		_ = adapter.Delete(ctx, uploadResult.Key)
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	// Create storage location
	location := &asset.StorageLocation{
		AssetID:     newAsset.ID,
		ProviderID:  provider.ID,
		ProviderKey: uploadResult.Key,
		URL:         uploadResult.URL,
		Metadata:    uploadResult.Metadata,
	}

	_, err = s.assetService.CreateLocation(ctx, location)
	if err != nil {
		// Try to cleanup
		_ = adapter.Delete(ctx, uploadResult.Key)
		_ = s.assetService.DeleteAsset(ctx, newAsset.ID, userID)
		return nil, fmt.Errorf("failed to create location: %w", err)
	}

	// Update provider usage
	newUsage := provider.Used + uploadResult.Size
	if err := s.repo.UpdateProviderUsage(ctx, provider.ID, newUsage); err != nil {
		// Non-fatal, log and continue
		fmt.Printf("Warning: failed to update provider usage: %v\n", err)
	}

	// TODO: Update user quota (Phase 5 enhancement)

	// Return asset with location
	newAsset.Locations = []*asset.StorageLocation{location}
	return newAsset, nil
}

// DownloadFile handles the complete file download workflow
func (s *Service) DownloadFile(ctx context.Context, assetID, userID string) (io.ReadCloser, error) {
	// Get asset metadata
	assetData, err := s.assetService.GetByID(ctx, assetID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	// Get storage location (prefer first available)
	if len(assetData.Locations) == 0 {
		return nil, fmt.Errorf("no storage location found for asset")
	}

	location := assetData.Locations[0]

	// Get provider
	provider, err := s.GetProviderByID(ctx, location.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage provider: %w", err)
	}

	// Create adapter
	adapter, err := s.CreateAdapter(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// Download from provider
	rc, err := adapter.Download(ctx, location.ProviderKey)
	if err != nil {
		return nil, fmt.Errorf("failed to download from provider: %w", err)
	}

	return rc, nil
}
