package storage

import "time"

// Provider is the domain view of a storage account. Credentials are never
// exposed in responses.
type Provider struct {
	ID        int64     `json:"id"`
	Layer     string    `json:"layer"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Quota     *int64    `json:"quota,omitempty"`
	Used      int64     `json:"used"`
	IsActive  bool      `json:"is_active"`
	CreatedBy *int64    `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AccountUsage is a per-account usage summary (from the storage_account_usage view).
type AccountUsage struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Layer         string  `json:"layer"`
	Type          string  `json:"type"`
	IsActive      bool    `json:"is_active"`
	Quota         *int64  `json:"quota,omitempty"`
	Used          int64   `json:"used"`
	UsedPercent   float64 `json:"used_percent"`
	LocationCount int64   `json:"location_count"`
	StoredCount   int64   `json:"stored_count"`
	PendingCount  int64   `json:"pending_count"`
	FailedCount   int64   `json:"failed_count"`
}

// --- inputs ---

type CreateProviderInput struct {
	Layer       string         `json:"layer" validate:"required,oneof=serving archive"`
	Name        string         `json:"name" validate:"required,min=1,max=255"`
	Type        string         `json:"type" validate:"required,oneof=cloudinary imagekit r2 gcs"`
	Credentials map[string]any `json:"credentials" validate:"required"`
	Quota       *int64         `json:"quota" validate:"omitempty,min=0"`
}

type UpdateProviderInput struct {
	Name        string         `json:"name" validate:"required,min=1,max=255"`
	Credentials map[string]any `json:"credentials" validate:"omitempty"`
	Quota       *int64         `json:"quota" validate:"omitempty,min=0"`
	IsActive    *bool          `json:"is_active"`
}
