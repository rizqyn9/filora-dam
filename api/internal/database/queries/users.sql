-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET name = $2, avatar_url = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdatePassword :exec
UPDATE users
SET password_hash = $2, updated_at = now()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ListUserRoles :many
SELECT role_name FROM user_roles WHERE user_id = $1;

-- name: AssignRole :exec
INSERT INTO user_roles (user_id, role_name)
VALUES ($1, $2)
ON CONFLICT (user_id, role_name) DO NOTHING;

-- name: RevokeRole :exec
DELETE FROM user_roles WHERE user_id = $1 AND role_name = $2;
