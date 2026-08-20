-- name: CreateCLISession :one
INSERT INTO cli_sessions (user_id, token_hash, label, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSessionByTokenHash :one
SELECT * FROM cli_sessions
WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now();

-- name: ListActiveSessions :many
SELECT * FROM cli_sessions
WHERE user_id = $1 AND revoked_at IS NULL
ORDER BY last_used_at DESC;

-- name: TouchSession :exec
UPDATE cli_sessions SET last_used_at = now() WHERE id = $1;

-- name: RevokeSession :exec
UPDATE cli_sessions SET revoked_at = now() WHERE id = $1;

-- name: RevokeAllUserSessions :exec
UPDATE cli_sessions SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;
