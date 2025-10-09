# Samedi Implementation Plan

**Last Updated:** 2025-10-09

## Overview

This document tracks the incremental implementation of samedi, broken into 5
stages. Each stage delivers a complete, tested slice of functionality.

## Principles

- **Incremental:** Each stage builds on the previous, compiles, and passes tests
- **Test-driven:** Write failing tests first, then implement
- **Boring tech:** SQLite, standard lib, proven libraries (Cobra, Bubble Tea, Viper)
- **Git hygiene:** Small commits with clear messages
- **No secrets:** All sensitive config via environment variables

## Stage 1: Foundation (Config + Storage)

**Goal:** Set up core infrastructure for configuration and data persistence

**Status:** Complete ✅

**Success Criteria:**

- [x] Config system loads TOML files with defaults
- [x] SQLite database initializes with schema
- [x] Filesystem structure created (~/.samedi/)
- [x] Repository interfaces defined
- [x] `samedi config list` command works
- [x] All tests pass
- [x] `make check` succeeds

**Deliverables:**

### 1.1 Configuration System

- [x] `internal/config/config.go` - Config struct and defaults
- [x] `internal/config/loader.go` - TOML loading with Viper
- [x] `internal/config/validator.go` - Validation logic
- [x] `internal/config/config_test.go` - Unit tests

### 1.2 SQLite Storage

- [x] `internal/storage/sqlite.go` - SQLite connection and operations
- [x] `internal/storage/migrations/001_initial_schema.sql` - Initial schema
- [x] `internal/storage/migrator.go` - Migration runner
- [x] `internal/storage/sqlite_test.go` - Database tests

### 1.3 Filesystem Storage

- [x] `internal/storage/filesystem.go` - File operations
- [x] `internal/storage/paths.go` - Path management
- [x] `internal/storage/filesystem_test.go` - Filesystem tests

### 1.4 Repository Interfaces

- [x] `internal/storage/repository.go` - Repository interfaces
- [x] Mock implementations for testing

### 1.5 CLI Integration

- [x] `internal/cli/root.go` - Root Cobra command
- [x] `internal/cli/config.go` - Config subcommands
- [x] Update `cmd/samedi/main.go` to use Cobra

**Tests:**

- Config loading from file and defaults
- Config validation (invalid LLM provider, invalid paths)
- SQLite schema creation and migrations
- Filesystem directory structure creation
- Repository CRUD operations

**Dependencies to Add:**

- `github.com/spf13/cobra` - Already in go.mod
- `github.com/spf13/viper` - Already in go.mod
- `github.com/mattn/go-sqlite3` - Need to add
- `gopkg.in/yaml.v3` - Already in go.mod

---

## Stage 2: Plan Management

**Goal:** Create and manage learning plans

**Status:** Complete ✅

**Success Criteria:**

- [x] Can generate plans via LLM (with mock provider)
- [x] Plans saved as markdown with frontmatter
- [x] Plans indexed in SQLite
- [x] Can list, view, and edit plans via CLI commands
- [x] All tests pass (146+ tests, combined coverage >80%)
- [x] `make check` succeeds

**Deliverables:**

### 2.1 Plan Domain Models ✅

- [x] `internal/plan/plan.go` - Plan and Chunk structs with validation
- [x] `internal/plan/plan_test.go` - Model tests (24 tests)

**Note:** Combined Plan and Chunk into single file (plan.go) per Go conventions

### 2.2 Markdown Parser ✅

- [x] `internal/plan/parser.go` - Parse markdown with frontmatter (bidirectional)
- [x] `internal/plan/parser_test.go` - Parser tests (14 tests with fixtures)

**Note:** Validation is integrated into Plan.Validate() method, no separate
validator.go needed

### 2.3 LLM Integration ✅

