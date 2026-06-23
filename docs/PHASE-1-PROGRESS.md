# Phase 1: Core Infrastructure - IN PROGRESS

**Date Started**: 2026-06-24

## Summary

Implementing core infrastructure with database migrations, sqlc code generation, and account module.

## Completed Tasks

### ✅ Database Setup

- [x] Create initial migration `000001_init_schema`
  - Users table with storage quota tracking
  - Storage providers table with credentials
  - Assets table with metadata and deduplication (hash)
  - Storage locations table (maps assets to providers)
  - Indexes for performance
  - Triggers for `updated_at` auto-update
- [x] Install golang-migrate tool
- [x] sqlc configuration ready

### ✅ sqlc Queries

- [x] Account queries (`internal/database/queries/account.sql`)
  - GetUserByID
  - GetUserByEmail
  - CreateUser
  - UpdateUserStorageUsed
  - GetUserQuota
- [x] Storage queries (`internal/database/queries/storage.sql`)
  - Provider management
  - Storage location tracking
- [x] Asset queries (`internal/database/queries/asset.sql`)
  - CRUD operations
  - Search and filtering
- [x] Generate sqlc code successfully

### ✅ Account Module (MVP)

- [x] Models (`internal/modules/account/models.go`)
  - User struct
  - QuotaInfo struct
  - Request/Response structs
- [x] Repository (`internal/modules/account/repository.go`)
  - GetByID, GetByEmail
  - Create user
  - UpdateStorageUsed
  - GetQuota
  - Proper pgtype.UUID handling
- [x] Service (`internal/modules/account/service.go`)
  - Business logic
  - Error handling
  - Quota checking
- [x] Handler (`internal/modules/account/handler.go`)
  - GET /api/v1/account/:id
  - GET /api/v1/account/:id/quota
- [x] Integrated into main.go

## Database Schema Created

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    password_hash VARCHAR(255),
    storage_quota BIGINT DEFAULT 5GB,
    storage_used BIGINT DEFAULT 0,
    created_at, updated_at TIMESTAMP
)

-- Storage Providers
CREATE TABLE storage_providers (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users,
    name VARCHAR(255),
    type VARCHAR(50), -- cloudinary, imagekit, r2
    credentials JSONB,
    quota BIGINT,
    used BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at, updated_at TIMESTAMP
)

-- Assets
CREATE TABLE assets (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users,
    name VARCHAR(500),
    type VARCHAR(50),
    mime_type VARCHAR(100),
    size BIGINT,
    hash VARCHAR(64), -- SHA-256 for deduplication
    tags TEXT[],
    metadata JSONB,
    created_at, updated_at TIMESTAMP
)

-- Storage Locations
CREATE TABLE storage_locations (
    id UUID PRIMARY KEY,
    asset_id UUID REFERENCES assets,
    provider_id UUID REFERENCES storage_providers,
    provider_key VARCHAR(500),
    url TEXT,
    metadata JSONB,
    created_at TIMESTAMP
)
```

## API Endpoints Available

### Health & Info
- `GET /` - API info
- `GET /health` - Health check

### Account Module
- `GET /api/v1/account/:id` - Get user info
  ```json
  {
    "success": true,
    "data": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "User Name",
      "storage_quota": 5368709120,
      "storage_used": 0,
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  }
  ```

- `GET /api/v1/account/:id/quota` - Get quota info
  ```json
  {
    "success": true,
    "data": {
      "quota": 5368709120,
      "used": 0,
      "free": 5368709120
    }
  }
  ```

## Build Status

✅ **Build successful**
✅ **sqlc code generation working**
✅ **Type conversions fixed** (pgtype.UUID ↔ google/uuid)

## Testing

To test (requires PostgreSQL database):

```bash
cd apps/api

# Set DATABASE_URL in .env
# Run migrations
make migrate-up

# Start server
make run

# Test endpoints
curl http://localhost:3000/health
curl http://localhost:3000/api/v1/account/{uuid}
```

## Remaining Tasks - Phase 1

### Configuration
- [x] Environment variables loaded
- [x] Validation working

### Database Setup
- [x] Migrations created
- [ ] Run migrations (needs real database)
- [x] sqlc configured and generated

### HTTP Server
- [x] Fiber initialized
- [x] Middleware configured
- [x] Health check working
- [ ] Enhanced health check with DB ping

### Response Utilities
- [x] Success/Error helpers
- [x] Standard format

## Technical Notes

### Type Conversions
sqlc generates `pgtype.UUID` while we use `google/uuid` for parsing. Conversion pattern:

```go
// String → pgtype.UUID
parsedUUID, _ := uuid.Parse(id)
var pgUUID pgtype.UUID
pgUUID.Scan(parsedUUID)

// pgtype.UUID → String
uuidString := uuid.UUID(pgUUID.Bytes).String()

// pgtype.Timestamp → time.Time
timeValue := timestamp.Time
```

### Migration Files
- `000001_init_schema.up.sql` - Create all tables
- `000001_init_schema.down.sql` - Drop all tables

### sqlc Configuration
- Queries in `internal/database/queries/*.sql`
- Generated code in `internal/database/db/`
- Using pgx/v5 driver

## Next Steps

1. **Get PostgreSQL database** (Neon recommended)
2. **Run migrations**: `make migrate-up`
3. **Test endpoints** with real database
4. **Add health check enhancement** (DB ping)
5. **Continue to Phase 2**: Complete account module with auth

## Blockers

⚠️  **Need PostgreSQL database URL** to:
- Run migrations
- Test database connection
- Test account endpoints

## Files Created/Modified

### Created
- `internal/database/migrations/000001_init_schema.{up,down}.sql`
- `internal/database/queries/{account,storage,asset}.sql`
- `internal/database/db/*.go` (generated by sqlc)
- `internal/modules/account/{models,repository,service,handler}.go`

### Modified
- `cmd/server/main.go` - Wire up account module

---

**Phase 1 Status**: 🔄 **IN PROGRESS** (80% complete)

**Completed**: Database schema, sqlc, account module MVP  
**Remaining**: Real database setup, migrations, testing  
**Ready for**: Testing with real PostgreSQL database
