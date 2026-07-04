# Filora CLI

Terminal client for the [Filora API](../api/README.md) — a thin HTTP client
(Go + Cobra). Supports terminal login with multiple concurrent, revocable
sessions.

## Install / run

```bash
cd apps/cli
go build -o bin/filora ./cmd/filora
./bin/filora --help
# or: go run ./cmd/filora --help
```

## Configuration

Credentials are stored in `~/.filora/config.json` (created by `login`).
Environment overrides:

- `FILORA_API_URL` — API base URL (default `http://localhost:3000`)
- `FILORA_TOKEN` — bearer token (overrides the stored token)

## Authentication

The CLI authenticates with an opaque, revocable Filora **CLI token**. To obtain
one you bootstrap with an existing token — either a **Clerk web session token**
or another CLI token — which is exchanged for a dedicated CLI token:

```bash
filora login --api-url http://localhost:3000 --token <clerk-or-cli-token> [--label my-laptop]
```

This calls `POST /api/v1/cli/sessions`, stores the returned token locally, and
records the session id (used by `logout`).

## Commands

```bash
filora login --token <token>        # authenticate, store a CLI token
filora logout                       # revoke this session + clear local creds
filora whoami                       # show the current user

filora sessions list                # list active CLI sessions (* = current)
filora sessions revoke <id>         # revoke a session

filora galleries list               # galleries you belong to

filora upload <file> --gallery <id> # upload a file
filora assets list --gallery <id> [--limit --offset]
filora download <asset-id> [-o out] # download an asset
```

## Notes

- `download` follows the API's redirect to the storage URL and fetches the object
  **without** sending your API token to the storage provider.
- All business logic lives in the API; the CLI only handles I/O and presentation.
- Uploads require the API to have an active storage account with a working
  adapter (currently the `r2` provider type).
