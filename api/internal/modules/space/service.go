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
