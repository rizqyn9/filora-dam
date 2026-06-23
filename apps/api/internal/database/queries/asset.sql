-- name: GetAssetByID :one
SELECT * FROM assets
WHERE id = $1 LIMIT 1;

-- name: ListAssetsByUser :many
SELECT * FROM assets
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAssetsByUser :one
SELECT COUNT(*) FROM assets
WHERE user_id = $1;

-- name: GetAssetByHash :one
SELECT * FROM assets
WHERE hash = $1 AND user_id = $2
LIMIT 1;

-- name: CreateAsset :one
INSERT INTO assets (
    user_id,
    name,
    type,
    mime_type,
    size,
    hash,
    tags,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateAssetTags :one
UPDATE assets
SET tags = $2
WHERE id = $1
RETURNING *;

-- name: DeleteAsset :exec
DELETE FROM assets
WHERE id = $1;

-- name: SearchAssetsByName :many
SELECT * FROM assets
WHERE user_id = $1 AND name ILIKE $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: FilterAssetsByType :many
SELECT * FROM assets
WHERE user_id = $1 AND type = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetDashboardStats :one
SELECT
    COUNT(*) as total_assets,
    COALESCE(SUM(size), 0) as total_size,
    COUNT(DISTINCT type) as unique_types
FROM assets
WHERE user_id = $1;

-- name: GetAssetsByTypeCount :many
SELECT type, COUNT(*) as count
FROM assets
WHERE user_id = $1
GROUP BY type
ORDER BY count DESC;

-- name: GetRecentAssets :many
SELECT * FROM assets
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;
