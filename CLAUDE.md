# Samedi Development Guide

## Project Philosophy

**Build a learning operating system that developers love to use.**

- **CLI-first**: Terminal is the primary interface
- **LLM-powered**: Delegate intelligence to Claude/Codex
- **Local-first**: Works offline, syncs optionally
- **Boring tech**: Proven tools over new frameworks
- **Test-driven**: Tests before features

## ⚠️ CRITICAL: Security Rules

**NEVER commit to the repository:**

- ❌ Passwords or passphrases
- ❌ API keys or tokens (OpenAI, Anthropic, etc.)
- ❌ Private keys or certificates
- ❌ Database credentials
- ❌ OAuth secrets
- ❌ Any `.env` files with real credentials

**ALWAYS:**

- ✅ Use environment variables for secrets
- ✅ Reference env vars in config files (e.g., `api_key_env = "OPENAI_API_KEY"`)
- ✅ Add sensitive file patterns to `.gitignore`
- ✅ Use `.env.example` with placeholder values
- ✅ Run `detect-secrets` before committing (pre-commit hook does this)

**If you accidentally commit a secret:**

1. **DO NOT** just delete it in a new commit (it's still in history!)
2. **Immediately** rotate/revoke the credential
3. Use `git filter-branch` or BFG Repo-Cleaner to remove from history
4. Force push (coordinate with team)
5. Report in #security channel

**Pre-commit hooks will catch most secrets, but you are the final line of defense.**

## Development Flow

### 1. Plan → Test → Implement → Document

```
1. Write spec/design doc (if major feature)
2. Write failing test (red)
3. Implement minimal code to pass (green)
4. Refactor and document (clean)
5. Commit with clear message
6. PR and review
```

### 2. Incremental Development

- **Small commits**: Each commit compiles and passes tests
- **Small PRs**: < 400 lines, single responsibility
- **Iterative**: Ship MVP → gather feedback → improve

### 3. Quality Gates

Every commit must:

- ✅ Pass `go test ./...`
- ✅ Pass `golangci-lint run`
- ✅ Pass `go vet ./...`
- ✅ Be formatted with `gofmt`
- ✅ Have meaningful commit message

## Code Standards

### License Headers

**All new Go files must include the MIT license header:**

```go
// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package yourpackage
```

Template available in `.license-header.txt`.

**Why?**

- Clarifies copyright and licensing
- SPDX identifier enables automated license scanning
- Industry standard for open source projects

### Go Style Guide

Follow [Uber's Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) with these additions:

**Package Organization**:

```go
// internal/plan/manager.go
package plan

import (
    "context"
    "fmt"

    "github.com/pezware/samedi.dev/internal/storage"  // Internal imports
    "github.com/pezware/samedi.dev/pkg/markdown"      // Public packages

    "github.com/spf13/cobra"                          // External imports (alphabetical)
)
```

**Error Handling**:

```go
// ✅ Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create plan %s: %w", planID, err)
}

// ❌ Bad: Lose context
if err != nil {
    return err
}
```

**Function Documentation**:

```go
// CreatePlan generates a new learning plan using the configured LLM.
// It validates the topic, calls the LLM provider, and saves the result
// to both filesystem and database.
//
// Returns the created plan or an error if:
//   - Topic is empty
//   - LLM call fails
//   - Storage operation fails
func CreatePlan(ctx context.Context, topic string, hours int) (*Plan, error) {
    // ...
}
```

**Struct Documentation**:

```go
// Plan represents a learning curriculum broken into time-boxed chunks.
type Plan struct {
    ID         string    `json:"id"`
    Title      string    `json:"title"`
    TotalHours int       `json:"total_hours"`
    Chunks     []Chunk   `json:"chunks"`
    CreatedAt  time.Time `json:"created_at"`
}
```

**When to Document**:

- ✅ Public functions/types (exported)
- ✅ Complex algorithms (e.g., SM-2)
- ✅ Non-obvious decisions (why, not what)
- ❌ Obvious getters/setters
- ❌ Self-explanatory code

### Testing Standards

**Test Organization**:

```
internal/plan/
├── manager.go
├── manager_test.go          # Unit tests
├── parser.go
├── parser_test.go
└── testdata/                # Test fixtures
    ├── valid_plan.md
    └── invalid_plan.md
```

**Test Naming**:

```go
// Format: Test<Function>_<Scenario>_<ExpectedResult>
func TestCreatePlan_ValidTopic_ReturnsPlan(t *testing.T) { }
func TestCreatePlan_EmptyTopic_ReturnsError(t *testing.T) { }
func TestCreatePlan_LLMFailure_ReturnsError(t *testing.T) { }
```

**Table-Driven Tests**:

```go
func TestParsePlan(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Plan
        wantErr bool
    }{
        {
            name:  "valid plan with frontmatter",
            input: "testdata/valid_plan.md",
            want:  &Plan{ID: "rust-async", ...},
        },
        {
            name:    "missing frontmatter",
            input:   "testdata/no_frontmatter.md",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParsePlan(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

**Test Coverage Goals**:

- **Unit tests**: 80%+ coverage for core logic
- **Integration tests**: Critical paths (plan generation, sync, review)
- **E2E tests**: Happy path for each command

**What to Test**:

- ✅ Business logic (plan creation, session tracking, SM-2)
- ✅ Error cases (invalid input, LLM failures)
- ✅ Edge cases (empty plans, extremely long sessions)
- ✅ Data transformations (markdown parsing, JSON serialization)
- ❌ External libraries (trust `cobra`, `bubble tea`)
- ❌ Trivial code (simple getters)

### Commit Messages

**🔒 SECURITY CHECK BEFORE EVERY COMMIT:**

- Review `git diff` - any API keys, passwords, or secrets?
- Check `.env` files are in `.gitignore`
- Verify no hardcoded credentials in test files
- Pre-commit hooks will scan, but **you are the final check**

**Format** (Conventional Commits):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding or updating tests
- `chore`: Build process, dependencies, tooling

**Examples**:

```
feat(plan): add LLM-powered plan generation

Implement plan generation flow that calls configured LLM provider,
parses markdown output, and saves to filesystem + database.

- Add PlanService.Generate() method
- Add LLM provider interface and Claude implementation
- Add markdown parser with frontmatter support
- Add integration test with mock LLM

Closes #12

---

fix(session): prevent duplicate active sessions

Check for existing active session before creating new one.
Return clear error message with suggestion to stop current session.

Fixes #45

---

test(flashcard): add SM-2 algorithm unit tests

Cover all rating scenarios (1-4) and edge cases:
- First review interval calculation
- Ease factor boundaries (min 1.3)
- Interval multiplication on success
- Reset on failure

---

refactor(storage): extract SQLite operations to repository

Move database logic from services to dedicated repository layer
for better testability and separation of concerns.

---

chore: add pre-commit hooks for linting

Install pre-commit framework and configure:
- gofmt
- golangci-lint
- go vet
- unit test runner
```

**Commit Hygiene**:

- One logical change per commit
- Commits should compile and pass tests
- Use `git commit --amend` to fix mistakes before pushing
- Squash "WIP" commits before merging

## Testing Strategy

### Unit Tests (Standard `testing` + `testify`)

**Setup**:

```go
// internal/plan/manager_test.go
import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

func TestPlanManager_Create(t *testing.T) {
    // Setup
    mockLLM := new(MockLLMProvider)
    mockLLM.On("Call", mock.Anything).Return("# Plan...", nil)

    svc := NewPlanService(mockLLM, mockDB)

    // Execute
    plan, err := svc.Create("rust-async", 40)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "rust-async", plan.ID)
    assert.Len(t, plan.Chunks, 40)
    mockLLM.AssertExpectations(t)
}
```

**Mocking with testify**:

```go
// internal/llm/mock.go (generated or manual)
type MockLLMProvider struct {
    mock.Mock
}

func (m *MockLLMProvider) Call(prompt string) (string, error) {
    args := m.Called(prompt)
    return args.String(0), args.Error(1)
}
```

### Integration Tests (`dockertest` for real dependencies)

**Setup**:

```go
// internal/storage/integration_test.go
// +build integration

import (
    "testing"

    "github.com/ory/dockertest/v3"
)

func TestSQLiteRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup: Real SQLite DB
    db, cleanup := setupTestDB(t)
    defer cleanup()

    repo := NewSQLiteRepository(db)

    // Test real database operations
    plan := &Plan{ID: "test", Title: "Test Plan"}
    err := repo.CreatePlan(plan)

    require.NoError(t, err)

    // Verify
    retrieved, err := repo.GetPlan("test")
    require.NoError(t, err)
    assert.Equal(t, plan.Title, retrieved.Title)
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)

    // Run migrations
    err = runMigrations(db)
    require.NoError(t, err)

    cleanup := func() {
        db.Close()
    }

    return db, cleanup
}
```

**Run integration tests**:

```bash
go test -tags=integration ./...
```

### E2E Tests (Real CLI execution)

**Setup**:

```go
// tests/e2e/plan_test.go
// +build e2e

