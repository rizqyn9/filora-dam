-- name: GalleryAssetStats :one
SELECT
    count(*)                        AS total_assets,
    COALESCE(sum(size), 0)::bigint  AS total_size,
    count(DISTINCT type)            AS type_count
FROM assets
WHERE gallery_id = $1 AND deleted_at IS NULL;

-- name: GalleryAssetCountsByType :many
SELECT type, count(*) AS count
FROM assets
WHERE gallery_id = $1 AND deleted_at IS NULL
GROUP BY type
ORDER BY count DESC;

-- name: GalleryRecentAssets :many
SELECT id, name, type, mime_type, size, created_at
FROM assets
WHERE gallery_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2;

-- name: ArchiveJobHealth :one
SELECT
    count(*) FILTER (WHERE status = 'pending')   AS pending,
    count(*) FILTER (WHERE status = 'running')   AS running,
    count(*) FILTER (WHERE status = 'completed') AS completed,
    count(*) FILTER (WHERE status = 'failed')    AS failed
FROM archive_sync_jobs;
