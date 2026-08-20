-- name: GetAssetByID :one
SELECT * FROM assets WHERE id = $1;

-- name: GetAssetByChecksum :one
SELECT * FROM assets WHERE checksum_sha256 = $1;

-- name: CreateAsset :one
INSERT INTO assets (original_filename, name, mime_type, size_bytes, checksum_sha256, width, height, uploaded_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateAssetName :exec
UPDATE assets SET name = $2, updated_at = now() WHERE id = $1;

-- name: CreateAssetReference :one
INSERT INTO asset_references (asset_id, space_id, folder_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: SoftDeleteAssetReference :exec
UPDATE asset_references
SET deleted_at = now()
WHERE id = $1;

-- name: RestoreAssetReference :exec
UPDATE asset_references
SET deleted_at = NULL
WHERE id = $1;

-- name: ListAssetsByFolder :many
SELECT a.* FROM assets a
JOIN asset_references ar ON ar.asset_id = a.id
WHERE ar.space_id = $1 AND ar.folder_id = $2 AND ar.deleted_at IS NULL
ORDER BY a.created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListAssetsBySpaceRoot :many
SELECT a.* FROM assets a
JOIN asset_references ar ON ar.asset_id = a.id
WHERE ar.space_id = $1 AND ar.folder_id IS NULL AND ar.deleted_at IS NULL
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAssetsBySpace :many
SELECT a.* FROM assets a
JOIN asset_references ar ON ar.asset_id = a.id
WHERE ar.space_id = $1 AND ar.deleted_at IS NULL
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountActiveReferences :one
SELECT count(*) FROM asset_references
WHERE asset_id = $1 AND deleted_at IS NULL;

-- name: ListTrashedReferences :many
SELECT ar.*, a.name, a.mime_type, a.size_bytes
FROM asset_references ar
JOIN assets a ON a.id = ar.asset_id
WHERE ar.space_id = $1 AND ar.deleted_at IS NOT NULL
ORDER BY ar.deleted_at DESC;

-- name: ListOrphanedAssets :many
SELECT a.id FROM assets a
WHERE NOT EXISTS (
    SELECT 1 FROM asset_references ar
    WHERE ar.asset_id = a.id AND ar.deleted_at IS NULL
)
AND a.created_at < $1;
