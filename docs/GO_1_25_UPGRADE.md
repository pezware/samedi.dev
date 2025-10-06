# Go 1.25.1 Upgrade Notes

This document explains the changes made to support Go 1.25.1 and resolve tooling compatibility issues.

## Issues Resolved

### Issue 1: Tool Installation Failure

**Error:**
```
# golang.org/x/tools/internal/tokeninternal
../../../go/pkg/mod/golang.org/x/tools@v0.25.0/internal/tokeninternal/tokeninternal.go:64:9: invalid array length -delta * delta (constant -256 of type int64)
```

**Root Cause:**
Pinned tool versions (e.g., `golang.org/x/tools@v0.25.0`) were not compatible with Go 1.25.1.

**Solution:**
Changed all tool installations in Makefile to use `@latest` instead of pinned versions:

```makefile
# Before (incompatible with Go 1.25.1)
$(GO) install golang.org/x/tools/cmd/goimports@v0.29.0

# After (compatible)
$(GO) install golang.org/x/tools/cmd/goimports@latest
```

### Issue 2: Import Ordering in tools.go

**Error:**
```
Go code is not formatted:
diff tools/tools.go.orig tools/tools.go
--- tools/tools.go.orig
+++ tools/tools.go
@@ -17,8 +17,8 @@

 import (
     _ "github.com/golangci/golangci-lint/cmd/golangci-lint" // Linter
-    _ "gotest.tools/gotestsum"                               // Better test output
-    _ "golang.org/x/vuln/cmd/govulncheck"                    // Vulnerability scanner
+    _ "golang.org/x/vuln/cmd/govulncheck"                   // Vulnerability scanner
+    _ "gotest.tools/gotestsum"                              // Better test output
 )
```

**Root Cause:**
Imports were not in alphabetical order, and comment alignment was inconsistent.

**Solution:**
- Reordered imports alphabetically
- Ran `gofmt -s -w tools/tools.go` to fix alignment

**Final format:**
```go
import (
    _ "github.com/golangci/golangci-lint/cmd/golangci-lint" // Linter
    _ "golang.org/x/vuln/cmd/govulncheck"                   // Vulnerability scanner
    _ "gotest.tools/gotestsum"                              // Better test output
)
```

## Changes Made

### 1. Version Management

**Two files for Go version:**

**`.tool-versions`** (for local development with asdf)
```
golang 1.25.1
```

**`.go-version`** (for GitHub Actions)
```
1.25.1
```

**Why two files?**
- `asdf` uses `.tool-versions` format: `<tool> <version>`
- `actions/setup-go` expects just the version number
- Both files ensure local and CI use the same Go version (1.25.1)

### 2. Makefile Updates

**Before:**
```makefile
install-tools:
    $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
    $(GO) install golang.org/x/tools/cmd/goimports@v0.29.0
    # ... pinned versions
```

**After:**
```makefile
install-tools:
    @echo "→ Go version: $(shell $(GO) version)"
    $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    $(GO) install golang.org/x/tools/cmd/goimports@latest
    # ... all @latest for Go 1.25+ compatibility
```

**Rationale:**
Using `@latest` ensures tools stay compatible with newer Go versions. The trade-off is less version pinning, but:
- ✅ Always compatible with current Go version
- ✅ Get latest bug fixes and features
- ✅ No manual version updates needed
- ⚠️ May introduce breaking changes (rare for stable tools)

**Mitigation:**
Team should run `make install-tools` together after Go version upgrades to ensure consistency.

### 3. CI/CD Updates

**`.github/workflows/ci.yml`**

**Before:**
```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.21'
    cache: true
```

**After:**
```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version-file: '.go-version'
    cache: true
```

**Benefits:**
- ✅ No hardcoded versions in CI
- ✅ Local and CI use same Go version (1.25.1)
- ✅ `.go-version` automatically read by actions/setup-go
- ✅ `.tool-versions` used by asdf locally

