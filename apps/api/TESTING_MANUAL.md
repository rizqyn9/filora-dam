# Manual Testing Guide for Filora API

> ⚠️ **Legacy.** These scenarios target the **current (legacy) API**
> (register/login with passwords, per-user assets). They do not reflect the
> target design in [`/docs`](../../docs/README.md). See the
> [roadmap](../../docs/product/roadmap.md).

This guide provides manual test scenarios to verify API functionality.

## Setup

1. Start the API:
```bash
cd apps/api
make run
# Server runs on http://localhost:9000
```

2. Set environment variables:
```bash
export API_BASE="http://localhost:9000"
```

## Test Scenarios

### 1. Health Check

```bash
curl -s $API_BASE/health | jq
```

Expected response:
```json
{
  "success": true,
  "data": {
    "status": "ok"
  }
}
```

### 2. User Registration

```bash
curl -X POST $API_BASE/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "name": "Test User",
    "password": "securepassword123"
  }' | jq
```

Response includes:
- User object with ID, email, name, storage info
- JWT token for subsequent requests

Save the token:
```bash
TOKEN="<token_from_response>"
```

### 3. User Login

```bash
curl -X POST $API_BASE/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "securepassword123"
  }' | jq
```

### 4. Get User Info

```bash
curl -s $API_BASE/api/v1/account/$USER_ID \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 5. Get Storage Quota

```bash
curl -s $API_BASE/api/v1/account/$USER_ID/quota \
  -H "Authorization: Bearer $TOKEN" | jq
```

Expected response:
```json
{
  "success": true,
  "data": {
    "quota": 5368709120,
    "used": 0,
    "free": 5368709120
  }
}
```

### 6. Create Storage Provider

For Cloudinary:
```bash
curl -X POST $API_BASE/api/v1/storage/providers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "My Cloudinary",
    "type": "cloudinary",
    "credentials": {
      "cloud_name": "your_cloud_name",
      "api_key": "your_api_key",
      "api_secret": "your_api_secret"
    }
  }' | jq
```

### 7. List Storage Providers

```bash
curl -s $API_BASE/api/v1/storage/providers \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 8. Upload File

```bash
curl -X POST $API_BASE/api/v1/storage/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/path/to/file.jpg" | jq
```

Response includes:
- Asset ID
- File name, type, size
- Hash for deduplication
- Storage location

### 9. List Assets

```bash
curl -s "$API_BASE/api/v1/assets?limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 10. Get Asset Details

```bash
curl -s $API_BASE/api/v1/assets/$ASSET_ID \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 11. Search Assets

```bash
curl -s "$API_BASE/api/v1/assets/search?q=photo&limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 12. Filter Assets by Type

```bash
curl -s "$API_BASE/api/v1/assets/filter/image?limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN" | jq
```

Asset types: `image`, `video`, `document`, `archive`, `file`

### 13. Update Asset Tags

```bash
curl -X PUT $API_BASE/api/v1/assets/$ASSET_ID/tags \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "tags": ["vacation", "2024", "important"]
  }' | jq
```

### 14. Download Asset

```bash
curl -s $API_BASE/api/v1/storage/download/$ASSET_ID \
  -H "Authorization: Bearer $TOKEN" \
  -o downloaded_file.jpg
```

### 15. Get Dashboard

```bash
curl -s $API_BASE/api/v1/dashboard/ \
  -H "Authorization: Bearer $TOKEN" | jq
```

Response includes:
- Total assets and storage usage
- Asset distribution by type
- Recent uploads
- Quota information

### 16. Delete Asset

```bash
curl -X DELETE $API_BASE/api/v1/assets/$ASSET_ID \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 17. Deactivate Provider

```bash
curl -X DELETE $API_BASE/api/v1/storage/providers/$PROVIDER_ID \
  -H "Authorization: Bearer $TOKEN" | jq
```

## Error Testing

### Missing Authentication