func TestE2E_PlanCreation(t *testing.T) {
    // Setup: Temp directory for test
    tmpDir := t.TempDir()
    os.Setenv("SAMEDI_DATA_DIR", tmpDir)

    // Execute CLI command
    cmd := exec.Command("samedi", "init", "test-plan", "--hours", "10")
    output, err := cmd.CombinedOutput()

    require.NoError(t, err)
    assert.Contains(t, string(output), "Plan created")

    // Verify: Check filesystem
    planPath := filepath.Join(tmpDir, "plans", "test-plan.md")
    assert.FileExists(t, planPath)

    // Verify: Check database
    db := openTestDB(t, tmpDir)
    defer db.Close()

    var count int
    db.QueryRow("SELECT COUNT(*) FROM plans WHERE id = ?", "test-plan").Scan(&count)
    assert.Equal(t, 1, count)
}
```

**Run e2e tests**:

```bash
go test -tags=e2e ./tests/e2e/...
```

### Golden Files (Snapshot testing)

**For complex outputs**:

```go
// internal/stats/reporter_test.go
import "github.com/sebdah/goldie/v2"

func TestReporter_GenerateMarkdown(t *testing.T) {
    g := goldie.New(t)

    reporter := NewReporter(testData)
    output := reporter.GenerateMarkdown()

    // Compare against golden file
    g.Assert(t, "report_full", []byte(output))
}
```

**Update golden files**:

```bash
go test ./... -update
```

### Test Helpers

**Common test utilities**:

```go
// internal/testutil/fixtures.go
package testutil

