# Filora Agent Operating Manual

Read this document before performing any task.

This repository is developed using multiple AI coding agents:

* Claude Code
* Kiro
* Antigravity

All agents must follow the same architecture, conventions, and development philosophy.

---

# Required Reading Order

Before making changes, read:

1. docs/database/README.md

---

# Project Overview

Filora is a multi-cloud Digital Asset Management (DAM) platform.

Users upload, organize, synchronize, and manage digital assets.

Storage complexity is abstracted from users.

Users never need to know:

* where files are stored
* which storage provider is used
* which account stores their files

Filora manages storage automatically.

---

# Project Stage

Current Stage: MVP

Filora is a solo-developed project.

Primary goals:

1. Ship working features quickly
2. Validate product-market fit
3. Keep code maintainable
4. Minimize operational complexity

Do not optimize for hypothetical future requirements.

---

# Technology Stack

## API & CLI (Backend)

* Go 1.23+
* Fiber v3 (Web Framework)
* sqlc (Type-safe SQL)
* PostgreSQL (Neon)
* validator v10 (Input validation)
* golang-migrate (Database migrations)

## Web (Frontend)

* React 19
* TypeScript
* TanStack Query (Data fetching)
* TanStack Router (Routing)
* Tailwind CSS v4
* Shadcn UI
* Zod v4

## Storage Providers

* Cloudinary
* ImageKit
* Cloudflare R2

## Backup Storage (Planned)

* Alibaba OSS Archive
* Google Cloud Storage Archive (alternative)

---

# Project Structure

Filora consists of three independent applications:

```
filora-dam/
├── apps/
│   ├── api/          # Go REST API server
│   ├── cli/          # Go CLI client (HTTP client)
│   └── web/          # React 19 frontend
├── AGENTS.md
├── CLAUDE.md
└── docs/
```

No shared packages between apps.

Each app is independent.

---

# API Architecture

API uses modular vertical slice architecture.

Each module owns:

* handler (HTTP routes)
* service (business logic)
* repository (database access)
* models (data structures)

Example structure:

```
apps/api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── modules/
│   │   ├── asset/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── models.go
│   │   ├── storage/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   ├── models.go
│   │   │   └── adapters/
│   │   │       ├── adapter.go
│   │   │       ├── cloudinary.go
│   │   │       ├── imagekit.go
│   │   │       └── r2.go
│   │   └── account/
│   │       ├── handler.go
│   │       ├── service.go
│   │       ├── repository.go
│   │       └── models.go
│   ├── database/
│   │   ├── db.go
│   │   ├── migrations/
│   │   └── queries/
│   ├── config/
│   │   └── config.go
│   └── lib/
│       └── response.go
├── go.mod
└── go.sum
```

Avoid layer-first architecture.

---

# CLI Architecture

CLI is a thin HTTP client.

CLI communicates with API via REST.

CLI does not contain business logic.

```
apps/cli/
├── cmd/
│   └── filora/
│       └── main.go
├── internal/
│   ├── client/
│   │   └── api_client.go
│   └── commands/
│       ├── upload.go
│       ├── download.go
│       ├── list.go
│       └── delete.go
├── go.mod
└── go.sum
```

---

# Decision Hierarchy

When making technical decisions prioritize:

1. Correctness
2. Simplicity
3. Readability
4. Developer Experience
5. Performance
6. Extensibility

Do not sacrifice simplicity for theoretical flexibility.

---

# Golden Rules

## Build First

Working software is more important than perfect architecture.

---

## Refactor Later

Do not introduce abstractions until at least two concrete implementations exist.

Duplication is acceptable.

Incorrect abstractions are expensive.

---

## Database First

Design order:

1. Database schema (SQL migration)
2. sqlc queries
3. Repository
4. Service
5. Handler (API)
6. UI

---

## Consistency First

Follow existing project patterns.

Do not introduce personal architectural preferences.

---

# Storage Rules

Storage providers are implementation details.

Business logic must never directly depend on:

* Cloudinary SDK
* ImageKit SDK
* R2 SDK

Only storage adapters may communicate with providers.

All adapters implement StorageAdapter interface:

