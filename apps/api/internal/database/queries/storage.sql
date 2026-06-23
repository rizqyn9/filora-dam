-- name: GetProviderByID :one
SELECT * FROM storage_providers
WHERE id = $1 LIMIT 1;

-- name: ListActiveProvidersByUser :many
SELECT * FROM storage_providers
WHERE user_id = $1 AND is_active = true
ORDER BY created_at;

-- name: ListAllProvidersByUser :many
SELECT * FROM storage_providers
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreateProvider :one
INSERT INTO storage_providers (
    user_id,
    name,
    type,
    credentials,
    quota
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: UpdateProviderUsage :one
UPDATE storage_providers
SET used = $2
WHERE id = $1
RETURNING *;

-- name: DeactivateProvider :one
UPDATE storage_providers
SET is_active = false
WHERE id = $1
RETURNING *;

-- name: CreateStorageLocation :one
INSERT INTO storage_locations (
    asset_id,
    provider_id,
    provider_key,
    url,
    metadata
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetLocationsByAssetID :many
SELECT * FROM storage_locations
WHERE asset_id = $1;

-- name: DeleteLocation :exec
DELETE FROM storage_locations
WHERE id = $1;
