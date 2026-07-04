# Filora DAM API Documentation

> ⚠️ **Legacy.** This documents the **current (legacy) API implementation**
> (JWT + password auth, per-user assets, single-layer storage). It does **not**
> reflect the target design (Clerk auth, RBAC, galleries/albums, two-layer
> storage). For the target, see [`/docs`](../../docs/README.md); for the gap and
> migration plan, see the [roadmap](../../docs/product/roadmap.md). Update this
> file when the API is rebuilt.

Complete REST API reference for Filora Digital Asset Management platform.

## Base URL

```
http://localhost:9000/api/v1
```

## Authentication

All protected endpoints require a JWT token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

Tokens expire after 24 hours and must be refreshed by logging in again.

---

## Authentication Endpoints

### Register User

Create a new user account.

**Endpoint:** `POST /auth/register`

**Request:**
```json
{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "securepassword123"
}
```

**Validation:**
- Email: required, valid email format
- Name: required, 2-255 characters
- Password: required, minimum 8 characters

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "John Doe",
      "storage_quota": 5368709120,
      "storage_used": 0,
      "created_at": "2026-06-24T10:00:00Z",
      "updated_at": "2026-06-24T10:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**Errors:**
- `400 Bad Request`: Invalid input (email taken, password too short)
- `409 Conflict`: Email already taken

---

### Login

Authenticate user and get JWT token.

**Endpoint:** `POST /auth/login`

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "John Doe",
      "storage_quota": 5368709120,
      "storage_used": 0,
      "created_at": "2026-06-24T10:00:00Z",
      "updated_at": "2026-06-24T10:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**Errors:**
- `400 Bad Request`: Missing email or password
- `401 Unauthorized`: Invalid email or password

---

## Account Endpoints

### Get User Info

Get authenticated user's profile information.

**Endpoint:** `GET /account/:id` ⚠️ Protected

**Parameters:**
- `id` (path): User ID

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "storage_quota": 5368709120,
    "storage_used": 1073741824,
    "created_at": "2026-06-24T10:00:00Z",
    "updated_at": "2026-06-24T10:00:00Z"
  }
}
```

**Errors:**
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: User not found

---

### Get Storage Quota

Get user's storage quota information.

**Endpoint:** `GET /account/:id/quota` ⚠️ Protected

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "quota": 5368709120,
    "used": 1073741824,
    "free": 4294967296
  }
}
```

All sizes in bytes.

---

## Storage Provider Endpoints

### List Providers

List all storage providers for authenticated user.

**Endpoint:** `GET /storage/providers` ⚠️ Protected

**Query Parameters:**
- `active` (optional): Filter by active status (true/false)

**Response:** `200 OK`
```json
{
  "success": true,
  "data": [
    {
      "id": "provider-uuid",
      "user_id": "user-uuid",
      "name": "My Cloudinary",
      "type": "cloudinary",
      "quota": null,
      "used": 1073741824,
      "is_active": true,
      "created_at": "2026-06-24T10:00:00Z",
      "updated_at": "2026-06-24T10:00:00Z"
    }
  ]
}
```

Note: Credentials are not returned in responses for security.

---

### Create Provider

Register a new storage provider.

**Endpoint:** `POST /storage/providers` ⚠️ Protected

**Request:**
```json
{
  "name": "My Cloudinary",
  "type": "cloudinary",
  "credentials": {
    "cloud_name": "your_cloud_name",
    "api_key": "your_api_key",
    "api_secret": "your_api_secret"
  },
  "quota": 10737418240
}
```

**Provider Types:**
- `cloudinary`: Cloudinary image/video hosting
- `imagekit`: ImageKit image optimization
- `r2`: Cloudflare R2 object storage

**Provider Credentials:**

**Cloudinary:**
```json
{
  "cloud_name": "string",
  "api_key": "string",
  "api_secret": "string"
}
```

**ImageKit:**
```json
{
  "public_key": "string",
  "private_key": "string",
  "url_endpoint": "string"
}
```

**R2:**
```json
{
  "account_id": "string",
  "access_key_id": "string",
  "secret_access_key": "string",
  "bucket_name": "string",
  "endpoint": "string"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "provider-uuid",
    "user_id": "user-uuid",
    "name": "My Cloudinary",
    "type": "cloudinary",
    "quota": null,
    "used": 0,
    "is_active": true,
    "created_at": "2026-06-24T10:00:00Z",
    "updated_at": "2026-06-24T10:00:00Z"
  }
}
```

**Errors:**
- `400 Bad Request`: Invalid provider type or credentials
- `409 Conflict`: Provider credentials already exist

