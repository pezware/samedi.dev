# Architecture

## Technology Stack

### Phase 1: Local CLI/TUI

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Language | **Go 1.21+** | Fast compile, single binary, great CLI libs, cross-platform |
| TUI Framework | **Bubble Tea** | Best-in-class Go TUI, active community, rich components |
| CLI Framework | **Cobra** | Industry standard, subcommands, flags, completions |
| Database | **SQLite** (go-sqlite3) | Embedded, reliable, no server, perfect for local data |
| Markdown Parser | **goldmark** | CommonMark compliant, extensible, fast |
| Frontmatter | **go-yaml/v3** | Parse YAML frontmatter in markdown |
| Testing | **testify** | Assertions, mocking, test suites |
| Config | **viper** | TOML/YAML/JSON, env vars, defaults |

### Phase 2: Cloud Sync

| Component | Technology | Rationale |
|-----------|------------|-----------|
| API | **Cloudflare Workers** | Edge computing, global distribution, free tier |
| Database | **Cloudflare D1** | SQLite at edge, serverless, syncs with local |
| Storage | **Cloudflare R2** | Object storage for backups, markdown files |
| Auth | **Cloudflare Access** or custom JWT | Email-based, simple, secure |
| Web Dashboard | **Hono + HTMX** | Lightweight, fast, works great on Workers |

### Testing & CI

| Tool | Purpose |
|------|---------|
| **Vitest** | Unit tests for Workers (Phase 2) |
| **Miniflare** | Local Cloudflare dev environment |
| **GitHub Actions** | CI/CD, cross-platform builds |
| **GoReleaser** | Release automation, binaries, Homebrew |

## System Architecture

### Phase 1: Local-Only

```
┌─────────────────────────────────────────────────────┐
│                     User                             │
└────────────┬────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│                  Samedi CLI                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │  Cobra   │  │BubbleTea │  │  Viper   │          │
│  │ Commands │  │   TUI    │  │  Config  │          │
│  └──────────┘  └──────────┘  └──────────┘          │
└────────────┬────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│                Business Logic                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │   Plan   │  │ Session  │  │ Flashcard│          │
│  │  Manager │  │  Tracker │  │  Manager │          │
│  └──────────┘  └──────────┘  └──────────┘          │
│  ┌──────────┐  ┌──────────┐                        │
│  │   LLM    │  │  Stats   │                        │
│  │Connector │  │Generator │                        │
│  └──────────┘  └──────────┘                        │
└────────────┬────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│                Data Layer                            │
│  ┌──────────────────┐  ┌──────────────────────┐    │
│  │   SQLite Store   │  │   Filesystem Store   │    │
│  │  (sessions.db)   │  │  (plans/, cards/)    │    │
│  └──────────────────┘  └──────────────────────┘    │
└────────────┬────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│            External Dependencies                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │  Claude  │  │  Codex   │  │  Custom  │          │
│  │   CLI    │  │   CLI    │  │  LLM CLI │          │
│  └──────────┘  └──────────┘  └──────────┘          │
└─────────────────────────────────────────────────────┘
```

### Phase 2: Cloud-Enabled

```
┌─────────────────────────────────────────────────────┐
│                    User                              │
│         (Terminal or Web Browser)                    │
└──────┬──────────────────────────────────────────┬───┘
       │                                          │
       ▼                                          ▼
┌──────────────┐                      ┌────────────────────┐
│  Samedi CLI  │◄────────────────────►│  Web Dashboard     │
│   (Local)    │      Sync API        │ (Cloudflare Pages)│
└──────┬───────┘                      └─────────┬──────────┘
       │                                        │
       ▼                                        ▼
┌──────────────────────────────────────────────────────────┐
│            Cloudflare Workers (API)                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │   Auth   │  │   Sync   │  │  Stats   │              │
│  │Middleware│  │  Engine  │  │   API    │              │
│  └──────────┘  └──────────┘  └──────────┘              │
└────────┬─────────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────────────┐
│             Cloudflare Storage                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │      D1      │  │      R2      │  │   KV Store   │  │
│  │  (Sessions)  │  │  (Backups)   │  │  (Auth)      │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└──────────────────────────────────────────────────────────┘
```

