-- name: ClaimArchiveJob :one
-- Atomically claims the next due archive job (skipping locked rows) and marks
-- it running. Returns no rows when the queue is empty.
UPDATE archive_sync_jobs
SET status = 'running', attempts = attempts + 1, updated_at = now()
WHERE id = (
    SELECT id FROM archive_sync_jobs
    WHERE status = 'pending'
      AND (next_retry_at IS NULL OR next_retry_at <= now())
    ORDER BY created_at
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING *;

-- name: MarkArchiveJobResult :exec
UPDATE archive_sync_jobs
SET status = $2, last_error = $3, next_retry_at = $4, updated_at = now()
WHERE id = $1;

-- name: GetArchiveSource :one
SELECT a.size, a.mime_type, sl.provider_id, sl.provider_key
FROM assets a
JOIN storage_locations sl
  ON sl.asset_id = a.id AND sl.layer = 'serving' AND sl.status = 'stored'
WHERE a.id = $1
LIMIT 1;
