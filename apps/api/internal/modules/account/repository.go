package account

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

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var userID pgtype.UUID
	if err := userID.Scan(parsedUUID); err != nil {
		return nil, err
	}

	user, err := r.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           uuid.UUID(user.ID.Bytes).String(),
		Email:        user.Email,
		Name:         user.Name,
		StorageQuota: user.StorageQuota,
		StorageUsed:  user.StorageUsed,
		CreatedAt:    user.CreatedAt.Time,
		UpdatedAt:    user.UpdatedAt.Time,
	}, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) Create(ctx context.Context, email, name, passwordHash string) (*User, error) {
	user, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           uuid.UUID(user.ID.Bytes).String(),
		Email:        user.Email,
		Name:         user.Name,
		StorageQuota: user.StorageQuota,
		StorageUsed:  user.StorageUsed,
		CreatedAt:    user.CreatedAt.Time,
		UpdatedAt:    user.UpdatedAt.Time,
	}, nil
}

func (r *Repository) UpdateStorageUsed(ctx context.Context, userID string, used int64) error {
	parsedUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	var id pgtype.UUID
	if err := id.Scan(parsedUUID); err != nil {
		return err
	}

	_, err = r.queries.UpdateUserStorageUsed(ctx, db.UpdateUserStorageUsedParams{
		ID:          id,
		StorageUsed: used,
	})
	return err
}

func (r *Repository) GetQuota(ctx context.Context, userID string) (*QuotaInfo, error) {
	parsedUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	var id pgtype.UUID
	if err := id.Scan(parsedUUID); err != nil {
		return nil, err
	}

	quota, err := r.queries.GetUserQuota(ctx, id)
	if err != nil {
		return nil, err
	}

	return &QuotaInfo{
		Quota: quota.StorageQuota,
		Used:  quota.StorageUsed,
		Free:  quota.StorageQuota - quota.StorageUsed,
	}, nil
}
