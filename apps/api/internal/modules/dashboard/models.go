package dashboard

import (
	"time"

	"github.com/google/uuid"
)

// GalleryDashboard is the per-gallery summary.
type GalleryDashboard struct {
	Stats      GalleryStats  `json:"stats"`
	TypeCounts []TypeCount   `json:"type_counts"`
	Recent     []RecentAsset `json:"recent_assets"`
}

type GalleryStats struct {
	TotalAssets  int64 `json:"total_assets"`
	TotalSize    int64 `json:"total_size"`
	UniqueTypes  int64 `json:"unique_types"`
	StorageQuota int64 `json:"storage_quota"`
	StorageUsed  int64 `json:"storage_used"`
	StorageFree  int64 `json:"storage_free"`
}

type TypeCount struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

type RecentAsset struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

// SystemDashboard is the admin-level summary.
type SystemDashboard struct {
	ArchiveJobs ArchiveJobHealth `json:"archive_jobs"`
}

type ArchiveJobHealth struct {
	Pending   int64 `json:"pending"`
	Running   int64 `json:"running"`
	Completed int64 `json:"completed"`
	Failed    int64 `json:"failed"`
}