```bash
curl -s $API_BASE/api/v1/assets
# Should return 401 Unauthorized
```

### Invalid Token

```bash
curl -s $API_BASE/api/v1/assets \
  -H "Authorization: Bearer invalid-token"
# Should return 401 Unauthorized
```

### Expired Token

```bash
# Generate token, wait 24 hours, then try to use it
curl -s $API_BASE/api/v1/assets \
  -H "Authorization: Bearer $OLD_TOKEN"
# Should return 401 Unauthorized
```

### Access Denied (Wrong User)

```bash
# Create 2 users, upload asset as user1
# Login as user2, try to download user1's asset
curl -s $API_BASE/api/v1/storage/download/$ASSET_ID \
  -H "Authorization: Bearer $TOKEN_USER2"
# Should return 403 Forbidden
```

### Invalid Request

```bash
curl -X POST $API_BASE/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalid-email",
    "name": "User",
    "password": "short"
  }' | jq
# Should return 400 Bad Request
```

## Load Testing

Using Apache Bench:

```bash
# Register user first
TOKEN="..."

# Test file upload concurrency
ab -n 100 -c 10 \
  -H "Authorization: Bearer $TOKEN" \
  -p file_upload.json \
  -T "application/json" \
  $API_BASE/api/v1/storage/upload

# Test asset listing
ab -n 1000 -c 20 \
  -H "Authorization: Bearer $TOKEN" \
  $API_BASE/api/v1/assets
```

## Deduplication Testing

```bash
# Upload same file twice
FILE_PATH="/path/to/file.jpg"

# First upload
RESPONSE1=$(curl -s -X POST $API_BASE/api/v1/storage/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@$FILE_PATH")

ASSET1_ID=$(echo $RESPONSE1 | jq -r '.data.id')

# Second upload (same file)
RESPONSE2=$(curl -s -X POST $API_BASE/api/v1/storage/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@$FILE_PATH")

ASSET2_ID=$(echo $RESPONSE2 | jq -r '.data.id')

# Check if same asset was returned (deduplication)
if [ "$ASSET1_ID" = "$ASSET2_ID" ]; then
  echo "✅ Deduplication working!"
else
  echo "❌ Deduplication failed!"
fi
```

## Quota Testing

```bash
# Get initial quota
curl -s $API_BASE/api/v1/account/$USER_ID/quota \
  -H "Authorization: Bearer $TOKEN" | jq '.data.free'

# Upload large file
curl -X POST $API_BASE/api/v1/storage/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@large_file.bin"

# Check quota again (should be less)
curl -s $API_BASE/api/v1/account/$USER_ID/quota \
  -H "Authorization: Bearer $TOKEN" | jq '.data'
```

## Batch Operations (Helper Script)

Create `test_api.sh`:

```bash
#!/bin/bash

API="http://localhost:9000"
EMAIL="test_$(date +%s)@example.com"

# Register
REGISTER=$(curl -s -X POST $API/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"name\":\"Test\",\"password\":\"pass123456\"}")

TOKEN=$(echo $REGISTER | jq -r '.data.token')
USER_ID=$(echo $REGISTER | jq -r '.data.user.id')

echo "Token: $TOKEN"
echo "User ID: $USER_ID"

# Get quota
curl -s $API/api/v1/account/$USER_ID/quota \
  -H "Authorization: Bearer $TOKEN" | jq '.data'

# Create provider (requires credentials)
# ... add provider creation code ...

# Upload test file
curl -X POST $API/api/v1/storage/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@test_file.txt"

# Get dashboard
curl -s $API/api/v1/dashboard/ \
  -H "Authorization: Bearer $TOKEN" | jq '.data'
```

Run with:
```bash
bash test_api.sh
```

## Monitoring

While running manual tests, monitor server logs:

```bash
# Check for errors
tail -f logs/api.log | grep ERROR

# Check request latency
tail -f logs/api.log | grep "duration"
```