## Project Structure

```
samedi/
├── cmd/
│   └── samedi/
│       └── main.go                  # Entry point
│
├── internal/
│   ├── cli/                         # CLI commands
│   │   ├── root.go                  # Root command
│   │   ├── init.go                  # Plan generation
│   │   ├── session.go               # Start/Stop/Status
│   │   ├── review.go                # Flashcard review
│   │   ├── stats.go                 # Statistics
│   │   └── config.go                # Configuration
│   │
│   ├── tui/                         # TUI components
│   │   ├── dashboard.go             # Main dashboard
│   │   ├── review.go                # Flashcard review UI
│   │   ├── stats.go                 # Stats visualization
│   │   └── components/              # Reusable components
│   │       ├── table.go
│   │       ├── progress.go
│   │       └── statusbar.go
│   │
│   ├── plan/                        # Plan management
│   │   ├── manager.go               # Plan CRUD
│   │   ├── parser.go                # Markdown parsing
│   │   └── validator.go             # Format validation
│   │
│   ├── session/                     # Session tracking
│   │   ├── tracker.go               # Session CRUD
│   │   └── timer.go                 # Duration calculation
│   │
│   ├── flashcard/                   # Flashcard system
│   │   ├── manager.go               # Card CRUD
│   │   ├── sm2.go                   # SM-2 algorithm
│   │   └── generator.go             # LLM extraction
│   │
│   ├── llm/                         # LLM integration
│   │   ├── provider.go              # Provider interface
│   │   ├── claude.go                # Claude implementation
│   │   ├── codex.go                 # Codex implementation
│   │   ├── custom.go                # Custom CLI
│   │   └── template.go              # Prompt templates
│   │
│   ├── stats/                       # Statistics
│   │   ├── calculator.go            # Aggregations
│   │   └── exporter.go              # Report generation
│   │
│   ├── storage/                     # Data layer
│   │   ├── sqlite.go                # SQLite operations
│   │   ├── filesystem.go            # File operations
│   │   └── migrations/              # Schema migrations
│   │
│   ├── sync/                        # Cloud sync (Phase 2)
│   │   ├── client.go                # API client
│   │   ├── conflict.go              # Conflict resolution
│   │   └── state.go                 # Sync state tracking
│   │
│   └── config/                      # Configuration
│       ├── config.go                # Config struct
│       └── defaults.go              # Default values
│
├── pkg/                             # Public packages
│   ├── markdown/                    # Markdown utilities
│   └── slug/                        # Slug generation
│
├── templates/                       # LLM prompt templates
│   ├── plan-generation.md
│   ├── flashcard-extraction.md
│   └── quiz-generation.md
│
├── web/                             # Web dashboard (Phase 2)
│   ├── src/
│   │   ├── index.ts                 # Hono app
│   │   ├── routes/                  # API routes
│   │   └── views/                   # HTMX templates
│   ├── wrangler.toml                # Cloudflare config
│   └── package.json
│
├── tests/
│   ├── integration/                 # Integration tests
│   ├── fixtures/                    # Test data
│   └── mocks/                       # Mock LLM providers
│
├── scripts/
│   └── install.sh                   # Installation script
│
├── docs/                            # Documentation
│   └── *.md                         # This file!
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Key Design Patterns

### 1. Repository Pattern (Storage)

```go
// Storage abstraction
type PlanRepository interface {
    Create(plan *Plan) error
    Get(id string) (*Plan, error)
    List(filter PlanFilter) ([]*Plan, error)
    Update(plan *Plan) error
    Delete(id string) error
}

// SQLite implementation
type SQLitePlanRepo struct {
    db *sql.DB
}

// Filesystem implementation (for markdown sync)
type FilesystemPlanRepo struct {
    baseDir string
}
```

### 2. Strategy Pattern (LLM Providers)

```go
type LLMProvider interface {
    Call(prompt string) (string, error)
}

type ClaudeProvider struct { ... }
type CodexProvider struct { ... }
type CustomProvider struct { ... }

