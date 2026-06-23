package asset

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	ErrAssetNotFound = errors.New("asset not found")
	ErrAccessDenied  = errors.New("access denied")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GetByID retrieves an asset by ID
func (s *Service) GetByID(ctx context.Context, id, userID string) (*Asset, error) {
	asset, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	// Verify ownership
	if asset.UserID != userID {
		return nil, ErrAccessDenied
	}

	// Load storage locations
	locations, err := s.repo.GetLocationsByAssetID(ctx, id)
	if err != nil {
		// Don't fail if locations can't be loaded
		locations = []*StorageLocation{}
	}
	asset.Locations = locations

	return asset, nil
}

// ListAssets lists assets for a user with pagination
func (s *Service) ListAssets(ctx context.Context, userID string, limit, offset int) (*AssetListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	assets, err := s.repo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}

	total, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count assets: %w", err)
	}

	return &AssetListResponse{
		Assets: assets,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// CreateAsset creates a new asset
func (s *Service) CreateAsset(ctx context.Context, userID string, req *CreateAssetRequest) (*Asset, error) {
	// Check for duplicate (same hash for same user)
	existing, err := s.repo.GetByHash(ctx, req.Hash, userID)
	if err == nil && existing != nil {
		// Asset with same hash already exists, return it
		return existing, nil
	}

	asset, err := s.repo.Create(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	return asset, nil
}

// UpdateTags updates asset tags
func (s *Service) UpdateTags(ctx context.Context, assetID, userID string, tags []string) error {
	// Verify ownership
	asset, err := s.repo.GetByID(ctx, assetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAssetNotFound
		}
		return fmt.Errorf("failed to get asset: %w", err)
	}

	if asset.UserID != userID {
		return ErrAccessDenied
	}

	if err := s.repo.UpdateTags(ctx, assetID, tags); err != nil {
		return fmt.Errorf("failed to update tags: %w", err)
	}

	return nil
}

// DeleteAsset deletes an asset and its locations
func (s *Service) DeleteAsset(ctx context.Context, assetID, userID string) error {
	// Verify ownership
	asset, err := s.repo.GetByID(ctx, assetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAssetNotFound
		}
		return fmt.Errorf("failed to get asset: %w", err)
	}

	if asset.UserID != userID {
		return ErrAccessDenied
	}

	// Delete asset (locations will be deleted via CASCADE)
	if err := s.repo.Delete(ctx, assetID); err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	// TODO: Delete from storage provider in Phase 5

	return nil
}

// CreateLocation creates a storage location for an asset
func (s *Service) CreateLocation(ctx context.Context, location *StorageLocation) (*StorageLocation, error) {
	loc, err := s.repo.CreateLocation(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create location: %w", err)
	}
	return loc, nil
}

// SearchAssets searches assets by name for a user
func (s *Service) SearchAssets(ctx context.Context, userID, query string, limit, offset int) (*AssetListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	assets, err := s.repo.SearchByName(ctx, userID, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search assets: %w", err)
	}

	// Count total matches
	total, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count assets: %w", err)
	}

	return &AssetListResponse{
		Assets: assets,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// FilterAssets filters assets by type for a user
func (s *Service) FilterAssets(ctx context.Context, userID, assetType string, limit, offset int) (*AssetListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	assets, err := s.repo.FilterByType(ctx, userID, assetType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to filter assets: %w", err)
	}

	// Count total by type
	total, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count assets: %w", err)
	}

	return &AssetListResponse{
		Assets: assets,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
