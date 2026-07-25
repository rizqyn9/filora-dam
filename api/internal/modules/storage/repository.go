package storage

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *Repository) CreateLocation(ctx context.Context, assetID uuid.UUID, providerID int64, layer, key string, url *string, status string) error {
	_, err := r.q.CreateStorageLocation(ctx, db.CreateStorageLocationParams{
		AssetID:     assetID,
		ProviderID:  providerID,
		Layer:       db.StorageLayer(layer),
		ProviderKey: key,
		Url:         url,
		Status:      db.LocationStatus(status),
	})
	return err
}

func (r *Repository) GetServingURL(ctx context.Context, assetID uuid.UUID) (string, bool, error) {
	url, err := r.q.GetServingLocationURL(ctx, assetID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if url == nil {
		return "", false, nil
	}
	return *url, true, nil
}

func (r *Repository) EnqueueArchive(ctx context.Context, assetID uuid.UUID) error {
	return r.q.EnqueueArchiveJob(ctx, assetID)
}

// --- archive worker ---

func (r *Repository) ClaimArchiveJob(ctx context.Context) (db.ArchiveSyncJob, bool, error) {
	job, err := r.q.ClaimArchiveJob(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.ArchiveSyncJob{}, false, nil
	}
	if err != nil {
		return db.ArchiveSyncJob{}, false, err
	}
	return job, true, nil
}

func (r *Repository) MarkJobResult(ctx context.Context, id int64, status string, lastErr *string, nextRetry *time.Time) error {
	return r.q.MarkArchiveJobResult(ctx, db.MarkArchiveJobResultParams{
		ID:          id,
		Status:      db.JobStatus(status),
		LastError:   lastErr,
		NextRetryAt: tsFromPtr(nextRetry),
	})
}

func (r *Repository) GetArchiveSource(ctx context.Context, assetID uuid.UUID) (db.GetArchiveSourceRow, bool, error) {
	row, err := r.q.GetArchiveSource(ctx, assetID)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.GetArchiveSourceRow{}, false, nil
	}
	if err != nil {
		return db.GetArchiveSourceRow{}, false, err
	}
	return row, true, nil
}

func tsFromPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
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