// Factory
func NewProvider(config Config) LLMProvider {
    switch config.Provider {
    case "claude":
        return &ClaudeProvider{...}
    case "codex":
        return &CodexProvider{...}
    default:
        return &CustomProvider{...}
    }
}
```

### 3. Command Pattern (CLI)

```go
// Cobra command structure
var rootCmd = &cobra.Command{
    Use:   "samedi",
    Short: "Learning tracking for the terminal",
}

var initCmd = &cobra.Command{
    Use:   "init <topic>",
    Short: "Generate a learning plan",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // Inject dependencies
        svc := NewPlanService(db, llm)
        return svc.Generate(args[0], hours)
    },
}

func init() {
    rootCmd.AddCommand(initCmd)
    initCmd.Flags().IntVar(&hours, "hours", 40, "Total hours")
}
```

### 4. Bubble Tea Model (TUI)

```go
type dashboardModel struct {
    plans    []Plan
    sessions []Session
    cursor   int
}

func (m dashboardModel) Init() tea.Cmd {
    return loadDataCmd
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q":
            return m, tea.Quit
        case "up", "k":
            m.cursor--
        case "down", "j":
            m.cursor++
        }
    }
    return m, nil
}

func (m dashboardModel) View() string {
    // Render UI
}
```

### 5. Dependency Injection

```go
// App struct holds dependencies
type App struct {
    db      *sql.DB
    fs      FileStorage
    llm     LLMProvider
    config  *Config
}

// Services use injected dependencies
type PlanService struct {
    repo PlanRepository
    llm  LLMProvider
}

func NewPlanService(repo PlanRepository, llm LLMProvider) *PlanService {
    return &PlanService{repo: repo, llm: llm}
}
```

## Data Flow Examples

### Plan Generation Flow

```
User: samedi init "french b1" --hours 50

1. CLI (init.go)
   ↓ Parse args, load config
2. PlanService.Generate()
   ↓ Load template, inject vars
3. LLMProvider.Call()
   ↓ Execute claude CLI
4. PlanValidator.Validate()
   ↓ Check markdown format
5. PlanRepository.Create()
   ↓ Save to filesystem + SQLite
6. FlashcardGenerator.Generate()
   ↓ Extract cards from plan
7. CLI output success message
```

### Session Tracking Flow

```
User: samedi start french-b1 chunk-003

1. CLI (session.go)
   ↓ Parse args
2. SessionService.Start()
   ↓ Check for active session
3. SessionRepository.Create()
   ↓ Insert with start_time=NOW(), end_time=NULL
4. PlanRepository.Get()
   ↓ Fetch chunk details
5. CLI output objectives + timer hint

---

User: samedi stop

1. CLI (session.go)
2. SessionService.Stop()
   ↓ Find active session
3. SessionRepository.Update()
   ↓ Set end_time, calculate duration
4. CLI prompt for notes/artifacts
5. SessionRepository.Update()
   ↓ Save metadata
6. CLI output session summary
```

### Flashcard Review Flow

```
User: samedi review

1. CLI (review.go)
   ↓ Launch Bubble Tea TUI
2. ReviewModel.Init()
   ↓ Load due cards from DB
3. ReviewModel.Update() [loop]
   ↓ Show card, wait for rating
4. SM2Algorithm.Update()
   ↓ Calculate new interval
5. FlashcardRepository.Update()
   ↓ Save ease_factor, next_review
6. ReviewModel.View()
   ↓ Render summary
```

## Build & Distribution

### Local Development

```bash
# Install dependencies
go mod download

# Run locally
go run cmd/samedi/main.go

# Run tests
go test ./...

# Run with race detection
go test -race ./...
```

### Building

```makefile
# Makefile
.PHONY: build
build:
	go build -o bin/samedi cmd/samedi/main.go

