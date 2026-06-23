# Architecture

## High Level

```
User
↓
Web (React 19) / CLI (Go)
↓
API (Go + Fiber)
↓
Services
↓
Repositories
↓
PostgreSQL
```

```
Services
↓
Storage Orchestrator
↓
Storage Adapters
↓
Cloudinary / ImageKit / R2
```

```
Background Workers (Future)
↓
Backup Storage
```

---

# Application Structure

Filora consists of three independent applications:

## apps/api

Go REST API server.

Contains all business logic.

Technology:
* Go 1.23+
* Fiber v3
* sqlc
* PostgreSQL

## apps/cli

Go CLI client.

Thin HTTP client that consumes API.

Technology:
* Go 1.23+
* Cobra (CLI framework)
* HTTP client

## apps/web

React frontend.

Thin client that consumes API.

Technology:
* React 19
* TypeScript
* TanStack Query
* TanStack Router
* Tailwind CSS v4
* Shadcn UI

---

# Core Domains

## Asset

Responsible for:

* asset metadata
* ownership
* organization
* tagging
* search
* versioning

## Storage

Responsible for:

* uploads
* downloads
* provider integrations
* storage account selection
* storage orchestration

## Account

Responsible for:

* users
* authentication
* quotas
* permissions

## Backup (Future)

Responsible for:

* backup scheduling
* backup uploads
* disaster recovery

## Dashboard

Responsible for:

* metrics
* reporting
* storage visibility

---

# API Module Structure

```
apps/api/internal/modules/

asset/
  handler.go
  service.go
  repository.go
  models.go

storage/
  handler.go
  service.go
  repository.go
  models.go
  adapters/
    adapter.go
    cloudinary.go
    imagekit.go
    r2.go

account/
  handler.go
  service.go
  repository.go
  models.go

backup/
  handler.go
  service.go
  repository.go
  models.go

dashboard/
  handler.go
  service.go
  repository.go
  models.go
```

Each module owns:

* handler - HTTP routes and request handling
* service - business logic and orchestration
* repository - database access
* models - data structures

---

# Storage Flow

## Upload

```
User
↓
API Handler (validation)
↓
Storage Service
↓
Storage Orchestrator (select provider)
↓
Storage Adapter (Cloudinary/ImageKit/R2)
↓
Upload to Provider
↓
Save Metadata to PostgreSQL
↓
Return Response
```

## Download

```
User
↓
API Handler
↓
Asset Service (check permissions)
↓
Storage Service
↓
Retrieve metadata from PostgreSQL
↓
Storage Adapter
↓
Generate signed URL or stream file
↓
Return to user
```

---

# Backup Flow (Future)

```
Upload Complete
↓
Queue Backup Job
↓
Background Worker
↓
Archive Storage (Alibaba OSS Archive)
↓
Update Backup Metadata
↓
Mark as Backed Up
```

User operations never wait for backup completion.

---

# Storage Adapter Pattern

All storage providers implement the same interface:

```go
type StorageAdapter interface {
    Upload(ctx context.Context, input UploadInput) (*UploadResult, error)
    Download(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

Implementations:

* `CloudinaryAdapter`
* `ImageKitAdapter`
* `R2Adapter`

Business logic never depends on specific provider SDKs.

---

# Data Flow

## Request Flow

```
HTTP Request
↓
Fiber Handler
↓
Validate Request (validator)
↓
Service Layer
↓
Repository Layer
↓
sqlc Generated Queries
↓
PostgreSQL
```

## Response Flow

```
PostgreSQL Result
↓
Repository (map to models)
↓
Service (business logic)
↓
Handler (format response)
↓
JSON Response
```

---

# Error Handling

Errors propagate up the stack:

```
Repository → Service → Handler → HTTP Response
```

Handlers convert errors to appropriate HTTP status codes.

---

# Authentication Flow (Future)

```
User Login
↓
API Handler
↓
Account Service
↓
Verify Credentials
↓
Generate JWT
↓
Return Token
```

```
Authenticated Request
↓
JWT Middleware
↓
Validate Token
↓
Extract User Context
↓
Handler (with user context)
```

---

# Key Principles

1. **API First** - All business logic in API
2. **Thin Clients** - Web and CLI only handle presentation
3. **Storage Abstraction** - Providers are implementation details
4. **Database as Truth** - PostgreSQL is source of truth
5. **Explicit Errors** - Handle errors explicitly in Go
6. **Context Propagation** - Use context.Context throughout
7. **Type Safety** - Use sqlc for type-safe SQL

---

# Non-Goals

This architecture does not support:

* Microservices
* Event sourcing
* CQRS
* Complex DI containers
* Plugin systems

Keep it simple.
