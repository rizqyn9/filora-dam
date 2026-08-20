# ADR-005: Self-Managed Auth with Opaque Tokens

Replace Clerk with self-managed authentication using opaque tokens and in-process cache.

---

## Status

Accepted (supersedes previous Clerk-based auth)

## Context

The initial design used Clerk for identity management (JWT verification, user
sync via webhook). This worked but introduced:

- External dependency for a 3-5 person family app
- Webhook complexity for user sync
- Reduced flexibility for token lifetime and session management
- Unnecessary vendor lock-in

The goal: manage auth manually with zero DB calls on most requests, while keeping
the ability to revoke sessions instantly.

## Decision

### Authentication flow

1. **Registration**: invite-only. Superuser creates invitation (opaque token) →
   shares link manually → invitee sets password → becomes user + space member.
2. **Login**: email + password → server returns opaque session token.
3. **Request auth**: Bearer token → SHA-256 hash → check in-process LRU cache →
   if miss, fallback to DB → cache result.
4. **Logout**: revoke session in DB + evict from cache.

### Token strategy

- Pure random opaque tokens (32 bytes, hex-encoded)
- Stored in DB as SHA-256 hash (never store raw)
- Sliding window TTL: web 7 days idle, CLI 90 days idle
- Token extended on every use (last_used_at + TTL = new expires_at)

### Cache strategy

- In-process Go map (`sync.RWMutex` + `map[string]*CachedSession`)
- No Redis or external cache (3-5 users = all tokens fit permanently in memory)
- Cache miss → DB lookup → populate cache
- Revocation → delete from cache + mark revoked in DB
- Server restart → cold cache, first request per user hits DB once

### Password

- bcrypt (via `golang.org/x/crypto/bcrypt`, DefaultCost)
- Change password revokes all sessions (force re-login)

### Superuser bootstrap

- `go run cmd/seed/main.go` — interactive, asks email + name + password
- Creates first user with superuser role
- Run once per installation

## Consequences

**Positive:**
- Zero external auth dependencies
- Zero DB calls for 99%+ of requests (cache always warm for 3-5 users)
- Full control over session lifecycle and token format
- Simpler deployment (no webhook, no Clerk keys, no third-party outage risk)

**Negative:**
- Must implement password reset ourselves (backlog)
- No social login (GitHub, Google) — acceptable for private family app
- No email verification for MVP (invite-only mitigates this)

**Removed:**
- `github.com/clerk/clerk-sdk-go/v2`
- `github.com/svix/svix-webhooks`
- Clerk webhook handler
- `CLERK_SECRET_KEY`, `CLERK_WEBHOOK_SECRET` env vars

---

**Previous:** [ADR-004](./004-archive-first-dam.md) — Archive-first DAM.