- [x] `internal/llm/provider.go` - LLM provider interface
- [x] `internal/llm/mock.go` - Mock provider for testing
- [x] `internal/llm/claude.go` - Claude CLI implementation (stub)
- [x] `templates/plan-generation.md` - LLM prompt template
- [x] `internal/llm/provider_test.go` - Provider tests (11 tests)

**Note:** Provider interface supports context-based timeouts and retry logic.
Claude CLI provider uses temp files for prompt passing.

### 2.4 Plan Service ✅

- [x] `internal/plan/service.go` - Business logic (290 lines)
- [x] `internal/plan/repository_sqlite.go` - SQLite implementation (279 lines)
- [x] `internal/plan/repository_filesystem.go` - Filesystem implementation (155 lines)
- [x] `internal/plan/service_test.go` - Service tests (21 tests, 550 lines)
- [x] `internal/plan/repository_sqlite_test.go` - SQLite tests
  (11 tests, 322 lines)
- [x] `internal/plan/repository_filesystem_test.go` - Filesystem tests
  (14 tests, 365 lines)

**Note:** Service layer orchestrates between SQLite metadata storage, filesystem
markdown storage, and LLM providers. Includes proper error handling, rollback on
failures, and comprehensive test coverage with mock LLM provider.

### 2.5 CLI Commands ✅

- [x] `internal/cli/init.go` - `samedi init` command (135 lines)
- [x] `internal/cli/plan.go` - `samedi plan` subcommands (379 lines)
- [x] `internal/cli/plan_test.go` - CLI tests (149 lines)
- [x] `internal/cli/root.go` - Updated with plan service initialization
  (98 lines added)

**Note:** CLI commands provide complete plan management workflow.
Integration with service layer enables LLM-powered plan generation,
listing with filters, detailed plan views, in-editor editing, and
archiving.

**Tests:**

- Markdown parsing with valid/invalid frontmatter
- Plan validation (missing fields, invalid chunks)
- LLM provider interface (mock calls)
- Plan creation and SQLite persistence
- Plan listing and filtering
- Plan editing workflow

**Dependencies Added:**

- ✅ `github.com/yuin/goldmark` - Markdown parsing (added v1.7.13)
- ✅ `gopkg.in/yaml.v3` - Already in go.mod

---

## Stage 3: Session Tracking

**Goal:** Track learning sessions with start/stop/status commands

**Status:** Complete ✅

**Success Criteria:**

- [x] Can start sessions linked to plans/chunks
- [x] Can stop sessions with notes
- [x] Can view active session status
- [x] Session duration calculated correctly
- [x] All tests pass (96+ tests for session tracking)
- [x] `make check` succeeds

**Deliverables:**

### 3.1 Session Domain Models ✅

- [x] `internal/session/session.go` - Session struct with validation (139 lines)
- [x] `internal/session/session_test.go` - Model tests (30 tests, 543 lines)

**Note:** Session model includes IsActive(), CalculateDuration(), ElapsedTime() helpers
and Complete() method for session lifecycle management.

### 3.2 Session Repository ✅

- [x] `internal/session/repository_sqlite.go` - SQLite implementation (277 lines)
- [x] `internal/session/repository_sqlite_test.go` - Repository tests
  (16 tests, 455 lines)

**Note:** Repository includes Create, Get, GetActive, Update, List, GetByPlan,
and Delete operations with proper JSON serialization for artifacts array.
Fixed List() to handle limit=0 correctly.

### 3.3 Session Service ✅

- [x] `internal/session/service.go` - Business logic (278 lines)
- [x] `internal/session/service_test.go` - Service tests (30 tests, 689 lines)

**Note:** Service layer handles Start/Stop/GetActive/GetStatus operations with validation
for duplicate active sessions. Duration calculation includes overnight and multi-day
session support. Note: Duration calculation is in session.go, not separate timer.go.

### 3.4 CLI Commands ✅

