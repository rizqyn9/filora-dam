# Filora DAM

Multi-cloud Digital Asset Management and backup platform.

## Project Structure

```
filora-dam/
├── apps/
│   ├── api/          # Go REST API server
│   ├── cli/          # Go CLI client (coming soon)
│   └── web/          # React 19 frontend (coming soon)
├── docs/             # Documentation
├── AGENTS.md         # Development guidelines
└── CLAUDE.md         # AI agent instructions
```

## Applications

### API (Backend)

Go REST API server with all business logic.

**Tech Stack**: Go, Fiber v3, sqlc, PostgreSQL

[View API README](apps/api/README.md)

### CLI (Coming Soon)

Command-line client for interacting with the API.

**Tech Stack**: Go, Cobra

### Web (Coming Soon)

React frontend for managing digital assets.

**Tech Stack**: React 19, TypeScript, TanStack Query/Router, Tailwind CSS v4

## Features

- Multi-cloud storage (Cloudinary, ImageKit, Cloudflare R2)
- Asset metadata management
- Automatic storage orchestration
- Deduplication
- Storage quota management
- Search and filtering
- Tags and organization
- Backup support (planned)

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL database
- Node.js 20+ (for web app)

### Quick Start

1. **Clone the repository**

```bash
git clone https://github.com/rizqynugroho9/filora-dam.git
cd filora-dam
```

2. **Set up API**

```bash
cd apps/api
cp .env.example .env
# Edit .env with your configuration
go mod download
make run
```

3. **Set up CLI** (coming soon)

4. **Set up Web** (coming soon)

## Documentation

- [Database Design](docs/database/README.md) - Database schema, ERD, rules, and RBAC
- [AGENTS.md](AGENTS.md) - Development guidelines for all agents

## Development

This project is developed using AI-assisted coding tools (Claude Code, Kiro, Antigravity).

All agents must follow the guidelines in [AGENTS.md](AGENTS.md).

### Core Principles

1. **Build First** - Working software over perfect architecture
2. **Refactor Later** - No abstractions until 2+ implementations exist
3. **Database First** - Schema → Repository → Service → API → UI
4. **Consistency First** - Follow existing patterns

### Tech Philosophy

- **API First** - All business logic in API
- **Thin Clients** - CLI and Web only handle presentation
- **Storage Abstraction** - Providers are implementation details
- **Database as Truth** - PostgreSQL is the source of truth

## Current Status

- [x] Project structure
- [x] API Phase 0: Project setup complete
- [x] Documentation
- [ ] API Phase 1: Core infrastructure (in progress)
- [ ] API Phase 2-12: Feature development
- [ ] CLI development
- [ ] Web development

## License

Private project - All rights reserved

## Support

For issues or questions, see [AGENTS.md](AGENTS.md) for development guidelines.
