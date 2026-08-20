package tag

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

type Repository struct {
	q *db.Queries
}

func NewRepository(q *db.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*db.Tag, error) {
	t, err := r.q.GetTagByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get tag by id: %w", err)
	}
	return &t, nil
}

func (r *Repository) ListBySpace(ctx context.Context, spaceID uuid.UUID) ([]db.Tag, error) {
	tags, err := r.q.ListTagsBySpace(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("list tags by space: %w", err)
	}
	return tags, nil
}

func (r *Repository) Create(ctx context.Context, spaceID uuid.UUID, name string) (*db.Tag, error) {
	t, err := r.q.CreateTag(ctx, db.CreateTagParams{SpaceID: spaceID, Name: name})
	if err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}
	return &t, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	if err := r.q.DeleteTag(ctx, id); err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	return nil
}

func (r *Repository) AddAssetTag(ctx context.Context, assetID uuid.UUID, tagID int64) error {
	if err := r.q.AddAssetTag(ctx, db.AddAssetTagParams{AssetID: assetID, TagID: tagID}); err != nil {
		return fmt.Errorf("add asset tag: %w", err)
	}
	return nil
}

func (r *Repository) RemoveAssetTag(ctx context.Context, assetID uuid.UUID, tagID int64) error {
	if err := r.q.RemoveAssetTag(ctx, db.RemoveAssetTagParams{AssetID: assetID, TagID: tagID}); err != nil {
		return fmt.Errorf("remove asset tag: %w", err)
	}
	return nil
}

func (r *Repository) ListByAsset(ctx context.Context, assetID uuid.UUID) ([]db.Tag, error) {
	tags, err := r.q.ListTagsByAsset(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("list tags by asset: %w", err)
	}
	return tags, nil
}