- [x] `internal/cli/start.go` - `samedi start` command (85 lines)
- [x] `internal/cli/stop.go` - `samedi stop` command (89 lines)
- [x] `internal/cli/status.go` - `samedi status` command (122 lines)
- [x] `internal/cli/session_test.go` - CLI tests (10 tests, 140 lines)
- [x] `internal/cli/plan.go` - Updated with session history display (37 lines added)
- [x] `internal/plan/service.go` - Integrated with session service (32 lines added)

**Note:** CLI commands provide complete session workflow. Session history
integrated into `samedi plan show` command. GetSessionService() helper added to
root.go for dependency injection.

### 3.5 Integration Tests ✅

- [x] `internal/session/integration_test.go` - Integration tests (4 tests, 289 lines)

**Note:** Comprehensive end-to-end tests covering complete session lifecycle, multiple
sessions per plan, sessions across different plans, and limit parameter behavior.

**Tests:**

- Session creation and validation (11 tests)
- Active session detection (7 tests)
- Duration calculation including overnight sessions (5 tests)
- Session notes and artifacts (5 tests)
- Error handling: no active session, duplicate start (4 tests)
- Repository operations: CRUD, filtering, artifacts (16 tests)
- Service operations: Start/Stop/GetStatus/List (30 tests)
- CLI command structure and flags (10 tests)
- Integration tests: full workflows (4 tests)

---

## Stage 4: Flashcard System

**Goal:** Spaced repetition flashcards with SM-2 algorithm

**Status:** Not Started

**Success Criteria:**

- [ ] Can create flashcards manually
- [ ] Can generate flashcards via LLM (stub)
- [ ] SM-2 algorithm calculates intervals correctly
- [ ] TUI review interface works
- [ ] Cards scheduled for future review
- [ ] All tests pass
- [ ] `make check` succeeds

**Deliverables:**

### 4.1 Flashcard Domain Models

- [ ] `internal/flashcard/card.go` - Card struct
- [ ] `internal/flashcard/sm2.go` - SM-2 algorithm
- [ ] `internal/flashcard/card_test.go` - Model tests
- [ ] `internal/flashcard/sm2_test.go` - Algorithm tests

### 4.2 Flashcard Repository

- [ ] `internal/flashcard/repository_sqlite.go` - SQLite implementation
- [ ] `internal/flashcard/repository_filesystem.go` - Markdown storage
- [ ] `internal/flashcard/repository_test.go` - Repository tests

### 4.3 Flashcard Service

- [ ] `internal/flashcard/service.go` - Business logic
- [ ] `internal/flashcard/generator.go` - LLM extraction (stub)
- [ ] `internal/flashcard/service_test.go` - Service tests

### 4.4 TUI Review Interface

- [ ] `internal/tui/review.go` - Bubble Tea review model
- [ ] `internal/tui/components/card.go` - Card display component
- [ ] `internal/tui/review_test.go` - TUI tests

### 4.5 CLI Commands

- [ ] `internal/cli/review.go` - `samedi review` command
- [ ] `internal/cli/cards.go` - `samedi cards` subcommands
- [ ] `internal/cli/cards_test.go` - CLI tests

**Tests:**

- SM-2 algorithm (all rating scenarios)
- Card scheduling logic
- Due card queries
- Review flow (show card, rate, update)
- TUI keyboard navigation

**Dependencies to Add:**

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - TUI styling

---

## Stage 5: Stats & Reporting

**Goal:** Visualize learning progress with stats dashboard and enable markdown
report export

**Status:** Complete ✅

**Success Criteria:**

- [x] Stats calculator computes totals, averages, and streaks correctly
- [x] Stats service integrates with session/plan repositories
- [x] CLI stats command provides text and JSON output
- [x] TUI stats dashboard shows interactive visualizations
- [x] Can export comprehensive markdown reports
- [x] All tests pass (196+ tests across all packages)
- [x] `make check` succeeds

**Deliverables:**

### 5.1 Stats Domain Models & Calculator ✅

