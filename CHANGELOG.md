# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project structure
- Comprehensive specification documents
- Development tooling (Makefile, pre-commit hooks, CI/CD)
- Testing framework setup
- Go ecosystem guide (docs/GO_ECOSYSTEM.md) with complete tooling reference
- `.tool-versions` file for asdf version management (golang 1.23.5)
- `tools/tools.go` for pinning development tool versions
- `make vuln` target for vulnerability scanning with govulncheck
- Pinned versions for all development tools in Makefile
- **Stage 1: Foundation** - Configuration and storage infrastructure
  - Configuration system with TOML loading, validation, and CLI commands
  - SQLite database with schema migrations
  - Filesystem storage with path management
  - Repository interfaces with mock implementations
  - Cobra CLI integration (`samedi config` subcommands)
  - 81.4% test coverage on config package
- **Stage 2 Phase 1-2: Plan Management** - Domain models and markdown parsing
  - Plan and Chunk domain models with comprehensive validation
  - Status enum (not-started, in-progress, completed, skipped, archived)
  - Bidirectional markdown parser with YAML frontmatter support
  - Duration parsing with flexible formats (hours/minutes)
  - Helper methods: Progress(), NextChunk(), duration calculations
  - 38 tests with 97.9% coverage
  - Test fixtures for valid/invalid/edge cases
  - `.markdownlintignore` to exclude test data from linting
- **Stage 2 Phase 3: LLM Integration** - Provider abstraction for plan generation
  - LLM provider interface with `Call(context, prompt)` method
  - Mock provider for testing with pattern-based canned responses
  - Claude CLI provider with timeout and error handling
  - Plan generation template with Go template variables (Topic, TotalHours, etc.)
  - ProviderError type with retry indication
  - 11 tests with 86.7% coverage
  - Security: Proper handling of dynamic commands with gosec annotations
- **Stage 2 Phase 4: Plan Service** - Complete plan management with hybrid storage
  - SQLite repository with ToRecord/RecordToPlan conversion functions
  - Filesystem repository for markdown file operations (Save/Load/Delete/List/LoadAll)
  - Service layer orchestrating SQLite metadata + filesystem markdown + LLM generation
  - Plan CRUD operations: Create (with LLM), Get, Update, Delete, List, GetMetadata
  - Template rendering with Go text/template for LLM prompts
  - Slugify utility for filesystem-safe plan IDs from topics
  - Atomic operations with proper rollback on failures
  - 46 new tests (14 filesystem, 11 SQLite, 21 service) = 126 total tests in plan package
  - Test coverage >85% across all components
  - Mock LLM provider for deterministic testing
- Dependency: `github.com/yuin/goldmark` v1.7.13 for markdown parsing

### Changed
- Updated to Go 1.25.1 (from 1.21)
- Updated CI/CD to read Go version from `.go-version` (GitHub Actions format)
- Changed tool installation to use `@latest` versions for Go 1.25+ compatibility
- Removed Go version matrix from CI (now uses single version from `.go-version`)
- Dependency review now only fails on HIGH severity (accepts MODERATE in dev tools)

### Deprecated
- TBD

### Removed
- TBD

### Fixed
- golangci-lint configuration updated to fix deprecated options and linters
  - Replaced deprecated `run.skip-dirs` and `run.skip-files` with `issues.exclude-dirs` and `issues.exclude-files`
  - Replaced deprecated `exportloopref` linter with `copyloopvar` (Go 1.22+ compatible)
  - Replaced deprecated `gomnd` linter with `mnd` (renamed)
  - Removed `output` section incompatible with v1.64.8
  - Pinned golangci-lint to v1.64.8 to avoid v2.x breaking changes
  - Resolves CI/CD exit code 7 error from golangci-lint
  - See docs/GOLANGCI_LINT_V2_ISSUE.md for details
- Tool installation errors with Go 1.25.1
  - Fixed incompatible tool versions by using `@latest` for most tools
  - Pinned golangci-lint to v1.64.8 (v2.x has breaking config changes)
  - Fixed import ordering in `tools/tools.go` to satisfy gofmt
  - Updated Makefile to use GOPATH binaries explicitly
  - See docs/GO_1_25_UPGRADE.md for details

### Security
- TBD

## [0.1.0] - TBD

### Added
- Initial release
- Core CLI commands (init, start, stop, status)
- Plan generation with LLM integration
- Session tracking
- Basic TUI dashboard
- SQLite storage
- Markdown plan format

[Unreleased]: https://github.com/pezware/samedi.dev/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/pezware/samedi.dev/releases/tag/v0.1.0
