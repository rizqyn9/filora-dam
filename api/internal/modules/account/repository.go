package account

import (
	"context"
	"fmt"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

type Repository struct {
	q *db.Queries
}

func NewRepository(q *db.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) GetByClerkID(ctx context.Context, clerkID string) (*db.User, error) {
	u, err := r.q.GetUserByClerkID(ctx, clerkID)
	if err != nil {
		return nil, fmt.Errorf("get user by clerk id: %w", err)
	}
	return &u, nil
}

func (r *Repository) Create(ctx context.Context, params db.CreateUserParams) (*db.User, error) {
	u, err := r.q.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (r *Repository) Update(ctx context.Context, params db.UpdateUserParams) (*db.User, error) {
	u, err := r.q.UpdateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return &u, nil
}

func (r *Repository) Delete(ctx context.Context, clerkID string) error {
	return r.q.DeleteUserByClerkID(ctx, clerkID)
}
