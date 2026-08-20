# Runbook: Setup Authentication

Step-by-step guide to bootstrap auth for a fresh Filora installation.

---

## Prerequisites

- Filora API compiled (`go build ./...`)
- Database migrated (schema applied)
- `.env` configured with `DATABASE_URL`

## 1. Create Superuser

Run the seed command (interactive):

```bash
cd api
go run cmd/seed/main.go
```

It will ask:
- Email
- Name
- Password (min 8 characters)

This creates the first user with `superuser` role.

## 2. Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your@email.com",
    "password": "yourpassword",
    "client": "web"
  }'
```

Response:
```json
{
  "success": true,
  "data": {
    "token": "a1b2c3d4...",
    "user": { "id": 1, "email": "your@email.com", "name": "Your Name" }
  }
}
```

Save the `token` — use it as `Authorization: Bearer <token>` for all API calls.

## 3. Invite a Family Member

```bash
# Create an invitation (as superuser)
curl -X POST http://localhost:8080/api/v1/invitations \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "space_id": "<your-space-id>",
    "email": "family@email.com",
    "role": "editor"
  }'
```

Response includes an `invite_token`. Share this link with the family member:

```
https://your-domain.com/register?invite=<invite_token>
```

## 4. Family Member Registers

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "invite_token": "<the-token-from-step-3>",
    "password": "theirpassword",
    "name": "Family Member"
  }'
```

They're now a user with the invited role on your space.

## Token Behavior

| Client | Idle TTL | Behavior |
|--------|----------|----------|
| Web | 7 days | Expires after 7 days of no API calls |
| CLI | 90 days | Expires after 90 days of no API calls |

Tokens use sliding window — every request extends the expiry. A user who uses
the app daily never gets logged out.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/login` | No | Login, get token |
| POST | `/auth/register` | No | Register with invite token |
| POST | `/api/v1/auth/logout` | Yes | Revoke current session |
| GET | `/api/v1/auth/me` | Yes | Get current user |
| POST | `/api/v1/auth/change-password` | Yes | Change password (revokes all sessions) |

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `UNAUTHORIZED: missing authorization token` | No Bearer header | Add `Authorization: Bearer <token>` |
| `UNAUTHORIZED: invalid or expired token` | Token expired or revoked | Login again |
| `AUTH_FAILED: invalid email or password` | Wrong credentials | Check email/password |
| `REGISTER_FAILED: invalid or expired invitation` | Invite token wrong or used | Create new invitation |
| First request slow after server restart | Cache cold, hits DB once | Normal — subsequent requests use cache |