.PHONY: install
install:
	go install cmd/samedi/main.go

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run
```

### Release (GoReleaser)

```yaml
# .goreleaser.yml
builds:
  - id: samedi
    main: ./cmd/samedi
    binary: samedi
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    files:
      - README.md
      - LICENSE
      - templates/*

brews:
  - name: samedi
    homepage: https://samedi.dev
    description: Learning tracking for the terminal
    tap:
      owner: arbeitandy
      name: homebrew-tap
```

### Installation

```bash
# Homebrew (macOS/Linux)
brew install samedi

# Go install
go install github.com/pezware/samedi.dev/cmd/samedi@latest

# Direct download
curl -fsSL https://samedi.dev/install.sh | sh
```

## Testing Strategy

### Unit Tests

```go
func TestPlanGeneration(t *testing.T) {
    // Mock LLM
    mockLLM := &MockLLM{
        Response: validPlanMarkdown,
    }

    // Mock repository
    mockRepo := &MockPlanRepo{}

    svc := NewPlanService(mockRepo, mockLLM)
    plan, err := svc.Generate("french b1", 50)

    assert.NoError(t, err)
    assert.Equal(t, "french-b1", plan.ID)
    assert.Len(t, plan.Chunks, 50)
}
```

### Integration Tests

```go
func TestEndToEnd_PlanGeneration(t *testing.T) {
    // Real SQLite + filesystem
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Mock LLM only
    mockLLM := &MockLLM{...}

    app := NewApp(db, mockLLM, testConfig)

    // Run command
    err := app.CLI.Execute("init", "french b1")
    assert.NoError(t, err)

    // Verify file exists
    assert.FileExists(t, "~/.samedi/plans/french-b1.md")

    // Verify DB record
    plan, _ := app.PlanRepo.Get("french-b1")
    assert.NotNil(t, plan)
}
```

### TUI Tests

```go
func TestReviewTUI(t *testing.T) {
    model := NewReviewModel(testCards)

    // Simulate keypresses
    model, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace}) // Reveal
    model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}) // Good

    // Assert state
    assert.Equal(t, 1, model.currentIndex)
}
```

## Performance Considerations

### Database

- **Indexes**: On frequently queried fields (plan_id, start_time, next_review)
- **Transactions**: Wrap multi-statement operations
- **Connection Pooling**: Single connection for CLI (short-lived), pool for TUI

### Filesystem

- **Lazy Loading**: Only parse markdown when needed
- **Caching**: Cache parsed plans in memory during TUI sessions
- **Buffered I/O**: Use bufio for large file reads

### LLM Calls

- **Timeouts**: Prevent hanging on slow APIs
- **Retries**: Exponential backoff for transient failures
- **Caching**: Cache LLM outputs for idempotent prompts (optional)

## Security

### Local

- **File Permissions**: `chmod 600` for sensitive files (sessions.db, config)
- **Secrets**: LLM API keys from env vars, never in config files
- **Input Validation**: Sanitize all user inputs (plan IDs, notes)

### Cloud (Phase 2)

- **Auth**: JWT tokens, short-lived (1 hour), refresh tokens in KV
- **TLS**: All API calls over HTTPS
- **Rate Limiting**: Per-user limits on Cloudflare Workers
- **Data Isolation**: Row-level security in D1

## Monitoring & Logging

### Logging

```go
// Structured logging with zerolog
log.Info().
    Str("plan_id", planID).
    Int("duration", duration).
    Msg("Session completed")
```

### Telemetry (Opt-in, Phase 2)

- Anonymous usage stats (command frequency, errors)
- Sent to Cloudflare Analytics
- User can disable in config

## Migration Path (v1 → v2)

### Schema Migrations

```go
// internal/storage/migrations/001_add_mood.sql
ALTER TABLE sessions ADD COLUMN mood TEXT DEFAULT 'neutral';

// Migration runner
func Migrate(db *sql.DB) error {
    for _, migration := range migrations {
        if err := migration.Up(db); err != nil {
            return err
        }
    }
    return nil
}
```

### Data Migrations

```go
// Migrate old plan format to new
func MigratePlansV1ToV2(dir string) error {
    files, _ := filepath.Glob(filepath.Join(dir, "*.md"))
    for _, file := range files {
        content, _ := os.ReadFile(file)
        newContent := transformV1ToV2(content)
        os.WriteFile(file, newContent, 0644)
    }
    return nil
}
```

## Future Considerations

### Phase 3+

- **Real-time Collaboration**: Shared plans (Cloudflare Durable Objects)
- **Voice Input**: Whisper integration for hands-free note-taking
- **Mobile App**: React Native or Flutter with sync API
- **AI Tutor**: Embedded LLM conversation in TUI
