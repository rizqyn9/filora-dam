-- name: CreateTag :one
INSERT INTO tags (gallery_id, name, created_by)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTagByID :one
SELECT * FROM tags WHERE id = $1;

-- name: ListTagsByGallery :many
SELECT * FROM tags WHERE gallery_id = $1 ORDER BY name;

-- name: UpdateTag :one
UPDATE tags SET name = $2 WHERE id = $1 RETURNING *;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = $1;

-- name: AttachTag :exec
INSERT INTO asset_tags (asset_id, tag_id)
VALUES ($1, $2)
ON CONFLICT (asset_id, tag_id) DO NOTHING;

-- name: DetachTag :execrows
DELETE FROM asset_tags WHERE asset_id = $1 AND tag_id = $2;

-- name: ListAssetTags :many
SELECT t.*
FROM asset_tags at
JOIN tags t ON t.id = at.tag_id
WHERE at.asset_id = $1
ORDER BY t.name;
