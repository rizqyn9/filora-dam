---
inclusion: fileMatch
fileMatchPattern: ["api/**/*.go"]
---

# Observability Standards

Rules for instrumenting Go code in Filora.

## Three Signals

| Signal | When | How |
|--------|------|-----|
| **Traces** | Every service-layer operation that does I/O or takes >1ms | `otel.Tracer("filora-api").Start(ctx, "module.operation")` |
| **Logs** | Business events, errors, state changes | `slog.InfoContext(ctx, ...)` — ctx links log to active span |
| **Metrics** | Counters/histograms for dashboards and alerting | `otel.Meter("filora-api").Int64Counter(...)` |

## Span Naming

Format: `module.operation` (lowercase, dot-separated).

```
asset.upload
asset.dedup_check
storage.elect_account
storage.upload_to_provider
auth.login
auth.verify_token
folder.move
archive.sync_asset
```

## Span Attributes

Add attributes that help filter/debug. Use semantic conventions where possible.

```go
span.SetAttributes(
    attribute.String("asset.id", id.String()),
    attribute.Int64("asset.size_bytes", size),
    attribute.String("storage.provider", "r2"),
    attribute.String("storage.account_id", "3"),
)
```

## Error Recording

Always record errors on spans AND log them:

```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    slog.ErrorContext(ctx, "operation failed", "error", err, "asset_id", id)
    return err
}
```

## Logging Rules

- Use `slog` (not zerolog) for application logs — these go to Axiom via OTel.
- Use `zerolog` only for server startup/shutdown (console-only).
- Always pass `ctx` as first arg: `slog.InfoContext(ctx, ...)` — this correlates log with the active trace.
- Log at appropriate levels:
  - `Info`: successful operations, state changes (uploaded, deleted, archived)
  - `Warn`: degraded state (archive job failed but asset usable, account near quota)
  - `Error`: failures that need investigation (upload failed, DB error, adapter error)

## What to Instrument

### Must instrument (service layer):

- `auth.login` / `auth.register` — with `user.email` attribute
- `asset.upload` — with `asset.size_bytes`, `asset.mime_type`, `storage.account_id`
- `asset.dedup_check` — with `asset.checksum`, `dedup.hit` (bool)
- `storage.elect_account` — with `storage.layer`, `storage.account_id`
- `storage.upload_to_provider` — with `storage.provider`, duration
- `archive.sync_asset` — with `asset.id`, `archive.account_id`, success/failure
- `space.check_quota` — with `space.id`, `space.used_bytes`, `space.quota_bytes`

### Don't instrument:

- Repository/DB layer (pgx has its own OTel instrumentation if needed later)
- Handlers (already covered by Fiber middleware span)
- Pure in-memory operations (cache lookups, hash computation)

## Metrics to Track

| Metric | Type | Labels |
|--------|------|--------|
| `filora.uploads.total` | Counter | provider, mime_type, dedup_hit |
| `filora.uploads.bytes` | Counter | provider |
| `filora.uploads.duration_ms` | Histogram | provider |
| `filora.archive.jobs.total` | Counter | status (completed/failed) |
| `filora.auth.logins.total` | Counter | client (web/cli), success |
| `filora.storage.used_bytes` | Gauge | account_id, provider |

## Pattern for Service Methods

```go
func (s *Service) Upload(ctx context.Context, input UploadInput) (*db.Asset, error) {
    ctx, span := otel.Tracer("filora-api").Start(ctx, "asset.upload")
    defer span.End()

    span.SetAttributes(
        attribute.String("asset.filename", input.Filename),
        attribute.Int64("asset.size_bytes", input.Size),
    )

    // ... business logic ...

    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "upload failed")
        slog.ErrorContext(ctx, "upload failed", "error", err)
        return nil, err
    }

    slog.InfoContext(ctx, "asset uploaded",
        "asset_id", asset.ID,
        "size_bytes", asset.SizeBytes,
        "dedup_hit", false,
    )
    return asset, nil
}
```
