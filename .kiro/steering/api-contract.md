---
inclusion: fileMatch
fileMatchPattern: ["api/**/*.go", "web/src/features/**/*"]
---

# API Contract

The canonical API spec lives at `#[[file:api/openapi.yaml]]`.

When implementing or modifying:
- **API handlers**: endpoints, request/response shapes, and status codes must match the spec.
- **Web API client code** (`web/src/features/*/api.ts`): request bodies and response parsing must match the spec. Use Zod schemas for runtime I/O validation — never trust raw responses.

## Web ↔ API I/O Pattern

```typescript
// 1. Define Zod schema matching the API response shape
const SpaceSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  owner_id: z.number(),
  storage_quota_bytes: z.number(),
  storage_used_bytes: z.number(),
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
});

// 2. Parse response through Zod — fail loud if API contract drifts
const data = SpaceSchema.parse(response.data);
```

Do NOT use generated TypeScript types from the spec. Zod schemas ARE the types (via `z.infer<>`). This gives runtime validation + compile-time types in one place.

## Response Envelope

All API responses follow:

```json
{ "success": true, "data": <payload> }
{ "success": false, "error": { "code": "...", "message": "..." } }
```

Web client should unwrap this envelope and throw on `success: false`.
