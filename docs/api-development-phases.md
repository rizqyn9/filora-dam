# API Development Phases

This document outlines the development phases for `apps/api`.

---

# Phase 0: Project Setup

**Goal**: Initialize Go project with proper structure

## Tasks

- [ ] Initialize Go module (`go mod init`)
- [ ] Create directory structure
- [ ] Set up `.gitignore`
- [ ] Create `.env.example`
- [ ] Set up `Makefile` for common tasks
- [ ] Install core dependencies:
  - Fiber v3
  - pgx (PostgreSQL driver)
  - sqlc
  - golang-migrate
  - validator v10
  - godotenv

## Deliverables

- Working Go project structure
- Dependency management configured
- Development tooling ready

---

# Phase 1: Core Infrastructure

**Goal**: Set up database, config, and basic HTTP server

## Tasks

### 1.1 Configuration

- [ ] Create `internal/config/config.go`
- [ ] Load environment variables
- [ ] Validate configuration with validator
- [ ] Support for:
  - Server port
  - Database URL
  - Storage provider credentials

### 1.2 Database Setup

- [ ] Create `internal/database/db.go`
- [ ] Set up PostgreSQL connection pool
- [ ] Create migrations directory
- [ ] Set up golang-migrate
- [ ] Create initial schema migration
- [ ] Set up sqlc configuration

### 1.3 HTTP Server

- [ ] Create `cmd/server/main.go`
- [ ] Initialize Fiber app
- [ ] Set up middleware:
  - Logger
  - Recovery
  - CORS
- [ ] Health check endpoint (`GET /health`)
- [ ] Root endpoint (`GET /`)

### 1.4 Response Utilities

- [ ] Create `internal/lib/response.go`
- [ ] Implement success response helper
- [ ] Implement error response helper
- [ ] Define standard response structures

## Deliverables

- Running HTTP server
- Database connection working
- Health check responding
- Configuration validated
- Standard response format

---

# Phase 2: Account Module (MVP)

**Goal**: Basic user account management

## Tasks

### 2.1 Database Schema

- [ ] Create migration: `users` table
  - id (UUID, primary key)
  - email (unique)
  - name
  - password_hash
  - storage_quota (default 5GB)
  - storage_used (default 0)
  - created_at
  - updated_at

### 2.2 sqlc Queries

- [ ] Define queries in `internal/database/queries/account.sql`:
  - FindUserByID
  - FindUserByEmail
  - CreateUser
  - UpdateStorageUsed

### 2.3 Repository

- [ ] Create `internal/modules/account/repository.go`
- [ ] Implement repository using sqlc generated code
- [ ] Add context support

### 2.4 Models

- [ ] Create `internal/modules/account/models.go`
- [ ] Define request/response structs
- [ ] Add validation tags

### 2.5 Service

- [ ] Create `internal/modules/account/service.go`
- [ ] Implement business logic:
  - Get user by ID
  - Check storage quota
  - Update storage usage

### 2.6 Handler

- [ ] Create `internal/modules/account/handler.go`
- [ ] Implement routes:
  - `GET /api/v1/account/:id` - Get user info
  - `GET /api/v1/account/:id/quota` - Get quota info

## Deliverables

- Working account endpoints
- User management foundation
- Quota tracking ready

---

# Phase 3: Storage Module Foundation

**Goal**: Storage abstraction layer and provider management

## Tasks

### 3.1 Database Schema

- [ ] Create migration: `storage_providers` table
  - id (UUID, primary key)
  - user_id (foreign key to users)
  - name
  - type (cloudinary, imagekit, r2)
  - credentials (jsonb, encrypted)
  - quota (nullable)
  - used (default 0)
  - is_active (boolean)
  - created_at
  - updated_at

### 3.2 Storage Adapter Interface

- [ ] Create `internal/modules/storage/adapters/adapter.go`
- [ ] Define `StorageAdapter` interface
- [ ] Define `UploadInput` struct
- [ ] Define `UploadResult` struct

### 3.3 Provider Adapters (Stubs)

- [ ] Create stub adapters (return not implemented):
  - `cloudinary.go`
  - `imagekit.go`
  - `r2.go`

