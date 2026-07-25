package folder

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

// Spaces exposes space membership checks (implemented by space.Service).
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

func (s *Service) Create(ctx context.Context, userID int64, spaceID int64, in CreateFolderInput) (Folder, error) {
	if err := lib.Validate(in); err != nil {
		return Folder{}, err
	}
	if err := s.access(ctx, userID, spaceID, "create", rankEditor); err != nil {
		return Folder{}, err
	}

	// Build materialized path for the new folder.
	path := "/"
	if in.ParentID != nil {
		parent, err := s.repo.GetByID(ctx, *in.ParentID)
		if errors.Is(err, ErrFolderNotFound) {
			return Folder{}, lib.ErrNotFound("parent folder not found")
		}
		if err != nil {
			return Folder{}, err
		}
		if parent.SpaceID != spaceID {
			return Folder{}, lib.ErrBadRequest("parent folder does not belong to this space")
		}
		path = parent.Path + strconv.FormatInt(parent.ID, 10) + "/"
	}

	f, err := s.repo.Create(ctx, spaceID, in.ParentID, userID, in.Name, path)
	if err != nil {
		return Folder{}, err
	}
	return f, nil
}

func (s *Service) Get(ctx context.Context, userID int64, id int64) (Folder, error) {
	f, err := s.load(ctx, id)
	if err != nil {
		return Folder{}, err
	}
	if err := s.access(ctx, userID, f.SpaceID, "read", rankViewer); err != nil {
		return Folder{}, err
	}
	return f, nil
}

func (s *Service) ListChildren(ctx context.Context, userID int64, spaceID int64, parentID *int64) ([]Folder, error) {
	if err := s.access(ctx, userID, spaceID, "read", rankViewer); err != nil {
		return nil, err
	}
	return s.repo.ListChildren(ctx, spaceID, parentID)
}

func (s *Service) Rename(ctx context.Context, userID int64, id int64, in UpdateFolderInput) (Folder, error) {
	if err := lib.Validate(in); err != nil {
		return Folder{}, err
	}
	f, err := s.load(ctx, id)
	if err != nil {
		return Folder{}, err
	}
	if err := s.access(ctx, userID, f.SpaceID, "update", rankEditor); err != nil {
		return Folder{}, err
	}
	return s.repo.Rename(ctx, id, in.Name)
}

func (s *Service) Move(ctx context.Context, userID int64, id int64, in MoveFolderInput) (Folder, error) {
	f, err := s.load(ctx, id)
	if err != nil {
		return Folder{}, err
	}
	if err := s.access(ctx, userID, f.SpaceID, "update", rankEditor); err != nil {
		return Folder{}, err
	}

	// Prevent moving folder into itself or its own subtree.
	if in.ParentID != nil {
		if *in.ParentID == id {
			return Folder{}, lib.ErrBadRequest("cannot move a folder into itself")
		}
		// Check the target parent is not a descendant of this folder.
		target, err := s.repo.GetByID(ctx, *in.ParentID)
		if errors.Is(err, ErrFolderNotFound) {
			return Folder{}, lib.ErrNotFound("target folder not found")
		}
		if err != nil {
			return Folder{}, err
		}
		if target.SpaceID != f.SpaceID {
			return Folder{}, lib.ErrBadRequest("target folder does not belong to the same space")
		}
		// If target's path contains our ID, it's a descendant → circular.
		selfSegment := "/" + strconv.FormatInt(id, 10) + "/"
		currentPath := target.Path + strconv.FormatInt(target.ID, 10) + "/"
		if contains(currentPath, selfSegment) {
			return Folder{}, lib.ErrBadRequest("cannot move a folder into its own subtree")
		}
	}

	// Compute old and new path prefixes for subtree update.
	oldPathPrefix := f.Path + strconv.FormatInt(f.ID, 10) + "/"

	newPath := "/"
	if in.ParentID != nil {
		parent, _ := s.repo.GetByID(ctx, *in.ParentID)
		newPath = parent.Path + strconv.FormatInt(parent.ID, 10) + "/"
	}
	newPathPrefix := newPath + strconv.FormatInt(f.ID, 10) + "/"

	// Move the folder itself.
	moved, err := s.repo.Move(ctx, id, in.ParentID, newPath)
	if err != nil {
		return Folder{}, err
	}

	// Update all descendant paths.
	if err := s.repo.UpdateSubtreePaths(ctx, f.SpaceID, oldPathPrefix, newPathPrefix); err != nil {
		return Folder{}, fmt.Errorf("failed to update subtree paths: %w", err)
	}

	return moved, nil
}

func (s *Service) Delete(ctx context.Context, userID int64, id int64) error {
	f, err := s.load(ctx, id)
	if err != nil {
		return err
	}
	if err := s.access(ctx, userID, f.SpaceID, "delete", rankEditor); err != nil {
		return err
	}
	// CASCADE in the DB will delete sub-folders. Assets get folder_id SET NULL.
	return s.repo.Delete(ctx, id)
}

func (s *Service) Breadcrumb(ctx context.Context, userID int64, id int64) ([]BreadcrumbItem, error) {
	f, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.access(ctx, userID, f.SpaceID, "read", rankViewer); err != nil {
		return nil, err
	}
	// Get ancestors from the path, then append the folder itself.
	items, err := s.repo.GetBreadcrumb(ctx, f.Path)
	if err != nil {
		return nil, err
	}
	items = append(items, BreadcrumbItem{ID: f.ID, Name: f.Name})
	return items, nil
}

// --- helpers ---

func (s *Service) load(ctx context.Context, id int64) (Folder, error) {
	f, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, ErrFolderNotFound) {
		return Folder{}, lib.ErrNotFound("folder not found")
	}
	return f, err
}

func (s *Service) access(ctx context.Context, userID, spaceID int64, action string, minRank int) error {
	dec, err := s.authz.Authorize(ctx, userID, "folder", action)
	if err != nil {
		return err
	}
	if !dec.Allowed {
		return lib.ErrForbidden("insufficient permission: folder:" + action)
	}
	if dec.Scope == auth.ScopeAll {
		return nil
	}
	// Own scope: check space membership.
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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
