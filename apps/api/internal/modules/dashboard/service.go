package dashboard

import (
	"context"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

const recentLimit = 10

// Galleries exposes gallery membership and quota (implemented by gallery.Service).
type Galleries interface {
	RoleOf(ctx context.Context, galleryID, userID int64) (string, bool, error)
	QuotaInfo(ctx context.Context, galleryID int64) (used, quota int64, err error)
}

type Service struct {
	repo      *Repository
	authz     *auth.Authorizer
	galleries Galleries
}

func NewService(repo *Repository, authz *auth.Authorizer, galleries Galleries) *Service {
	return &Service{repo: repo, authz: authz, galleries: galleries}
}

// Gallery returns the per-gallery dashboard for a member (or admin).
func (s *Service) Gallery(ctx context.Context, userID, galleryID int64) (GalleryDashboard, error) {
	if err := s.accessGallery(ctx, userID, galleryID); err != nil {
		return GalleryDashboard{}, err
	}

	stats, err := s.repo.GalleryStats(ctx, galleryID)
	if err != nil {
		return GalleryDashboard{}, err
	}
	used, quota, err := s.galleries.QuotaInfo(ctx, galleryID)
	if err != nil {
		return GalleryDashboard{}, err
	}
	stats.StorageQuota = quota
	stats.StorageUsed = used
	stats.StorageFree = quota - used
	if stats.StorageFree < 0 {
		stats.StorageFree = 0
	}

	types, err := s.repo.TypeCounts(ctx, galleryID)
	if err != nil {
		return GalleryDashboard{}, err
	}
	recent, err := s.repo.RecentAssets(ctx, galleryID, recentLimit)
	if err != nil {
		return GalleryDashboard{}, err
	}

	return GalleryDashboard{Stats: stats, TypeCounts: types, Recent: recent}, nil
}

// System returns the admin-level dashboard (requires workspace-wide dashboard:read).
func (s *Service) System(ctx context.Context, userID int64) (SystemDashboard, error) {
	dec, err := s.authz.Authorize(ctx, userID, "dashboard", "read")
	if err != nil {
		return SystemDashboard{}, err
	}
	if !dec.Allowed || dec.Scope != auth.ScopeAll {
		return SystemDashboard{}, lib.ErrForbidden("requires workspace-wide dashboard access")
	}
	jobs, err := s.repo.ArchiveJobHealth(ctx)
	if err != nil {
		return SystemDashboard{}, err
	}
	return SystemDashboard{ArchiveJobs: jobs}, nil
}

// accessGallery enforces dashboard:read globally, then (own scope) gallery membership.
func (s *Service) accessGallery(ctx context.Context, userID, galleryID int64) error {
	dec, err := s.authz.Authorize(ctx, userID, "dashboard", "read")
	if err != nil {
		return err
	}
	if !dec.Allowed {
		return lib.ErrForbidden("insufficient permission: dashboard:read")
	}
	if dec.Scope == auth.ScopeAll {
		return nil
	}
	_, ok, err := s.galleries.RoleOf(ctx, galleryID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrForbidden("you are not a member of this gallery")
	}
	return nil
}
