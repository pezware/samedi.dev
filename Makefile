# Samedi Makefile
# Common development tasks

.PHONY: help build test lint fmt clean install install-tools check coverage run

# Variables
BINARY_NAME=samedi
BUILD_DIR=bin
COVERAGE_DIR=coverage
GO=go
GOLANGCI_LINT=golangci-lint
MAIN_PATH=./cmd/samedi

# Default target
.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "Samedi Development Tasks"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@grep -E '^## [a-zA-Z_-]+:.*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = "## "}; {printf "  %-20s %s\n", $$2, $$3}'

## build: Build the samedi binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build for all platforms (macOS, Linux, Windows)
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✓ All binaries built"

## test: Run unit tests
test:
	@echo "Running unit tests..."
	$(GO) test -short -race -timeout=5m ./...
	@echo "✓ Tests passed"

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GO) test -tags=integration -race -timeout=10m ./...
	@echo "✓ Integration tests passed"

## test-e2e: Run end-to-end tests
test-e2e:
	@echo "Running E2E tests..."
	$(GO) test -tags=e2e -timeout=15m ./tests/e2e/...
	@echo "✓ E2E tests passed"

## test-all: Run all tests (unit + integration + e2e)
test-all:
	@echo "Running all tests..."
	$(GO) test -tags="integration e2e" -race -timeout=20m ./...
	@echo "✓ All tests passed"

## coverage: Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test -short -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "✓ Coverage report: $(COVERAGE_DIR)/coverage.html"
	@$(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out | grep total | awk '{print "Total coverage: " $$3}'

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

## lint: Run golangci-lint
lint:
	@echo "Running linters..."
	$(GOLANGCI_LINT) run --fix
	@echo "✓ Linting passed"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	goimports -w .
	@echo "✓ Code formatted"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "✓ Vet passed"

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "✓ All checks passed"

## clean: Remove build artifacts and test caches
clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)
	$(GO) clean -testcache
	$(GO) clean -cache
	@echo "✓ Cleaned"

## install: Install samedi to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(MAIN_PATH)
	@echo "✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

## install-tools: Install development tools at pinned versions
install-tools:
	@echo "Installing development tools..."
	@echo "→ Installing Go tools..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
	$(GO) install golang.org/x/tools/cmd/goimports@v0.29.0
	$(GO) install github.com/securego/gosec/v2/cmd/gosec@v2.21.4
	$(GO) install gotest.tools/gotestsum@v1.12.0
	$(GO) install golang.org/x/vuln/cmd/govulncheck@latest
	$(GO) install github.com/air-verse/air@latest
	@echo "→ Installing pre-commit hooks..."
	@command -v pre-commit >/dev/null 2>&1 || { \
		echo "Installing pre-commit via pip..."; \
		pip3 install pre-commit || pip install pre-commit; \
	}
	pre-commit install
	pre-commit install --hook-type commit-msg
	@echo "✓ All development tools installed"
	@echo ""
	@echo "Installed tools:"
	@echo "  golangci-lint v1.64.8   - Linter"
	@echo "  goimports       v0.29.0 - Import formatter"
	@echo "  gosec           v2.21.4 - Security scanner"
	@echo "  gotestsum       v1.12.0 - Better test output"
	@echo "  govulncheck     latest  - Vulnerability scanner"
	@echo "  air             latest  - Hot reload"
	@echo "  pre-commit      latest  - Git hooks"

## deps: Download and tidy dependencies
deps:
	@echo "Managing dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	$(GO) mod verify
	@echo "✓ Dependencies updated"

## run: Run samedi locally
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

## run-dev: Run with live reload (requires air: go install github.com/cosmtrek/air@latest)
run-dev:
	@echo "Running with live reload..."
	air

## security: Run security checks
security:
	@echo "Running security checks..."
	gosec -quiet ./...
	@echo "✓ Security checks passed"

## vuln: Check for vulnerable dependencies
vuln:
	@echo "Checking for vulnerable dependencies..."
	@command -v govulncheck >/dev/null 2>&1 || { \
		echo "govulncheck not found. Installing..."; \
		$(GO) install golang.org/x/vuln/cmd/govulncheck@latest; \
	}
	govulncheck ./...
	@echo "✓ No vulnerabilities found"

## todo: List TODO comments
todo:
	@echo "TODOs in codebase:"
	@grep -rn "TODO" --include="*.go" . || echo "No TODOs found"

## lines: Count lines of code
lines:
	@echo "Lines of code:"
	@find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t samedi:latest .
	@echo "✓ Docker image built"

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run --rm -it samedi:latest

## proto: Generate protobuf code (if needed in future)
proto:
	@echo "Generating protobuf code..."
	# protoc --go_out=. --go_opt=paths=source_relative \
	#        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	#        api/proto/*.proto
	@echo "✓ Protobuf code generated"

## release: Create a new release (requires goreleaser)
release:
	@echo "Creating release..."
	goreleaser release --clean
	@echo "✓ Release created"

## release-snapshot: Create a snapshot release (no publish)
release-snapshot:
	@echo "Creating snapshot release..."
	goreleaser release --snapshot --clean
	@echo "✓ Snapshot release created"

## pre-commit: Run pre-commit hooks manually
pre-commit:
	@echo "Running pre-commit hooks..."
	pre-commit run --all-files
	@echo "✓ Pre-commit hooks passed"

## update-docs: Update documentation
update-docs:
	@echo "Updating documentation..."
	$(GO) run tools/gendocs/main.go
	@echo "✓ Documentation updated"

## init-secrets: Initialize secrets baseline for detect-secrets
init-secrets:
	@echo "Initializing secrets baseline..."
	detect-secrets scan > .secrets.baseline
	@echo "✓ Secrets baseline created"

# Development workflow shortcuts

## dev: Quick development check (fmt + test)
dev: fmt test
	@echo "✓ Development check passed"

## ci: Run full CI pipeline locally
ci: deps check test-all coverage security vuln
	@echo "✓ CI pipeline passed"

## quick: Quick build and run
quick: build run

# Version information
version:
	@echo "Go version: $(shell $(GO) version)"
	@echo "golangci-lint version: $(shell $(GOLANGCI_LINT) --version 2>/dev/null || echo 'not installed')"
	@echo "pre-commit version: $(shell pre-commit --version 2>/dev/null || echo 'not installed')"