- [x] `internal/stats/types.go` - Domain types (169 lines)
  - TotalStats (total hours, session count, streak, active plans)
  - PlanStats (plan-specific hours, progress, session count)
  - DailyStats (date, duration, session count)
  - TimeRange (start, end, preset filters)
- [x] `internal/stats/calculator.go` - Aggregation logic (283 lines)
  - CalculateTotalStats, CalculatePlanStats, CalculateDailyStats
  - AggregateByPlan helper
  - sumDurations, countSessions, calculateProgress helpers
- [x] `internal/stats/streak.go` - Streak detection (120 lines)
  - CalculateStreak (current and longest)
  - GetActiveDays
  - Timezone and overnight session handling
- [x] `internal/stats/types_test.go` - Domain tests (14 tests)
- [x] `internal/stats/calculator_test.go` - Calculator tests (23 tests)
- [x] `internal/stats/streak_test.go` - Streak tests (9 tests)

**Note:** Pure calculation functions with comprehensive edge case coverage.
Test coverage: >85%. Fixed bug where plan counting was skipped when no sessions exist.

### 5.2 Stats Service ✅

- [x] `internal/stats/service.go` - Business logic (245 lines)
  - GetTotalStats, GetPlanStats, GetDailyStats, GetAllPlanStats, GetStreakInfo, GetActiveDays
  - Repository integration (session + plan)
  - TimeRange filtering applied at service layer
  - Error handling and validation
- [x] `internal/stats/service_test.go` - Service tests (17 tests)
  - All service methods with various time ranges
  - Repository error handling
  - Empty results handling
  - Mock repositories

**Note:** Service orchestrates calculator functions with repository data loading.
Includes context-based operations for cancellation. Fixed bugs where TimeRange
was not passed to service methods.

### 5.3 CLI Stats Command (Text Output) ✅

- [x] `internal/cli/stats.go` - Stats command (442 lines)
  - Time range flags (--range: all, today, this-week, this-month)
  - JSON output support (--json)
  - Breakdown flag (--breakdown) for daily stats
  - TUI flag (--tui) for interactive mode
  - Text formatting with Unicode progress bars
  - Helper functions: printPlanStatsJSON, printPlanBreakdown
  - Error handling and validation
- [x] `internal/cli/stats_test.go` - CLI tests (18 tests)
  - Command structure and flags
  - Text formatting (formatPlanStatus, buildProgressBar)
  - Time range parsing
  - Edge cases (zero/full/negative progress, Unicode width)
- [x] `internal/cli/root.go` - Updated
  - getStatsService() helper for dependency injection
  - Register stats command

**Note:** Text-based output for terminal use, JSON for scripting/automation.
Progress bars use Unicode characters (█░) for better visualization.
Fixed bugs where --range and --breakdown flags were not wired up.

### 5.4 TUI Stats Dashboard ✅

- [x] `internal/tui/stats.go` - Bubble Tea model (230 lines)
  - Interactive dashboard with keyboard navigation (q to quit)
  - View modes: total stats, plan-specific stats
  - Keyboard input handling (Ctrl+C, q)
  - Window size adaptation
- [x] `internal/tui/components/progress.go` - Progress bar (72 lines)
  - Styled progress bars with Lipgloss
  - Color-coded (green/yellow/red based on progress)
  - Percentage labels
- [x] `internal/tui/components/table.go` - Table component (121 lines)
  - Aligned columns with headers
  - Border rendering
  - Style support
- [x] `internal/tui/stats_test.go` - TUI tests (8 tests)
  - Model initialization
  - Keyboard input handling (quit, Ctrl+C)
  - View rendering (total stats, plan stats)
  - Help text verification
- [x] `internal/cli/stats.go` - Updated
  - --tui flag for interactive mode
  - launchTUI() function to initialize Bubble Tea program

**Note:** Interactive TUI provides rich visualization. Fixed bug where non-functional
time-range toggle controls were present - removed them and direct users to use
CLI --range flag before launching TUI.

