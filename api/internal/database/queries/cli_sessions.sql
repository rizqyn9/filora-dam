-- name: CreateCliSession :one
INSERT INTO cli_sessions (user_id, token_hash, label, ip_address, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetActiveSessionByTokenHash :one
SELECT * FROM cli_sessions
WHERE token_hash = $1
  AND revoked_at IS NULL
  AND (expires_at IS NULL OR expires_at > now());

-- name: ListActiveSessionsByUser :many
SELECT * FROM cli_sessions
WHERE user_id = $1 AND revoked_at IS NULL
ORDER BY created_at DESC;

-- name: TouchSessionLastUsed :exec
UPDATE cli_sessions
SET last_used_at = now()
WHERE id = $1;

-- name: RevokeSession :execrows
UPDATE cli_sessions
SET revoked_at = now()
WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL;

-- name: RevokeAllUserSessions :exec
UPDATE cli_sessions
SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;
