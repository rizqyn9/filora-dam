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

// Helper to convert string to pgtype.UUID
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

// Helper to convert pgtype.UUID to string
func pgUUIDToString(pgUUID pgtype.UUID) string {
	return uuid.UUID(pgUUID.Bytes).String()
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	userID, err := stringToPgUUID(id)
	if err != nil {
		return nil, err
	}

	user, err := r.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           pgUUIDToString(user.ID),
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
		ID:           pgUUIDToString(user.ID),
		Email:        user.Email,
		Name:         user.Name,
		StorageQuota: user.StorageQuota,
		StorageUsed:  user.StorageUsed,
		CreatedAt:    user.CreatedAt.Time,
		UpdatedAt:    user.UpdatedAt.Time,
	}, nil
}

func (r *Repository) UpdateStorageUsed(ctx context.Context, userID string, used int64) error {
	id, err := stringToPgUUID(userID)
	if err != nil {
		return err
	}

	_, err = r.queries.UpdateUserStorageUsed(ctx, db.UpdateUserStorageUsedParams{
		ID:          id,
		StorageUsed: used,
	})
	return err
}

func (r *Repository) GetQuota(ctx context.Context, userID string) (*QuotaInfo, error) {
	id, err := stringToPgUUID(userID)
	if err != nil {
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
