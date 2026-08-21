.PHONY: build lint fmt vet check clean sqlc run worker

# --- Build ---
build:
	go build -o bin/server cmd/server/main.go
	go build -o bin/worker cmd/worker/main.go

run:
	go run cmd/server/main.go

worker:
	go run cmd/worker/main.go

# --- Code quality ---
lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .
	goimports -w -local github.com/rizqyn9/filora-dam .

vet:
	go vet ./...

# Run all checks (use in CI or pre-commit)
check: fmt vet lint build
	@echo "All checks passed."

# --- Database ---
sqlc:
	sqlc generate

seed:
	go run cmd/seed/main.go

# --- Clean ---
clean:
	rm -rf bin/
