-- name: GetFolderByID :one
SELECT * FROM folders WHERE id = $1 AND deleted_at IS NULL;

-- name: ListFoldersByParent :many
SELECT * FROM folders
WHERE space_id = $1 AND parent_id = $2 AND deleted_at IS NULL
ORDER BY name;

-- name: ListRootFolders :many
SELECT * FROM folders
WHERE space_id = $1 AND parent_id IS NULL AND deleted_at IS NULL
ORDER BY name;

-- name: CreateFolder :one
INSERT INTO folders (space_id, parent_id, name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: RenameFolder :exec
UPDATE folders SET name = $2, updated_at = now() WHERE id = $1;

-- name: MoveFolder :exec
UPDATE folders SET parent_id = $2, updated_at = now() WHERE id = $1;

-- name: SoftDeleteFolder :exec
UPDATE folders SET deleted_at = now(), updated_at = now() WHERE id = $1;

-- name: RestoreFolder :exec
UPDATE folders SET deleted_at = NULL, updated_at = now() WHERE id = $1;

-- name: GetFolderAncestors :many
WITH RECURSIVE ancestors AS (
    SELECT f.id, f.space_id, f.parent_id, f.name, 0 AS depth
    FROM folders f WHERE f.id = $1
    UNION ALL
    SELECT f.id, f.space_id, f.parent_id, f.name, a.depth + 1
    FROM folders f
    JOIN ancestors a ON f.id = a.parent_id
)
SELECT ancestors.id, ancestors.name, ancestors.depth FROM ancestors ORDER BY ancestors.depth DESC;
