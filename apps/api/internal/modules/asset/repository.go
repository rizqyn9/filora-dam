package asset

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

func (r *Repository) GetByID(ctx context.Context, id string) (*Asset, error) {
	assetID, err := stringToPgUUID(id)
	if err != nil {
		return nil, err
	}

	asset, err := r.queries.GetAssetByID(ctx, assetID)
	if err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	if len(asset.Metadata) > 0 {
		if err := json.Unmarshal(asset.Metadata, &metadata); err != nil {
			metadata = nil
		}
	}

	return &Asset{
		ID:        pgUUIDToString(asset.ID),
		UserID:    pgUUIDToString(asset.UserID),
		Name:      asset.Name,
		Type:      asset.Type,
		MimeType:  asset.MimeType,
		Size:      asset.Size,
		Hash:      asset.Hash,
		Tags:      asset.Tags,
		Metadata:  metadata,
		CreatedAt: asset.CreatedAt.Time,
		UpdatedAt: asset.UpdatedAt.Time,
	}, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*Asset, error) {
	uid, err := stringToPgUUID(userID)
	if err != nil {
		return nil, err
	}

	assets, err := r.queries.ListAssetsByUser(ctx, db.ListAssetsByUserParams{
		UserID: uid,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	result := make([]*Asset, 0, len(assets))
	for _, a := range assets {
		var metadata map[string]interface{}
		if len(a.Metadata) > 0 {
			if err := json.Unmarshal(a.Metadata, &metadata); err != nil {
				metadata = nil
			}
		}

		result = append(result, &Asset{
			ID:        pgUUIDToString(a.ID),
			UserID:    pgUUIDToString(a.UserID),
			Name:      a.Name,
			Type:      a.Type,
			MimeType:  a.MimeType,
			Size:      a.Size,
			Hash:      a.Hash,
			Tags:      a.Tags,
			Metadata:  metadata,
			CreatedAt: a.CreatedAt.Time,
			UpdatedAt: a.UpdatedAt.Time,
		})
	}

	return result, nil
}

func (r *Repository) CountByUser(ctx context.Context, userID string) (int64, error) {
	uid, err := stringToPgUUID(userID)
	if err != nil {
		return 0, err
	}

	return r.queries.CountAssetsByUser(ctx, uid)
}

func (r *Repository) GetByHash(ctx context.Context, hash, userID string) (*Asset, error) {
	uid, err := stringToPgUUID(userID)
	if err != nil {
		return nil, err
	}

	asset, err := r.queries.GetAssetByHash(ctx, db.GetAssetByHashParams{
		Hash:   hash,
		UserID: uid,
	})
	if err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	if len(asset.Metadata) > 0 {
		if err := json.Unmarshal(asset.Metadata, &metadata); err != nil {
			metadata = nil
		}
	}

	return &Asset{
		ID:        pgUUIDToString(asset.ID),
		UserID:    pgUUIDToString(asset.UserID),
		Name:      asset.Name,
		Type:      asset.Type,
		MimeType:  asset.MimeType,
		Size:      asset.Size,
		Hash:      asset.Hash,
		Tags:      asset.Tags,
		Metadata:  metadata,
		CreatedAt: asset.CreatedAt.Time,
		UpdatedAt: asset.UpdatedAt.Time,
	}, nil
}

func (r *Repository) Create(ctx context.Context, userID string, req *CreateAssetRequest) (*Asset, error) {
	uid, err := stringToPgUUID(userID)
	if err != nil {
		return nil, err
	}

	var metadataJSON []byte
	if req.Metadata != nil {
		metadataJSON, err = json.Marshal(req.Metadata)
		if err != nil {
			return nil, err
		}
	}

	asset, err := r.queries.CreateAsset(ctx, db.CreateAssetParams{
		UserID:   uid,
		Name:     req.Name,
		Type:     req.Type,
		MimeType: req.MimeType,
		Size:     req.Size,
		Hash:     req.Hash,
		Tags:     req.Tags,
		Metadata: metadataJSON,
	})
	if err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	if len(asset.Metadata) > 0 {
		if err := json.Unmarshal(asset.Metadata, &metadata); err != nil {
			metadata = nil
		}
	}

	return &Asset{
		ID:        pgUUIDToString(asset.ID),
		UserID:    pgUUIDToString(asset.UserID),
		Name:      asset.Name,
		Type:      asset.Type,
		MimeType:  asset.MimeType,
		Size:      asset.Size,
		Hash:      asset.Hash,
		Tags:      asset.Tags,
		Metadata:  metadata,
		CreatedAt: asset.CreatedAt.Time,
		UpdatedAt: asset.UpdatedAt.Time,
	}, nil
}

func (r *Repository) UpdateTags(ctx context.Context, assetID string, tags []string) error {
	aid, err := stringToPgUUID(assetID)
	if err != nil {
		return err
	}

	_, err = r.queries.UpdateAssetTags(ctx, db.UpdateAssetTagsParams{
		ID:   aid,
		Tags: tags,
	})
	return err
}

func (r *Repository) Delete(ctx context.Context, assetID string) error {
	aid, err := stringToPgUUID(assetID)
	if err != nil {
		return err
	}

	return r.queries.DeleteAsset(ctx, aid)
}

func (r *Repository) SearchByName(ctx context.Context, userID, query string, limit, offset int) ([]*Asset, error) {
	uid, err := stringToPgUUID(userID)
	if err != nil {
		return nil, err
	}

	assets, err := r.queries.SearchAssetsByName(ctx, db.SearchAssetsByNameParams{
		UserID: uid,
		Name:   "%" + query + "%",
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	result := make([]*Asset, 0, len(assets))
	for _, a := range assets {
		var metadata map[string]interface{}
		if len(a.Metadata) > 0 {
			if err := json.Unmarshal(a.Metadata, &metadata); err != nil {
				metadata = nil
			}
		}

		result = append(result, &Asset{
			ID:        pgUUIDToString(a.ID),
			UserID:    pgUUIDToString(a.UserID),
			Name:      a.Name,
			Type:      a.Type,
			MimeType:  a.MimeType,
			Size:      a.Size,
			Hash:      a.Hash,
			Tags:      a.Tags,
			Metadata:  metadata,
			CreatedAt: a.CreatedAt.Time,
			UpdatedAt: a.UpdatedAt.Time,
		})
	}

	return result, nil
}

func (r *Repository) FilterByType(ctx context.Context, userID, assetType string, limit, offset int) ([]*Asset, error) {
	uid, err := stringToPgUUID(userID)
	if err != nil {
		return nil, err
	}

	assets, err := r.queries.FilterAssetsByType(ctx, db.FilterAssetsByTypeParams{
		UserID: uid,
		Type:   assetType,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	result := make([]*Asset, 0, len(assets))
	for _, a := range assets {
		var metadata map[string]interface{}
		if len(a.Metadata) > 0 {
			if err := json.Unmarshal(a.Metadata, &metadata); err != nil {
				metadata = nil
			}
		}

		result = append(result, &Asset{
			ID:        pgUUIDToString(a.ID),
			UserID:    pgUUIDToString(a.UserID),
			Name:      a.Name,
			Type:      a.Type,
			MimeType:  a.MimeType,
			Size:      a.Size,
			Hash:      a.Hash,
			Tags:      a.Tags,
			Metadata:  metadata,
			CreatedAt: a.CreatedAt.Time,
			UpdatedAt: a.UpdatedAt.Time,
		})
	}

	return result, nil
}

// Storage location methods

func (r *Repository) GetLocationsByAssetID(ctx context.Context, assetID string) ([]*StorageLocation, error) {
	aid, err := stringToPgUUID(assetID)
	if err != nil {
		return nil, err
	}

	locations, err := r.queries.GetLocationsByAssetID(ctx, aid)
	if err != nil {
		return nil, err
	}

	result := make([]*StorageLocation, 0, len(locations))
	for _, l := range locations {
		var metadata map[string]interface{}
		if len(l.Metadata) > 0 {
			if err := json.Unmarshal(l.Metadata, &metadata); err != nil {
				metadata = nil
			}
		}

		result = append(result, &StorageLocation{
			ID:          pgUUIDToString(l.ID),
			AssetID:     pgUUIDToString(l.AssetID),
			ProviderID:  pgUUIDToString(l.ProviderID),
			ProviderKey: l.ProviderKey,
			URL:         l.Url,
			Metadata:    metadata,
			CreatedAt:   l.CreatedAt.Time,
		})
	}

	return result, nil
}

func (r *Repository) CreateLocation(ctx context.Context, location *StorageLocation) (*StorageLocation, error) {
	aid, err := stringToPgUUID(location.AssetID)
	if err != nil {
		return nil, err
	}

	pid, err := stringToPgUUID(location.ProviderID)
	if err != nil {
		return nil, err
	}

	var metadataJSON []byte
	if location.Metadata != nil {
		metadataJSON, err = json.Marshal(location.Metadata)
		if err != nil {
			return nil, err
		}
	}

	loc, err := r.queries.CreateStorageLocation(ctx, db.CreateStorageLocationParams{
		AssetID:     aid,
		ProviderID:  pid,
		ProviderKey: location.ProviderKey,
		Url:         location.URL,
		Metadata:    metadataJSON,
	})
	if err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	if len(loc.Metadata) > 0 {
		if err := json.Unmarshal(loc.Metadata, &metadata); err != nil {
			metadata = nil
		}
	}

	return &StorageLocation{
		ID:          pgUUIDToString(loc.ID),
		AssetID:     pgUUIDToString(loc.AssetID),
		ProviderID:  pgUUIDToString(loc.ProviderID),
		ProviderKey: loc.ProviderKey,
		URL:         loc.Url,
		Metadata:    metadata,
		CreatedAt:   loc.CreatedAt.Time,
	}, nil
}

func (r *Repository) DeleteLocation(ctx context.Context, locationID string) error {
	lid, err := stringToPgUUID(locationID)
	if err != nil {
		return err
	}

	return r.queries.DeleteLocation(ctx, lid)
}
