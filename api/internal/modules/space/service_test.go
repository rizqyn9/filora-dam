package space

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

func TestCheckQuota_UnderLimit_Passes(t *testing.T) {
	repo := &fakeSpaceRepo{space: &db.Space{
		ID: uuid.New(), StorageQuotaBytes: 1000, StorageUsedBytes: 500,
	}}
	svc := NewService(repo)

	err := svc.CheckQuota(context.Background(), repo.space.ID, 400)
	if err != nil {
		t.Fatalf("expected pass, got: %v", err)
	}
}

func TestCheckQuota_OverLimit_Rejects(t *testing.T) {
	repo := &fakeSpaceRepo{space: &db.Space{
		ID: uuid.New(), StorageQuotaBytes: 1000, StorageUsedBytes: 900,
	}}
	svc := NewService(repo)

	err := svc.CheckQuota(context.Background(), repo.space.ID, 200)
	if err == nil {
		t.Fatal("expected quota exceeded error")
	}
	if !strings.Contains(err.Error(), "quota exceeded") {
		t.Fatalf("expected quota exceeded, got: %v", err)
	}
}

func TestCheckQuota_UnlimitedQuota_AlwaysPasses(t *testing.T) {
	repo := &fakeSpaceRepo{space: &db.Space{
		ID: uuid.New(), StorageQuotaBytes: 0, StorageUsedBytes: 999999,
	}}
	svc := NewService(repo)

	err := svc.CheckQuota(context.Background(), repo.space.ID, 999999)
	if err != nil {
		t.Fatalf("expected unlimited quota to pass, got: %v", err)
	}
}

// --- Fake ---

type fakeSpaceRepo struct {
	space *db.Space
}

func (r *fakeSpaceRepo) GetByID(_ context.Context, _ uuid.UUID) (*db.Space, error) {
	return r.space, nil
}

func (r *fakeSpaceRepo) ListByOwner(_ context.Context, _ int64) ([]db.Space, error) { return nil, nil }

func (r *fakeSpaceRepo) ListByMember(_ context.Context, _ int64) ([]db.Space, error) { return nil, nil }

func (r *fakeSpaceRepo) Create(_ context.Context, _ db.CreateSpaceParams) (*db.Space, error) {
	return nil, nil
}

func (r *fakeSpaceRepo) Update(_ context.Context, _ db.UpdateSpaceParams) (*db.Space, error) {
	return nil, nil
}

func (r *fakeSpaceRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }

func (r *fakeSpaceRepo) IncrementUsage(_ context.Context, _ uuid.UUID, _ int64) error { return nil }

func (r *fakeSpaceRepo) DecrementUsage(_ context.Context, _ uuid.UUID, _ int64) error { return nil }

func (r *fakeSpaceRepo) GetMember(_ context.Context, _ uuid.UUID, _ int64) (*db.SpaceMember, error) {
	return nil, nil
}
