package dashboard

import (
	"context"
	"fmt"

	"github.com/rizqynugroho9/filora-dam/api/internal/modules/account"
)

type Service struct {
	repo         *Repository
	accountRepo  *account.Repository
}

func NewService(repo *Repository, accountRepo *account.Repository) *Service {
	return &Service{
		repo:         repo,
		accountRepo:  accountRepo,
	}
}

// GetDashboard retrieves complete dashboard data for a user
func (s *Service) GetDashboard(ctx context.Context, userID string) (*DashboardResponse, error) {
	// Get statistics
	stats, err := s.repo.GetStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Get user for quota info
	user, err := s.accountRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get asset counts by type
	typeCounts, err := s.repo.GetAssetsByType(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get type counts: %w", err)
	}

	// Get recent assets
	recentRows, err := s.repo.GetRecentAssets(ctx, userID, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent assets: %w", err)
	}

	// Convert total size to int64
	var totalSize int64
	if ts, ok := stats.TotalSize.(int64); ok {
		totalSize = ts
	}

	// Convert to response models
	dashboardStats := &DashboardStats{
		TotalAssets:  stats.TotalAssets,
		TotalSize:    totalSize,
		TotalSizeGB:  formatBytes(totalSize),
		UniqueTypes:  stats.UniqueTypes,
		StorageQuota: user.StorageQuota,
		StorageUsed:  user.StorageUsed,
		StorageFree:  user.StorageQuota - user.StorageUsed,
	}

	typeCountsResponse := make([]*AssetTypeCount, 0, len(typeCounts))
	for _, tc := range typeCounts {
		typeCountsResponse = append(typeCountsResponse, &AssetTypeCount{
			Type:  tc.Type,
			Count: tc.Count,
		})
	}

	recentAssets := make([]*RecentAsset, 0, len(recentRows))
	for _, ra := range recentRows {
		recentAssets = append(recentAssets, &RecentAsset{
			ID:        ra.ID.String(),
			Name:      ra.Name,
			Type:      ra.Type,
			MimeType:  ra.MimeType,
			Size:      ra.Size,
			CreatedAt: ra.CreatedAt.Time,
		})
	}

	return &DashboardResponse{
		Stats:        dashboardStats,
		TypeCounts:   typeCountsResponse,
		RecentAssets: recentAssets,
	}, nil
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
