-- name: GetStorageAccountByID :one
SELECT * FROM storage_accounts WHERE id = $1;

-- name: ListActiveAccountsByLayer :many
SELECT * FROM storage_accounts
WHERE layer = $1 AND is_active = true
ORDER BY used_bytes ASC;

-- name: ListAllStorageAccounts :many
SELECT * FROM storage_accounts ORDER BY created_at;

-- name: CreateStorageAccount :one
INSERT INTO storage_accounts (provider, label, layer, credentials_encrypted, quota_bytes)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateStorageAccount :exec
UPDATE storage_accounts
SET label = $2, is_active = $3, quota_bytes = $4, updated_at = now()
WHERE id = $1;

-- name: IncrementAccountUsage :exec
UPDATE storage_accounts
SET used_bytes = used_bytes + $2, updated_at = now()
WHERE id = $1;

-- name: DecrementAccountUsage :exec
UPDATE storage_accounts
SET used_bytes = GREATEST(used_bytes - $2, 0), updated_at = now()
WHERE id = $1;

-- name: DeactivateAccount :exec
UPDATE storage_accounts SET is_active = false, updated_at = now() WHERE id = $1;

-- name: CreateStorageLocation :one
INSERT INTO storage_locations (asset_id, account_id, layer, status, remote_path, remote_url)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateStorageLocationStatus :exec
UPDATE storage_locations
SET status = $2, remote_path = $3, remote_url = $4, error = $5, updated_at = now()
WHERE id = $1;

-- name: GetServingLocation :one
SELECT * FROM storage_locations
WHERE asset_id = $1 AND layer = 'serving' AND status = 'stored'
LIMIT 1;

-- name: GetArchiveLocation :one
SELECT * FROM storage_locations
WHERE asset_id = $1 AND layer = 'archive' AND status = 'stored'
LIMIT 1;

-- name: ListLocationsByAsset :many
SELECT * FROM storage_locations WHERE asset_id = $1 ORDER BY created_at;
