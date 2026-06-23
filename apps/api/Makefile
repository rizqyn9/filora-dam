.PHONY: help run build test fmt lint clean migrate-up migrate-down migrate-create sqlc

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the server
	@go run cmd/server/main.go

build: ## Build the binary
	@go build -o bin/server cmd/server/main.go

test: ## Run tests
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

fmt: ## Format code
	@go fmt ./...

lint: ## Run linter
	@golangci-lint run

clean: ## Clean build artifacts
	@rm -rf bin/
	@rm -f coverage.out

migrate-up: ## Run database migrations up
	@migrate -path internal/database/migrations -database "$(DATABASE_URL)" up

migrate-down: ## Run database migrations down
	@migrate -path internal/database/migrations -database "$(DATABASE_URL)" down

migrate-create: ## Create a new migration (use: make migrate-create name=create_users)
	@migrate create -ext sql -dir internal/database/migrations -seq $(name)

sqlc: ## Generate sqlc code
	@sqlc generate

install-tools: ## Install development tools
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest

deps: ## Download dependencies
	@go mod download
	@go mod tidy
