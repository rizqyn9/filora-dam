-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByClerkID :one
SELECT * FROM users WHERE clerk_id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (clerk_id, email, name, avatar_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET email = $2, name = $3, avatar_url = $4, updated_at = now()
WHERE clerk_id = $1
RETURNING *;

-- name: DeleteUserByClerkID :exec
DELETE FROM users WHERE clerk_id = $1;

-- name: ListUserRoles :many
SELECT role_name FROM user_roles WHERE user_id = $1;

-- name: AssignRole :exec
INSERT INTO user_roles (user_id, role_name)
VALUES ($1, $2)
ON CONFLICT (user_id, role_name) DO NOTHING;

-- name: RevokeRole :exec
DELETE FROM user_roles WHERE user_id = $1 AND role_name = $2;
