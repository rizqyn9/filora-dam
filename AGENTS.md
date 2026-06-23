# Filora Agent Context

> Read this document before making any code changes.
>
> Filora is developed primarily using AI-assisted coding tools including Claude Code, Kiro, and Antigravity. This document exists to ensure all agents follow the same principles, architecture, and coding standards.

---

# Project Overview

Filora is a multi-cloud Digital Asset Management (DAM) and backup platform.

Users can upload, organize, search, manage, and synchronize digital assets through a single interface while Filora automatically manages storage across multiple providers.

Supported asset types:

* Images
* Videos
* Documents
* Archives
* Generic files

Supported storage providers:

* Cloudinary
* ImageKit
* Cloudflare R2

Future providers may be added without changing business logic.

---

# Project Stage

Current Stage: MVP

Filora is a solo-developed project.

Primary goals:

1. Ship working features quickly.
2. Validate product-market fit.
3. Keep the codebase simple.
4. Maintain high code quality.

Do not optimize for:

* Massive scale
* Enterprise requirements
* Multi-region deployments
* Hypothetical future features

---

# Vision

Users should never need to know:

* Which storage provider is used
* Which account stores their files
* How storage distribution works

Filora abstracts storage complexity.

Users manage assets.

Filora manages storage.

---

# Technology Stack

## Backend

* TypeScript
* Bun
* Elysia
* Drizzle ORM
* Neon PostgreSQL
* Zod v4
* Oxlint
* Oxfmt

## Frontend

* React 19
* TypeScript
* Tailwind CSS v4
* Shadcn UI
* Zod v4

## Storage

* Cloudinary
* ImageKit
* Cloudflare R2

---

# Development Philosophy

Prioritize:

1. Correctness
2. Simplicity
3. Readability
4. Developer Experience
5. Performance
6. Extensibility

Do not sacrifice simplicity for theoretical flexibility.

Filora is a product, not an architecture exercise.

---

# Golden Rules

## Build First

Prefer shipping working software over designing perfect systems.

---

## Refactor Later

Do not introduce abstractions until at least two concrete implementations exist.

Duplication is acceptable.

Incorrect abstractions are expensive.

---

## Database First

Design database schema first.

Implementation order:

1. Database schema
2. Repository
3. Service
4. API
5. UI

---

## Reuse Existing Patterns

Before introducing new patterns:

1. Search existing code.
2. Follow established conventions.
3. Reuse existing solutions.

Consistency is more important than personal preference.

---

# Architecture

Filora uses a modular vertical-slice architecture.

Each module owns its own:

* router
* service
* repository
* schema
* types

Example:

```text
src/modules/

  asset/
    router.ts
    service.ts
    repository.ts
    schema.ts
    types.ts

  storage/
    router.ts
    service.ts
    repository.ts
    schema.ts
    types.ts

  account/
    router.ts
    service.ts
    repository.ts
    schema.ts
    types.ts
```

Avoid layer-first structures such as:

```text
src/
  routes/
  services/
  repositories/
  schemas/
```

---

# Domain Boundaries

## Asset Module

Responsible for:

* asset metadata
* ownership
* tagging
* search
* asset versioning

Not responsible for:

* storage provider logic

---

## Storage Module

Responsible for:

* uploads
* downloads
* deletion
* storage orchestration
* provider integrations

Not responsible for:

* user management
* permissions

---

## Account Module

Responsible for:

* users
* quotas
* permissions
* subscription plans

---

## Dashboard Module

Responsible for:

* metrics
* statistics
* reporting
* storage visibility

---

# Storage Philosophy

Storage providers are implementation details.

Business logic must never depend directly on:

* Cloudinary SDK
* ImageKit SDK
* R2 SDK

Only storage adapters may communicate with providers.

---

## Required Abstraction

```ts
interface StorageAdapter {
  upload(input: UploadInput): Promise<UploadResult>

  download(key: string): Promise<ReadableStream>

  delete(key: string): Promise<void>

  exists(key: string): Promise<boolean>
}
```

