# Filora API

Go backend for Filora DAM. All business logic lives here.

---

## Prerequisites

- Go 1.23+
- PostgreSQL (Neon or local)
- [sqlc](https://sqlc.dev) (for query generation)
- [golang-migrate](https://github.com/golang-migrate/migrate) (for migrations)

## Setup

```bash
cp .env.example .env
# Edit .env with your DATABASE_URL and other values
```

## Run migrations

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

## Run the server

```bash
go run cmd/server/main.go
```

## Run the archive worker

```bash
go run cmd/worker/main.go
```

## Regenerate sqlc

```bash
sqlc generate
```

## Lint

```bash
golangci-lint run ./...
```

## Build

```bash
go build -o bin/server cmd/server/main.go
go build -o bin/worker cmd/worker/main.go
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `PORT` | No | HTTP port (default: 8080) |
| `ENVIRONMENT` | No | `development` or `production` |
| `ENCRYPTION_KEY` | No | 32-byte hex for AES-256-GCM credential encryption |

## API Endpoints

All protected endpoints require `Authorization: Bearer <token>` header.

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Health check |
| POST | `/auth/login` | No | Login (email + password) |
| POST | `/auth/register` | No | Register with invite token |
| POST | `/api/v1/auth/logout` | Yes | Revoke current session |
| GET | `/api/v1/auth/me` | Yes | Get current user |
| POST | `/api/v1/auth/change-password` | Yes | Change password |
| GET/POST | `/api/v1/spaces` | Yes | List/create spaces |
| GET/PUT/DELETE | `/api/v1/spaces/:id` | Yes | Get/update/delete space |
| GET/POST | `/api/v1/folders` | Yes | List/create folders |
| GET/DELETE | `/api/v1/folders/:id` | Yes | Get/delete folder |
| PATCH | `/api/v1/folders/:id/rename` | Yes | Rename folder |
| PATCH | `/api/v1/folders/:id/move` | Yes | Move folder |
| POST | `/api/v1/folders/:id/restore` | Yes | Restore trashed folder |
| GET | `/api/v1/folders/:id/breadcrumbs` | Yes | Get folder ancestors |
| GET/POST/DELETE | `/api/v1/tags` | Yes | List/create/delete tags |
| POST/DELETE | `/api/v1/tags/asset` | Yes | Tag/untag an asset |
| GET | `/api/v1/tags/asset/:asset_id` | Yes | List tags for asset |
| POST/GET/DELETE | `/api/v1/sessions` | Yes | CLI session management |
| GET/POST | `/api/v1/storage/accounts` | Yes | List/create storage accounts |
| GET/PUT | `/api/v1/storage/accounts/:id` | Yes | Get/update account |
| POST | `/api/v1/storage/accounts/:id/deactivate` | Yes | Deactivate account |
| GET | `/api/v1/assets` | Yes | List assets in space/folder |
| POST | `/api/v1/assets/upload` | Yes | Upload file (multipart) |
| GET | `/api/v1/assets/:id` | Yes | Get asset metadata |
| PATCH | `/api/v1/assets/:id/rename` | Yes | Rename asset |
| POST | `/api/v1/assets/references` | Yes | Create asset reference |
| DELETE | `/api/v1/assets/references/:ref_id` | Yes | Remove asset reference |
