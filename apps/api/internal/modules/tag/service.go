package tag

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

// GalleryAccess exposes gallery membership (implemented by gallery.Service).
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

func (s *Service) Create(ctx context.Context, userID, galleryID int64, in CreateTagInput) (Tag, error) {
	if err := lib.Validate(in); err != nil {
		return Tag{}, err
	}
	if err := s.requireGalleryRole(ctx, userID, galleryID, "create", rankEditor); err != nil {
		return Tag{}, err
	}
	return s.repo.Create(ctx, galleryID, in.Name, &userID)
}

func (s *Service) ListByGallery(ctx context.Context, userID, galleryID int64) ([]Tag, error) {
	if err := s.requireGalleryRole(ctx, userID, galleryID, "read", rankViewer); err != nil {
		return nil, err
	}
	return s.repo.ListByGallery(ctx, galleryID)
}

func (s *Service) Update(ctx context.Context, userID, tagID int64, in UpdateTagInput) (Tag, error) {
	if err := lib.Validate(in); err != nil {
		return Tag{}, err
	}
	t, err := s.load(ctx, tagID)
	if err != nil {
		return Tag{}, err
	}
	if err := s.requireGalleryRole(ctx, userID, t.GalleryID, "update", rankEditor); err != nil {
		return Tag{}, err
	}
	return s.repo.Update(ctx, tagID, in.Name)
}

func (s *Service) Delete(ctx context.Context, userID, tagID int64) error {
	t, err := s.load(ctx, tagID)
	if err != nil {
		return err
	}
	if err := s.requireGalleryRole(ctx, userID, t.GalleryID, "delete", rankEditor); err != nil {
		return err
	}
	return s.repo.Delete(ctx, tagID)
}

func (s *Service) Attach(ctx context.Context, userID, tagID int64, in AttachInput) error {
	if err := lib.Validate(in); err != nil {
		return err
	}
	t, err := s.load(ctx, tagID)
	if err != nil {
		return err
	}
	if err := s.requireGalleryRole(ctx, userID, t.GalleryID, "create", rankEditor); err != nil {
		return err
	}
	assetID, err := uuid.Parse(in.AssetID)
	if err != nil {
		return lib.ErrBadRequest("invalid asset id")
	}
	return s.repo.Attach(ctx, assetID, tagID)
}

func (s *Service) Detach(ctx context.Context, userID, tagID int64, assetID uuid.UUID) error {
	t, err := s.load(ctx, tagID)
	if err != nil {
		return err
	}
	if err := s.requireGalleryRole(ctx, userID, t.GalleryID, "delete", rankEditor); err != nil {
		return err
	}
	ok, err := s.repo.Detach(ctx, assetID, tagID)
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrNotFound("tag not attached to asset")
	}
	return nil
}

func (s *Service) load(ctx context.Context, tagID int64) (Tag, error) {
	t, err := s.repo.GetByID(ctx, tagID)
	if errors.Is(err, ErrTagNotFound) {
		return Tag{}, lib.ErrNotFound("tag not found")
	}
	return t, err
}

// requireGalleryRole enforces tag:<action> globally, then (own scope) a minimum
// gallery membership role.
func (s *Service) requireGalleryRole(ctx context.Context, userID, galleryID int64, action string, minRank int) error {
	dec, err := s.authz.Authorize(ctx, userID, "tag", action)
	if err != nil {
		return err
	}
	if !dec.Allowed {
		return lib.ErrForbidden("insufficient permission: tag:" + action)
	}
	if dec.Scope == auth.ScopeAll {
		return nil
	}
	role, ok, err := s.galleries.RoleOf(ctx, galleryID, userID)
	if err != nil {
		return err
	}
	if !ok || rank(role) < minRank {
		return lib.ErrForbidden("insufficient access to this gallery")
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
