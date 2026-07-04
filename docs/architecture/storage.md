# Storage Architecture

How Filora turns "many free-tier cloud accounts" into one transparent,
safe storage pool. For the data model, see
[database/schema.md → Storage](../database/schema.md#storage).

---

## Two layers

Every asset is stored in **both** layers:

| Layer | Purpose | Providers | Characteristics |
|-------|---------|-----------|-----------------|
| **serving** | Fast, viewable copy the app serves | Cloudinary, ImageKit | Hot, public URLs, free-tier limits per account |
| **archive** | Cheap, durable backup copy | GCS Archive, Cloudflare R2 | Cold, cheap, not directly served |

Rationale: serving providers are great for delivery but limited and not a safe
backup; archive-class storage is cheap and durable but slow/awkward to serve.
Using both gives fast delivery **and** a safety copy.

## Accounts are global and pooled

- Storage accounts (`storage_providers`) are **global**, managed by
  admins/superuser — not owned by end users.
- **Many accounts per layer** are supported. When one free-tier account fills up,
  add another; together they form one logical pool per layer.
- A convenience view, `storage_account_usage`, summarizes usage and copy counts
  per account for the management UI.

## The adapter pattern

Business logic never imports a provider SDK. Each provider implements a common
interface:

```go
type StorageAdapter interface {
    Upload(ctx context.Context, input UploadInput) (*UploadResult, error)
    Download(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

Implementations: `CloudinaryAdapter`, `ImageKitAdapter`, `R2Adapter`, and
(planned) `GCSAdapter`. The storage service selects an adapter based on the
account's `type` and never leaks provider-specific types upward.

## Upload flow

```
User uploads a file
  → API handler (validate, auth)
  → Storage service
      → compute SHA-256; per-gallery dedup check
      → check gallery quota
      → pick a serving account       (account election = backlog)
      → StorageAdapter.Upload(...)     → serving provider
      → persist: assets row + storage_location (layer=serving, status=stored)
      → enqueue archive_sync_job(asset, layer=archive)
  ← return asset (available immediately from the serving layer)
```

The user's upload **does not wait** for archiving.

## Archive replication flow (async)

```
Background worker
  → claim due archive_sync_jobs (status pending/failed, next_retry_at ≤ now)
  → mark running
  → pick an archive account          (account election = backlog)
  → StorageAdapter.Upload(...)         → archive provider
  → create/update storage_location (layer=archive, status=stored)
  → mark job completed
On error: increment attempts, set next_retry_at (backoff), status failed
          until max_attempts is reached
```

Per-copy state lives on `storage_locations.status` (`pending`→`stored`/`failed`);
retry bookkeeping lives on `archive_sync_jobs`.

## Download flow

```
User requests an asset
  → API handler (auth + permission check)
  → Storage service: find a serving storage_location
  → StorageAdapter: signed URL or streamed bytes
  ← return to client
```

Archive copies are for durability/recovery, not day-to-day serving.

## Deletion

- Deleting an asset is a **soft delete** (moves to trash via `deleted_at`); bytes
  are retained so it can be recovered.
- Purging trash (future) will delete the underlying storage locations across both
  layers via adapters.
- A storage account that still hosts files **cannot be deleted**
  (`storage_locations.provider_id` uses `ON DELETE RESTRICT`); deactivate it with
  `is_active = false` instead.

## Quota

Quota is tracked **per gallery** (`galleries.storage_quota` / `storage_used`),
independent of any single account's free-tier cap — that's the whole point of
pooling multiple accounts.

## Backlog

- **Account election** — the strategy that chooses which account within a layer
  receives a new upload (round-robin, most-free-space, weighted). Until then, a
  simple "first active account with space" rule is enough.
- **GCS Archive adapter** and **Alibaba OSS Archive** as an alternative.
- **Trash purge** and disaster-recovery restore-from-archive tooling.

## See also

- [Database: storage tables](../database/schema.md#storage)
- [Architecture overview](./overview.md)
- [Product: storage concepts](../product/concepts.md#storage)
