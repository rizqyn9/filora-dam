package dashboard

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
)

type Repository struct {
	q *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

func (r *Repository) GalleryStats(ctx context.Context, galleryID int64) (GalleryStats, error) {
	row, err := r.q.GalleryAssetStats(ctx, galleryID)
	if err != nil {
		return GalleryStats{}, err
	}
	return GalleryStats{
		TotalAssets: row.TotalAssets,
		TotalSize:   row.TotalSize,
		UniqueTypes: row.TypeCount,
	}, nil
}

func (r *Repository) TypeCounts(ctx context.Context, galleryID int64) ([]TypeCount, error) {
	rows, err := r.q.GalleryAssetCountsByType(ctx, galleryID)
	if err != nil {
		return nil, err
	}
	out := make([]TypeCount, 0, len(rows))
	for _, row := range rows {
		out = append(out, TypeCount{Type: row.Type, Count: row.Count})
	}
	return out, nil
}

func (r *Repository) RecentAssets(ctx context.Context, galleryID int64, limit int32) ([]RecentAsset, error) {
	rows, err := r.q.GalleryRecentAssets(ctx, db.GalleryRecentAssetsParams{GalleryID: galleryID, Limit: limit})
	if err != nil {
		return nil, err
	}
	out := make([]RecentAsset, 0, len(rows))
	for _, row := range rows {
		out = append(out, RecentAsset{
			ID:        row.ID,
			Name:      row.Name,
			Type:      row.Type,
			MimeType:  row.MimeType,
			Size:      row.Size,
			CreatedAt: row.CreatedAt,
		})
	}
	return out, nil
}

func (r *Repository) ArchiveJobHealth(ctx context.Context) (ArchiveJobHealth, error) {
	row, err := r.q.ArchiveJobHealth(ctx)
	if err != nil {
		return ArchiveJobHealth{}, err
	}
	return ArchiveJobHealth{
		Pending:   row.Pending,
		Running:   row.Running,
		Completed: row.Completed,
		Failed:    row.Failed,
	}, nil
}
