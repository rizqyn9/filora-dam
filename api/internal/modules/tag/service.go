package tag

import (
	"context"

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

func (s *Service) TagAsset(ctx context.Context, assetID uuid.UUID, tagID int64) error {
	return s.repo.AddAssetTag(ctx, assetID, tagID)
}

func (s *Service) UntagAsset(ctx context.Context, assetID uuid.UUID, tagID int64) error {
	return s.repo.RemoveAssetTag(ctx, assetID, tagID)
}

func (s *Service) ListAssetTags(ctx context.Context, assetID uuid.UUID) ([]db.Tag, error) {
	return s.repo.ListByAsset(ctx, assetID)
}
