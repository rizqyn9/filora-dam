-- name: CreateAsset :one
INSERT INTO assets (space_id, folder_id, uploaded_by, name, type, mime_type, size, hash, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetAssetByID :one
SELECT * FROM assets WHERE id = $1;

-- name: GetActiveAssetBySpaceHash :one
SELECT * FROM assets
WHERE space_id = $1 AND hash = $2 AND deleted_at IS NULL;

-- name: ListActiveAssetsBySpace :many
SELECT * FROM assets
WHERE space_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveAssetsByFolder :many
SELECT * FROM assets
WHERE space_id = $1 AND folder_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListActiveRootAssets :many
SELECT * FROM assets
WHERE space_id = $1 AND folder_id IS NULL AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountActiveAssetsBySpace :one
SELECT count(*) FROM assets
WHERE space_id = $1 AND deleted_at IS NULL;

-- name: CountActiveAssetsByFolder :one
SELECT count(*) FROM assets
WHERE space_id = $1 AND folder_id = $2 AND deleted_at IS NULL;

-- name: CountActiveRootAssets :one
SELECT count(*) FROM assets
WHERE space_id = $1 AND folder_id IS NULL AND deleted_at IS NULL;

-- name: SearchAssetsByName :many
SELECT * FROM assets
WHERE space_id = $1 AND deleted_at IS NULL AND name ILIKE $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: FilterAssetsByType :many
SELECT * FROM assets
WHERE space_id = $1 AND deleted_at IS NULL AND type = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListTrashedAssetsBySpace :many
SELECT * FROM assets
WHERE space_id = $1 AND deleted_at IS NOT NULL
ORDER BY deleted_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateAssetName :one
UPDATE assets SET name = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: MoveAssetToFolder :one
UPDATE assets SET folder_id = $2
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
