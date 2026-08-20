-- name: CreateArchiveSyncJob :one
INSERT INTO archive_sync_jobs (asset_id)
VALUES ($1)
RETURNING *;

-- name: ClaimNextArchiveJob :one
UPDATE archive_sync_jobs
SET status = 'processing', attempts = attempts + 1, updated_at = now()
WHERE id = (
    SELECT id FROM archive_sync_jobs
    WHERE status IN ('pending', 'failed')
      AND attempts < max_attempts
      AND (next_retry_at IS NULL OR next_retry_at <= now())
    ORDER BY created_at
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: CompleteArchiveJob :exec
UPDATE archive_sync_jobs
SET status = 'completed', completed_at = now(), updated_at = now()
WHERE id = $1;

-- name: FailArchiveJob :exec
UPDATE archive_sync_jobs
SET status = 'failed', last_error = $2, next_retry_at = $3, updated_at = now()
WHERE id = $1;

-- name: ListPendingArchiveJobs :many
SELECT * FROM archive_sync_jobs
WHERE status IN ('pending', 'failed') AND attempts < max_attempts
ORDER BY created_at
LIMIT $1;
