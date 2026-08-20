# Runbook: Setup Clerk Authentication for Filora

Step-by-step guide to configure Clerk as the auth provider.

---

## Prerequisites

- Clerk account (free tier: 10,000 MAUs)
- Filora API running with database migrated
- A domain or localhost for development

## 1. Create a Clerk Application

1. Go to [Clerk Dashboard](https://dashboard.clerk.com)
2. Click **Create application**
3. Name: `Filora`
4. Sign-in methods: pick what you want (Email, Google, GitHub, etc.)
5. Click **Create application**

## 2. Get API Keys

Go to **API Keys** in the sidebar. You need:

| Key | Where to use |
|-----|-------------|
| **Publishable Key** (`pk_...`) | Web frontend (public, safe to expose) |
| **Secret Key** (`sk_...`) | API server (private, never expose) |

Add to `api/.env`:

```bash
CLERK_SECRET_KEY=sk_test_...
```

Add to `web/.env`:

```bash
VITE_CLERK_PUBLISHABLE_KEY=pk_test_...
```

## 3. Configure Webhook for User Sync

Filora mirrors Clerk users locally. The webhook keeps them in sync.

1. Go to **Webhooks** in the Clerk sidebar
2. Click **Add Endpoint**
3. URL: `https://<your-domain>/webhooks/clerk`
   - For local dev: use a tunnel like [ngrok](https://ngrok.com) or [cloudflared](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/get-started/)
   - Example: `https://abc123.ngrok.io/webhooks/clerk`
4. Events to subscribe:
   - `user.created`
   - `user.updated`
   - `user.deleted`
5. Click **Create**
6. Copy the **Signing Secret** (`whsec_...`)

Add to `api/.env`:

```bash
CLERK_WEBHOOK_SECRET=whsec_...
```

## 4. Local Development with Tunnel

For local webhook testing:

```bash
# Option A: ngrok
ngrok http 8080
# Copy the https URL → update webhook endpoint in Clerk

# Option B: cloudflared (free, no account needed)
cloudflared tunnel --url http://localhost:8080
# Copy the https URL → update webhook endpoint in Clerk
```

## 5. Configure JWT Template (Optional)

By default, Clerk JWTs work out of the box. If you need custom claims:

1. Go to **JWT Templates** in Clerk sidebar
2. Create a template or edit the default
3. Filora only uses `sub` (Clerk user ID) from the JWT — no custom claims needed

## 6. Configure Allowed Origins

For the web frontend to work with Clerk:

1. Go to **Domains** in Clerk sidebar
2. Add your frontend URL:
   - Development: `http://localhost:5173` (Vite default)
   - Production: `https://your-domain.com`

## 7. Test the Integration

### Test webhook (user sync):

1. Sign up a new user via your frontend or Clerk's test user feature
2. Check the users table:

```bash
psql "$DATABASE_URL" -c "SELECT id, clerk_id, email, name FROM users;"
```

Should show the newly created user.

### Test JWT auth:

```bash
# Get a session token from Clerk (via frontend or Clerk Dashboard → Users → Sessions)
TOKEN="your-clerk-session-token"

curl http://localhost:8080/api/v1/spaces \
  -H "Authorization: Bearer $TOKEN"
```

Should return `200` with your spaces (empty list initially).

## 8. Grant Superuser Role

After your first sign-in triggers the webhook:

```sql
-- Find your user
SELECT id, email FROM users WHERE email = 'your@email.com';

-- Grant superuser
INSERT INTO user_roles (user_id, role_name) VALUES (<id>, 'superuser');
```

## Clerk Free Tier Limits

| Resource | Limit |
|----------|-------|
| Monthly active users | 10,000 |
| Social connections | Unlimited |
| Custom domains | 1 (production) |
| Webhooks | Unlimited |

For Filora's 3-5 family members, free tier is permanent.

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `UNAUTHORIZED: invalid token` | Wrong `CLERK_SECRET_KEY` or expired token | Verify key in .env matches dashboard |
| Webhook returns 401 | Wrong `CLERK_WEBHOOK_SECRET` | Copy signing secret again from Clerk |
| User not created after signup | Webhook not firing | Check webhook logs in Clerk dashboard |
| `user not found` after webhook | Webhook arrives before first API call | Normal race; JIT sync handles this |
| CORS errors on frontend | Origin not in allowed list | Add frontend URL to Clerk Domains |

## Session Token Lifecycle

```
Frontend                   Clerk                    Filora API
    |                        |                          |
    |--- Sign In ----------->|                          |
    |<-- Session Token ------|                          |
    |                        |--- Webhook: user.created -->|
    |                        |                          |-- Create user row
    |--- API call + Bearer token ---------------------->|
    |                        |                          |-- Verify JWT (sub=clerk_id)
    |                        |                          |-- Resolve clerk_id → user_id
    |<------------------------------- Response ---------|
```

## Multiple Environments

Use separate Clerk applications for dev/staging/production:

| Environment | Clerk App | Keys |
|-------------|-----------|------|
| Development | `Filora Dev` | `pk_test_...` / `sk_test_...` |
| Production | `Filora` | `pk_live_...` / `sk_live_...` |

Switch via `.env` per environment. Never use test keys in production.
