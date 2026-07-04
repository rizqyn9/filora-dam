-- name: InsertWebhookEvent :one
-- Returns the new row id only when inserted; ON CONFLICT (duplicate delivery)
-- returns no rows, which the caller treats as "already seen" (idempotent).
INSERT INTO clerk_webhook_events (event_id, event_type, payload)
VALUES ($1, $2, $3)
ON CONFLICT (event_id) DO NOTHING
RETURNING id;

-- name: MarkWebhookProcessed :exec
UPDATE clerk_webhook_events
SET processed_at = now(),
    error = NULL
WHERE id = $1;

-- name: MarkWebhookFailed :exec
UPDATE clerk_webhook_events
SET error = $2
WHERE id = $1;
