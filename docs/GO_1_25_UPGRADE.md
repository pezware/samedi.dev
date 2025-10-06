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

**`.tool-versions`** (already updated by user)
```
golang 1.25.1
```

This file is used by `asdf` to automatically switch Go versions per project.

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
    go-version-file: '.tool-versions'
    cache: true
```

**Benefits:**
- ✅ Single source of truth (`.tool-versions`)
- ✅ No hardcoded versions in CI
- ✅ Local and CI always use same Go version
- ✅ Automatically updates when `.tool-versions` changes

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
# Verify .tool-versions is committed
git status .tool-versions

# Should show:
# On branch main
# nothing to commit, working tree clean
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
1. ✅ `.tool-versions` → go 1.25.1 (user updated)
2. ✅ Makefile → use `@latest` for all tools
3. ✅ CI workflow → read Go version from `.tool-versions`
4. ✅ `tools/tools.go` → fixed import ordering and formatting
5. ✅ `go.mod`/`go.sum` → added tool dependencies

**What to remember:**
- `make install-tools` installs tools compatible with current Go version
- `.tool-versions` is the single source of truth for Go version
- CI automatically uses the Go version from `.tool-versions`
- Run `gofmt -s -w .` before committing to fix formatting

**Next steps:**
```bash
# Install tools with new Go version
make install-tools

# Verify everything works
make ci

# Commit changes
git add .github/workflows/ci.yml Makefile go.mod go.sum tools/tools.go
git commit -m "fix: update tooling for Go 1.25.1 compatibility"
```