func NewTestPlan(t *testing.T) *plan.Plan {
    return &plan.Plan{
        ID:    "test-plan",
        Title: "Test Plan",
        Chunks: []plan.Chunk{
            {ID: "chunk-001", Duration: 60},
        },
        CreatedAt: time.Now(),
    }
}

func NewTestDB(t *testing.T) (*sql.DB, func()) {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)

    return db, func() { db.Close() }
}

// Golden file helper
func LoadFixture(t *testing.T, name string) string {
    data, err := os.ReadFile(filepath.Join("testdata", name))
    require.NoError(t, err)
    return string(data)
}
```

## Development Workflow

### Daily Development

```bash
# 1. Start new feature/fix
git checkout -b feat/flashcard-review

# 2. Write failing test
vim internal/flashcard/review_test.go
go test ./internal/flashcard/...  # Should fail

# 3. Implement
vim internal/flashcard/review.go
go test ./internal/flashcard/...  # Should pass

# 4. Run all checks (pre-commit will do this automatically)
make check

# 5. Commit
git add .
git commit -m "feat(flashcard): implement review interface"

# 6. Push and create PR
git push origin feat/flashcard-review
gh pr create
```

### Pre-Commit Checks

These run automatically via git hooks:

```bash
make check
# Runs:
# - gofmt
# - golangci-lint
# - go vet
# - go test ./...
# - go mod tidy
```

### Running Tests

```bash
# Unit tests only (fast)
go test -short ./...

