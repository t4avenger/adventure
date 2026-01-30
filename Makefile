.PHONY: help test test-js lint lint-js fmt vet build run clean install-tools install-js check

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-tools: ## Install linting and code quality tools
	@echo "Installing code quality tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@GOPATH_BIN=$$(go env GOPATH)/bin; \
	echo ""; \
	echo "Tools installed successfully!"; \
	echo "Installation location: $$GOPATH_BIN"; \
	echo ""; \
	if echo "$$PATH" | grep -q "$$GOPATH_BIN"; then \
		echo "✓ GOPATH/bin is already in your PATH"; \
	else \
		echo "⚠ GOPATH/bin is NOT in your PATH"; \
		echo "Add this to your ~/.bashrc or ~/.zshrc:"; \
		echo "  export PATH=$$PATH:$$GOPATH_BIN"; \
		echo ""; \
		echo "Or run this command now:"; \
		echo "  export PATH=$$PATH:$$GOPATH_BIN"; \
	fi

test: ## Run all tests (with race detection if CGO is available)
	@if command -v gcc > /dev/null 2>&1; then \
		echo "Running tests with race detector..."; \
		CGO_ENABLED=1 go test -v -race -coverprofile=coverage.out ./...; \
	else \
		echo "gcc not found, running tests without race detector..."; \
		go test -v -coverprofile=coverage.out ./...; \
	fi
	@echo ""
	@echo "Coverage report:"
	@go tool cover -func=coverage.out

test-race: ## Run tests with race detector (requires CGO/gcc)
	CGO_ENABLED=1 go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "Coverage report:"
	go tool cover -func=coverage.out

test-short: ## Run tests without race detection (faster)
	go test -v ./...

test-js: ## Run JavaScript unit tests (Jest). Run 'make install-js' first if needed.
	@if ! command -v npm > /dev/null 2>&1; then \
		echo "npm not found; install Node.js and npm to run JS tests"; exit 1; \
	fi; \
	if [ ! -d node_modules ]; then \
		echo "node_modules not found; run 'make install-js' first"; exit 1; \
	fi; \
	npm test

lint-js: ## Run ESLint on static/js. Run 'make install-js' first if needed.
	@if ! command -v npm > /dev/null 2>&1; then \
		echo "npm not found; install Node.js and npm to run JS lint"; exit 1; \
	fi; \
	if [ ! -d node_modules ]; then \
		echo "node_modules not found; run 'make install-js' first"; exit 1; \
	fi; \
	npm run lint

install-js: ## Install JavaScript dependencies (npm install)
	@if command -v npm > /dev/null 2>&1; then \
		npm install; \
		echo "JS dependencies installed. Run 'npm test' for tests, 'npm run lint' for lint."; \
	else \
		echo "npm not found; install Node.js and npm first"; exit 1; \
	fi

lint: ## Run golangci-lint (requires install-tools)
	@GOPATH_BIN=$$(go env GOPATH 2>/dev/null || echo "$$HOME/go")/bin; \
	if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run --out-format=colored-line-number ./...; \
	elif [ -f "$$GOPATH_BIN/golangci-lint" ]; then \
		$$GOPATH_BIN/golangci-lint run --out-format=colored-line-number ./...; \
	elif [ -f "$$HOME/go/bin/golangci-lint" ]; then \
		$$HOME/go/bin/golangci-lint run --out-format=colored-line-number ./...; \
	else \
		echo "golangci-lint not found in PATH or common Go bin locations"; \
		echo "Run: make install-tools"; \
		echo "Then ensure $$HOME/go/bin is in your PATH, or run: export PATH=$$PATH:$$HOME/go/bin"; \
		exit 1; \
	fi

fmt: ## Format code with gofmt and goimports
	go fmt ./...
	@GOPATH_BIN=$$(go env GOPATH 2>/dev/null || echo "$$HOME/go")/bin; \
	if command -v goimports > /dev/null 2>&1; then \
		goimports -w .; \
	elif [ -f "$$GOPATH_BIN/goimports" ]; then \
		$$GOPATH_BIN/goimports -w .; \
	elif [ -f "$$HOME/go/bin/goimports" ]; then \
		$$HOME/go/bin/goimports -w .; \
	else \
		echo "goimports not found in PATH or common Go bin locations (optional, skipping)"; \
	fi

vet: ## Run go vet
	go vet ./...

staticcheck: ## Run staticcheck (requires install-tools)
	@GOPATH_BIN=$$(go env GOPATH 2>/dev/null || echo "$$HOME/go")/bin; \
	if command -v staticcheck > /dev/null 2>&1; then \
		staticcheck ./...; \
	elif [ -f "$$GOPATH_BIN/staticcheck" ]; then \
		$$GOPATH_BIN/staticcheck ./...; \
	elif [ -f "$$HOME/go/bin/staticcheck" ]; then \
		$$HOME/go/bin/staticcheck ./...; \
	else \
		echo "staticcheck not found in PATH or common Go bin locations"; \
		echo "Run: make install-tools"; \
		echo "Then ensure $$HOME/go/bin is in your PATH, or run: export PATH=$$PATH:$$HOME/go/bin"; \
		exit 1; \
	fi

build: ## Build the application
	go build -o bin/adventure cmd/server/main.go

run: ## Run the application
	go run cmd/server/main.go

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out

check: fmt vet lint test install-js test-js lint-js ## Run all checks (Go + JS: format, vet, lint, test)
