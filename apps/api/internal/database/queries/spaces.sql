-- name: CreateSpace :one
INSERT INTO spaces (owner_id, name, description, is_default)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSpaceByID :one
SELECT * FROM spaces WHERE id = $1;

-- name: GetDefaultSpace :one
SELECT * FROM spaces WHERE owner_id = $1 AND is_default = TRUE;

-- name: ListSpacesForUser :many
SELECT s.*
FROM spaces s
JOIN space_members sm ON sm.space_id = s.id
WHERE sm.user_id = $1
ORDER BY s.created_at;

-- name: UpdateSpace :one
UPDATE spaces
SET name = $2, description = $3
WHERE id = $1
RETURNING *;

-- name: DeleteSpace :exec
DELETE FROM spaces WHERE id = $1;

-- name: UpsertSpaceMember :exec
INSERT INTO space_members (space_id, user_id, role, invited_by)
VALUES ($1, $2, $3, $4)
ON CONFLICT (space_id, user_id) DO UPDATE SET role = EXCLUDED.role;

-- name: GetSpaceMember :one
SELECT * FROM space_members WHERE space_id = $1 AND user_id = $2;

-- name: ListSpaceMembers :many
SELECT sm.space_id, sm.user_id, sm.role, sm.created_at,
       u.email, u.name, u.avatar_url
FROM space_members sm
JOIN users u ON u.id = sm.user_id
WHERE sm.space_id = $1
ORDER BY sm.created_at;

-- name: UpdateSpaceMemberRole :execrows
UPDATE space_members SET role = $3 WHERE space_id = $1 AND user_id = $2;

-- name: RemoveSpaceMember :execrows
DELETE FROM space_members WHERE space_id = $1 AND user_id = $2;

-- name: CreateSpaceInvitation :one
INSERT INTO invitations (space_id, email, role, token, invited_by, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetInvitationByToken :one
SELECT * FROM invitations WHERE token = $1;

-- name: ListSpaceInvitations :many
SELECT * FROM invitations
WHERE space_id = $1 AND status = 'pending'
ORDER BY created_at DESC;

-- name: MarkInvitationAccepted :exec
UPDATE invitations
SET status = 'accepted', accepted_user_id = $2, accepted_at = now()
WHERE id = $1;

-- name: RevokeSpaceInvitation :execrows
UPDATE invitations
SET status = 'revoked'
WHERE id = $1 AND space_id = $2 AND status = 'pending';

-- name: AddSpaceUsed :exec
UPDATE spaces SET storage_used = storage_used + $2 WHERE id = $1;