### 3.4 sqlc Queries

- [ ] Define queries in `internal/database/queries/storage.sql`:
  - FindProviderByID
  - FindActiveProviders
  - CreateProvider
  - UpdateProviderUsage

### 3.5 Repository

- [ ] Create `internal/modules/storage/repository.go`
- [ ] Implement provider management

### 3.6 Models

- [ ] Create `internal/modules/storage/models.go`
- [ ] Define provider models
- [ ] Define upload/download models

### 3.7 Service

- [ ] Create `internal/modules/storage/service.go`
- [ ] Implement:
  - Provider registration
  - Provider selection logic (stub)
  - Adapter initialization

### 3.8 Handler

- [ ] Create `internal/modules/storage/handler.go`
- [ ] Implement routes:
  - `GET /api/v1/storage/providers` - List providers
  - `POST /api/v1/storage/providers` - Add provider
  - `GET /api/v1/storage/providers/:id` - Get provider

## Deliverables

- Storage provider management
- Adapter interface defined
- Foundation for uploads ready

---

# Phase 4: Asset Module

**Goal**: Asset metadata management

## Tasks

### 4.1 Database Schema

- [ ] Create migration: `assets` table
  - id (UUID, primary key)
  - user_id (foreign key to users)
  - name
  - type (image, video, document, archive, file)
  - mime_type
  - size
  - hash (for deduplication)
  - tags (text array)
  - metadata (jsonb)
  - created_at
  - updated_at
- [ ] Add indexes:
  - user_id
  - hash
  - created_at

- [ ] Create migration: `storage_locations` table
  - id (UUID, primary key)
  - asset_id (foreign key to assets)
  - provider_id (foreign key to storage_providers)
  - provider_key (storage key/path)
  - url (public URL)
  - metadata (jsonb)
  - created_at
- [ ] Add indexes:
  - asset_id
  - provider_id

### 4.2 sqlc Queries

- [ ] Define queries in `internal/database/queries/asset.sql`:
  - FindAssetByID
  - FindAssetsByUserID
  - FindAssetByHash
  - CreateAsset
  - UpdateAsset
  - DeleteAsset
  - FindLocationsByAssetID
  - CreateLocation
  - DeleteLocation

### 4.3 Repository

- [ ] Create `internal/modules/asset/repository.go`
- [ ] Implement asset queries
- [ ] Implement location queries

### 4.4 Models

- [ ] Create `internal/modules/asset/models.go`
- [ ] Define asset models
- [ ] Define location models

### 4.5 Service

- [ ] Create `internal/modules/asset/service.go`
- [ ] Implement:
  - List user assets
  - Get asset details
  - Delete asset (with storage cleanup)
  - Tag management

### 4.6 Handler

- [ ] Create `internal/modules/asset/handler.go`
- [ ] Implement routes:
  - `GET /api/v1/assets` - List user assets
  - `GET /api/v1/assets/:id` - Get asset
  - `DELETE /api/v1/assets/:id` - Delete asset
  - `PUT /api/v1/assets/:id/tags` - Update tags

## Deliverables

- Asset metadata management
- Storage location tracking
- Asset listing and details

---

# Phase 5: Upload Implementation

**Goal**: Complete upload workflow

## Tasks

### 5.1 Implement Cloudinary Adapter

- [ ] Install Cloudinary Go SDK
- [ ] Implement `Upload()`
- [ ] Implement `Download()`
- [ ] Implement `Delete()`
- [ ] Implement `Exists()`
- [ ] Handle errors properly

### 5.2 File Processing

- [ ] Create `internal/lib/hash.go` - File hashing (SHA-256)
- [ ] Create `internal/lib/mime.go` - MIME type detection
- [ ] Implement file type detection

### 5.3 Upload Service Logic

- [ ] Implement in `storage/service.go`:
  - Calculate file hash
  - Check for duplicates
  - Select storage provider
  - Upload to provider via adapter
  - Save asset metadata
  - Save storage location
  - Update storage usage
  - Update user quota

### 5.4 Upload Handler

