# Samedi

**A learning operating system for the terminal.**

[![CI](https://github.com/pezware/samedi.dev/workflows/CI/badge.svg)](https://github.com/pezware/samedi.dev/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/pezware/samedi.dev)](https://goreportcard.com/report/github.com/pezware/samedi.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## What is Samedi?

Samedi is a CLI-native tool for tracking and managing your learning journey across any domain—programming, languages, music, or anything else you want to master.

### Key Features

- **🤖 LLM-Powered Curricula**: Generate learning plans with Claude, Codex, or any LLM CLI
- **⏱️ Time Tracking**: Track sessions with simple `start`/`stop` commands
- **🎴 Spaced Repetition**: Built-in flashcard system with SM-2 algorithm
- **📊 Progress Dashboard**: Beautiful TUI to visualize your learning journey
- **📝 Markdown-Based**: All plans in git-trackable markdown
- **🔄 Cloud Sync**: Optional multi-device sync via Cloudflare (Phase 2)

### The Philosophy

> "Samedi doesn't teach. LLMs do. Samedi orchestrates, tracks, and motivates."

Learn anything with the same workflow:
```bash
samedi init "rust async programming" --hours 40
samedi start rust-async chunk-001
# ... learn for 1 hour ...
samedi stop
samedi review  # Flashcard review
samedi stats   # See your progress
```

---

## Quick Start

### Installation

**Homebrew** (macOS/Linux):
```bash
brew install samedi
```

**Go Install**:
```bash
go install github.com/pezware/samedi.dev/cmd/samedi@latest
```

**From Source**:
```bash
git clone https://github.com/pezware/samedi.dev.git
cd samedi.dev
make install
```

### First Use

1. **Generate a learning plan**:
   ```bash
   samedi init "french b1" --hours 50
   ```

2. **Start learning**:
   ```bash
   samedi start french-b1 chunk-001
   # ... study for an hour ...
   samedi stop
   ```

3. **Review flashcards**:
   ```bash
   samedi review french-b1
   ```

4. **Check your progress**:
   ```bash
   samedi stats
   ```

---

## Documentation

- **[Product Vision](./docs/00-product-vision.md)** - Why Samedi exists
- **[User Journeys](./docs/01-user-journeys.md)** - Real-world usage scenarios
- **[CLI Reference](./docs/05-cli-tui-design.md)** - Complete command documentation
- **[Architecture](./docs/04-architecture.md)** - Technical design
- **[Development Guide](./CLAUDE.md)** - Contributing guidelines

Full documentation: [docs/](./docs/)

---

## Features

### Phase 1: Local MVP (Current)

- ✅ **Plan Generation**: LLM-powered curriculum design
- ✅ **Session Tracking**: Start/stop learning sessions
- ✅ **Flashcards**: SM-2 spaced repetition
- ✅ **Stats Dashboard**: TUI with progress visualization
- ✅ **Markdown Plans**: Human-readable, git-trackable
- ✅ **SQLite Storage**: Fast, local, reliable

### Phase 2: Cloud Sync (Planned)

- 🔄 **Multi-Device Sync**: Cloudflare Workers + D1
- 📱 **Web Dashboard**: Mobile-friendly stats viewer
- ☁️ **Cloud Backups**: Automatic to Cloudflare R2
- 🔐 **Email Auth**: Magic link authentication

### Phase 3: Intelligence (Future)

- 🧠 **Adaptive Learning**: LLM-powered insights
- 🎯 **Smart Quizzing**: Personalized tests
- 📅 **Calendar Integration**: iCal export
- 🏆 **Achievements**: Gamification (optional)

---

## Examples

### Learning Rust
```bash
samedi init "rust async programming" --hours 40
samedi start rust-async chunk-001
# Code along with Claude Code...
samedi stop --note "Built async web server"
samedi cards generate rust-async chunk-001  # Extract flashcards
```

### Learning French
```bash
samedi init "french b1" --hours 50
samedi start french-b1 chunk-003
# Study with Duolingo, practice with Codex...
samedi stop
samedi review french-b1  # Review vocab flashcards
```

### Cross-Domain Learning
```bash
samedi plan list
# rust-async     100%  ✓
# french-b1       68%  in-progress
# music-theory    25%  in-progress

samedi stats --all
# Total: 187.5 hours across 3 domains
# Streak: 42 days 🔥
```

---

## Architecture

**Tech Stack**:
- **Language**: Go 1.21+
- **TUI**: Bubble Tea
- **CLI**: Cobra
- **Database**: SQLite
- **LLM Integration**: Configurable (Claude, Codex, llm, etc.)

**Project Structure**:
```
samedi/
├── cmd/samedi/           # Main entry point
├── internal/             # Core application logic
│   ├── cli/             # CLI commands
│   ├── tui/             # TUI components
│   ├── plan/            # Plan management
│   ├── session/         # Session tracking
│   ├── flashcard/       # Spaced repetition
│   └── llm/             # LLM integration
├── docs/                 # Documentation
└── templates/            # LLM prompts
```

See [Architecture Documentation](./docs/04-architecture.md) for details.

---

## Contributing

We welcome contributions! Please read:

- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - Contribution guidelines
- **[CLAUDE.md](./CLAUDE.md)** - Development workflow and standards

### Quick Development Setup

```bash
# Clone and setup
git clone https://github.com/pezware/samedi.dev.git
cd samedi
make install-tools

# Run tests
make test

# Build and run
make build
./bin/samedi
```

---

## License

MIT License - see [LICENSE](./LICENSE) for details.

---

## Acknowledgments

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [SQLite](https://www.sqlite.org/) - Database
- [goldmark](https://github.com/yuin/goldmark) - Markdown parser

---

**Start your learning journey today:**

```bash
brew install samedi
samedi init "your next skill"
```

Happy learning! 🎓
