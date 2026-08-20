-- name: GetTagByID :one
SELECT * FROM tags WHERE id = $1;

-- name: GetTagByName :one
SELECT * FROM tags WHERE space_id = $1 AND name = $2;

-- name: ListTagsBySpace :many
SELECT * FROM tags WHERE space_id = $1 ORDER BY name;

-- name: CreateTag :one
INSERT INTO tags (space_id, name)
VALUES ($1, $2)
RETURNING *;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = $1;

-- name: AddAssetTag :exec
INSERT INTO asset_tags (asset_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveAssetTag :exec
DELETE FROM asset_tags WHERE asset_id = $1 AND tag_id = $2;

-- name: ListTagsByAsset :many
SELECT t.* FROM tags t
JOIN asset_tags at ON at.tag_id = t.id
WHERE at.asset_id = $1
ORDER BY t.name;

-- name: ListAssetsByTag :many
SELECT a.* FROM assets a
JOIN asset_tags at ON at.asset_id = a.id
WHERE at.tag_id = $1
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;
