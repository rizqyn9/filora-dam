-- name: GetSpaceByID :one
SELECT * FROM spaces WHERE id = $1;

-- name: ListSpacesByOwner :many
SELECT * FROM spaces WHERE owner_id = $1 ORDER BY created_at;

-- name: ListSpacesByMember :many
SELECT s.* FROM spaces s
JOIN space_members sm ON sm.space_id = s.id
WHERE sm.user_id = $1
ORDER BY s.created_at;

-- name: CreateSpace :one
INSERT INTO spaces (name, owner_id, storage_quota_bytes)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateSpace :one
UPDATE spaces
SET name = $2, storage_quota_bytes = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteSpace :exec
DELETE FROM spaces WHERE id = $1;

-- name: IncrementSpaceUsage :exec
UPDATE spaces
SET storage_used_bytes = storage_used_bytes + $2, updated_at = now()
WHERE id = $1;

-- name: DecrementSpaceUsage :exec
UPDATE spaces
SET storage_used_bytes = GREATEST(storage_used_bytes - $2, 0), updated_at = now()
WHERE id = $1;

-- name: GetSpaceMember :one
SELECT * FROM space_members WHERE space_id = $1 AND user_id = $2;

-- name: ListSpaceMembers :many
SELECT sm.*, u.email, u.name, u.avatar_url
FROM space_members sm
JOIN users u ON u.id = sm.user_id
WHERE sm.space_id = $1
ORDER BY sm.joined_at;

-- name: AddSpaceMember :one
INSERT INTO space_members (space_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateSpaceMemberRole :exec
UPDATE space_members SET role = $3 WHERE space_id = $1 AND user_id = $2;

-- name: RemoveSpaceMember :exec
DELETE FROM space_members WHERE space_id = $1 AND user_id = $2;
