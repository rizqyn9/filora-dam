package space

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

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*db.Space, error) {
	space, err := r.q.GetSpaceByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get space by id: %w", err)
	}
	return &space, nil
}

func (r *Repository) ListByOwner(ctx context.Context, ownerID int64) ([]db.Space, error) {
	spaces, err := r.q.ListSpacesByOwner(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list spaces by owner: %w", err)
	}
	return spaces, nil
}

func (r *Repository) ListByMember(ctx context.Context, userID int64) ([]db.Space, error) {
	spaces, err := r.q.ListSpacesByMember(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list spaces by member: %w", err)
	}
	return spaces, nil
}

func (r *Repository) Create(ctx context.Context, params db.CreateSpaceParams) (*db.Space, error) {
	space, err := r.q.CreateSpace(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create space: %w", err)
	}
	return &space, nil
}

func (r *Repository) Update(ctx context.Context, params db.UpdateSpaceParams) (*db.Space, error) {
	space, err := r.q.UpdateSpace(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update space: %w", err)
	}
	return &space, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteSpace(ctx, id); err != nil {
		return fmt.Errorf("delete space: %w", err)
	}
	return nil
}

func (r *Repository) IncrementUsage(ctx context.Context, id uuid.UUID, bytes int64) error {
	return r.q.IncrementSpaceUsage(ctx, db.IncrementSpaceUsageParams{ID: id, StorageUsedBytes: bytes})
}

func (r *Repository) DecrementUsage(ctx context.Context, id uuid.UUID, bytes int64) error {
	return r.q.DecrementSpaceUsage(ctx, db.DecrementSpaceUsageParams{ID: id, StorageUsedBytes: bytes})
}

func (r *Repository) GetMember(ctx context.Context, spaceID uuid.UUID, userID int64) (*db.SpaceMember, error) {
	m, err := r.q.GetSpaceMember(ctx, db.GetSpaceMemberParams{SpaceID: spaceID, UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("get space member: %w", err)
	}
	return &m, nil
}