All provider implementations must conform to this interface.

---

# Metadata Philosophy

Files live in storage.

Truth lives in PostgreSQL.

Business logic must rely on metadata stored in the database.

Never use storage providers as the source of truth.

---

# API Philosophy

The API is the primary interface.

Web and CLI are both API clients.

Business logic belongs in services.

Business logic must not live in:

* routers
* React components
* CLI commands

---

# Validation Rules

All external input must be validated.

Sources include:

* HTTP requests
* Query parameters
* Route parameters
* CLI arguments
* Environment variables
* Webhook payloads

Use Zod v4.

Never trust external data.

---

# TypeScript Rules

Enable strict mode.

Avoid:

```ts
any
```

Prefer:

```ts
unknown
```

and validate appropriately.

Use explicit types for public APIs.

Prefer inferred types internally.

---

# Database Rules

Drizzle schema is the source of truth.

Every schema change must include:

1. Schema update
2. Migration
3. Repository update

Never modify production schema without migrations.

Avoid raw SQL unless absolutely necessary.

---

# Repository Rules

Repositories are responsible for:

* database access
* queries
* persistence

Repositories must not contain business logic.

---

# Service Rules

Services are responsible for:

* business rules
* orchestration
* workflows

Services should coordinate repositories and adapters.

---

# Router Rules

Routers are responsible for:

* request validation
* calling services
* returning responses

Routers must remain thin.

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

Maintain consistent response structures across all endpoints.

---

# Performance Rules

Avoid:

* N+1 queries
* unnecessary joins
* full table scans
* loading unused relations

Prefer explicit selects.

Do not optimize prematurely.

Measure before optimizing.

---

# Error Handling

Prefer simple, understandable errors.

Avoid creating excessive error hierarchies.

Start simple.

Only introduce specialized errors when a real need exists.

---

# Testing Philosophy

Prioritize testing:

1. Repository behavior
2. Service behavior
3. Critical workflows

Prefer integration tests over heavily mocked unit tests.

Avoid testing implementation details.

Test observable behavior.

---

# Logging

Log important events:

* uploads
* downloads
* deletions
* storage failures
* provider failures

Logs should help diagnose problems.

Avoid excessive logging noise.

---

# Observability

Future observability stack:

* OpenTelemetry
* Grafana LGTM

Current MVP should prioritize simplicity.

Do not introduce complex observability frameworks unless required.

---

# Code Style

Prefer:

```ts
const asset = await repository.findById(id)
```

Over:

```ts
const assetEntityAggregateRoot =
  await repository.findById(id)
```

Use concise, descriptive names.

Avoid unnecessary verbosity.

---

# Avoid These Patterns

Do not introduce:

* GenericRepository<T>
* BaseRepository
* BaseService
* ServiceFactory
* RepositoryFactory
* AbstractFactory
* CQRS
* Event Sourcing
* Domain Events
* Dependency Injection Containers
* Microservices
* Plugin Frameworks

Unless explicitly requested.

---

# File Organization

Prefer:

```text
asset/
  repository.ts
  service.ts
  router.ts
```

Over:

```text
repositories/
  asset.repository.ts

services/
  asset.service.ts

routers/
  asset.router.ts
```

Modules should own their code.

---

# Agent Behavior

Before making changes:

1. Understand the feature request.
2. Search for existing implementations.
3. Follow existing conventions.
4. Minimize code changes.
5. Avoid unrelated refactors.

When unsure:

Prefer the simplest implementation.

---

# Non Goals

Filora is not:

* Google Drive clone
* Dropbox clone
* Social media platform
* Image editor
* Video editor
* CDN replacement

Filora focuses on:

* Asset management
* Storage orchestration
* Backup
* Synchronization
* Multi-cloud storage abstraction

---

# Final Rule

Every abstraction must justify its existence.

If a solution feels overly clever, choose the simpler alternative.

Working software beats perfect architecture.
