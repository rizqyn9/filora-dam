.PHONY: help run run-worker build test fmt lint clean db-apply db-seed sqlc

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the server
	@go run cmd/server/main.go

run-worker: ## Run the archive worker
	@go run cmd/worker/main.go

build: ## Build the binaries (server + worker)
	@go build -o bin/server cmd/server/main.go
	@go build -o bin/worker cmd/worker/main.go

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

db-apply: ## Apply the canonical schema (manual, no migrations)
	@test -f .env && export $$(cat .env | grep -v '^#' | xargs) && \
	psql "$$DATABASE_URL" -f internal/database/schema.sql

db-seed: ## Seed baseline RBAC roles/permissions
	@test -f .env && export $$(cat .env | grep -v '^#' | xargs) && \
	psql "$$DATABASE_URL" -f internal/database/seed.sql

sqlc: ## Generate sqlc code
	@sqlc generate

install-tools: ## Install development tools
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest

deps: ## Download dependencies
	@go mod download
	@go mod tidy

kill: ## Kill port 8000
	lsof -ti:8000 | xargs kill -9 2>/dev/null; echo "done"; lsof -i:8000