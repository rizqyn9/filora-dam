-- name: GetUserPermissions :many
-- Effective (resource, action, scope) grants for a user across all their roles.
SELECT p.resource, p.action, rp.scope
FROM user_roles ur
JOIN role_permissions rp ON rp.role_id = ur.role_id
JOIN permissions p ON p.id = rp.permission_id
WHERE ur.user_id = $1;

-- name: ListRoles :many
SELECT * FROM roles ORDER BY id;

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1;

-- name: CreateRole :one
INSERT INTO roles (slug, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateRole :one
UPDATE roles
SET name = $2, description = $3
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;

-- name: ListPermissions :many
SELECT * FROM permissions ORDER BY resource, action;

-- name: CreatePermission :one
INSERT INTO permissions (resource, action, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListRolePermissions :many
SELECT sqlc.embed(p), rp.scope
FROM role_permissions rp
JOIN permissions p ON p.id = rp.permission_id
WHERE rp.role_id = $1
ORDER BY p.resource, p.action;

-- name: GrantPermission :exec
INSERT INTO role_permissions (role_id, permission_id, scope)
VALUES ($1, $2, $3)
ON CONFLICT (role_id, permission_id) DO UPDATE SET scope = EXCLUDED.scope;

-- name: RevokePermission :exec
DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2;

-- name: ListUserRoles :many
SELECT r.*
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = $1
ORDER BY r.id;

-- name: AssignRole :exec
INSERT INTO user_roles (user_id, role_id, granted_by)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: RevokeRole :exec
DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2;
