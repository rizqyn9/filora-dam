-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    email,
    name,
    password_hash
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: UpdateUserStorageUsed :one
UPDATE users
SET storage_used = $2
WHERE id = $1
RETURNING *;

-- name: GetUserQuota :one
SELECT storage_quota, storage_used
FROM users
WHERE id = $1;
