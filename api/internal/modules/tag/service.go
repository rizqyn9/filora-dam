package tag

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

// Spaces exposes space membership (implemented by space.Service).
type Spaces interface {
	RoleOf(ctx context.Context, spaceID, userID int64) (string, bool, error)
}

type Service struct {
	repo   *Repository
	authz  *auth.Authorizer
	spaces Spaces
}

func NewService(repo *Repository, authz *auth.Authorizer, spaces Spaces) *Service {
	return &Service{repo: repo, authz: authz, spaces: spaces}
}

func (s *Service) Create(ctx context.Context, userID, spaceID int64, in CreateTagInput) (Tag, error) {
	if err := lib.Validate(in); err != nil {
		return Tag{}, err
	}
	if err := s.requireSpaceRole(ctx, userID, spaceID, "create", rankEditor); err != nil {
		return Tag{}, err
	}
	return s.repo.Create(ctx, spaceID, in.Name, &userID)
}

func (s *Service) ListBySpace(ctx context.Context, userID, spaceID int64) ([]Tag, error) {
	if err := s.requireSpaceRole(ctx, userID, spaceID, "read", rankViewer); err != nil {
		return nil, err
	}
	return s.repo.ListBySpace(ctx, spaceID)
}

func (s *Service) Update(ctx context.Context, userID, tagID int64, in UpdateTagInput) (Tag, error) {
	if err := lib.Validate(in); err != nil {
		return Tag{}, err
	}
	t, err := s.load(ctx, tagID)
	if err != nil {
		return Tag{}, err
	}
	if err := s.requireSpaceRole(ctx, userID, t.SpaceID, "update", rankEditor); err != nil {
		return Tag{}, err
	}
	return s.repo.Update(ctx, tagID, in.Name)
}

func (s *Service) Delete(ctx context.Context, userID, tagID int64) error {
	t, err := s.load(ctx, tagID)
	if err != nil {
		return err
	}
	if err := s.requireSpaceRole(ctx, userID, t.SpaceID, "delete", rankEditor); err != nil {
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
	if err := s.requireSpaceRole(ctx, userID, t.SpaceID, "create", rankEditor); err != nil {
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
	if err := s.requireSpaceRole(ctx, userID, t.SpaceID, "delete", rankEditor); err != nil {
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

// requireSpaceRole enforces tag:<action> globally, then (own scope) a minimum
// space membership role.
func (s *Service) requireSpaceRole(ctx context.Context, userID, spaceID int64, action string, minRank int) error {
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
	role, ok, err := s.spaces.RoleOf(ctx, spaceID, userID)
	if err != nil {
		return err
	}
	if !ok || rank(role) < minRank {
		return lib.ErrForbidden("insufficient access to this space")
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
