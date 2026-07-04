package asset

import (
	"encoding/json"
	"io"
	"time"

	"github.com/google/uuid"
)

// Asset is the domain view of an asset.
type Asset struct {
	ID         uuid.UUID       `json:"id"`
	GalleryID  int64           `json:"gallery_id"`
	UploadedBy *int64          `json:"uploaded_by,omitempty"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	MimeType   string          `json:"mime_type"`
	Size       int64           `json:"size"`
	Hash       string          `json:"hash"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	DeletedAt  *time.Time      `json:"deleted_at,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// UploadInput carries a validated file ready to be stored.
type UploadInput struct {
	GalleryID int64
	Name      string
	Type      string
	MimeType  string
	Size      int64
	Hash      string
	Reader    io.Reader
}

// UpdateAssetInput updates editable metadata.
type UpdateAssetInput struct {
	Name string `json:"name" validate:"required,min=1,max=500"`
}

// ListResult is a paginated list of assets.
type ListResult struct {
	Assets []Asset `json:"assets"`
	Total  int64   `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}