**Dependencies Added:**

- `github.com/charmbracelet/bubbletea` - TUI framework ✅
- `github.com/charmbracelet/lipgloss` - TUI styling ✅

### 5.5 Report Exporter & Command ✅

- [x] `internal/stats/exporter.go` - Report generator (274 lines)
  - ExportTotalStats, ExportPlanStats, ExportFullReport methods
  - Section generators (summary, plans, daily breakdown)
  - Markdown formatting with headers and tables
  - Time-based filtering support
- [x] `internal/stats/exporter_test.go` - Exporter tests (15 tests)
  - Full report generation
  - Plan-specific reports
  - Summary reports
  - Markdown formatting validation
- [x] `internal/cli/report.go` - Report command (166 lines)
  - Type selection (--type: summary, full)
  - Filtering flags (--range: all, today, this-week, this-month)
  - Output to stdout or file (--output)
  - Error handling and validation
- [x] `internal/cli/report_test.go` - Report CLI tests (10 tests)
  - Command structure and flags
  - Output destinations (stdout, file)
  - Filtering options
  - Error handling

**Note:** Markdown report generation with comprehensive statistics.
Fixed bug where TimeRange was not applied to service method calls in report command.

**Tests:**

- Aggregation calculations (sum, average, count) - 25 tests
- Streak detection (consecutive days, gaps, overnight) - 20 tests
- Progress percentage calculations - 10 tests
- Service integration with repositories - 30 tests
- CLI text and JSON output formatting - 18 tests
- TUI rendering and keyboard navigation - 12 tests
- Markdown report generation and validation - 20 tests
- Template rendering edge cases - 15 tests

**Total:** 150+ tests, ~2,450 lines of test code, ~3,200 lines of implementation

**Implementation Timeline:**

- **Phase 1** (Days 1-5): Stats domain models & calculator (60 tests)
- **Phase 2** (Days 6-10): Stats service layer (30 tests)
- **Phase 3** (Days 11-15): CLI stats command (18 tests)
- **Phase 4** (Days 16-22): TUI stats dashboard (12 tests)
- **Phase 5** (Days 23-27): Report exporter & command (30 tests)

**Git Branches:**

- `feat/stage-5-stats-calculator` (Phase 1)
- `feat/stage-5-stats-service` (Phase 2)
- `feat/stage-5-cli-stats` (Phase 3)
- `feat/stage-5-tui-dashboard` (Phase 4)
- `feat/stage-5-report-exporter` (Phase 5)

---

## Definition of Done (Each Stage)

Before marking a stage complete:

- [ ] All deliverables implemented
- [ ] All tests written and passing
- [ ] Code coverage > 80% for core logic
- [ ] `make check` passes (fmt, vet, lint, test)
- [ ] License headers on all new files
- [ ] Code reviewed against Uber Go Style Guide
- [ ] Commit messages follow Conventional Commits
- [ ] IMPLEMENTATION_PLAN.md updated with status

---

## Risk Mitigation

### Risks

1. **LLM CLI integration complexity** → Mitigation: Use mock provider,
   defer real integration
2. **SQLite concurrency issues** → Mitigation: WAL mode, proper locking
3. **Markdown parsing edge cases** → Mitigation: Comprehensive test fixtures
4. **TUI rendering on different terminals** → Mitigation: Test on multiple
   terminals, graceful degradation

### Decisions

- **Mock LLM first:** Stage 2 uses mock provider, real LLM integration in Phase 2
- **No network calls in MVP:** All local-first, cloud sync deferred to
  Phase 2
- **Simple TUI:** Focus on functionality over fancy visuals in MVP

---

## Stage 1 Follow-Up Improvements

**Priority:** Low - Nice-to-have enhancements for future iterations

### 1. Enhanced Config Validation

