package rbac

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
)

// ErrNotFound is returned when a role lookup finds nothing.
var ErrNotFound = errors.New("not found")

// Repository provides persistence for roles, permissions, grants, and assignments.
type Repository struct {
	q *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

// --- roles ---

func (r *Repository) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := r.q.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Role, 0, len(rows))
	for _, row := range rows {
		out = append(out, toRole(row))
	}
	return out, nil
}

func (r *Repository) GetRole(ctx context.Context, id int64) (Role, error) {
	row, err := r.q.GetRoleByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Role{}, ErrNotFound
	}
	if err != nil {
		return Role{}, err
	}
	return toRole(row), nil
}

func (r *Repository) CreateRole(ctx context.Context, in CreateRoleInput) (Role, error) {
	row, err := r.q.CreateRole(ctx, db.CreateRoleParams{
		Slug:        in.Slug,
		Name:        in.Name,
		Description: in.Description,
	})
	if err != nil {
		return Role{}, err
	}
	return toRole(row), nil
}

func (r *Repository) UpdateRole(ctx context.Context, id int64, in UpdateRoleInput) (Role, error) {
	row, err := r.q.UpdateRole(ctx, db.UpdateRoleParams{
		ID:          id,
		Name:        in.Name,
		Description: in.Description,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Role{}, ErrNotFound
	}
	if err != nil {
		return Role{}, err
	}
	return toRole(row), nil
}

func (r *Repository) DeleteRole(ctx context.Context, id int64) error {
	return r.q.DeleteRole(ctx, id)
}

// --- permissions ---

func (r *Repository) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := r.q.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Permission, 0, len(rows))
	for _, row := range rows {
		out = append(out, toPermission(row))
	}
	return out, nil
}

func (r *Repository) CreatePermission(ctx context.Context, in CreatePermissionInput) (Permission, error) {
	row, err := r.q.CreatePermission(ctx, db.CreatePermissionParams{
		Resource:    in.Resource,
		Action:      in.Action,
		Description: in.Description,
	})
	if err != nil {
		return Permission{}, err
	}
	return toPermission(row), nil
}

// --- grants ---

func (r *Repository) ListRolePermissions(ctx context.Context, roleID int64) ([]RolePermission, error) {
	rows, err := r.q.ListRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	out := make([]RolePermission, 0, len(rows))
	for _, row := range rows {
		out = append(out, RolePermission{
			Permission: toPermission(row.Permission),
			Scope:      string(row.Scope),
		})
	}
	return out, nil
}

func (r *Repository) GrantPermission(ctx context.Context, roleID, permissionID int64, scope string) error {
	return r.q.GrantPermission(ctx, db.GrantPermissionParams{
		RoleID:       roleID,
		PermissionID: permissionID,
		Scope:        db.PermissionScope(scope),
	})
}

func (r *Repository) RevokePermission(ctx context.Context, roleID, permissionID int64) error {
	return r.q.RevokePermission(ctx, db.RevokePermissionParams{RoleID: roleID, PermissionID: permissionID})
}

// --- assignments ---

func (r *Repository) ListUserRoles(ctx context.Context, userID int64) ([]Role, error) {
	rows, err := r.q.ListUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Role, 0, len(rows))
	for _, row := range rows {
		out = append(out, toRole(row))
	}
	return out, nil
}

func (r *Repository) AssignRole(ctx context.Context, userID, roleID int64, grantedBy *int64) error {
	return r.q.AssignRole(ctx, db.AssignRoleParams{UserID: userID, RoleID: roleID, GrantedBy: grantedBy})
}

func (r *Repository) RevokeRole(ctx context.Context, userID, roleID int64) error {
	return r.q.RevokeRole(ctx, db.RevokeRoleParams{UserID: userID, RoleID: roleID})
}

func toRole(r db.Role) Role {
	return Role{
		ID:          r.ID,
		Slug:        r.Slug,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func toPermission(p db.Permission) Permission {
	return Permission{
		ID:          p.ID,
		Resource:    p.Resource,
		Action:      p.Action,
		Description: p.Description,
	}
}
