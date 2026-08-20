package space

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetSpace(ctx context.Context, id uuid.UUID) (*db.Space, error) {
	space, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("space not found: %w", err)
		}
		return nil, err
	}
	return space, nil
}

func (s *Service) ListSpaces(ctx context.Context, userID int64) ([]db.Space, error) {
	owned, err := s.repo.ListByOwner(ctx, userID)
	if err != nil {
		return nil, err
	}

	shared, err := s.repo.ListByMember(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Merge owned + shared, dedup by ID
	seen := make(map[uuid.UUID]bool, len(owned))
	result := make([]db.Space, 0, len(owned)+len(shared))
	for _, sp := range owned {
		seen[sp.ID] = true
		result = append(result, sp)
	}
	for _, sp := range shared {
		if !seen[sp.ID] {
			result = append(result, sp)
		}
	}
	return result, nil
}

func (s *Service) CreateSpace(ctx context.Context, ownerID int64, req CreateSpaceRequest) (*db.Space, error) {
	return s.repo.Create(ctx, db.CreateSpaceParams{
		Name:              req.Name,
		OwnerID:           ownerID,
		StorageQuotaBytes: req.StorageQuotaBytes,
	})
}

func (s *Service) UpdateSpace(ctx context.Context, id uuid.UUID, req UpdateSpaceRequest) (*db.Space, error) {
	return s.repo.Update(ctx, db.UpdateSpaceParams{
		ID:                id,
		Name:              req.Name,
		StorageQuotaBytes: req.StorageQuotaBytes,
	})
}

func (s *Service) DeleteSpace(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// CheckQuota implements asset.SpaceQuota.
func (s *Service) CheckQuota(ctx context.Context, spaceID uuid.UUID, additionalBytes int64) error {
	sp, err := s.repo.GetByID(ctx, spaceID)
	if err != nil {
		return fmt.Errorf("get space for quota check: %w", err)
	}
	if sp.StorageQuotaBytes > 0 && sp.StorageUsedBytes+additionalBytes > sp.StorageQuotaBytes {
		return fmt.Errorf("space %s quota exceeded (%d/%d bytes)", spaceID, sp.StorageUsedBytes+additionalBytes, sp.StorageQuotaBytes)
	}
	return nil
}

// IncrementUsage implements asset.SpaceQuota.
func (s *Service) IncrementUsage(ctx context.Context, spaceID uuid.UUID, bytes int64) error {
	return s.repo.IncrementUsage(ctx, spaceID, bytes)
}

// DecrementUsage implements asset.SpaceQuota.
func (s *Service) DecrementUsage(ctx context.Context, spaceID uuid.UUID, bytes int64) error {
	return s.repo.DecrementUsage(ctx, spaceID, bytes)
}

// HasMember checks if a user has access to a space (owner or member).
func (s *Service) HasMember(ctx context.Context, spaceID uuid.UUID, userID int64) (bool, error) {
	sp, err := s.repo.GetByID(ctx, spaceID)
	if err != nil {
		return false, err
	}
	if sp.OwnerID == userID {
		return true, nil
	}
	_, err = s.repo.GetMember(ctx, spaceID, userID)
	if err != nil {
		return false, nil //nolint:nilerr // ErrNoRows means "not a member", not an error
	}
	return true, nil
}
