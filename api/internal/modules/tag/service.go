package tag

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListTags(ctx context.Context, spaceID uuid.UUID) ([]db.Tag, error) {
	return s.repo.ListBySpace(ctx, spaceID)
}

func (s *Service) CreateTag(ctx context.Context, req CreateTagRequest) (*db.Tag, error) {
	return s.repo.Create(ctx, req.SpaceID, req.Name)
}

func (s *Service) DeleteTag(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// TagAsset validates the tag belongs to the given space before linking.
func (s *Service) TagAsset(ctx context.Context, spaceID uuid.UUID, assetID uuid.UUID, tagID int64) error {
	tag, err := s.repo.GetByID(ctx, tagID)
	if err != nil {
		return fmt.Errorf("tag not found: %w", err)
	}
	if tag.SpaceID != spaceID {
		return fmt.Errorf("tag %d does not belong to space %s", tagID, spaceID)
	}
	return s.repo.AddAssetTag(ctx, assetID, tagID)
}

func (s *Service) UntagAsset(ctx context.Context, assetID uuid.UUID, tagID int64) error {
	return s.repo.RemoveAssetTag(ctx, assetID, tagID)
}

func (s *Service) ListAssetTags(ctx context.Context, assetID uuid.UUID) ([]db.Tag, error) {
	return s.repo.ListByAsset(ctx, assetID)
}
