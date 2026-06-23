package storage

import (
	"io"
	"time"
)

// Provider represents a storage provider
type Provider struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // cloudinary, imagekit, r2
	Credentials map[string]interface{} `json:"-"`     // Hidden from JSON
	Quota       *int64                 `json:"quota"`
	Used        int64                  `json:"used"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CreateProviderRequest represents request to create a provider
type CreateProviderRequest struct {
	Name        string                 `json:"name" validate:"required,min=2,max=255"`
	Type        string                 `json:"type" validate:"required,oneof=cloudinary imagekit r2"`
	Credentials map[string]interface{} `json:"credentials" validate:"required"`
	Quota       *int64                 `json:"quota"`
}

// UpdateProviderRequest represents request to update a provider
type UpdateProviderRequest struct {
	Name        *string                `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Credentials map[string]interface{} `json:"credentials,omitempty"`
	Quota       *int64                 `json:"quota,omitempty"`
	IsActive    *bool                  `json:"is_active,omitempty"`
}

// UploadInput represents file upload input
type UploadInput struct {
	File     io.Reader
	Filename string
	MimeType string
	Size     int64
}

// UploadResult represents upload result from provider
type UploadResult struct {
	Key      string                 `json:"key"`       // Provider's file key/ID
	URL      string                 `json:"url"`       // Public URL
	Size     int64                  `json:"size"`      // Actual file size
	Metadata map[string]interface{} `json:"metadata"`  // Provider-specific metadata
}

// StorageLocation represents where an asset is stored
type StorageLocation struct {
	ID          string                 `json:"id"`
	AssetID     string                 `json:"asset_id"`
	ProviderID  string                 `json:"provider_id"`
	ProviderKey string                 `json:"provider_key"`
	URL         string                 `json:"url"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}
