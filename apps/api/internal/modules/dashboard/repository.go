package dashboard

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func stringToPgUUID(id string) (pgtype.UUID, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, err
	}

	return pgtype.UUID{
		Bytes: parsedUUID,
		Valid: true,
	}, nil
}

func pgUUIDToString(pgUUID pgtype.UUID) string {
	return uuid.UUID(pgUUID.Bytes).String()
}

// GetStats retrieves dashboard statistics for a user
func (r *Repository) GetStats(ctx context.Context, userID string) (*db.GetDashboardStatsRow, error) {
	uid, err := stringToPgUUID(userID)
	if err != nil {
		return nil, err
	}

	stats, err := r.queries.GetDashboardStats(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetAssetsByType retrieves asset counts by type
func (r *Repository) GetAssetsByType(ctx context.Context, userID string) ([]db.GetAssetsByTypeCountRow, error) {
	uid, err := stringToPgUUID(userID)
	if err != nil {
		return nil, err
	}

	counts, err := r.queries.GetAssetsByTypeCount(ctx, uid)
	if err != nil {
		return nil, err
	}

	return counts, nil
}

// GetRecentAssets retrieves recent assets for a user
func (r *Repository) GetRecentAssets(ctx context.Context, userID string, limit int) ([]db.Asset, error) {
	uid, err := stringToPgUUID(userID)
	if err != nil {
		return nil, err
	}

	assets, err := r.queries.GetRecentAssets(ctx, db.GetRecentAssetsParams{
		UserID: uid,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, err
	}

	return assets, nil
}
