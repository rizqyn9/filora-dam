package storage

import (
	"context"
	"encoding/json"

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

func (r *Repository) GetProviderByID(ctx context.Context, id string) (*Provider, error) {
	providerID, err := stringToPgUUID(id)
	if err != nil {
		return nil, err
	}

	provider, err := r.queries.GetProviderByID(ctx, providerID)
	if err != nil {
		return nil, err
	}

	// Parse credentials from JSONB
	var credentials map[string]interface{}
	if err := json.Unmarshal(provider.Credentials, &credentials); err != nil {
		return nil, err
	}

	return &Provider{
		ID:          pgUUIDToString(provider.ID),
		UserID:      pgUUIDToString(provider.UserID),
		Name:        provider.Name,
		Type:        provider.Type,
		Credentials: credentials,
		Quota:       provider.Quota,
		Used:        provider.Used,
		IsActive:    provider.IsActive,
		CreatedAt:   provider.CreatedAt.Time,
		UpdatedAt:   provider.UpdatedAt.Time,
	}, nil
}

func (r *Repository) ListActiveProviders(ctx context.Context, userID string) ([]*Provider, error) {
	uid, err := stringToPgUUID(userID)
	if err != nil {
		return nil, err
	}

	providers, err := r.queries.ListActiveProvidersByUser(ctx, uid)
	if err != nil {
		return nil, err
	}

	result := make([]*Provider, 0, len(providers))
	for _, p := range providers {
		var credentials map[string]interface{}
		if err := json.Unmarshal(p.Credentials, &credentials); err != nil {
			continue
		}

		result = append(result, &Provider{
			ID:          pgUUIDToString(p.ID),
			UserID:      pgUUIDToString(p.UserID),
			Name:        p.Name,
			Type:        p.Type,
			Credentials: credentials,
			Quota:       p.Quota,
			Used:        p.Used,
			IsActive:    p.IsActive,
			CreatedAt:   p.CreatedAt.Time,
			UpdatedAt:   p.UpdatedAt.Time,
		})
	}

	return result, nil
}

func (r *Repository) ListAllProviders(ctx context.Context, userID string) ([]*Provider, error) {
	uid, err := stringToPgUUID(userID)
	if err != nil {
		return nil, err
	}

	providers, err := r.queries.ListAllProvidersByUser(ctx, uid)
	if err != nil {
		return nil, err
	}

	result := make([]*Provider, 0, len(providers))
	for _, p := range providers {
		var credentials map[string]interface{}
		if err := json.Unmarshal(p.Credentials, &credentials); err != nil {
			continue
		}

		result = append(result, &Provider{
			ID:          pgUUIDToString(p.ID),
			UserID:      pgUUIDToString(p.UserID),
			Name:        p.Name,
			Type:        p.Type,
			Credentials: credentials,
			Quota:       p.Quota,
			Used:        p.Used,
			IsActive:    p.IsActive,
			CreatedAt:   p.CreatedAt.Time,
			UpdatedAt:   p.UpdatedAt.Time,
		})
	}

	return result, nil
}

func (r *Repository) CreateProvider(ctx context.Context, userID string, req *CreateProviderRequest) (*Provider, error) {
	uid, err := stringToPgUUID(userID)
	if err != nil {
		return nil, err
	}

	// Marshal credentials to JSONB
	credentialsJSON, err := json.Marshal(req.Credentials)
	if err != nil {
		return nil, err
	}

	var quota *int64
	if req.Quota != nil {
		quota = req.Quota
	}

	provider, err := r.queries.CreateProvider(ctx, db.CreateProviderParams{
		UserID:      uid,
		Name:        req.Name,
		Type:        req.Type,
		Credentials: credentialsJSON,
		Quota:       quota,
	})
	if err != nil {
		return nil, err
	}

	var credentials map[string]interface{}
	if err := json.Unmarshal(provider.Credentials, &credentials); err != nil {
		return nil, err
	}

	return &Provider{
		ID:          pgUUIDToString(provider.ID),
		UserID:      pgUUIDToString(provider.UserID),
		Name:        provider.Name,
		Type:        provider.Type,
		Credentials: credentials,
		Quota:       provider.Quota,
		Used:        provider.Used,
		IsActive:    provider.IsActive,
		CreatedAt:   provider.CreatedAt.Time,
		UpdatedAt:   provider.UpdatedAt.Time,
	}, nil
}

func (r *Repository) UpdateProviderUsage(ctx context.Context, providerID string, used int64) error {
	pid, err := stringToPgUUID(providerID)
	if err != nil {
		return err
	}

	_, err = r.queries.UpdateProviderUsage(ctx, db.UpdateProviderUsageParams{
		ID:   pid,
		Used: used,
	})
	return err
}

func (r *Repository) DeactivateProvider(ctx context.Context, providerID string) error {
	pid, err := stringToPgUUID(providerID)
	if err != nil {
		return err
	}

	_, err = r.queries.DeactivateProvider(ctx, pid)
	return err
}
