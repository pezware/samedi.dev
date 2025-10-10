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
- **Stage 2 Phase 5: CLI Commands** - User-facing plan management commands
  - `samedi init <topic>` - LLM-powered plan generation with flags (--hours, --level, --goals, --edit)
  - `samedi plan list` - List all plans with filtering (--status, --tag) and JSON output
  - `samedi plan show <plan-id>` - Display plan details with progress and recent chunks (--chunks for all)
  - `samedi plan edit <plan-id>` - Open plan in $EDITOR with validation and metadata update
  - `samedi plan archive <plan-id>` - Archive completed or abandoned plans
  - Service initialization helper with LLM provider selection (claude/mock)
  - Template installation to ~/.samedi/templates/ on first use
  - Helper functions: formatStatus, formatDuration, truncate, formatProgress
  - 20 CLI tests (command structure, argument validation, helper functions)
  - Complete plan management workflow from CLI
- **Stage 3: Session Tracking - Smart Inference & Enhanced Display**
  - Smart inference: automatic chunk/plan status updates based on session activity
    - Starting a session marks chunk as "in-progress" if "not-started"
    - Stopping a session auto-completes chunk when total time ≥ chunk duration
    - Plan status recalculated based on chunk statuses (not-started → in-progress → completed)
    - Best-effort updates: session operations succeed even if status update fails
  - Session service methods: `GetChunkSessions()`, `GetChunkStats()` with `ChunkStats` struct
  - Shared chunk display utility (`internal/cli/chunk_display.go`) with comprehensive formatting
  - `samedi show <plan-id> <chunk-id>` command for detailed chunk information
    - Title, ID, status with icons (○ not-started, ◐ in-progress, ● completed, ⊘ skipped)
    - Duration and progress (X/Y min, Z%)
    - Session count and recent session history (last 3)
    - Objectives, resources, and deliverables
  - Enhanced `samedi status` to display full chunk details when session has active chunk
  - Enhanced `samedi start <plan-id> <chunk-id>` to show progress and session history
  - All smart inference updates properly tested with mock services
- Dependency: `github.com/yuin/goldmark` v1.7.13 for markdown parsing
- **Stage 6: TUI Enhancements** - Interactive stats dashboard with multi-view navigation
  - Multi-view navigation with 5 interactive views (overview, plan-list, plan-detail, session-history, export-dialog)
  - View state management with history stack for back navigation ([Esc] key)
  - Plan list view with cursor navigation and drill-down capability
    - Table display with progress bars for all plans
    - Keyboard navigation (↑/↓/j/k) with wraparound
    - Enter to drill into plan details
    - Empty state handling with graceful messages
  - Plan detail view showing comprehensive plan statistics
    - Visual progress bar with percentage and chunk completion
    - Time investment metrics (total, sessions, average session duration)
    - Last session timestamp
    - Navigation shortcuts: [s] for sessions, [Esc] to return
  - Session history view with context-aware filtering
    - List sessions with date, plan ID, duration, notes preview
    - Automatic filtering by plan when accessed from plan detail view
    - Pagination with centered window (max 20 sessions visible)
    - Helper functions: filterSessionsByPlan(), paginateSessions(), formatSessionRow()
    - Cursor highlighting and wraparound navigation
  - Export dialog for quick report generation
    - Two export types: Summary Report, Full Report
    - Cursor navigation between options (↑/↓/j/k)
    - Enter to select, Esc to cancel
    - User guidance on shell redirection for file export
    - Integration with CLI `samedi report` command
  - Keyboard shortcuts: [p] plans, [s] sessions, [e] export, [q] quit, [Esc] back, [↑/↓/j/k] navigation
  - Comprehensive documentation in `docs/08-stats-analytics.md` (lines 186-353)
    - All 5 views documented with examples
    - Navigation patterns and keyboard shortcuts reference
    - Context-aware behavior documented
  - Implementation: Single-file architecture in `internal/tui/stats.go` (830 lines)
    - All views consolidated rather than split into separate files
    - Shared state management (cursors, selected plan, view history stack)
    - Follows "avoid premature abstractions" principle
    - Benefits: Explicit state, simpler testing, better cohesion
  - Tests: View switching, navigation, rendering for all views
    - `internal/tui/stats_test.go` - Core view tests
    - `internal/tui/stats_navigation_bug_test.go` - Edge case coverage
    - Empty state handling verified
  - All tests passing, `make check` succeeds

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
- **Stage 2 CLI Gaps** - Fixed 7 functional requirement violations identified in review
  - Fixed `samedi init --model` flag to properly override LLM model (FR-001)
  - Implemented progress calculation in `samedi plan list` showing percentage and chunk counts (FR-009)
  - Implemented `--sort` functionality in `samedi plan list` with field validation (created/updated/title/status/hours) (FR-009)
  - Added confirmation prompt to `samedi plan archive` requiring plan ID to confirm, with `--yes` flag to skip (FR-018)
  - Filtered archived plans from default `samedi plan list` output, added `--all` flag to show all (FR-018)
  - Added session history display to `samedi plan show` with `--sessions` flag (FR-010, Stage 3 placeholder)
  - Added flashcard count display to `samedi plan show` with `--cards` flag (FR-010, Stage 4 placeholder)
  - Refactored `planShowCmd` to reduce cyclomatic complexity by extracting display helpers
  - All tests passing, all FR violations resolved
- **LLM CLI Integration** - Fixed broken Claude provider and added working implementations
  - Fixed `--prompt-file` error: Claude CLI provider was using non-existent flag
  - Added `auto` provider for automatic CLI detection (claude → codex → gemini → llm → mock)
  - Added `claude` provider for Claude Code CLI (recommended for Claude users)
  - Added `codex` provider for Codex CLI (OpenAI-focused)
  - Added `gemini` provider for Gemini CLI (Google models)
  - Added `llm` provider for Simon Willison's llm CLI tool (universal fallback)
  - Added `stdin` provider for generic CLI tools (aichat, mods, ollama, etc.)
  - Added `mock` provider as explicit option for testing
  - Updated default LLM provider from "claude" to "llm" with model "claude-3-5-sonnet"
  - Updated config validator to accept new providers (auto, claude, codex, gemini, llm, stdin, mock)
  - All providers use stdin-based interface for better compatibility
  - Installation: `uv pip install llm && llm install llm-claude-3` (using modern uv package manager)
  - Auto-detection allows zero-configuration setup for users with existing CLIs
  - 34+ new tests for LLM providers with CI-safe design
  - All existing tests passing with updated defaults
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
