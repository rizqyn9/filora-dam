# Filora DAM API

Go REST API server for Filora Digital Asset Management platform.

## Tech Stack

- **Go** 1.23+
- **Fiber v3** - Web framework
- **sqlc** - Type-safe SQL
- **PostgreSQL** (Neon) - Database
- **pgx/v5** - PostgreSQL driver
- **golang-migrate** - Database migrations
- **validator/v10** - Input validation

## Prerequisites

- Go 1.23 or higher
- PostgreSQL database
- Make (optional)

## Getting Started

### 1. Install dependencies

```bash
go mod download
```

### 2. Install development tools

```bash
make install-tools
```

This installs:
- sqlc (SQL code generator)
- golang-migrate (migration tool)
- golangci-lint (linter)

### 3. Set up environment

```bash
cp .env.example .env
```

Edit `.env` with your configuration:
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for JWT (min 32 characters)
- Storage provider credentials (optional for now)

### 4. Run database migrations

```bash
make migrate-up
```

### 5. Run the server

```bash
make run
```

The API will start on `http://localhost:3000`

## Available Commands

```bash
make help              # Show all available commands
make run               # Run the server
make build             # Build binary
make test              # Run tests
make test-coverage     # Run tests with coverage
make fmt               # Format code
make lint              # Run linter
make clean             # Clean build artifacts
make migrate-up        # Run migrations up
make migrate-down      # Run migrations down
make migrate-create    # Create new migration
make sqlc              # Generate sqlc code
make deps              # Download and tidy dependencies
```

## Project Structure

```
apps/api/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration
│   ├── database/
│   │   ├── db.go                # Database connection
│   │   ├── migrations/          # SQL migrations
│   │   └── queries/             # sqlc query files
│   ├── lib/
│   │   └── response.go          # Response helpers
│   ├── middleware/              # HTTP middleware
│   └── modules/                 # Feature modules
│       ├── account/             # User accounts
│       ├── asset/               # Asset metadata
│       ├── storage/             # Storage & providers
│       └── dashboard/           # Metrics
├── .env.example
├── .gitignore
├── .golangci.yml
├── Makefile
├── sqlc.yaml
├── go.mod
└── go.sum
```

## API Endpoints

### Health Check

```
GET /           # API info
GET /health     # Health check
```

More endpoints will be added as modules are implemented.

## Development Workflow

### Creating a new migration

```bash
make migrate-create name=create_users_table
```

This creates two files:
- `internal/database/migrations/000001_create_users_table.up.sql`
- `internal/database/migrations/000001_create_users_table.down.sql`

### Writing SQL queries

1. Write queries in `internal/database/queries/*.sql`
2. Run `make sqlc` to generate Go code
3. Use generated code in repositories

### Adding a new module

1. Create directory: `internal/modules/{module_name}/`
2. Add files:
   - `handler.go` - HTTP routes
   - `service.go` - Business logic
   - `repository.go` - Database access
   - `models.go` - Data structures

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## Building

```bash
# Build binary
make build

# Run binary
./bin/server
```

## Deployment

### Using Docker (TODO)

```bash
docker build -t filora-api .
docker run -p 3000:3000 --env-file .env filora-api
```

### Environment Variables

Required:
- `PORT` - Server port (default: 3000)
- `ENV` - Environment (development, production, test)
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - JWT secret key (min 32 chars)

Optional (storage providers):
- Cloudinary: `CLOUDINARY_CLOUD_NAME`, `CLOUDINARY_API_KEY`, `CLOUDINARY_API_SECRET`
- ImageKit: `IMAGEKIT_PUBLIC_KEY`, `IMAGEKIT_PRIVATE_KEY`, `IMAGEKIT_URL_ENDPOINT`
- R2: `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET_NAME`, `R2_ENDPOINT`

## Development Phases

See [../../docs/api-development-phases.md](../../docs/api-development-phases.md) for the complete development roadmap.

## Architecture

See [../../docs/architecture.md](../../docs/architecture.md) for architecture documentation.

## Contributing

Follow the guidelines in [../../AGENTS.md](../../AGENTS.md).
