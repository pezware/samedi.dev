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

**Status:** In Progress

**Success Criteria:**

- [ ] Config system loads TOML files with defaults
- [ ] SQLite database initializes with schema
- [ ] Filesystem structure created (~/.samedi/)
- [ ] Repository interfaces defined
- [ ] `samedi config list` command works
- [ ] All tests pass
- [ ] `make check` succeeds

**Deliverables:**

### 1.1 Configuration System

- [ ] `internal/config/config.go` - Config struct and defaults
- [ ] `internal/config/loader.go` - TOML loading with Viper
- [ ] `internal/config/validator.go` - Validation logic
- [ ] `internal/config/config_test.go` - Unit tests

### 1.2 SQLite Storage

- [ ] `internal/storage/sqlite.go` - SQLite connection and operations
- [ ] `internal/storage/migrations/001_initial_schema.sql` - Initial schema
- [ ] `internal/storage/migrator.go` - Migration runner
- [ ] `internal/storage/sqlite_test.go` - Database tests

### 1.3 Filesystem Storage

- [ ] `internal/storage/filesystem.go` - File operations
- [ ] `internal/storage/paths.go` - Path management
- [ ] `internal/storage/filesystem_test.go` - Filesystem tests

### 1.4 Repository Interfaces

- [ ] `internal/storage/repository.go` - Repository interfaces
- [ ] Mock implementations for testing

### 1.5 CLI Integration

- [ ] `internal/cli/root.go` - Root Cobra command
- [ ] `internal/cli/config.go` - Config subcommands
- [ ] Update `cmd/samedi/main.go` to use Cobra

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

**Status:** Not Started

**Success Criteria:**

- [ ] Can generate plans via LLM (with mock provider)
- [ ] Plans saved as markdown with frontmatter
- [ ] Plans indexed in SQLite
- [ ] Can list, view, and edit plans
- [ ] All tests pass
- [ ] `make check` succeeds

**Deliverables:**

### 2.1 Plan Domain Models

- [ ] `internal/plan/plan.go` - Plan struct
- [ ] `internal/plan/chunk.go` - Chunk struct
- [ ] `internal/plan/plan_test.go` - Model tests

### 2.2 Markdown Parser

- [ ] `internal/plan/parser.go` - Parse markdown with frontmatter
- [ ] `internal/plan/validator.go` - Validate plan structure
- [ ] `internal/plan/parser_test.go` - Parser tests

### 2.3 LLM Integration

- [ ] `internal/llm/provider.go` - LLM provider interface
- [ ] `internal/llm/mock.go` - Mock provider for testing
- [ ] `internal/llm/claude.go` - Claude CLI implementation (stub)
- [ ] `templates/plan-generation.md` - LLM prompt template
- [ ] `internal/llm/provider_test.go` - Provider tests

### 2.4 Plan Service

- [ ] `internal/plan/service.go` - Business logic
- [ ] `internal/plan/repository_sqlite.go` - SQLite implementation
- [ ] `internal/plan/repository_filesystem.go` - Filesystem implementation
- [ ] `internal/plan/service_test.go` - Service tests

### 2.5 CLI Commands

- [ ] `internal/cli/init.go` - `samedi init` command
- [ ] `internal/cli/plan.go` - `samedi plan` subcommands
- [ ] `internal/cli/plan_test.go` - CLI tests

**Tests:**

- Markdown parsing with valid/invalid frontmatter
- Plan validation (missing fields, invalid chunks)
- LLM provider interface (mock calls)
- Plan creation and SQLite persistence
- Plan listing and filtering
- Plan editing workflow

**Dependencies to Add:**

- `github.com/yuin/goldmark` - Markdown parsing
- `gopkg.in/yaml.v3` - Already in go.mod

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

## Progress Log

### 2025-10-06

- Created implementation plan
- Started Stage 1: Foundation
- Created todo list for Stage 1 tasks

---

## Next Actions

1. ✅ Create IMPLEMENTATION_PLAN.md (this file)
2. → Create internal package structure
3. → Implement config system
4. → Create SQLite schema
5. → Build repository interfaces
