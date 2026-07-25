-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByClerkID :one
SELECT * FROM users
WHERE clerk_user_id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (clerk_user_id, email, name, avatar_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpsertUserByClerkID :one
INSERT INTO users (clerk_user_id, email, name, avatar_url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (clerk_user_id) DO UPDATE
SET email = EXCLUDED.email,
    name = EXCLUDED.name,
    avatar_url = EXCLUDED.avatar_url,
    is_active = TRUE
RETURNING *;

-- name: UpdateUserProfile :one
UPDATE users
SET name = $2,
    avatar_url = $3
WHERE id = $1
RETURNING *;

-- name: DeactivateUserByClerkID :exec
UPDATE users
SET is_active = FALSE
WHERE clerk_user_id = $1;

-- name: TouchUserLastSeen :exec
UPDATE users
SET last_seen_at = now()
WHERE id = $1;