---

### Get Provider

Get specific provider details.

**Endpoint:** `GET /storage/providers/:id` ⚠️ Protected

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "provider-uuid",
    "user_id": "user-uuid",
    "name": "My Cloudinary",
    "type": "cloudinary",
    "quota": null,
    "used": 1073741824,
    "is_active": true,
    "created_at": "2026-06-24T10:00:00Z",
    "updated_at": "2026-06-24T10:00:00Z"
  }
}
```

**Errors:**
- `403 Forbidden`: Provider not owned by user
- `404 Not Found`: Provider not found

---

### Deactivate Provider

Deactivate a storage provider.

**Endpoint:** `DELETE /storage/providers/:id` ⚠️ Protected

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "message": "Provider deactivated successfully"
  }
}
```

---

## Upload Endpoint

### Upload File

Upload a file to active storage provider.

**Endpoint:** `POST /storage/upload` ⚠️ Protected

**Request:** `multipart/form-data`
- `file` (required): File to upload (max 5GB)

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "asset-uuid",
    "user_id": "user-uuid",
    "name": "photo.jpg",
    "type": "image",
    "mime_type": "image/jpeg",
    "size": 2048576,
    "hash": "abc123def456...",
    "tags": [],
    "metadata": {
      "width": 1920,
      "height": 1080
    },
    "locations": [
      {
        "id": "location-uuid",
        "asset_id": "asset-uuid",
        "provider_id": "provider-uuid",
        "provider_key": "public_id",
        "url": "https://res.cloudinary.com/...",
        "metadata": {},
        "created_at": "2026-06-24T10:00:00Z"
      }
    ],
    "created_at": "2026-06-24T10:00:00Z",
    "updated_at": "2026-06-24T10:00:00Z"
  }
}
```

**Features:**
- Automatic MIME type detection
- SHA-256 hash-based deduplication
- Automatic asset type classification
- Storage location tracking
- Quota enforcement

**Errors:**
- `400 Bad Request`: No file uploaded
- `400 Bad Request`: No active storage provider
- `413 Payload Too Large`: File exceeds size limit
- `507 Insufficient Storage`: User quota exceeded

---

## Download Endpoint

### Download File

Download a file from storage.

**Endpoint:** `GET /storage/download/:id` ⚠️ Protected

**Parameters:**
- `id` (path): Asset ID

**Response:** `200 OK` (file stream)

**Errors:**
- `403 Forbidden`: Asset not owned by user
- `404 Not Found`: Asset not found
- `500 Internal Server Error`: Storage provider error

---

## Asset Endpoints

### List Assets

List all assets for authenticated user.

**Endpoint:** `GET /assets` ⚠️ Protected

**Query Parameters:**
- `limit` (optional): Results per page (1-100, default 20)
- `offset` (optional): Page offset (default 0)

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "assets": [
      {
        "id": "asset-uuid",
        "name": "photo.jpg",
        "type": "image",
        "mime_type": "image/jpeg",
        "size": 2048576,
        "hash": "abc123...",
        "tags": ["vacation", "2024"],
        "metadata": {},
        "created_at": "2026-06-24T10:00:00Z",
        "updated_at": "2026-06-24T10:00:00Z"
      }
    ],
    "total": 42,
    "limit": 20,
    "offset": 0
  }
}
```

---

### Get Asset

Get specific asset details.

**Endpoint:** `GET /assets/:id` ⚠️ Protected

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "asset-uuid",
    "name": "photo.jpg",
    "type": "image",
    "mime_type": "image/jpeg",
    "size": 2048576,
    "hash": "abc123...",
    "tags": ["vacation", "2024"],
    "metadata": {
      "width": 1920,
      "height": 1080
    },
    "locations": [
      {
        "id": "location-uuid",
        "asset_id": "asset-uuid",
        "provider_id": "provider-uuid",
        "provider_key": "public_id",
        "url": "https://...",
        "metadata": {},
        "created_at": "2026-06-24T10:00:00Z"
      }
    ],
    "created_at": "2026-06-24T10:00:00Z",
    "updated_at": "2026-06-24T10:00:00Z"
  }
}
```

---

### Search Assets

Search assets by name.

**Endpoint:** `GET /assets/search` ⚠️ Protected

**Query Parameters:**
- `q` (required): Search query
- `limit` (optional): Results per page (default 20)
- `offset` (optional): Page offset (default 0)

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "assets": [...],
    "total": 5,
    "limit": 20,
    "offset": 0
  }
}
```

---

### Filter Assets by Type

Filter assets by type.

