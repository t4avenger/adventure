.PHONY: help test lint fmt vet build run clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run all tests
	go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "Coverage report:"
	go tool cover -func=coverage.out

test-short: ## Run tests without race detection (faster)
	go test -v ./...

lint: ## Run golangci-lint
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

fmt: ## Format code with gofmt
	go fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	fi

vet: ## Run go vet
	go vet ./...

staticcheck: ## Run staticcheck
	@if command -v staticcheck > /dev/null; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

build: ## Build the application
	go build -o bin/adventure cmd/server/main.go

run: ## Run the application
	go run cmd/server/main.go

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out

check: fmt vet lint test ## Run all checks (format, vet, lint, test)