- [ ] Add to `storage/handler.go`:
  - `POST /api/v1/storage/upload` - Upload file
  - Handle multipart form
  - Validate file size
  - Validate file type
  - Return asset info

## Deliverables

- Working file upload
- Cloudinary integration
- Deduplication support
- Quota enforcement

---

# Phase 6: Download Implementation

**Goal**: File download and retrieval

## Tasks

### 6.1 Download Service

- [ ] Implement in `storage/service.go`:
  - Get asset metadata
  - Check permissions
  - Get storage location
  - Generate signed URL (if supported)
  - Or proxy download

### 6.2 Download Handler

- [ ] Add to `storage/handler.go`:
  - `GET /api/v1/storage/download/:id` - Download file
  - Support signed URLs
  - Support direct streaming

## Deliverables

- Working file download
- Signed URL generation
- Permission checking

---

# Phase 7: Additional Storage Providers

**Goal**: Implement ImageKit and R2 adapters

## Tasks

### 7.1 ImageKit Adapter

- [ ] Install ImageKit Go SDK
- [ ] Implement all interface methods
- [ ] Test upload/download/delete

### 7.2 Cloudflare R2 Adapter

- [ ] Install AWS S3 SDK
- [ ] Configure for R2
- [ ] Implement all interface methods
- [ ] Test upload/download/delete

### 7.3 Provider Selection Logic

- [ ] Implement in `storage/service.go`:
  - Round-robin selection
  - Quota-aware selection
  - Health checking (future)

## Deliverables

- Multi-provider support working
- Automatic provider selection
- All three providers functional

---

# Phase 8: Authentication (Basic)

**Goal**: Simple authentication for MVP

## Tasks

### 8.1 Password Handling

- [ ] Install bcrypt package
- [ ] Create `internal/lib/password.go`
- [ ] Implement hash/verify functions

### 8.2 JWT

- [ ] Install JWT package
- [ ] Create `internal/lib/jwt.go`
- [ ] Implement generate/validate functions

### 8.3 Auth Service

- [ ] Add to `account/service.go`:
  - Register user
  - Login user
  - Validate token

### 8.4 Auth Handler

