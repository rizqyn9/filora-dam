-- name: CreateInvitation :one
INSERT INTO invitations (space_id, email, role, invited_by, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetInvitationByID :one
SELECT * FROM invitations WHERE id = $1;

-- name: ListPendingInvitationsBySpace :many
SELECT * FROM invitations
WHERE space_id = $1 AND status = 'pending'
ORDER BY created_at DESC;

-- name: ListPendingInvitationsByEmail :many
SELECT * FROM invitations
WHERE email = $1 AND status = 'pending'
ORDER BY created_at DESC;

-- name: AcceptInvitation :exec
UPDATE invitations
SET status = 'accepted', updated_at = now()
WHERE id = $1;

-- name: RevokeInvitation :exec
UPDATE invitations
SET status = 'revoked', updated_at = now()
WHERE id = $1;

-- name: ExpirePendingInvitations :exec
UPDATE invitations
SET status = 'expired', updated_at = now()
WHERE status = 'pending' AND expires_at < now();
