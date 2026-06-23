# Filora DAM API

A high-performance, production-ready REST API for Digital Asset Management built with Go.

**Status:** MVP Complete ✅ (Phases 0-11)

## Key Features

- 🔐 **JWT Authentication** - Secure user registration and login with bcrypt password hashing
- ☁️ **Multi-Cloud Storage** - Support for Cloudinary, ImageKit, Cloudflare R2
- 📦 **Deduplication** - SHA-256 hash-based file deduplication per user
- 💾 **Quota Management** - Configurable per-user storage quotas (default 5GB)
- 🔍 **Search & Filtering** - Find assets by name or filter by type
- 📊 **Dashboard** - Real-time statistics and recent activity tracking
- 🚀 **Type-Safe** - sqlc for compile-time SQL verification
- 📝 **Well Documented** - Comprehensive API documentation with examples

## Tech Stack

- **Go** 1.23+ 
- **Fiber v3** - Lightweight web framework
- **PostgreSQL** - Persistent data storage
- **sqlc** - Type-safe SQL code generation
- **pgx/v5** - High-performance PostgreSQL driver
- **JWT** - Token-based authentication
- **bcrypt** - Secure password hashing

## Prerequisites

- Go 1.23+
- PostgreSQL 14+
- Git

## Quick Start

### 1. Clone and setup

```bash
git clone <repo-url>
cd filora-dam/apps/api
go mod download
```

### 2. Environment configuration

```bash
cp .env.example .env
```

Update `.env`:
```bash
PORT=9000
DATABASE_URL=postgres://user:pass@localhost/filora
JWT_SECRET=your-32-character-minimum-secret-key
ENVIRONMENT=development
```

### 3. Database setup

```bash
createdb filora
make migrate-up
```

### 4. Run the server

```bash
make run
```

Server starts on `http://localhost:9000`

Test: `curl http://localhost:9000/health`

## Available Commands

```bash
make run               # Run development server
make build             # Build production binary
make test              # Run tests
make fmt               # Format code with gofmt
make lint              # Run linter (golangci-lint)
make migrate-up        # Run database migrations
make migrate-down      # Rollback database
make sqlc              # Generate sqlc code from queries
make deps              # Update dependencies
```

## Project Structure

```
apps/api/
├── cmd/server/
│   └── main.go                  # Entry point (28 endpoints registered)
├── internal/
│   ├── config/
│   │   └── config.go            # Environment configuration
│   ├── database/
│   │   ├── db.go                # PostgreSQL connection pool
│   │   ├── migrations/          # Database schema (sqlc)
│   │   └── queries/             # SQL queries for sqlc
│   ├── lib/
│   │   ├── response.go          # HTTP response helpers
│   │   ├── jwt.go               # JWT token management (24h)
│   │   ├── password.go          # bcrypt password hashing
│   │   ├── hash.go              # SHA-256 file hashing
│   │   └── mime.go              # MIME type detection
│   ├── middleware/
│   │   └── auth.go              # JWT middleware
│   └── modules/                 # Feature modules (vertical slice)
│       ├── account/             # User auth/registration
│       ├── asset/               # Asset metadata management
│       ├── storage/             # Storage provider management
│       │   └── adapters/        # Cloudinary, ImageKit, R2
│       └── dashboard/           # Statistics & metrics
├── API.md                       # Complete API documentation
├── README.md                    # This file
├── TESTING.md                   # Testing strategy
├── TESTING_MANUAL.md            # Manual test examples
├── Makefile
├── sqlc.yaml
├── go.mod
└── go.sum
```

## Complete API

See [API.md](API.md) for comprehensive documentation.

**Quick Summary:**
- `POST /auth/register` - Create account
- `POST /auth/login` - Get JWT token
- `POST /storage/upload` - Upload file
- `GET /storage/download/:id` - Download file
- `GET /assets` - List assets
- `GET /assets/search?q=...` - Search by name
- `GET /assets/filter/:type` - Filter by type
- `GET /dashboard/` - View statistics