- [ ] Add to `account/handler.go`:
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`

### 8.5 Auth Middleware

- [ ] Create `internal/middleware/auth.go`
- [ ] Validate JWT
- [ ] Extract user context
- [ ] Protect endpoints

## Deliverables

- User registration
- User login
- JWT authentication
- Protected endpoints

---

# Phase 9: Search and Filtering

**Goal**: Asset search capabilities

## Tasks

### 9.1 Search Queries

- [ ] Add to `asset.sql`:
  - SearchAssetsByName
  - FilterAssetsByType
  - FilterAssetsByTags
  - FilterAssetsByDateRange

### 9.2 Search Service

- [ ] Add to `asset/service.go`:
  - Search by name
  - Filter by type
  - Filter by tags
  - Combine filters

### 9.3 Search Handler

- [ ] Add to `asset/handler.go`:
  - `GET /api/v1/assets/search?q=...&type=...&tags=...`

## Deliverables

- Asset search working
- Multiple filter support
- Fast queries with indexes

---

# Phase 10: Dashboard Module

**Goal**: Basic metrics and statistics

## Tasks

### 10.1 Metrics Queries

- [ ] Create `internal/database/queries/dashboard.sql`:
  - GetTotalAssets
  - GetStorageUsage
  - GetAssetsByType
  - GetRecentUploads

### 10.2 Dashboard Service

- [ ] Create `internal/modules/dashboard/service.go`
- [ ] Implement metric calculations

### 10.3 Dashboard Handler

- [ ] Create `internal/modules/dashboard/handler.go`
- [ ] Routes:
  - `GET /api/v1/dashboard/stats`
  - `GET /api/v1/dashboard/recent`

## Deliverables

- Dashboard statistics
- Storage usage metrics
- Recent activity tracking

---

# Phase 11: Testing

**Goal**: Core test coverage

## Tasks

### 11.1 Repository Tests

- [ ] Test account repository
- [ ] Test asset repository
- [ ] Test storage repository

### 11.2 Service Tests

- [ ] Test upload workflow
- [ ] Test download workflow
- [ ] Test delete workflow
- [ ] Test deduplication

### 11.3 Integration Tests

- [ ] Test full upload flow
- [ ] Test authentication flow
- [ ] Test quota enforcement

## Deliverables

- Test coverage > 70%
- Critical paths tested
- Integration tests passing

---

# Phase 12: Documentation & Deployment

**Goal**: Production readiness

## Tasks

### 12.1 API Documentation

- [ ] Install Swagger/OpenAPI
- [ ] Document all endpoints
- [ ] Add request/response examples

### 12.2 README

- [ ] Setup instructions
- [ ] Environment variables
- [ ] Database migrations
- [ ] Running the server

### 12.3 Docker

- [ ] Create `Dockerfile`
- [ ] Create `docker-compose.yml` (with PostgreSQL)
- [ ] Test containerized deployment

### 12.4 Deployment

- [ ] Deploy to production (Fly.io / Railway / VPS)
- [ ] Set up environment variables
- [ ] Run migrations
- [ ] Health checks

## Deliverables

- Complete API documentation
- Deployment ready
- Production running

---

# Development Guidelines

## Order of Implementation

For each feature:

1. **Database First** - Create migration
2. **Queries** - Write SQL in sqlc
3. **Repository** - Use generated code
4. **Service** - Business logic
5. **Handler** - HTTP routes
6. **Test** - Write tests

## Testing as You Go

After each phase:

- Test manually with curl/Postman
- Write automated tests
- Verify error handling
- Check edge cases

## Commit Strategy

Commit after each completed task:

- Keep commits small and focused
- Write clear commit messages
- Reference phase number in commits

Example: `feat(phase-2): add user repository with sqlc queries`

---

# Success Criteria

## Phase Completion

A phase is complete when:

- [ ] All tasks checked off
- [ ] Code compiles without errors
- [ ] Tests pass
- [ ] Endpoints respond correctly
- [ ] Errors handled properly
- [ ] Code reviewed (if team)

## MVP Complete

API MVP is complete when:

- All 12 phases done
- Upload/download working
- Multiple providers functional
- Authentication working
- Basic search implemented
- Tests passing
- Documentation complete
- Deployed to production

---

# Notes

## Technology Choices Rationale

- **Fiber v3**: Fast, Express-like API, good for REST
- **sqlc**: Type-safe SQL, better control than ORM
- **golang-migrate**: Standard migration tool
- **validator**: Struct validation with tags
- **pgx**: Modern, fast PostgreSQL driver

## Future Phases (Post-MVP)

- Phase 13: Folders/Organization
- Phase 14: Asset Versions
- Phase 15: Backup Module
- Phase 16: Webhooks
- Phase 17: Rate Limiting
- Phase 18: Advanced Search (Full-text)
- Phase 19: Observability (OpenTelemetry)
- Phase 20: Performance Optimization

---

# Quick Reference

## Common Commands

```bash
# Run server
make run

# Run migrations
make migrate-up

# Generate sqlc
make sqlc

# Run tests
make test

# Format code
make fmt

# Lint
make lint
```

## File Structure Reference

```
apps/api/
├── cmd/server/main.go              # Entry point
├── internal/
│   ├── config/config.go            # Configuration
│   ├── database/
│   │   ├── db.go                   # DB connection
│   │   ├── migrations/             # SQL migrations
│   │   └── queries/                # sqlc queries
│   ├── middleware/
│   │   └── auth.go                 # Auth middleware
│   ├── lib/
│   │   ├── response.go             # Response helpers
│   │   ├── hash.go                 # File hashing
│   │   ├── mime.go                 # MIME detection
│   │   ├── password.go             # Password helpers
│   │   └── jwt.go                  # JWT helpers
│   └── modules/
│       ├── account/                # User accounts
│       ├── asset/                  # Asset metadata
│       ├── storage/                # Storage & adapters
│       └── dashboard/              # Metrics
├── .env.example
├── Makefile
├── go.mod
└── go.sum
```