# All tests including integration
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Specific package
go test ./internal/plan/...

# Verbose output
go test -v ./...

# Run single test
go test -run TestCreatePlan_ValidTopic ./internal/plan/...
```

### Benchmarking

```go
func BenchmarkSM2_Review(b *testing.B) {
    card := &Card{EaseFactor: 2.5, Interval: 6}

    for i := 0; i < b.N; i++ {
        card.Review(3)
    }
}
```

```bash
go test -bench=. -benchmem ./internal/flashcard/...
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./internal/flashcard/...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./internal/flashcard/...
go tool pprof mem.prof
```

## Recommended Tools

### Development Tools

```bash
# Install all at once
make install-tools

# Or individually:
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/kisielk/errcheck@latest
```

### Testing Frameworks

- **Unit Tests**: `testing` (stdlib) + `testify` (assertions/mocking)
- **Integration Tests**: `dockertest` (real dependencies)
- **E2E Tests**: `testing` (stdlib with real CLI execution)
- **Golden Files**: `goldie` (snapshot testing)
- **HTTP Mocking**: `go-vcr` (record/replay HTTP)
- **Fuzzing**: `go test -fuzz` (Go 1.18+)

### VS Code Extensions

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.testOnSave": false,
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  }
}
```

## Documentation

### Code Documentation

**When to write docs**:

```go
// ✅ Good: Public API needs godoc
// ParsePlan parses a markdown file containing a learning plan.
// It validates frontmatter and chunk structure.
func ParsePlan(path string) (*Plan, error)

// ❌ Unnecessary: Self-explanatory
// GetID returns the plan ID.
func (p *Plan) GetID() string { return p.ID }

// ✅ Good: Complex logic needs explanation
// SM-2 algorithm calculates next review interval based on rating.
// See: https://www.supermemo.com/en/archives1990-2015/english/ol/sm2
func (c *Card) Review(rating int) { ... }
```

### README Updates

Update README when:

- Adding new commands
- Changing installation steps
- Adding new configuration options

### Changelog

Update `CHANGELOG.md` on every PR:

```markdown
## [Unreleased]

### Added
- Flashcard review TUI interface (#23)
- SM-2 algorithm for spaced repetition (#24)

### Fixed
- Session duration calculation for sessions > 24h (#25)

### Changed
- Improved error messages for LLM failures (#26)
```

## Troubleshooting

### Tests Failing Locally

```bash
# Clean and rebuild
go clean -testcache
go test ./...

# Check Go version (need 1.21+)
go version

# Update dependencies
go mod tidy
go mod verify
```

### Linter Errors

```bash
# See what will be checked
golangci-lint linters

# Run with auto-fix
golangci-lint run --fix

# Run on specific file
golangci-lint run internal/plan/manager.go
```

### Pre-Commit Hook Not Running

```bash
# Reinstall hooks
pre-commit install

# Run manually
pre-commit run --all-files
```

## Resources

- **Go Style**: <https://go.dev/doc/effective_go>
- **Uber Go Guide**: <https://github.com/uber-go/guide>
- **Testing Best Practices**: <https://go.dev/doc/tutorial/add-a-test>
- **Project Layout**: <https://github.com/golang-standards/project-layout>
- **Conventional Commits**: <https://www.conventionalcommits.org/>

## Questions?

- Check existing issues: <https://github.com/pezware/samedi.dev/issues>
- Ask in discussions: <https://github.com/pezware/samedi.dev/discussions>
- Read the specs: `docs/*.md`