- [ ] Check if LLM CLI command exists in PATH
- [ ] Verify data directory is writable before accepting config
- [ ] Validate timezone string format (e.g., "America/Los_Angeles")
- [ ] Check backup directory permissions when backup is enabled

**Rationale:** Catch configuration errors early rather than at runtime

### 2. Improved Error Messages

- [ ] Add actionable suggestions to error messages
- [ ] Example: "invalid LLM provider" → suggest correct command to fix

**Rationale:** Better UX per NFR-006 (actionable error messages)

### 3. Config Set Validation Timing

- [ ] Move validation from `Save()` to `setConfigValue()` for immediate feedback

**Rationale:** User shouldn't see "success" message if value will fail validation

### 4. Integration Tests for Config CLI

- [ ] Test `samedi config list` output formatting
- [ ] Test `samedi config list --json` produces valid JSON
- [ ] Test `samedi config edit` opens $EDITOR correctly

**Rationale:** Increase confidence in CLI command wiring

### 5. Test Coverage Analysis

- [ ] Run coverage report and document current percentage
- [ ] Add tests to reach >80% coverage goal

**Rationale:** Meet Definition of Done requirement

---

## Progress Log

### 2025-10-06

- Created implementation plan
- Started Stage 1: Foundation
- Completed Stage 1 deliverables:
  - Configuration system (config, loader, validator, tests)
  - SQLite storage with migrations
  - Filesystem storage (paths, operations, tests)
  - Repository interfaces with mocks
  - Cobra CLI integration with config subcommands
- All Stage 1 success criteria met
- Documented 5 follow-up improvements for future iteration
- Started Stage 2: Plan Management
- Completed Stage 2 Phase 1-2:
  - Plan and Chunk domain models with validation (24 tests)
  - Markdown parser with YAML frontmatter support (14 tests)
  - 97.9% test coverage, all quality checks passing
  - PR #5 created and merged
- Completed Stage 2 Phase 3:
  - LLM provider interface with mock and Claude CLI implementations
  - Plan generation template with Go template variables
  - 11 tests with 86.7% coverage
  - Context-based timeout handling and error wrapping
  - Branch: feat/stage-2-llm-integration

### 2025-10-07

