package storage

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
)

var ErrProviderNotFound = errors.New("storage provider not found")

type Repository struct {
	q *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

func (r *Repository) Create(ctx context.Context, layer, name, typ string, credentials []byte, quota, createdBy *int64) (Provider, error) {
	p, err := r.q.CreateStorageProvider(ctx, db.CreateStorageProviderParams{
		Layer:       db.StorageLayer(layer),
		Name:        name,
		Type:        typ,
		Credentials: credentials,
		Quota:       quota,
		CreatedBy:   createdBy,
	})
	if err != nil {
		return Provider{}, err
	}
	return toProvider(p), nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (Provider, error) {
	p, err := r.q.GetStorageProvider(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Provider{}, ErrProviderNotFound
	}
	if err != nil {
		return Provider{}, err
	}
	return toProvider(p), nil
}

// GetRaw returns the full row including credentials (for adapter construction).
func (r *Repository) GetRaw(ctx context.Context, id int64) (db.StorageProvider, error) {
	p, err := r.q.GetStorageProvider(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.StorageProvider{}, ErrProviderNotFound
	}
	return p, err
}

func (r *Repository) List(ctx context.Context) ([]Provider, error) {
	rows, err := r.q.ListStorageProviders(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Provider, 0, len(rows))
	for _, p := range rows {
		out = append(out, toProvider(p))
	}
	return out, nil
}

func (r *Repository) ListActiveByLayer(ctx context.Context, layer string) ([]db.StorageProvider, error) {
	return r.q.ListActiveProvidersByLayer(ctx, db.StorageLayer(layer))
}

func (r *Repository) Update(ctx context.Context, id int64, name string, credentials []byte, quota *int64, isActive bool) (Provider, error) {
	p, err := r.q.UpdateStorageProvider(ctx, db.UpdateStorageProviderParams{
		ID:          id,
		Name:        name,
		Credentials: credentials,
		Quota:       quota,
		IsActive:    isActive,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Provider{}, ErrProviderNotFound
	}
	if err != nil {
		return Provider{}, err
	}
	return toProvider(p), nil
}

func (r *Repository) Deactivate(ctx context.Context, id int64) error {
	return r.q.DeactivateStorageProvider(ctx, id)
}

func (r *Repository) AddUsed(ctx context.Context, id, delta int64) error {
	return r.q.AddStorageProviderUsed(ctx, db.AddStorageProviderUsedParams{ID: id, Used: delta})
}

func (r *Repository) Usage(ctx context.Context) ([]AccountUsage, error) {
	rows, err := r.q.ListStorageAccountUsage(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AccountUsage, 0, len(rows))
	for _, u := range rows {
		out = append(out, AccountUsage{
			ID:            u.ID,
			Name:          u.Name,
			Layer:         string(u.Layer),
			Type:          u.Type,
			IsActive:      u.IsActive,
			Quota:         u.Quota,
			Used:          u.Used,
			UsedPercent:   u.UsedPercent,
			LocationCount: u.LocationCount,
			StoredCount:   u.StoredCount,
			PendingCount:  u.PendingCount,
			FailedCount:   u.FailedCount,
		})
	}
	return out, nil
}

func toProvider(p db.StorageProvider) Provider {
	return Provider{
		ID:        p.ID,
		Layer:     string(p.Layer),
		Name:      p.Name,
		Type:      p.Type,
		Quota:     p.Quota,
		Used:      p.Used,
		IsActive:  p.IsActive,
		CreatedBy: p.CreatedBy,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func marshalCredentials(m map[string]any) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}
