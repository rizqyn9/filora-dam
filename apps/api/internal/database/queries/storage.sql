-- name: CreateStorageProvider :one
INSERT INTO storage_providers (layer, name, type, credentials, quota, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetStorageProvider :one
SELECT * FROM storage_providers WHERE id = $1;

-- name: ListStorageProviders :many
SELECT * FROM storage_providers ORDER BY layer, name;

-- name: ListActiveProvidersByLayer :many
SELECT * FROM storage_providers
WHERE layer = $1 AND is_active = TRUE
ORDER BY id;

-- name: UpdateStorageProvider :one
UPDATE storage_providers
SET name = $2, credentials = $3, quota = $4, is_active = $5
WHERE id = $1
RETURNING *;

-- name: DeactivateStorageProvider :exec
UPDATE storage_providers SET is_active = FALSE WHERE id = $1;

-- name: AddStorageProviderUsed :exec
UPDATE storage_providers SET used = used + $2 WHERE id = $1;

-- name: ListStorageAccountUsage :many
SELECT id, name, layer, type, is_active, quota, used,
       COALESCE(used_percent, 0)::float8 AS used_percent,
       location_count, stored_count, pending_count, failed_count
FROM storage_account_usage
ORDER BY layer, name;

-- name: CreateStorageLocation :one
INSERT INTO storage_locations (asset_id, provider_id, layer, provider_key, url, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetServingLocationURL :one
SELECT url FROM storage_locations
WHERE asset_id = $1 AND layer = 'serving' AND status = 'stored' AND url IS NOT NULL
LIMIT 1;

-- name: EnqueueArchiveJob :exec
INSERT INTO archive_sync_jobs (asset_id, target_layer)
VALUES ($1, 'archive')
ON CONFLICT (asset_id, target_layer) WHERE status IN ('pending', 'running') DO NOTHING;
