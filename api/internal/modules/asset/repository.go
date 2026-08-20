package asset

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

type Repository struct {
	q *db.Queries
}

func NewRepository(q *db.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*db.Asset, error) {
	a, err := r.q.GetAssetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get asset: %w", err)
	}
	return &a, nil
}

func (r *Repository) GetByChecksum(ctx context.Context, hash string) (*db.Asset, error) {
	a, err := r.q.GetAssetByChecksum(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("get asset by checksum: %w", err)
	}
	return &a, nil
}

func (r *Repository) Create(ctx context.Context, params db.CreateAssetParams) (*db.Asset, error) {
	a, err := r.q.CreateAsset(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create asset: %w", err)
	}
	return &a, nil
}

func (r *Repository) CreateReference(ctx context.Context, assetID, spaceID uuid.UUID, folderID *uuid.UUID) (*db.AssetReference, error) {
	ref, err := r.q.CreateAssetReference(ctx, db.CreateAssetReferenceParams{
		AssetID:  assetID,
		SpaceID:  spaceID,
		FolderID: uuidPtrToPgtype(folderID),
	})
	if err != nil {
		return nil, fmt.Errorf("create asset reference: %w", err)
	}
	return &ref, nil
}

func (r *Repository) SoftDeleteReference(ctx context.Context, id int64) error {
	return r.q.SoftDeleteAssetReference(ctx, id)
}

func (r *Repository) RestoreReference(ctx context.Context, id int64) error {
	return r.q.RestoreAssetReference(ctx, id)
}

func (r *Repository) ListByFolder(ctx context.Context, spaceID, folderID uuid.UUID, limit, offset int32) ([]db.Asset, error) {
	return r.q.ListAssetsByFolder(ctx, db.ListAssetsByFolderParams{
		SpaceID:  spaceID,
		FolderID: uuidToPgtype(folderID),
		Limit:    limit,
		Offset:   offset,
	})
}

func (r *Repository) ListBySpaceRoot(ctx context.Context, spaceID uuid.UUID, limit, offset int32) ([]db.Asset, error) {
	return r.q.ListAssetsBySpaceRoot(ctx, db.ListAssetsBySpaceRootParams{
		SpaceID: spaceID,
		Limit:   limit,
		Offset:  offset,
	})
}

func (r *Repository) ListBySpace(ctx context.Context, spaceID uuid.UUID, limit, offset int32) ([]db.Asset, error) {
	return r.q.ListAssetsBySpace(ctx, db.ListAssetsBySpaceParams{
		SpaceID: spaceID,
		Limit:   limit,
		Offset:  offset,
	})
}

func (r *Repository) CountActiveReferences(ctx context.Context, assetID uuid.UUID) (int64, error) {
	return r.q.CountActiveReferences(ctx, assetID)
}

func (r *Repository) ListOrphaned(ctx context.Context, before time.Time) ([]uuid.UUID, error) {
	return r.q.ListOrphanedAssets(ctx, before)
}

func (r *Repository) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	return r.q.UpdateAssetName(ctx, db.UpdateAssetNameParams{ID: id, Name: name})
}

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidPtrToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}
