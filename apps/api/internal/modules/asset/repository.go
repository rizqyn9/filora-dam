package asset

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
)

var ErrAssetNotFound = errors.New("asset not found")

type Repository struct {
	q *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

func (r *Repository) Create(ctx context.Context, galleryID int64, uploadedBy *int64, name, typ, mime string, size int64, hash string) (Asset, error) {
	a, err := r.q.CreateAsset(ctx, db.CreateAssetParams{
		GalleryID:  galleryID,
		UploadedBy: uploadedBy,
		Name:       name,
		Type:       typ,
		MimeType:   mime,
		Size:       size,
		Hash:       hash,
		Metadata:   nil,
	})
	if err != nil {
		return Asset{}, err
	}
	return toAsset(a), nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Asset, error) {
	a, err := r.q.GetAssetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Asset{}, ErrAssetNotFound
	}
	if err != nil {
		return Asset{}, err
	}
	return toAsset(a), nil
}

func (r *Repository) GetActiveByHash(ctx context.Context, galleryID int64, hash string) (Asset, error) {
	a, err := r.q.GetActiveAssetByGalleryHash(ctx, db.GetActiveAssetByGalleryHashParams{GalleryID: galleryID, Hash: hash})
	if errors.Is(err, pgx.ErrNoRows) {
		return Asset{}, ErrAssetNotFound
	}
	if err != nil {
		return Asset{}, err
	}
	return toAsset(a), nil
}

func (r *Repository) ListActive(ctx context.Context, galleryID int64, limit, offset int32) ([]Asset, error) {
	rows, err := r.q.ListActiveAssetsByGallery(ctx, db.ListActiveAssetsByGalleryParams{GalleryID: galleryID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return toAssets(rows), nil
}

func (r *Repository) CountActive(ctx context.Context, galleryID int64) (int64, error) {
	return r.q.CountActiveAssetsByGallery(ctx, galleryID)
}

func (r *Repository) Search(ctx context.Context, galleryID int64, pattern string, limit, offset int32) ([]Asset, error) {
	rows, err := r.q.SearchAssetsByName(ctx, db.SearchAssetsByNameParams{GalleryID: galleryID, Name: pattern, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return toAssets(rows), nil
}

func (r *Repository) FilterByType(ctx context.Context, galleryID int64, typ string, limit, offset int32) ([]Asset, error) {
	rows, err := r.q.FilterAssetsByType(ctx, db.FilterAssetsByTypeParams{GalleryID: galleryID, Type: typ, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return toAssets(rows), nil
}

func (r *Repository) ListTrashed(ctx context.Context, galleryID int64, limit, offset int32) ([]Asset, error) {
	rows, err := r.q.ListTrashedAssetsByGallery(ctx, db.ListTrashedAssetsByGalleryParams{GalleryID: galleryID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return toAssets(rows), nil
}

func (r *Repository) UpdateName(ctx context.Context, id uuid.UUID, name string) (Asset, error) {
	a, err := r.q.UpdateAssetName(ctx, db.UpdateAssetNameParams{ID: id, Name: name})
	if errors.Is(err, pgx.ErrNoRows) {
		return Asset{}, ErrAssetNotFound
	}
	if err != nil {
		return Asset{}, err
	}
	return toAsset(a), nil
}

func (r *Repository) SoftDelete(ctx context.Context, id uuid.UUID, by *int64) (bool, error) {
	n, err := r.q.SoftDeleteAsset(ctx, db.SoftDeleteAssetParams{ID: id, DeletedBy: by})
	return n > 0, err
}

func (r *Repository) Restore(ctx context.Context, id uuid.UUID) (bool, error) {
	n, err := r.q.RestoreAsset(ctx, id)
	return n > 0, err
}

func (r *Repository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return r.q.HardDeleteAsset(ctx, id)
}

func toAssets(rows []db.Asset) []Asset {
	out := make([]Asset, 0, len(rows))
	for _, a := range rows {
		out = append(out, toAsset(a))
	}
	return out
}

func toAsset(a db.Asset) Asset {
	return Asset{
		ID:         a.ID,
		GalleryID:  a.GalleryID,
		UploadedBy: a.UploadedBy,
		Name:       a.Name,
		Type:       a.Type,
		MimeType:   a.MimeType,
		Size:       a.Size,
		Hash:       a.Hash,
		Metadata:   a.Metadata,
		DeletedAt:  tsPtr(a.DeletedAt),
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}

func tsPtr(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		v := t.Time
		return &v
	}
	return nil
}