- Completed Stage 2 Phase 4 (Plan Service):
  - Fixed SQLite repository interface to match storage.PlanRepository
  - Added ToRecord() and RecordToPlan() conversion functions
  - Implemented FilesystemRepository with Save/Load/Delete/List operations (14 tests)
  - Implemented Service layer orchestrating SQLite + Filesystem + LLM (21 tests)
  - Service includes Create/Get/Update/Delete/List/GetMetadata operations
  - Template rendering with Go templates for LLM prompt generation
  - Slugify utility for generating filesystem-safe plan IDs
  - Comprehensive error handling with rollback on failures
  - 126 total tests in plan package, all passing
  - All quality checks passing (fmt, vet, lint, tests with race detection)
  - Test coverage >85% across all plan package components
  - Branch: feat/stage-2-plan-service-cli (PR #7, merged)

- Completed Stage 2 Phase 5 (CLI Commands):
  - Implemented `samedi init` command for LLM-powered plan generation (135 lines)
  - Implemented `samedi plan list/show/edit/archive` commands (379 lines)
  - Added getPlanService() helper for dependency injection in CLI
  - Template installation with ensureTemplate() helper
  - 20 CLI tests covering command structure and helper functions (149 lines)
  - Full plan management workflow: init → list → show → edit → archive
  - All quality checks passing (make check)
  - Branch: feat/stage-2-cli-commands

- Fixed Stage 2 CLI gaps (identified in review):
  - Fixed `samedi init --model` flag to wire through to LLM provider
  - Implemented progress calculation in `samedi plan list` (% and chunk counts)
  - Implemented `--sort` functionality in `samedi plan list` with field validation
  - Added confirmation prompt to `samedi plan archive` with --yes flag
  - Filtered archived plans from default `samedi plan list` output
  - Added session history display to `samedi plan show` (Stage 3 placeholder)
  - Added flashcard count display to `samedi plan show` (Stage 4 placeholder)
  - Extracted display helpers to reduce cyclomatic complexity
  - 6 commits, all tests passing, all FR violations resolved
  - Branch: fix/stage-2-cli-gaps

- Completed Stage 3: Session Tracking (full implementation):
  - **Phase 1-3**: Session Domain, Repository, and Service layers
    - Session model with validation and lifecycle management (30 tests)
    - SQLite repository with CRUD operations and JSON artifact storage (16 tests)
    - Service layer with Start/Stop/GetActive/GetStatus operations (30 tests)
    - Duration calculation supporting overnight and multi-day sessions
    - Prevention of duplicate active sessions
    - 76 total tests passing with >85% coverage
  - **Phase 4**: CLI Commands
    - `samedi start` command with plan-id, chunk-id, and --note flag (85 lines)
    - `samedi stop` command with --note and --artifact flags (89 lines)
    - `samedi status` command showing active session or recent history (122 lines)
    - 10 CLI tests for command structure, flags, and argument validation
    - getSessionService() helper for dependency injection
  - **Phase 5**: Integration & Polish
    - Session history integrated into `samedi plan show` command
    - Plan service SetSessionService() method for optional session integration
    - displaySessionSummary() helper for formatted session display
    - 4 comprehensive integration tests covering full session workflows
    - Fixed List() repository bug where limit=0 returned no results
    - All 96+ tests passing (unit + integration)
  - Branch: feat/stage-3-session-tracking (ready for merge)

### 2025-10-09

- Completed Stage 5: Stats & Reporting (full implementation):
  - **Phase 1-2**: Stats Domain Models, Calculator, and Service
    - Stats domain types: TotalStats, PlanStats, DailyStats, TimeRange
      (169 lines)
    - Calculator functions: CalculateTotalStats, CalculatePlanStats,
      CalculateDailyStats (283 lines)
    - Streak detection: CalculateStreak, GetActiveDays (120 lines)
    - Stats service with repository integration (245 lines)
    - 46 domain/calculator tests + 17 service tests, >85% coverage
  - **Phase 3**: CLI Stats Command
    - `samedi stats` command with text and JSON output (442 lines)
    - Time range filtering (--range: all, today, this-week, this-month)
    - Daily breakdown flag (--breakdown)
    - TUI launcher (--tui)
    - Unicode progress bars and status formatting
    - 18 CLI tests for command structure, flags, and formatting
  - **Phase 4**: TUI Stats Dashboard
    - Bubble Tea interactive dashboard (230 lines)
    - Progress bar and table components (193 lines combined)
    - View modes: total stats, plan-specific stats
    - Keyboard navigation (q to quit, Ctrl+C)
    - 8 TUI tests for model, input handling, and rendering
  - **Phase 5**: Report Exporter & Command
    - Markdown report generator (274 lines)
    - `samedi report` command with filtering (166 lines)
    - Export types: summary, full with daily breakdown
    - Time range filtering and output to file/stdout
    - 15 exporter tests + 10 CLI tests
  - **Bug Fixes** (from code review):
    - Fixed plan counting with 0 sessions (calculator.go)
    - Added TimeRange parameters to service methods
    - Wired up --range and --breakdown flags in CLI
    - Removed non-functional TUI time-range controls
    - Applied time-range filtering to report command
    - Reduced cyclomatic complexity in CLI helpers
  - **Total**: 119 tests across stats, TUI, and CLI packages
  - All quality checks passing (make check)
  - Branch: feat/stage-5-stats-calculator

---

## Next Actions

1. ✅ Create IMPLEMENTATION_PLAN.md (this file)
2. → Create internal package structure
3. → Implement config system
4. → Create SQLite schema
5. → Build repository interfaces
