package rbac

import (
	"context"
	"errors"

	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

// Service holds RBAC administration logic.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// --- roles ---

func (s *Service) ListRoles(ctx context.Context) ([]Role, error) {
	return s.repo.ListRoles(ctx)
}

func (s *Service) GetRole(ctx context.Context, id int64) (Role, error) {
	role, err := s.repo.GetRole(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return Role{}, lib.ErrNotFound("role not found")
	}
	return role, err
}

func (s *Service) CreateRole(ctx context.Context, in CreateRoleInput) (Role, error) {
	if err := lib.Validate(in); err != nil {
		return Role{}, err
	}
	return s.repo.CreateRole(ctx, in)
}

func (s *Service) UpdateRole(ctx context.Context, id int64, in UpdateRoleInput) (Role, error) {
	if err := lib.Validate(in); err != nil {
		return Role{}, err
	}
	role, err := s.repo.UpdateRole(ctx, id, in)
	if errors.Is(err, ErrNotFound) {
		return Role{}, lib.ErrNotFound("role not found")
	}
	return role, err
}

// DeleteRole removes a non-system role.
func (s *Service) DeleteRole(ctx context.Context, id int64) error {
	role, err := s.repo.GetRole(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return lib.ErrNotFound("role not found")
	}
	if err != nil {
		return err
	}
	if role.IsSystem {
		return lib.ErrForbidden("system roles cannot be deleted")
	}
	return s.repo.DeleteRole(ctx, id)
}

// --- permissions ---

func (s *Service) ListPermissions(ctx context.Context) ([]Permission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *Service) CreatePermission(ctx context.Context, in CreatePermissionInput) (Permission, error) {
	if err := lib.Validate(in); err != nil {
		return Permission{}, err
	}
	return s.repo.CreatePermission(ctx, in)
}

// --- grants ---

func (s *Service) ListRolePermissions(ctx context.Context, roleID int64) ([]RolePermission, error) {
	if _, err := s.GetRole(ctx, roleID); err != nil {
		return nil, err
	}
	return s.repo.ListRolePermissions(ctx, roleID)
}

func (s *Service) GrantPermission(ctx context.Context, roleID int64, in GrantInput) error {
	if err := lib.Validate(in); err != nil {
		return err
	}
	if _, err := s.GetRole(ctx, roleID); err != nil {
		return err
	}
	return s.repo.GrantPermission(ctx, roleID, in.PermissionID, in.Scope)
}

func (s *Service) RevokePermission(ctx context.Context, roleID, permissionID int64) error {
	return s.repo.RevokePermission(ctx, roleID, permissionID)
}

// --- assignments ---

func (s *Service) ListUserRoles(ctx context.Context, userID int64) ([]Role, error) {
	return s.repo.ListUserRoles(ctx, userID)
}

func (s *Service) AssignRole(ctx context.Context, userID int64, in AssignRoleInput, grantedBy *int64) error {
	if err := lib.Validate(in); err != nil {
		return err
	}
	if _, err := s.GetRole(ctx, in.RoleID); err != nil {
		return err
	}
	return s.repo.AssignRole(ctx, userID, in.RoleID, grantedBy)
}

func (s *Service) RevokeRole(ctx context.Context, userID, roleID int64) error {
	return s.repo.RevokeRole(ctx, userID, roleID)
}
