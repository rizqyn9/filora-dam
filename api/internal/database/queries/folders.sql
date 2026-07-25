-- name: CreateFolder :one
INSERT INTO folders (space_id, parent_id, owner_id, name, path)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFolderByID :one
SELECT * FROM folders WHERE id = $1;

-- name: ListFoldersByParent :many
SELECT * FROM folders
WHERE space_id = $1 AND parent_id = $2
ORDER BY name;

-- name: ListRootFolders :many
SELECT * FROM folders
WHERE space_id = $1 AND parent_id IS NULL
ORDER BY name;

-- name: UpdateFolderName :one
UPDATE folders SET name = $2 WHERE id = $1 RETURNING *;

-- name: UpdateFolderParent :one
UPDATE folders SET parent_id = $2, path = $3 WHERE id = $1 RETURNING *;

-- name: UpdateFolderPath :exec
UPDATE folders SET path = $2 WHERE id = $1;

-- name: ListFolderSubtree :many
-- Returns all descendants of a folder by matching the materialized path prefix.
-- The path pattern should be passed as '/parentId/%' (e.g. '/5/%' or '/5/12/%').
SELECT * FROM folders
WHERE space_id = $1 AND path LIKE $2
ORDER BY path, name;

-- name: DeleteFolder :exec
DELETE FROM folders WHERE id = $1;

-- name: GetFolderBreadcrumb :many
-- Returns ancestor folders for breadcrumb display. Caller passes comma-separated
-- IDs parsed from the folder's path field.
SELECT id, name, parent_id, path FROM folders
WHERE id = ANY($1::bigint[])
ORDER BY path;