```go
type StorageAdapter interface {
    Upload(ctx context.Context, input UploadInput) (*UploadResult, error)
    Download(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

---

# Metadata Rules

Files live in storage.

Truth lives in PostgreSQL.

Business logic must never use cloud storage as the source of truth.

---

# API Rules

API is the source of business logic.

Web and CLI are thin clients.

Business logic must not live in:

* React components
* CLI commands
* Handlers (keep handlers thin)

---

# Validation Rules

All external input must be validated.

Sources:

* HTTP requests
* Query parameters
* Route parameters
* Environment variables
* Webhook payloads
* CLI arguments

Use Go validator v10 for struct validation.

Never trust external data.

---

# Go Rules

Follow standard Go conventions:

* Use `gofmt` for formatting
* Use `golangci-lint` for linting
* Handle errors explicitly - never ignore errors
* Use interfaces for abstraction (StorageAdapter, Repository)
* Prefer composition over inheritance
* Use `context.Context` for cancellation and timeouts
* Follow Go naming conventions (PascalCase for exported, camelCase for unexported)

Avoid:

* `panic()` except for unrecoverable startup errors
* `interface{}` prefer concrete types or specific interfaces
* Global variables except for config

Keep functions small and focused.

---

# Database Rules

SQL migrations are the source of truth.

Every schema change requires:

1. SQL migration file (timestamp_name.up.sql)
2. sqlc query definitions
3. Repository update

Never modify production schemas manually.

Use sqlc for type-safe SQL queries.

Write explicit SQL - avoid ORMs for complex operations.

---

# Repository Rules

Repositories are responsible for:

* database access
* queries
* persistence

Repositories must not contain business logic.

Repositories use sqlc generated code.

---

# Service Rules

Services are responsible for:

* business rules
* orchestration
* workflows

Services coordinate repositories and adapters.

---

# Handler Rules

Handlers are responsible for:

* request validation
* calling services
* returning responses

Handlers must remain thin.

---

# API Response Format

Success:

```json
{
  "success": true,
  "data": {}
}
```

Error:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message"
  }
}
```

Maintain consistent response structure.

---

# Performance Rules

Avoid:

* N+1 queries
* unnecessary joins
* loading unused relations
* full table scans

Prefer explicit selects.

Do not optimize prematurely.

Measure before optimizing.

---

# Error Handling

In Go, handle errors explicitly:

```go
result, err := service.DoSomething(ctx)
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

Use error wrapping with `%w` for error chains.

Return errors up the stack.

Convert errors to HTTP responses in handlers.

---

# Testing Philosophy

Prioritize:

* repository tests
* service tests
* workflow tests

Prefer integration tests.

Avoid excessive mocking.

Test observable behavior, not implementation details.

---

# Logging

Log important events:

* uploads
* downloads
* deletions
* storage failures
* backup failures

Use structured logging (e.g., zerolog, zap).

Avoid noisy logs.

---

# Observability

Future stack:

* OpenTelemetry
* Grafana LGTM

Current MVP:

Keep observability simple.

Do not introduce complex observability frameworks unless required.

---

# Code Style

Prefer:

```go
asset, err := repo.FindByID(ctx, id)
```

Over:

```go
assetEntityAggregateRoot, err := repo.FindByID(ctx, id)
```

Use concise, descriptive names.

Avoid unnecessary verbosity.

---

# File Organization

Prefer:

```
asset/
  handler.go
  service.go
  repository.go
  models.go
```

Over:

```
handlers/
  asset_handler.go

services/
  asset_service.go

repositories/
  asset_repository.go
```

Modules should own their code.

---

# Forbidden Patterns

Do not introduce:

* Generic repositories
* Base repositories
* Base services
* Repository factories
* Service factories
* CQRS
* Event Sourcing
* Domain Events
* Dependency Injection Containers
* Plugin Frameworks
* Microservices

Unless explicitly requested.

---

# Agent Behaviour

Before making changes:

1. Understand the request
2. Search existing code
3. Follow conventions
4. Reuse existing implementations
5. Minimize changes
6. Avoid unrelated refactors

When uncertain:

Choose the simplest implementation.

---

# Final Rule

Filora is a product.

Not an architecture exercise.

Every abstraction must justify its existence.

Working software beats perfect architecture.
