package storage

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

func (r *Repository) GetByID(ctx context.Context, id int64) (*db.StorageAccount, error) {
	a, err := r.q.GetStorageAccountByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get storage account: %w", err)
	}
	return &a, nil
}

func (r *Repository) ListAll(ctx context.Context) ([]db.StorageAccount, error) {
	accounts, err := r.q.ListAllStorageAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list storage accounts: %w", err)
	}
	return accounts, nil
}

func (r *Repository) ListActiveByLayer(ctx context.Context, layer db.StorageLayer) ([]db.StorageAccount, error) {
	accounts, err := r.q.ListActiveAccountsByLayer(ctx, layer)
	if err != nil {
		return nil, fmt.Errorf("list active accounts by layer: %w", err)
	}
	return accounts, nil
}

func (r *Repository) Create(ctx context.Context, params db.CreateStorageAccountParams) (*db.StorageAccount, error) {
	a, err := r.q.CreateStorageAccount(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create storage account: %w", err)
	}
	return &a, nil
}

func (r *Repository) Update(ctx context.Context, params db.UpdateStorageAccountParams) error {
	if err := r.q.UpdateStorageAccount(ctx, params); err != nil {
		return fmt.Errorf("update storage account: %w", err)
	}
	return nil
}

func (r *Repository) Deactivate(ctx context.Context, id int64) error {
	return r.q.DeactivateAccount(ctx, id)
}

func (r *Repository) IncrementUsage(ctx context.Context, id int64, bytes int64) error {
	return r.q.IncrementAccountUsage(ctx, db.IncrementAccountUsageParams{ID: id, UsedBytes: bytes})
}

func (r *Repository) CreateLocation(ctx context.Context, params db.CreateStorageLocationParams) (*db.StorageLocation, error) {
	loc, err := r.q.CreateStorageLocation(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create storage location: %w", err)
	}
	return &loc, nil
}

func (r *Repository) UpdateLocationStatus(ctx context.Context, params db.UpdateStorageLocationStatusParams) error {
	return r.q.UpdateStorageLocationStatus(ctx, params)
}

func (r *Repository) GetServingLocation(ctx context.Context, assetID uuid.UUID) (*db.StorageLocation, error) {
	loc, err := r.q.GetServingLocation(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("get serving location: %w", err)
	}
	return &loc, nil
}
