package space

import (
	"time"

	"github.com/google/uuid"
)

type Space struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	OwnerID           int64     `json:"owner_id"`
	StorageQuotaBytes int64     `json:"storage_quota_bytes"`
	StorageUsedBytes  int64     `json:"storage_used_bytes"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CreateSpaceRequest struct {
	Name              string `json:"name" validate:"required,min=1,max=255"`
	StorageQuotaBytes int64  `json:"storage_quota_bytes"`
}

type UpdateSpaceRequest struct {
	Name              string `json:"name" validate:"required,min=1,max=255"`
	StorageQuotaBytes int64  `json:"storage_quota_bytes"`
}