**Endpoint:** `GET /assets/filter/:type` ⚠️ Protected

**Parameters:**
- `type` (path): Asset type (image, video, document, archive, file)

**Query Parameters:**
- `limit` (optional): Results per page (default 20)
- `offset` (optional): Page offset (default 0)

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "assets": [...],
    "total": 12,
    "limit": 20,
    "offset": 0
  }
}
```

---

### Update Asset Tags

Update tags for an asset.

**Endpoint:** `PUT /assets/:id/tags` ⚠️ Protected

**Request:**
```json
{
  "tags": ["vacation", "2024", "important"]
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "message": "Tags updated successfully"
  }
}
```

---

### Delete Asset

Delete an asset and remove from storage.

**Endpoint:** `DELETE /assets/:id` ⚠️ Protected

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "message": "Asset deleted successfully"
  }
}
```

---

## Dashboard Endpoint

### Get Dashboard

Get complete dashboard statistics.

**Endpoint:** `GET /dashboard/` ⚠️ Protected

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "stats": {
      "total_assets": 42,
      "total_size": 1073741824,
      "total_size_gb": "1.00 GB",
      "unique_types": 4,
      "storage_quota": 5368709120,
      "storage_used": 1073741824,
      "storage_free": 4294967296
    },
    "type_counts": [
      {
        "type": "image",
        "count": 28
      },
      {
        "type": "video",
        "count": 8
      },
      {
        "type": "document",
        "count": 5
      },
      {
        "type": "archive",
        "count": 1
      }
    ],
    "recent_assets": [
      {
        "id": "asset-uuid",
        "name": "latest.jpg",
        "type": "image",
        "mime_type": "image/jpeg",
        "size": 2048576,
        "created_at": "2026-06-24T15:30:00Z"
      }
    ]
  }
}
```

---

## Error Responses

All errors follow the standard format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message"
  }
}
```

### Common HTTP Status Codes

| Status | Meaning |
|--------|---------|
| 200 | OK - Request successful |
| 201 | Created - Resource created |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Missing/invalid token |
| 403 | Forbidden - Access denied |
| 404 | Not Found - Resource not found |
| 409 | Conflict - Resource already exists |
| 413 | Payload Too Large - File too big |
| 500 | Internal Server Error - Server error |
| 507 | Insufficient Storage - Quota exceeded |

---

## Asset Types

Assets are automatically classified:

| MIME Type Pattern | Asset Type |
|-------------------|-----------|
| `image/*` | image |
| `video/*` | video |
| `application/pdf`, `text/*`, `application/msword`, etc. | document |
| `application/zip`, `application/x-rar-compressed`, etc. | archive |
| Others | file |

---

## Rate Limiting

Currently unlimited. Future versions will implement:
- 1000 requests/hour per user
- 100 MB/sec upload bandwidth
- 500 concurrent downloads

---

## Pagination

All list endpoints support pagination:

```
GET /assets?limit=20&offset=0
GET /assets?limit=50&offset=50  # Second page
```

- Default limit: 20
- Max limit: 100
- Offset is 0-based

---

## Timestamps

All timestamps are in ISO 8601 format (UTC):

```
2026-06-24T10:00:00Z
```

---

## Size Units

All file sizes in API are in **bytes**:
- 1 KB = 1024 bytes
- 1 MB = 1048576 bytes
- 1 GB = 1073741824 bytes
- 5 GB default quota = 5368709120 bytes

---

## Example Workflows

### Complete Upload & Download

```bash
# 1. Register
TOKEN=$(curl -X POST http://localhost:9000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","name":"User","password":"pass123456"}' \
  | jq -r '.data.token')

# 2. Create storage provider
curl -X POST http://localhost:9000/api/v1/storage/providers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"Cloudinary","type":"cloudinary",
    "credentials":{"cloud_name":"...","api_key":"...","api_secret":"..."}
  }'

# 3. Upload file
ASSET_ID=$(curl -X POST http://localhost:9000/api/v1/storage/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@photo.jpg" \
  | jq -r '.data.id')

# 4. Get asset details
curl http://localhost:9000/api/v1/assets/$ASSET_ID \
  -H "Authorization: Bearer $TOKEN" | jq

# 5. Download file
curl http://localhost:9000/api/v1/storage/download/$ASSET_ID \
  -H "Authorization: Bearer $TOKEN" -o downloaded.jpg

# 6. View dashboard
curl http://localhost:9000/api/v1/dashboard/ \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## Support

For issues or questions:
- Check TESTING_MANUAL.md for endpoint examples
- Review error responses for detailed messages
- Check application logs for debugging
