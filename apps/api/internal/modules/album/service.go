package album

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

// GalleryAccess exposes the parent gallery's membership (implemented by gallery.Service).
type GalleryAccess interface {
	RoleOf(ctx context.Context, galleryID, userID int64) (string, bool, error)
}

type Service struct {
	repo      *Repository
	authz     *auth.Authorizer
	galleries GalleryAccess
}

func NewService(repo *Repository, authz *auth.Authorizer, galleries GalleryAccess) *Service {
	return &Service{repo: repo, authz: authz, galleries: galleries}
}

func (s *Service) Create(ctx context.Context, userID, galleryID int64, in CreateAlbumInput) (Album, error) {
	if err := lib.Validate(in); err != nil {
		return Album{}, err
	}
	if err := s.authz.Require(ctx, userID, "album", "create"); err != nil {
		return Album{}, err
	}
	// Must be at least an editor of the parent gallery.
	role, ok, err := s.galleries.RoleOf(ctx, galleryID, userID)
	if err != nil {
		return Album{}, err
	}
	if !ok || rank(role) < rankEditor {
		return Album{}, lib.ErrForbidden("requires editor access to the gallery")
	}
	return s.repo.CreateWithOwner(ctx, galleryID, userID, in.Name, in.Description)
}

func (s *Service) ListByGallery(ctx context.Context, userID, galleryID int64) ([]Album, error) {
	role, ok, err := s.galleries.RoleOf(ctx, galleryID, userID)
	if err != nil {
		return nil, err
	}
	if !ok && !s.hasAll(ctx, userID, "read") {
		return nil, lib.ErrForbidden("you are not a member of this gallery")
	}
	_ = role
	return s.repo.ListByGallery(ctx, galleryID)
}

func (s *Service) Get(ctx context.Context, userID, albumID int64) (Album, error) {
	a, err := s.load(ctx, albumID)
	if err != nil {
		return Album{}, err
	}
	if err := s.access(ctx, userID, a, "read", rankViewer); err != nil {
		return Album{}, err
	}
	return a, nil
}

func (s *Service) Update(ctx context.Context, userID, albumID int64, in UpdateAlbumInput) (Album, error) {
	if err := lib.Validate(in); err != nil {
		return Album{}, err
	}
	a, err := s.load(ctx, albumID)
	if err != nil {
		return Album{}, err
	}
	if err := s.access(ctx, userID, a, "update", rankEditor); err != nil {
		return Album{}, err
	}
	updated, err := s.repo.Update(ctx, albumID, in.Name, in.Description, in.CoverAssetID)
	if errors.Is(err, ErrAlbumNotFound) {
		return Album{}, lib.ErrNotFound("album not found")
	}
	return updated, err
}

func (s *Service) Delete(ctx context.Context, userID, albumID int64) error {
	a, err := s.load(ctx, albumID)
	if err != nil {
		return err
	}
	if err := s.access(ctx, userID, a, "delete", rankOwner); err != nil {
		return err
	}
	return s.repo.Delete(ctx, albumID)
}

func (s *Service) ListMembers(ctx context.Context, userID, albumID int64) ([]Member, error) {
	a, err := s.load(ctx, albumID)
	if err != nil {
		return nil, err
	}
	if err := s.access(ctx, userID, a, "read", rankViewer); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(ctx, albumID)
}

func (s *Service) AddMember(ctx context.Context, actorID, albumID int64, in AddMemberInput) error {
	if err := lib.Validate(in); err != nil {
		return err
	}
	a, err := s.load(ctx, albumID)
	if err != nil {
		return err
	}
	if err := s.access(ctx, actorID, a, "invite", rankOwner); err != nil {
		return err
	}
	return s.repo.UpsertMember(ctx, albumID, in.UserID, db.MemberRole(in.Role), &actorID)
}

func (s *Service) RemoveMember(ctx context.Context, actorID, albumID, targetUserID int64) error {
	a, err := s.load(ctx, albumID)
	if err != nil {
		return err
	}
	if err := s.access(ctx, actorID, a, "invite", rankOwner); err != nil {
		return err
	}
	if a.OwnerID == targetUserID {
		return lib.ErrForbidden("the album owner cannot be removed")
	}
	ok, err := s.repo.RemoveMember(ctx, albumID, targetUserID)
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrNotFound("member not found")
	}
	return nil
}

func (s *Service) AddAsset(ctx context.Context, userID, albumID int64, in AddAssetInput) error {
	if err := lib.Validate(in); err != nil {
		return err
	}
	a, err := s.load(ctx, albumID)
	if err != nil {
		return err
	}
	if err := s.access(ctx, userID, a, "update", rankEditor); err != nil {
		return err
	}
	assetID, err := uuid.Parse(in.AssetID)
	if err != nil {
		return lib.ErrBadRequest("invalid asset id")
	}
	return s.repo.AddAsset(ctx, albumID, assetID, &userID, in.SortOrder)
}

func (s *Service) RemoveAsset(ctx context.Context, userID, albumID int64, assetID uuid.UUID) error {
	a, err := s.load(ctx, albumID)
	if err != nil {
		return err
	}
	if err := s.access(ctx, userID, a, "update", rankEditor); err != nil {
		return err
	}
	ok, err := s.repo.RemoveAsset(ctx, albumID, assetID)
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrNotFound("asset not in album")
	}
	return nil
}

func (s *Service) ListAssetIDs(ctx context.Context, userID, albumID int64) ([]uuid.UUID, error) {
	a, err := s.load(ctx, albumID)
	if err != nil {
		return nil, err
	}
	if err := s.access(ctx, userID, a, "read", rankViewer); err != nil {
		return nil, err
	}
	return s.repo.ListAssetIDs(ctx, albumID)
}

// --- helpers ---

func (s *Service) load(ctx context.Context, albumID int64) (Album, error) {
	a, err := s.repo.GetByID(ctx, albumID)
	if errors.Is(err, ErrAlbumNotFound) {
		return Album{}, lib.ErrNotFound("album not found")
	}
	return a, err
}

func (s *Service) hasAll(ctx context.Context, userID int64, action string) bool {
	dec, err := s.authz.Authorize(ctx, userID, "album", action)
	return err == nil && dec.Allowed && dec.Scope == auth.ScopeAll
}

// access enforces album:<action> globally, then (own scope) the highest of the
// user's album membership and parent-gallery membership must meet minRank.
func (s *Service) access(ctx context.Context, userID int64, a Album, action string, minRank int) error {
	dec, err := s.authz.Authorize(ctx, userID, "album", action)
	if err != nil {
		return err
	}
	if !dec.Allowed {
		return lib.ErrForbidden("insufficient permission: album:" + action)
	}
	if dec.Scope == auth.ScopeAll {
		return nil
	}

	best := 0
	if role, err := s.repo.GetMemberRole(ctx, a.ID, userID); err == nil {
		best = max(best, rank(string(role)))
	} else if !errors.Is(err, ErrNotMember) {
		return err
	}
	if grole, ok, err := s.galleries.RoleOf(ctx, a.GalleryID, userID); err != nil {
		return err
	} else if ok {
		best = max(best, rank(grole))
	}

	if best < minRank {
		return lib.ErrForbidden("insufficient access to this album")
	}
	return nil
}

const (
	rankViewer = 1
	rankEditor = 2
	rankOwner  = 3
)

func rank(role string) int {
	switch role {
	case "owner":
		return rankOwner
	case "editor":
		return rankEditor
	case "viewer":
		return rankViewer
	}
	return 0
}
