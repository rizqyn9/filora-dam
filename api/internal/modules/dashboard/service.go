package dashboard

import (
	"context"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

const recentLimit = 10

// Spaces exposes space membership and quota (implemented by space.Service).
type Spaces interface {
	RoleOf(ctx context.Context, spaceID, userID int64) (string, bool, error)
	QuotaInfo(ctx context.Context, spaceID int64) (used, quota int64, err error)
}

type Service struct {
	repo   *Repository
	authz  *auth.Authorizer
	spaces Spaces
}

func NewService(repo *Repository, authz *auth.Authorizer, spaces Spaces) *Service {
	return &Service{repo: repo, authz: authz, spaces: spaces}
}

// Space returns the per-space dashboard for a member (or admin).
func (s *Service) Space(ctx context.Context, userID, spaceID int64) (SpaceDashboard, error) {
	if err := s.accessSpace(ctx, userID, spaceID); err != nil {
		return SpaceDashboard{}, err
	}

	stats, err := s.repo.SpaceStats(ctx, spaceID)
	if err != nil {
		return SpaceDashboard{}, err
	}
	used, quota, err := s.spaces.QuotaInfo(ctx, spaceID)
	if err != nil {
		return SpaceDashboard{}, err
	}
	stats.StorageQuota = quota
	stats.StorageUsed = used
	stats.StorageFree = quota - used
	if stats.StorageFree < 0 {
		stats.StorageFree = 0
	}

	types, err := s.repo.TypeCounts(ctx, spaceID)
	if err != nil {
		return SpaceDashboard{}, err
	}
	recent, err := s.repo.RecentAssets(ctx, spaceID, recentLimit)
	if err != nil {
		return SpaceDashboard{}, err
	}

	return SpaceDashboard{Stats: stats, TypeCounts: types, Recent: recent}, nil
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

// accessSpace enforces dashboard:read globally, then (own scope) space membership.
func (s *Service) accessSpace(ctx context.Context, userID, spaceID int64) error {
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
	_, ok, err := s.spaces.RoleOf(ctx, spaceID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrForbidden("you are not a member of this space")
	}
	return nil
}
