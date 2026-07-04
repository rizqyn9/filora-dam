-- name: CreateGallery :one
INSERT INTO galleries (owner_id, name, description, is_default)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetGalleryByID :one
SELECT * FROM galleries WHERE id = $1;

-- name: GetDefaultGallery :one
SELECT * FROM galleries WHERE owner_id = $1 AND is_default = TRUE;

-- name: ListGalleriesForUser :many
SELECT g.*
FROM galleries g
JOIN gallery_members gm ON gm.gallery_id = g.id
WHERE gm.user_id = $1
ORDER BY g.created_at;

-- name: UpdateGallery :one
UPDATE galleries
SET name = $2, description = $3
WHERE id = $1
RETURNING *;

-- name: DeleteGallery :exec
DELETE FROM galleries WHERE id = $1;

-- name: UpsertGalleryMember :exec
INSERT INTO gallery_members (gallery_id, user_id, role, invited_by)
VALUES ($1, $2, $3, $4)
ON CONFLICT (gallery_id, user_id) DO UPDATE SET role = EXCLUDED.role;

-- name: GetGalleryMember :one
SELECT * FROM gallery_members WHERE gallery_id = $1 AND user_id = $2;

-- name: ListGalleryMembers :many
SELECT gm.gallery_id, gm.user_id, gm.role, gm.created_at,
       u.email, u.name, u.avatar_url
FROM gallery_members gm
JOIN users u ON u.id = gm.user_id
WHERE gm.gallery_id = $1
ORDER BY gm.created_at;

-- name: UpdateGalleryMemberRole :execrows
UPDATE gallery_members SET role = $3 WHERE gallery_id = $1 AND user_id = $2;

-- name: RemoveGalleryMember :execrows
DELETE FROM gallery_members WHERE gallery_id = $1 AND user_id = $2;

-- name: CreateGalleryInvitation :one
INSERT INTO invitations (gallery_id, email, role, token, invited_by, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetInvitationByToken :one
SELECT * FROM invitations WHERE token = $1;

-- name: ListGalleryInvitations :many
SELECT * FROM invitations
WHERE gallery_id = $1 AND status = 'pending'
ORDER BY created_at DESC;

-- name: MarkInvitationAccepted :exec
UPDATE invitations
SET status = 'accepted', accepted_user_id = $2, accepted_at = now()
WHERE id = $1;

-- name: RevokeGalleryInvitation :execrows
UPDATE invitations
SET status = 'revoked'
WHERE id = $1 AND gallery_id = $2 AND status = 'pending';

-- name: AddGalleryUsed :exec
UPDATE galleries SET storage_used = storage_used + $2 WHERE id = $1;
