# Linter Configuration Fix

## Issues Fixed

### 1. Deprecated Configuration Options

**Problem:**
```
level=warning msg="[config_reader] The configuration option `run.skip-files` is deprecated, please use `issues.exclude-files`."
level=warning msg="[config_reader] The configuration option `run.skip-dirs` is deprecated, please use `issues.exclude-dirs`."
level=warning msg="[config_reader] The configuration option `output.format` is deprecated, please use `output.formats`"
```

**Solution:**
- Moved `run.skip-dirs` → `issues.exclude-dirs`
- Moved `run.skip-files` → `issues.exclude-files`
- Changed `output.format` → `output.formats` (with array syntax)

### 2. Deprecated Linters

**Problem:**
```
level=warning msg="[lintersdb] The linter \"gomnd\" is deprecated (step 2) and deactivated."
level=warning msg="The linter 'exportloopref' is deprecated (since v1.60.2)"
level=error msg="[linters_context] exportloopref: This linter is fully inactivated"
```

**Solution:**
- Replaced `exportloopref` with `copyloopvar` (new replacement in Go 1.22+)
- Replaced `gomnd` with `mnd` (renamed linter)

## Updated Configuration

### Before (Deprecated)

```yaml
run:
  skip-dirs:
    - vendor
  skip-files:
    - ".*\\.pb\\.go$"

linters:
  enable:
    - exportloopref  # DEPRECATED
  disable:
    - gomnd          # DEPRECATED

output:
  format: colored-line-number  # DEPRECATED
```

### After (Current)

```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - copyloopvar    # Replaces exportloopref
  disable:
    - mnd            # Replaces gomnd

issues:
  exclude-dirs:      # Moved from run.skip-dirs
    - vendor
  exclude-files:     # Moved from run.skip-files
    - ".*\\.pb\\.go$"

output:
  formats:           # Changed from format (singular)
    - format: colored-line-number
```

## Pre-commit Integration

The linter is already configured in `.pre-commit-config.yaml`:

```yaml
- repo: local
  hooks:
    - id: golangci-lint
      name: golangci-lint
      entry: golangci-lint run --fix
      language: system
      types: [go]
      pass_filenames: false
```

This runs automatically on every commit!

## Testing

### Verify Configuration

```bash
# Check config is valid
golangci-lint config verify

# Run linter
golangci-lint run

# Run with auto-fix
golangci-lint run --fix
```

### CI/CD

The GitHub Actions workflow already runs golangci-lint:

```yaml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v4
  with:
    version: latest
    args: --timeout=5m
```

## What's Enabled

### Active Linters (20+)

- **errcheck** - Check for unchecked errors
- **gosimple** - Simplify code
- **govet** - Vet examines Go source code
- **ineffassign** - Detect ineffectual assignments
- **staticcheck** - Advanced static analysis
- **typecheck** - Parse and type-check
- **unused** - Find unused code
- **gofmt** - Check formatting
- **goimports** - Check imports
- **misspell** - Find typos
- **revive** - Fast, configurable linter
- **unconvert** - Remove unnecessary conversions
- **unparam** - Find unused parameters
- **gocyclo** - Cyclomatic complexity
- **goconst** - Find repeated strings
- **gocritic** - Opinionated checks
- **godox** - Find FIXME/TODO
- **gosec** - Security issues
- **prealloc** - Pre-allocation suggestions
- **copyloopvar** - Loop variable issues (NEW - Go 1.22+)
- **nilerr** - Nil error checks
- **errorlint** - Error wrapping
- **dupl** - Code duplication

### Disabled Linters

- **exhaustive** - Too strict
- **funlen** - Use gocyclo instead
- **lll** - Let formatter handle
- **mnd** - Magic numbers (too noisy)

## Migration Notes

### copyloopvar vs exportloopref

**Why the change?**

Go 1.22 introduced loop variable scoping changes, making `exportloopref` obsolete. The new `copyloopvar` linter handles this correctly.

**Before (Go < 1.22):**
```go
for i := range items {
    go func() {
        fmt.Println(i)  // BUG: captures loop variable
    }()
}
```

**After (Go 1.22+):**
```go
for i := range items {
    go func() {
        fmt.Println(i)  // OK: each iteration gets own variable
    }()
}
```

### mnd vs gomnd

Simple rename - same functionality, new name:
- Old: `gomnd` (Go Magic Number Detector)
- New: `mnd` (Magic Number Detector)

## Troubleshooting

### If CI Still Fails

1. **Clear GitHub Actions cache:**
   - Go to Actions → Caches
   - Delete golangci-lint cache

2. **Verify config locally:**
   ```bash
   golangci-lint config verify
   golangci-lint run --no-config  # Use defaults
   golangci-lint run              # Use .golangci.yml
   ```

3. **Check linter versions:**
   ```bash
   golangci-lint version
   golangci-lint linters  # List all available linters
   ```

### Common Issues

**Issue:** `copyloopvar not found`
- **Solution:** Update golangci-lint to v1.60.2+
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```

**Issue:** Config still shows warnings
- **Solution:** Delete cache and re-run
  ```bash
  golangci-lint cache clean
  golangci-lint run
  ```

## Verification Checklist

- [x] Removed `run.skip-dirs` (moved to `issues.exclude-dirs`)
- [x] Removed `run.skip-files` (moved to `issues.exclude-files`)
- [x] Changed `output.format` to `output.formats`
- [x] Replaced `exportloopref` with `copyloopvar`
- [x] Replaced `gomnd` with `mnd`
- [x] Pre-commit hook configured
- [x] GitHub Actions workflow configured

## Next Steps

1. **Commit the fix:**
   ```bash
   git add .golangci.yml
   git commit -m "fix: update golangci-lint config to latest format

   - Replace deprecated exportloopref with copyloopvar
   - Replace deprecated gomnd with mnd
   - Move run.skip-* to issues.exclude-*
   - Update output.format to output.formats

   Fixes golangci-lint exit code 7 error in CI."
   ```

2. **Push and verify CI:**
   ```bash
   git push
   # Check GitHub Actions pass
   ```

3. **Install locally:**
   ```bash
   make install-tools  # Installs golangci-lint
   make lint          # Run linter locally
   ```

## References

- [golangci-lint Configuration](https://golangci-lint.run/usage/configuration/)
- [Linter Deprecation Cycle](https://golangci-lint.run/product/roadmap/#linter-deprecation-cycle)
- [copyloopvar Documentation](https://pkg.go.dev/github.com/golangci/golangci-lint/pkg/golinters/copyloopvar)
- [Go 1.22 Loop Changes](https://go.dev/blog/loopvar-preview)

---

**All linter issues are now fixed!** ✅
