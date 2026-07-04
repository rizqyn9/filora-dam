package tag

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
)

var ErrTagNotFound = errors.New("tag not found")

type Repository struct {
	q *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

func (r *Repository) Create(ctx context.Context, galleryID int64, name string, createdBy *int64) (Tag, error) {
	t, err := r.q.CreateTag(ctx, db.CreateTagParams{GalleryID: galleryID, Name: name, CreatedBy: createdBy})
	if err != nil {
		return Tag{}, err
	}
	return toTag(t), nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (Tag, error) {
	t, err := r.q.GetTagByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tag{}, ErrTagNotFound
	}
	if err != nil {
		return Tag{}, err
	}
	return toTag(t), nil
}

func (r *Repository) ListByGallery(ctx context.Context, galleryID int64) ([]Tag, error) {
	rows, err := r.q.ListTagsByGallery(ctx, galleryID)
	if err != nil {
		return nil, err
	}
	out := make([]Tag, 0, len(rows))
	for _, t := range rows {
		out = append(out, toTag(t))
	}
	return out, nil
}

func (r *Repository) Update(ctx context.Context, id int64, name string) (Tag, error) {
	t, err := r.q.UpdateTag(ctx, db.UpdateTagParams{ID: id, Name: name})
	if errors.Is(err, pgx.ErrNoRows) {
		return Tag{}, ErrTagNotFound
	}
	if err != nil {
		return Tag{}, err
	}
	return toTag(t), nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteTag(ctx, id)
}

func (r *Repository) Attach(ctx context.Context, assetID uuid.UUID, tagID int64) error {
	return r.q.AttachTag(ctx, db.AttachTagParams{AssetID: assetID, TagID: tagID})
}

func (r *Repository) Detach(ctx context.Context, assetID uuid.UUID, tagID int64) (bool, error) {
	n, err := r.q.DetachTag(ctx, db.DetachTagParams{AssetID: assetID, TagID: tagID})
	return n > 0, err
}

func toTag(t db.Tag) Tag {
	return Tag{
		ID:        t.ID,
		GalleryID: t.GalleryID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
