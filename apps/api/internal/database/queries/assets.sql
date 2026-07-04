-- name: CreateAsset :one
INSERT INTO assets (gallery_id, uploaded_by, name, type, mime_type, size, hash, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetAssetByID :one
SELECT * FROM assets WHERE id = $1;

-- name: GetActiveAssetByGalleryHash :one
SELECT * FROM assets
WHERE gallery_id = $1 AND hash = $2 AND deleted_at IS NULL;

-- name: ListActiveAssetsByGallery :many
SELECT * FROM assets
WHERE gallery_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountActiveAssetsByGallery :one
SELECT count(*) FROM assets
WHERE gallery_id = $1 AND deleted_at IS NULL;

-- name: SearchAssetsByName :many
SELECT * FROM assets
WHERE gallery_id = $1 AND deleted_at IS NULL AND name ILIKE $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: FilterAssetsByType :many
SELECT * FROM assets
WHERE gallery_id = $1 AND deleted_at IS NULL AND type = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListTrashedAssetsByGallery :many
SELECT * FROM assets
WHERE gallery_id = $1 AND deleted_at IS NOT NULL
ORDER BY deleted_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateAssetName :one
UPDATE assets SET name = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteAsset :execrows
UPDATE assets SET deleted_at = now(), deleted_by = $2
WHERE id = $1 AND deleted_at IS NULL;

-- name: RestoreAsset :execrows
UPDATE assets SET deleted_at = NULL, deleted_by = NULL
WHERE id = $1 AND deleted_at IS NOT NULL;

-- name: HardDeleteAsset :exec
DELETE FROM assets WHERE id = $1;
