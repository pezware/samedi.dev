# golangci-lint v2 Compatibility Issue

## Problem

When upgrading to Go 1.25.1 and using `@latest` for tool installation, golangci-lint upgraded from v1.64.8 to v2.5.0, which has breaking changes in configuration format.

## Error Messages

### Initial Error
```
Error: can't load config: can't unmarshal config by viper (flags, file): 1 error(s) decoding:

* 'output.formats' expected a map, got 'slice'
```

### After Adding version: "2"
```
jsonschema: "issues" does not validate with "/properties/issues/additionalProperties": additional properties 'exclude-dirs', 'exclude-files', 'exclude-rules' not allowed
jsonschema: "output" does not validate with "/properties/output/additionalProperties": additional properties 'print-issued-lines', 'print-linter-name', 'sort-results', 'format' not allowed
jsonschema: "" does not validate with "/additionalProperties": additional properties 'linters-settings' not allowed
```

## Root Cause

golangci-lint v2.0+ introduced breaking changes to the configuration schema:

**v1.x (supported):**
- `linters-settings` - Configures individual linters
- `issues.exclude-dirs`, `issues.exclude-files`, `issues.exclude-rules`
- `output.format`, `output.print-issued-lines`, etc.

**v2.x (breaking changes):**
- Requires `version: "2"` field
- Different schema for `issues`, `output`, and `linters-settings`
- Many options removed or restructured
- `typecheck` is built-in and cannot be enabled/disabled

## Solution

**Pin golangci-lint to v1.64.8** instead of using `@latest`.

### Changes Made

#### 1. Makefile
```makefile
# Pin golangci-lint to v1.64.8 (stable, compatible)
$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

# Use GOPATH version explicitly to avoid conflicts
GOPATH_BIN=$(shell go env GOPATH)/bin
GOLANGCI_LINT=$(GOPATH_BIN)/golangci-lint
```

**Why pin instead of @latest?**
- ✅ Stable, predictable configuration
- ✅ Team uses same version
- ✅ No surprise breaking changes
- ✅ Compatible with existing config
- ⚠️ Need to manually upgrade (but worth it for stability)

#### 2. .golangci.yml
```yaml
# Removed problematic output section (uses defaults)
# Removed: version field (not needed for v1.x)
```

The `output` section was removed because:
- Not essential for linting functionality
- Defaults work fine (colored output to terminal)
- Can be overridden via command-line flags if needed

### Path Conflict Issue

**Problem:** Homebrew installed golangci-lint v2.5.0 at `/opt/homebrew/bin/golangci-lint`, which took precedence over our installed v1.64.8 at `~/go/bin/golangci-lint`.

**Solution:** Makefile now explicitly uses `$(GOPATH)/bin/golangci-lint`.

**For users:** Add to your shell profile:
```bash
# Prioritize GOPATH binaries
export PATH="$(go env GOPATH)/bin:$PATH"
```

Or uninstall Homebrew version:
```bash
brew uninstall golangci-lint
```

## Verification

```bash
# Check installed version
$(go env GOPATH)/bin/golangci-lint version
# Should show: v1.64.8

# Verify config
$(go env GOPATH)/bin/golangci-lint config verify
# Should succeed with no errors

# Run checks
make check
# Should pass: fmt, vet, lint, test
```

## When to Upgrade to v2

golangci-lint v2 will eventually be necessary, but requires:

1. **Migration effort:**
   - Rewrite `.golangci.yml` for v2 schema
   - Test all linter configurations
   - Update team documentation

2. **Stability assessment:**
   - Wait for v2.x to mature
   - Check community feedback
   - Ensure critical linters are supported

3. **Coordinated upgrade:**
   - Plan team-wide upgrade
   - Update CI/CD simultaneously
   - Provide migration guide

**For now:** v1.64.8 is stable, well-documented, and sufficient for our needs.

## Lessons Learned

### 1. Don't Use @latest for Critical Tools

**Bad:**
```makefile
$(GO) install github.com/tool/cmd/tool@latest
```

**Good:**
```makefile
$(GO) install github.com/tool/cmd/tool@v1.2.3
```

**Exception:** Tools with stable APIs (goimports, gotestsum) can use @latest safely.

### 2. Pin Major Version Numbers

```makefile
# Development tools - pinned versions
golangci-lint@v1.64.8   # Pin to avoid v2 breaking changes
gosec@v2.21.4           # Pin major version
goimports@latest        # Safe - stable API
```

### 3. Test Locally Before CI

```bash
# Always test locally first
make install-tools
make check

# Then commit and push
git commit -m "fix: update tools"
```

### 4. Document Version Requirements

In `CLAUDE.md` or `CONTRIBUTING.md`:
```markdown
## Required Tool Versions

- golangci-lint: v1.64.8 (not v2.x - breaking changes)
- Go: 1.25.1+ (see .tool-versions)
- Install: `make install-tools`
```

## Summary

**Problem:** golangci-lint v2 broke our configuration
**Solution:** Pin to v1.64.8, use GOPATH version explicitly
**Result:** `make check` passes, stable tooling

**Files changed:**
- `Makefile` - Pin golangci-lint to v1.64.8, use GOPATH explicitly
- `.golangci.yml` - Remove problematic `output` section
- Keep using v1.x config format (proven, stable)

**Recommendation:** Stay on v1.x until v2 matures and we have time for proper migration.
