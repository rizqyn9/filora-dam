package asset

import "time"

// Asset represents a digital asset
type Asset struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"` // image, video, document, archive, file
	MimeType  string                 `json:"mime_type"`
	Size      int64                  `json:"size"`
	Hash      string                 `json:"hash"`
	Tags      []string               `json:"tags"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Locations []*StorageLocation     `json:"locations,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// StorageLocation represents where an asset is stored
type StorageLocation struct {
	ID          string                 `json:"id"`
	AssetID     string                 `json:"asset_id"`
	ProviderID  string                 `json:"provider_id"`
	ProviderKey string                 `json:"provider_key"`
	URL         string                 `json:"url"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// CreateAssetRequest represents request to create an asset
type CreateAssetRequest struct {
	Name     string                 `json:"name" validate:"required"`
	Type     string                 `json:"type" validate:"required,oneof=image video document archive file"`
	MimeType string                 `json:"mime_type" validate:"required"`
	Size     int64                  `json:"size" validate:"required,gt=0"`
	Hash     string                 `json:"hash" validate:"required"`
	Tags     []string               `json:"tags"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTagsRequest represents request to update asset tags
type UpdateTagsRequest struct {
	Tags []string `json:"tags" validate:"required"`
}

// AssetListResponse represents paginated list of assets
type AssetListResponse struct {
	Assets []*Asset `json:"assets"`
	Total  int64    `json:"total"`
	Limit  int      `json:"limit"`
	Offset int      `json:"offset"`
}