Also removed the Go version matrix (`go: ['1.21', '1.22']`) since we now use a single version from `.tool-versions`.

### 4. Dependency Management

**go.mod and go.sum**

After running `go mod tidy`, tool dependencies are now tracked:

```go
// tools/tools.go imports these packages
require (
    github.com/golangci/golangci-lint v1.64.8
    golang.org/x/vuln v1.1.4
    gotest.tools/gotestsum v1.13.0
)
```

This ensures the team uses consistent tool versions (even though we install `@latest`, the versions are recorded).

## Verification

### Test Installation

```bash
# Should complete without errors
make install-tools
```

Expected output:
```
Installing development tools...
→ Go version: go version go1.25.1 darwin/arm64
→ Installing Go tools...
# ... installations
✓ All development tools installed

Note: Tools are installed at @latest to ensure Go 1.25+ compatibility
Run 'make version' to see installed versions
```

### Test Formatting

```bash
# Should show no output (all files formatted)
gofmt -s -l .
```

### Test Build

```bash
# Should compile successfully
make build
./bin/samedi
```

Expected output:
```
samedi version dev (commit: none, built: unknown)

A learning operating system for the terminal.
```

## Best Practices for Go Version Upgrades

### 1. Coordinate Team Upgrades

When upgrading Go versions:

```bash
# 1. Update .tool-versions
echo "golang X.Y.Z" > .tool-versions

# 2. Install new Go version
asdf install golang X.Y.Z

# 3. Reinstall tools (may get new versions)
make install-tools

# 4. Test everything
make ci

# 5. Commit together: .tool-versions + go.mod + go.sum
git add .tool-versions go.mod go.sum
git commit -m "chore: upgrade to Go X.Y.Z"
```

### 2. CI/CD Will Auto-Update

Because CI reads from `.tool-versions`, it automatically uses the new Go version when you push the commit.

### 3. Check Tool Compatibility

After upgrading:

```bash
# Verify all tools work
make install-tools
make check
make ci
```

## Troubleshooting

### Problem: Tool still fails to install

```bash
# Clear Go module cache
go clean -modcache

# Reinstall tools
make install-tools
```

### Problem: Import formatting issues

```bash
# Auto-fix all files
gofmt -s -w .

# Or use make target
make fmt
```

### Problem: CI fails with different Go version

```bash
# Verify both version files are committed
git status .tool-versions .go-version

# Both should be committed
```

### Problem: GitHub Actions can't find Go 1.25.1

```bash
# Error: Unable to find Go version 'golang 1.25.1'
# This happens if CI uses .tool-versions instead of .go-version

# Fix: Ensure CI workflow uses .go-version
grep "go-version-file" .github/workflows/ci.yml
# Should show: go-version-file: '.go-version'
```

### Problem: Team has inconsistent tool versions

```bash
# Everyone run this after pulling
make install-tools

# Check installed versions
make version
```

## Summary

**What changed:**
1. ✅ `.tool-versions` → golang 1.25.1 (for asdf local)
2. ✅ `.go-version` → 1.25.1 (for GitHub Actions)
3. ✅ Makefile → use `@latest` for most tools, pinned golangci-lint v1.64.8
4. ✅ CI workflow → read Go version from `.go-version`
5. ✅ `tools/tools.go` → fixed import ordering and formatting
6. ✅ `go.mod`/`go.sum` → added tool dependencies

**What to remember:**
- `make install-tools` installs tools compatible with current Go version
- `.tool-versions` for local (asdf), `.go-version` for CI (GitHub Actions)
- Keep both files in sync when upgrading Go
- CI automatically uses the Go version from `.go-version`
- Run `gofmt -s -w .` before committing to fix formatting

**Next steps:**
```bash
# Install tools with new Go version
make install-tools

# Verify everything works
make ci

# Commit changes
git add .go-version .github/workflows/ci.yml Makefile go.mod go.sum tools/tools.go
git commit -m "fix: update tooling for Go 1.25.1 compatibility"
```
