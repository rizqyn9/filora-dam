# Runbook: Setup Neon PostgreSQL for Filora

Step-by-step guide to configure Neon as the database for Filora.

---

## Prerequisites

- Neon account (free tier: 0.5 GB storage, 1 project)
- `psql` or Neon SQL Editor
- golang-migrate CLI (optional, for migration management)

## 1. Create a Neon Project

1. Go to [Neon Console](https://console.neon.tech)
2. Click **New Project**
3. Project name: `filora`
4. Region: pick closest to you (or closest to your server)
5. PostgreSQL version: **16** (latest stable)
6. Click **Create Project**

## 2. Get Connection String

After project creation, Neon shows the connection string:

```
postgresql://<user>:<password>@<host>.neon.tech/<dbname>?sslmode=require
```

Copy this — it's your `DATABASE_URL`.

Alternative: go to **Dashboard** → **Connection Details** → copy the connection string.

## 3. Apply Schema Migration

Option A — using `psql`:

```bash
psql "$DATABASE_URL" < api/migrations/001_initial.up.sql
```

Option B — using golang-migrate:

```bash
# Install if needed
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path api/migrations -database "$DATABASE_URL" up
```

Option C — using Neon SQL Editor:

1. Go to **SQL Editor** in the Neon console
2. Paste the contents of `api/migrations/001_initial.up.sql`
3. Click **Run**

## 4. Verify Schema

```bash
psql "$DATABASE_URL" -c "\dt"
```

Expected tables (14):

```
 Schema |       Name         | Type  
--------+--------------------+-------
 public | archive_sync_jobs  | table
 public | asset_references   | table
 public | asset_tags         | table
 public | assets             | table
 public | cli_sessions       | table
 public | folders            | table
 public | invitations        | table
 public | space_members      | table
 public | spaces             | table
 public | storage_accounts   | table
 public | storage_locations  | table
 public | tags               | table
 public | user_roles         | table
 public | users              | table
```

## 5. Configure Filora

Add to your `api/.env`:

```bash
DATABASE_URL=postgresql://<user>:<password>@<host>.neon.tech/filora?sslmode=require
```

## 6. Bootstrap Superuser

Run the seed command:

```bash
cd api
go run cmd/seed/main.go
```

Follow the prompts (email, name, password). This creates the first user with
superuser role. See [setup-auth.md](./setup-auth.md) for the full auth flow.

## 7. Create Default Space

After bootstrap, create your personal space:

```sql
INSERT INTO spaces (name, owner_id) VALUES ('Personal', <your-user-id>);
```

Or use the API after auth is configured:

```bash
curl -X POST http://localhost:8080/api/v1/spaces \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Personal", "storage_quota_bytes": 0}'
```

## Neon Free Tier Limits

| Resource | Limit |
|----------|-------|
| Storage | 0.5 GB |
| Compute hours | 191.9 hours/month |
| Branches | 10 |
| Projects | 1 |

For Filora's family-scale usage, free tier is more than sufficient. The 0.5 GB
limit is for the database itself (metadata), not for file storage (that's in R2).

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `failed to ping database` | Wrong connection string or Neon project suspended | Check URL, wake project in console |
| `relation "users" does not exist` | Migration not applied | Run step 3 |
| `SSL required` | Missing `?sslmode=require` | Add it to DATABASE_URL |
| `too many connections` | Connection pool exhausted | Neon free tier has limited connections; ensure pgxpool max is ≤5 |
| Slow first query after inactivity | Neon suspends after 5 min idle (free tier) | First request wakes it (~1-2s cold start) |

## Branching (Optional)

Neon supports git-like database branching — useful for testing:

```bash
# Create a dev branch from main
neonctl branches create --name dev --parent main

# Get the branch connection string
neonctl connection-string --branch dev
```

Each branch is a copy-on-write clone. Free tier allows up to 10 branches.
