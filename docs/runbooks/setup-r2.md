# Runbook: Setup Cloudflare R2 for Filora

Step-by-step guide to configure a Cloudflare R2 bucket as a storage account.

---

## Prerequisites

- Cloudflare account (free tier includes 10 GB R2 storage)
- Filora API running with database migrated
- `curl` or similar for API calls

## 1. Create an R2 Bucket

1. Go to [Cloudflare Dashboard](https://dash.cloudflare.com) → **R2 Object Storage**
2. Click **Create bucket**
3. Name it something like `filora-serving` (or `filora-archive` for the archive layer)
4. Region: **Automatic** (or pick closest to you)
5. Click **Create bucket**

## 2. Create R2 API Token

1. In the R2 page, click **Manage R2 API Tokens** (right sidebar)
2. Click **Create API token**
3. Permissions: **Object Read & Write**
4. Specify bucket: select the bucket you just created
5. TTL: no expiry (or set as needed)
6. Click **Create API Token**
7. Save these values — they appear only once:

```
Access Key ID:     <your-access-key-id>
Secret Access Key: <your-secret-access-key>
```

## 3. Get Your Account ID

Your Cloudflare Account ID is visible in:
- Dashboard URL: `https://dash.cloudflare.com/<account-id>/...`
- Or: R2 bucket overview → the S3 API endpoint shows it:
  `https://<account-id>.r2.cloudflarestorage.com`

## 4. (Optional) Enable Public Access for Serving Layer

If this bucket is for the **serving layer** (L1), you want files publicly accessible:

1. Go to your bucket → **Settings** → **Public Access**
2. Option A: **R2.dev subdomain** — enable it. You get a URL like:
   `https://pub-<hash>.r2.dev`
3. Option B: **Custom domain** — connect your own domain (needs Cloudflare DNS)

For the **archive layer** (L2), skip this — files don't need public URLs.

## 5. Register the Account in Filora

Call the storage accounts API to register this R2 bucket:

```bash
# Serving layer example
curl -X POST http://localhost:8080/api/v1/storage/accounts \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "r2",
    "label": "R2 Serving #1",
    "layer": "serving",
    "credentials": {
      "account_id": "<your-cloudflare-account-id>",
      "access_key_id": "<your-access-key-id>",
      "secret_key": "<your-secret-access-key>",
      "bucket": "filora-serving",
      "public_url": "https://pub-<hash>.r2.dev"
    },
    "quota_bytes": 10737418240
  }'
```

```bash
# Archive layer example
curl -X POST http://localhost:8080/api/v1/storage/accounts \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "r2",
    "label": "R2 Archive #1",
    "layer": "archive",
    "credentials": {
      "account_id": "<your-cloudflare-account-id>",
      "access_key_id": "<your-access-key-id>",
      "secret_key": "<your-secret-access-key>",
      "bucket": "filora-archive",
      "public_url": ""
    },
    "quota_bytes": 10737418240
  }'
```

Notes:
- `quota_bytes`: 10737418240 = 10 GB (R2 free tier). Adjust as needed.
- `public_url`: only needed for serving layer. Leave empty for archive.
- Credentials are encrypted at rest (AES-256-GCM) if `ENCRYPTION_KEY` is set.

## 6. Verify

```bash
# List accounts — should show your new R2 account(s)
curl http://localhost:8080/api/v1/storage/accounts \
  -H "Authorization: Bearer <your-token>"
```

## 7. Test Upload

```bash
curl -X POST http://localhost:8080/api/v1/assets/upload \
  -H "Authorization: Bearer <your-token>" \
  -F "space_id=<your-space-id>" \
  -F "file=@/path/to/test-photo.jpg"
```

If successful, the response returns the asset record. The file is now:
- Stored in the R2 serving bucket
- A `storage_locations` record exists with status `stored`
- An `archive_sync_job` is enqueued (will copy to archive layer if configured)

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `no available storage account for layer serving` | No active serving account registered | Register one (step 5) |
| `upload to storage: r2 upload: AccessDenied` | Wrong API token permissions | Recreate token with Object Read & Write |
| `elect storage account: quota exceeded` | Bucket quota full | Register another account or increase quota |
| Upload succeeds but no public URL | `public_url` empty or R2.dev not enabled | Enable public access (step 4) |

## Multiple R2 Accounts

Filora supports pooling. Register as many R2 accounts as you want — the election
strategy picks the one with most remaining quota. Each can be a different bucket
under the same Cloudflare account, or buckets across different Cloudflare accounts.

```bash
# Second serving account
curl -X POST http://localhost:8080/api/v1/storage/accounts \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "r2",
    "label": "R2 Serving #2",
    "layer": "serving",
    "credentials": { ... },
    "quota_bytes": 10737418240
  }'
```

## Credential JSON Schema (R2)

| Field | Required | Description |
|-------|----------|-------------|
| `account_id` | Yes | Cloudflare Account ID |
| `access_key_id` | Yes | R2 API Token Access Key ID |
| `secret_key` | Yes | R2 API Token Secret Access Key |
| `bucket` | Yes | Bucket name |
| `public_url` | No | Public base URL for serving (R2.dev or custom domain) |
