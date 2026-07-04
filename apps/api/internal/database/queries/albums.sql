-- name: CreateAlbum :one
INSERT INTO albums (gallery_id, owner_id, name, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAlbumByID :one
SELECT * FROM albums WHERE id = $1;

-- name: ListAlbumsByGallery :many
SELECT * FROM albums WHERE gallery_id = $1 ORDER BY created_at DESC;

-- name: UpdateAlbum :one
UPDATE albums
SET name = $2, description = $3, cover_asset_id = $4
WHERE id = $1
RETURNING *;

-- name: DeleteAlbum :exec
DELETE FROM albums WHERE id = $1;

-- name: UpsertAlbumMember :exec
INSERT INTO album_members (album_id, user_id, role, invited_by)
VALUES ($1, $2, $3, $4)
ON CONFLICT (album_id, user_id) DO UPDATE SET role = EXCLUDED.role;

-- name: GetAlbumMember :one
SELECT * FROM album_members WHERE album_id = $1 AND user_id = $2;

-- name: ListAlbumMembers :many
SELECT am.album_id, am.user_id, am.role, am.created_at,
       u.email, u.name, u.avatar_url
FROM album_members am
JOIN users u ON u.id = am.user_id
WHERE am.album_id = $1
ORDER BY am.created_at;

-- name: RemoveAlbumMember :execrows
DELETE FROM album_members WHERE album_id = $1 AND user_id = $2;

-- name: AddAssetToAlbum :exec
INSERT INTO album_assets (album_id, asset_id, added_by, sort_order)
VALUES ($1, $2, $3, $4)
ON CONFLICT (album_id, asset_id) DO NOTHING;

-- name: RemoveAssetFromAlbum :execrows
DELETE FROM album_assets WHERE album_id = $1 AND asset_id = $2;

-- name: ListAlbumAssetIDs :many
SELECT asset_id FROM album_assets
WHERE album_id = $1
ORDER BY sort_order, created_at;
