package dashboard

import "time"

// DashboardStats represents user dashboard statistics
type DashboardStats struct {
	TotalAssets  int64  `json:"total_assets"`
	TotalSize    int64  `json:"total_size"`
	TotalSizeGB  string `json:"total_size_gb"`
	UniqueTypes  int64  `json:"unique_types"`
	StorageQuota int64  `json:"storage_quota"`
	StorageUsed  int64  `json:"storage_used"`
	StorageFree  int64  `json:"storage_free"`
}

// AssetTypeCount represents asset count by type
type AssetTypeCount struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

// RecentAsset represents a recent asset
type RecentAsset struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

// DashboardResponse represents complete dashboard data
type DashboardResponse struct {
	Stats       *DashboardStats    `json:"stats"`
	TypeCounts  []*AssetTypeCount  `json:"type_counts"`
	RecentAssets []*RecentAsset    `json:"recent_assets"`
}