**28 total endpoints** across 5 modules.

## Testing

Tests available for core utilities (100% coverage):

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run lib tests only (passing ✅)
go test -v ./internal/lib

# View coverage
go test -cover ./...
```

**Current Coverage:**
- Password hashing/verification: ✅
- JWT token management: ✅
- File hashing (SHA-256): ✅

See [TESTING.md](TESTING.md) for strategy and [TESTING_MANUAL.md](TESTING_MANUAL.md) for curl examples.

## Manual Testing

Test all endpoints with provided curl examples:

```bash
# Start server
make run

# In another terminal
bash test_api.sh  # See TESTING_MANUAL.md for setup
```

## Examples

### Upload and Download
```bash
# Get token
TOKEN=$(curl -X POST http://localhost:9000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"u@example.com","name":"User","password":"pass123456"}' \
  | jq -r '.data.token')

# Create provider
curl -X POST http://localhost:9000/api/v1/storage/providers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Cloud","type":"cloudinary","credentials":{...}}'

# Upload
ASSET_ID=$(curl -X POST http://localhost:9000/api/v1/storage/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@photo.jpg" | jq -r '.data.id')

# Download
curl http://localhost:9000/api/v1/storage/download/$ASSET_ID \
  -H "Authorization: Bearer $TOKEN" -o downloaded.jpg
```

See [TESTING_MANUAL.md](TESTING_MANUAL.md) for 17 complete endpoint examples.

## Development

### Module Structure (Vertical Slice)

Each module owns its layer stack:
```
internal/modules/{module}/
├── handler.go       # HTTP routes
├── service.go       # Business logic
├── repository.go    # Database access
└── models.go        # Data structures
```

### Adding a new feature

1. **Database first** - Create migration
2. **SQL queries** - Write in `queries/*.sql`
3. **Repository** - Implement database access
4. **Service** - Add business logic
5. **Handler** - Create HTTP endpoints
6. **Tests** - Add unit tests

### Generate sqlc code

After modifying SQL queries:
```bash
make sqlc
```

## Environment Variables

**Required:**
```
PORT=9000                          # Server port
DATABASE_URL=postgres://user:pass@localhost/filora
JWT_SECRET=min-32-characters-secret
ENVIRONMENT=development            # or production
```

**Optional (Storage Providers):**
```
# Cloudinary
CLOUDINARY_CLOUD_NAME=...
CLOUDINARY_API_KEY=...
CLOUDINARY_API_SECRET=...

# ImageKit
IMAGEKIT_PUBLIC_KEY=...
IMAGEKIT_PRIVATE_KEY=...
IMAGEKIT_URL_ENDPOINT=...

# R2
R2_ACCOUNT_ID=...
R2_ACCESS_KEY_ID=...
R2_SECRET_ACCESS_KEY=...
R2_BUCKET_NAME=...
R2_ENDPOINT=...
```

## Performance

- **Connection Pooling**: pgx with 25 default pool size
- **Query Optimization**: Indexes on user_id, hash, type, created_at
- **Deduplication**: Saves storage via hash-based detection
- **Streaming**: Direct download from provider (low memory)
- **Caching**: JWT validation without DB lookup

## Security

✅ Implemented:
- JWT authentication (HS256, 24h expiration)
- Bcrypt password hashing (cost 10)
- User isolation (all queries filtered by user_id)
- Authorization checks (ownership verification)
- CORS support
- Input validation (struct tags)
- SQL injection prevention (prepared statements)

## Troubleshooting

**Connection refused**
```bash
# Verify PostgreSQL is running
psql -U postgres -d filora
```

**JWT expired**
- Tokens expire after 24 hours, login again

**Provider credentials error**
- Check provider configuration in .env
- Verify API credentials are correct

## Architecture & Design

See [../../docs/architecture.md](../../docs/architecture.md) for:
- System design
- Database schema
- API design principles
- Deployment architecture

See [../../AGENTS.md](../../AGENTS.md) for development guidelines and coding standards.

## License

Proprietary - Filora Digital Asset Management Platform
