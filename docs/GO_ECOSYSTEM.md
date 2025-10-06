# Go Development Ecosystem Guide

This guide covers the contemporary Go development ecosystem and tooling practices for the Samedi project.

## Table of Contents

1. [Go Version Management](#go-version-management)
2. [Dependency Management](#dependency-management)
3. [Tool Management](#tool-management)
4. [Essential Development Tools](#essential-development-tools)
5. [Testing Ecosystem](#testing-ecosystem)
6. [Code Quality Tools](#code-quality-tools)
7. [Security Tools](#security-tools)
8. [Development Workflow](#development-workflow)

---

## Go Version Management

### Option 1: asdf (Recommended) ⭐

**Best for:** Multi-language projects (Go, Node, Python, Ruby, etc.)

```bash
# Install asdf
brew install asdf

# Add to shell (already done if using .zshrc/.bashrc)
echo -e "\n. $(brew --prefix asdf)/libexec/asdf.sh" >> ~/.zshrc

# Install Go plugin
asdf plugin add golang

# Install specific version
asdf install golang 1.23.5

# Set global default
asdf global golang 1.23.5

# Or set project-specific (creates .tool-versions)
asdf local golang 1.23.5
```

**Why asdf?**
- ✅ Single tool for all languages
- ✅ `.tool-versions` file ensures team consistency
- ✅ Automatic version switching per directory
- ✅ No shell modifications needed
- ✅ Works with direnv for environment variables

**Project Setup:**
```bash
# Already created in samedi.dev
cat .tool-versions
# golang 1.23.5

# When entering directory, asdf auto-switches Go version
cd samedi.dev
go version  # Uses 1.23.5
```

### Option 2: gvm (Go Version Manager)

**Best for:** Go-only projects

```bash
# Install gvm
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)

# Install and use Go version
gvm install go1.23.5
gvm use go1.23.5 --default
```

### Option 3: Official Installer

**Best for:** Single Go version, simple setup

```bash
# macOS
brew install go

# Or download from golang.org
# https://go.dev/dl/

# Check version
go version
```

### Option 4: Docker (Isolation)

**Best for:** CI/CD, reproducible builds

```dockerfile
FROM golang:1.23.5-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /samedi ./cmd/samedi
CMD ["/samedi"]
```

---

## Dependency Management

Go modules are **built-in** since Go 1.11 (no external tool needed).

### Key Commands

```bash
# Initialize module (already done)
go mod init github.com/pezware/samedi.dev

# Add dependency (automatically adds to go.mod)
go get github.com/spf13/cobra@v1.8.1

# Upgrade dependency
go get -u github.com/spf13/cobra

# Upgrade all dependencies
go get -u ./...

# Remove unused dependencies
go mod tidy

# Verify dependencies (checksums)
go mod verify

# Download dependencies (for CI)
go mod download

# Vendor dependencies (optional)
go mod vendor

# Show dependency graph
go mod graph

# Explain why package is needed
go mod why github.com/spf13/cobra
```

### Key Files

- **`go.mod`** - Direct dependencies (commit to git)
- **`go.sum`** - Cryptographic checksums for security (commit to git)
- **`vendor/`** - Optional vendored deps (usually .gitignore)

### Best Practices

**DO:**
- ✅ Commit `go.mod` and `go.sum`
- ✅ Run `go mod tidy` before committing
- ✅ Pin to specific versions for stability
- ✅ Use `go mod verify` in CI/CD

**DON'T:**
- ❌ Edit `go.mod` manually (use `go get`)
- ❌ Delete `go.sum` (security!)
- ❌ Use `go get -u` without testing

---

## Tool Management

### The `tools.go` Pattern ⭐

**Problem:** How to pin development tool versions (golangci-lint, mockgen, etc.)?

**Solution:** Use a build-tagged Go file to declare tool dependencies.

**File: `tools/tools.go`**

```go
// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

//go:build tools
// +build tools

package tools

// This file ensures tool dependencies are tracked in go.mod.
// Import tools here to pin their versions.

import (
    _ "github.com/golangci/golangci-lint/cmd/golangci-lint"
    _ "gotest.tools/gotestsum"
    _ "golang.org/x/vuln/cmd/govulncheck"
)
```

**Install tools:**

```bash
# Install all tools at once
make install-tools

# Or manually
go install $(go list -f '{{.ImportPath}}' -tags=tools ./tools)
```

**Why this pattern?**
- ✅ Tools are versioned in `go.mod`
- ✅ Team uses same tool versions
- ✅ Reproducible across environments
- ✅ CI/CD uses exact versions

### Makefile Tool Management

See `Makefile` for pinned versions:

```makefile
install-tools:
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
    go install gotest.tools/gotestsum@v1.12.0
    # ... more tools
```

---

## Essential Development Tools

### 1. Code Quality

#### golangci-lint (Already Configured)

**The meta-linter** - runs 20+ linters in one command.

```bash
# Run all enabled linters
golangci-lint run

# Run with auto-fix
golangci-lint run --fix

# Check specific files
golangci-lint run ./pkg/...

# See enabled linters
golangci-lint linters

# Verify configuration
golangci-lint config verify
```

**Configuration:** `.golangci.yml` (already set up with 20+ linters)

#### goimports

**Import formatter** - like `gofmt` but handles imports.

```bash
# Format all files
goimports -w .

# Check without modifying
goimports -l .
```

### 2. Testing Tools

#### gotestsum ⭐

**Better test output** - replaces `go test` with nicer formatting.

```bash
# Install
go install gotest.tools/gotestsum@latest

# Run tests with better output
gotestsum --format testname

# With coverage
gotestsum --format testname -- -coverprofile=coverage.out ./...

# Watch mode (re-run on file changes)
gotestsum --watch
```

#### testify (Already in go.mod)

**Assertion library** - makes tests more readable.

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
)

func TestSomething(t *testing.T) {
    // assert continues on failure
    assert.Equal(t, expected, actual, "should be equal")
    assert.NoError(t, err)

    // require stops on failure
    require.NotNil(t, obj)
    require.NoError(t, err)
}
```

#### mockery

**Mock generator** - auto-generates mocks from interfaces.

```bash
# Install
go install github.com/vektra/mockery/v2@latest

# Generate mocks
mockery --name=LLMProvider --dir=pkg/llm --output=pkg/llm/mocks

# Or use go:generate
//go:generate mockery --name=LLMProvider
```

### 3. Live Reload

#### air ⭐

**Hot reload** - automatically rebuilds on file changes.

```bash
# Install
go install github.com/air-verse/air@latest

# Run with hot reload
air

# Or via Makefile
make run-dev
```

**Configuration:** `.air.toml` (create if needed)

```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ./cmd/samedi"
  bin = "tmp/main"
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  delay = 1000
```

### 4. Code Generation

#### go generate

**Built-in** - runs code generators.

```bash
# Run all go:generate directives
go generate ./...

# In specific package
go generate ./pkg/models
```

**Example usage:**

```go
//go:generate mockery --name=Repository
//go:generate stringer -type=Status
//go:generate go-enum --file=$GOFILE
```

---

## Security Tools

### 1. gosec (Security Scanner)

```bash
# Run security audit
gosec ./...

# With verbose output
gosec -fmt=json -out=results.json ./...

# Or via Makefile
make security
```

### 2. govulncheck (Vulnerability Scanner) ⭐

```bash
# Check for known vulnerabilities
govulncheck ./...

# Or via Makefile
make vuln
```

### 3. detect-secrets (Pre-commit Hook)

Already configured in `.pre-commit-config.yaml`.

```bash
# Scan for secrets
detect-secrets scan

# Audit findings
detect-secrets audit .secrets.baseline

# Run via pre-commit
pre-commit run detect-secrets --all-files
```

---

## Development Workflow

### Daily Development

```bash
# 1. Pull latest changes
git pull origin main

# 2. Create feature branch
git checkout -b feat/my-feature

# 3. Install/update tools (if needed)
make install-tools

# 4. Write tests (TDD)
# ... write test in pkg/foo/bar_test.go ...
gotestsum --format testname ./pkg/foo

# 5. Implement feature
# ... write code in pkg/foo/bar.go ...
make test

# 6. Run all checks
make check  # runs: fmt, vet, lint, test

# 7. Commit (pre-commit hooks run automatically)
git add .
git commit -m "feat(foo): add bar feature"

# 8. Push
git push origin feat/my-feature
```

### Pre-commit Hooks

Automatically run before every commit:

1. **Trailing whitespace** - Removes trailing spaces
2. **End of file fixer** - Ensures newline at EOF
3. **YAML validator** - Checks YAML syntax
4. **gofmt** - Formats Go code
5. **goimports** - Formats imports
6. **golangci-lint** - Runs all linters
7. **detect-secrets** - Scans for secrets
8. **commit message** - Validates format

**Run manually:**
```bash
pre-commit run --all-files
```

**Skip hooks** (NOT RECOMMENDED):
```bash
git commit --no-verify
```

### CI/CD Pipeline

**GitHub Actions:** `.github/workflows/ci.yml`

Runs on every push and PR:

1. **Multi-platform builds** (Linux, macOS, Windows)
2. **Unit tests** with race detector
3. **Integration tests**
4. **golangci-lint**
5. **gosec** security scan
6. **Code coverage** (Codecov)
7. **Dependency check**

**Run CI locally:**
```bash
make ci
```

---

## Quick Reference

### Common Commands

```bash
# Build
make build                  # Build binary
make build-all              # Build for all platforms

# Test
make test                   # Unit tests
make test-integration       # Integration tests
make test-e2e               # End-to-end tests
make test-all               # All tests
make coverage               # Coverage report

# Quality
make fmt                    # Format code
make lint                   # Run linters
make vet                    # Run go vet
make check                  # All checks (fmt + vet + lint + test)

# Security
make security               # Security scan (gosec)
make vuln                   # Vulnerability check (govulncheck)

# Tools
make install-tools          # Install all dev tools
make deps                   # Update dependencies

# Run
make run                    # Build and run
make run-dev                # Run with hot reload (air)

# CI
make ci                     # Run full CI pipeline locally
```

### Useful Go Commands

```bash
# Build
go build ./cmd/samedi                    # Build main package
go build -o bin/samedi ./cmd/samedi      # Build with output path
go install ./cmd/samedi                  # Install to $GOPATH/bin

# Test
go test ./...                            # All tests
go test -short ./...                     # Skip slow tests
go test -race ./...                      # With race detector
go test -cover ./...                     # With coverage
go test -bench=. ./...                   # Benchmarks
go test -run TestSpecific ./pkg/foo      # Specific test

# Modules
go mod tidy                              # Clean up dependencies
go mod download                          # Download dependencies
go mod verify                            # Verify checksums
go mod why github.com/pkg/foo            # Why is package needed
go mod graph                             # Dependency graph

# Tools
go install github.com/pkg/foo@latest     # Install tool
go install github.com/pkg/foo@v1.2.3     # Install specific version
go list -m all                           # List all dependencies
go list -m -u all                        # Check for updates

# Misc
go fmt ./...                             # Format code
go vet ./...                             # Static analysis
go clean -cache                          # Clear build cache
go env                                   # Show Go environment
go version                               # Show Go version
```

### File Structure

```
samedi.dev/
├── .tool-versions           # asdf version pinning (golang 1.23.5)
├── go.mod                   # Direct dependencies
├── go.sum                   # Dependency checksums
├── tools/
│   └── tools.go             # Tool dependencies
├── cmd/
│   └── samedi/
│       └── main.go          # Main entry point
├── pkg/                     # Public packages
├── internal/                # Private packages
└── Makefile                 # Common tasks
```

---

## Resources

### Official Documentation
- [Go Documentation](https://go.dev/doc/)
- [Go Modules Reference](https://go.dev/ref/mod)
- [Effective Go](https://go.dev/doc/effective_go)

### Tools
- [golangci-lint](https://golangci-lint.run/)
- [asdf](https://asdf-vm.com/)
- [testify](https://github.com/stretchr/testify)
- [air](https://github.com/air-verse/air)
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)

### Best Practices
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

---

## Summary

**For Samedi Development:**

1. **Version Control:** Use `asdf` with `.tool-versions` (golang 1.23.5)
2. **Dependencies:** Use `go mod` (go.mod + go.sum)
3. **Tools:** Pin versions in `tools/tools.go` + Makefile
4. **Quality:** Pre-commit hooks + golangci-lint + gotestsum
5. **Security:** gosec + govulncheck + detect-secrets
6. **Workflow:** TDD → make check → commit → CI/CD

**One Command Setup:**
```bash
git clone https://github.com/pezware/samedi.dev
cd samedi.dev
make install-tools  # Installs everything
make check          # Verifies setup
```

**Daily Commands:**
```bash
make test      # Run tests
make check     # All quality checks
make run-dev   # Run with hot reload
```
