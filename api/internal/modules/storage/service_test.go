package storage

import (
	"context"
	"strings"
	"testing"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

type fakeStorageRepo struct {
	accounts []db.StorageAccount
}

func (r *fakeStorageRepo) ListActiveByLayer(_ context.Context, layer db.StorageLayer) ([]db.StorageAccount, error) {
	var result []db.StorageAccount
	for _, a := range r.accounts {
		if a.Layer == layer && a.IsActive {
			result = append(result, a)
		}
	}
	return result, nil
}

func (r *fakeStorageRepo) ListAll(_ context.Context) ([]db.StorageAccount, error) {
	return r.accounts, nil
}

func (r *fakeStorageRepo) GetByID(_ context.Context, _ int64) (*db.StorageAccount, error) {
	return nil, nil
}

func (r *fakeStorageRepo) Create(_ context.Context, _ db.CreateStorageAccountParams) (*db.StorageAccount, error) {
	return nil, nil
}

func (r *fakeStorageRepo) Update(_ context.Context, _ db.UpdateStorageAccountParams) error {
	return nil
}

func (r *fakeStorageRepo) Deactivate(_ context.Context, _ int64) error { return nil }

func (r *fakeStorageRepo) IncrementUsage(_ context.Context, _ int64, _ int64) error { return nil }

func TestElectAccount_PicksFirstWithCapacity(t *testing.T) {
	repo := &fakeStorageRepo{accounts: []db.StorageAccount{
		{ID: 1, Provider: db.StorageProviderR2, Layer: db.StorageLayerServing, IsActive: true, QuotaBytes: 100, UsedBytes: 95}, // only 5 left
		{ID: 2, Provider: db.StorageProviderR2, Layer: db.StorageLayerServing, IsActive: true, QuotaBytes: 100, UsedBytes: 10}, // 90 left
		{ID: 3, Provider: db.StorageProviderR2, Layer: db.StorageLayerServing, IsActive: true, QuotaBytes: 100, UsedBytes: 50}, // 50 left
	}}

	svc := NewService(repo, func(b []byte) ([]byte, error) { return b, nil })

	got, err := svc.ElectAccount(context.Background(), db.StorageLayerServing, 20)
	if err != nil {
		t.Fatalf("expected account, got error: %v", err)
	}
	// First account with capacity >= 20 is ID=2 (90 left)
	if got.ID != 2 {
		t.Fatalf("expected account 2 (first with capacity), got %d", got.ID)
	}
}

func TestElectAccount_AllFull_ReturnsError(t *testing.T) {
	repo := &fakeStorageRepo{accounts: []db.StorageAccount{
		{ID: 1, Layer: db.StorageLayerServing, IsActive: true, QuotaBytes: 100, UsedBytes: 100},
		{ID: 2, Layer: db.StorageLayerServing, IsActive: true, QuotaBytes: 50, UsedBytes: 50},
	}}

	svc := NewService(repo, func(b []byte) ([]byte, error) { return b, nil })

	_, err := svc.ElectAccount(context.Background(), db.StorageLayerServing, 1)
	if err == nil {
		t.Fatal("expected error when all accounts full")
	}
	if !strings.Contains(err.Error(), "no available storage account") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestElectAccount_UnlimitedQuota_AlwaysFits(t *testing.T) {
	repo := &fakeStorageRepo{accounts: []db.StorageAccount{
		{ID: 1, Layer: db.StorageLayerServing, IsActive: true, QuotaBytes: 0, UsedBytes: 999999}, // 0 = unlimited
	}}

	svc := NewService(repo, func(b []byte) ([]byte, error) { return b, nil })

	got, err := svc.ElectAccount(context.Background(), db.StorageLayerServing, 99999)
	if err != nil {
		t.Fatalf("expected unlimited account to fit, got: %v", err)
	}
	if got.ID != 1 {
		t.Fatalf("expected account 1, got %d", got.ID)
	}
}

func TestElectAccount_SkipsInactive(t *testing.T) {
	repo := &fakeStorageRepo{accounts: []db.StorageAccount{
		{ID: 1, Layer: db.StorageLayerServing, IsActive: false, QuotaBytes: 1000, UsedBytes: 0}, // inactive
		{ID: 2, Layer: db.StorageLayerServing, IsActive: true, QuotaBytes: 1000, UsedBytes: 0},  // active
	}}

	svc := NewService(repo, func(b []byte) ([]byte, error) { return b, nil })

	got, err := svc.ElectAccount(context.Background(), db.StorageLayerServing, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 2 {
		t.Fatalf("expected active account 2, got %d", got.ID)
	}
}
