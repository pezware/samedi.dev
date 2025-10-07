# Samedi Implementation Plan

**Last Updated:** 2025-10-06

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

**Status:** Not Started

**Success Criteria:**

- [ ] Can start sessions linked to plans/chunks
- [ ] Can stop sessions with notes
- [ ] Can view active session status
- [ ] Session duration calculated correctly
- [ ] All tests pass
- [ ] `make check` succeeds

**Deliverables:**

### 3.1 Session Domain Models

- [ ] `internal/session/session.go` - Session struct
- [ ] `internal/session/session_test.go` - Model tests

### 3.2 Session Repository

- [ ] `internal/session/repository_sqlite.go` - SQLite implementation
- [ ] `internal/session/repository_test.go` - Repository tests

### 3.3 Session Service

- [ ] `internal/session/service.go` - Business logic
- [ ] `internal/session/timer.go` - Duration calculation
- [ ] `internal/session/service_test.go` - Service tests

### 3.4 CLI Commands

- [ ] `internal/cli/start.go` - `samedi start` command
- [ ] `internal/cli/stop.go` - `samedi stop` command
- [ ] `internal/cli/status.go` - `samedi status` command
- [ ] `internal/cli/session_test.go` - CLI tests

**Tests:**

- Session creation and validation
- Active session detection
- Duration calculation (including overnight sessions)
- Session notes and artifacts
- Error handling (no active session, duplicate start)

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

**Goal:** Visualize learning progress and export reports

**Status:** Not Started

**Success Criteria:**

- [ ] TUI stats dashboard shows total hours, streaks, progress
- [ ] Can export markdown reports
- [ ] Streak calculation works correctly
- [ ] Progress bars and charts display correctly
- [ ] All tests pass
- [ ] `make check` succeeds

**Deliverables:**

### 5.1 Stats Calculator

- [ ] `internal/stats/calculator.go` - Aggregation logic
- [ ] `internal/stats/streak.go` - Streak detection
- [ ] `internal/stats/calculator_test.go` - Calculator tests

### 5.2 Stats Service

- [ ] `internal/stats/service.go` - Business logic
- [ ] `internal/stats/service_test.go` - Service tests

### 5.3 TUI Stats Dashboard

- [ ] `internal/tui/stats.go` - Bubble Tea stats model
- [ ] `internal/tui/components/progress.go` - Progress bar
- [ ] `internal/tui/components/chart.go` - Simple charts
- [ ] `internal/tui/stats_test.go` - TUI tests

### 5.4 Report Exporter

- [ ] `internal/stats/exporter.go` - Markdown generation
- [ ] `internal/stats/exporter_test.go` - Exporter tests

### 5.5 CLI Commands

- [ ] `internal/cli/stats.go` - `samedi stats` command
- [ ] `internal/cli/report.go` - `samedi report` command
- [ ] `internal/cli/stats_test.go` - CLI tests

**Tests:**

- Aggregation calculations (total hours, session counts)
- Streak detection (gaps, multi-day streaks)
- Progress percentage calculations
- Markdown report generation
- TUI rendering (text mode tests)

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

---

## Next Actions

1. ✅ Create IMPLEMENTATION_PLAN.md (this file)
2. → Create internal package structure
3. → Implement config system
4. → Create SQLite schema
5. → Build repository interfaces
