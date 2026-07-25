package rbac

import "time"

// Role is the domain view of an RBAC role.
type Role struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission is a catalog entry.
type Permission struct {
	ID          int64   `json:"id"`
	Resource    string  `json:"resource"`
	Action      string  `json:"action"`
	Description *string `json:"description,omitempty"`
}

// RolePermission is a permission granted to a role at a scope.
type RolePermission struct {
	Permission
	Scope string `json:"scope"`
}

// --- inputs ---

type CreateRoleInput struct {
	Slug        string  `json:"slug" validate:"required,min=2,max=64,lowercase"`
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

type UpdateRoleInput struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

type CreatePermissionInput struct {
	Resource    string  `json:"resource" validate:"required,min=1,max=64"`
	Action      string  `json:"action" validate:"required,min=1,max=64"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

type GrantInput struct {
	PermissionID int64  `json:"permission_id" validate:"required"`
	Scope        string `json:"scope" validate:"required,oneof=own all"`
}

type AssignRoleInput struct {
	RoleID int64 `json:"role_id" validate:"required"`
}
