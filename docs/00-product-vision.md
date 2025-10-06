# Product Vision: Samedi

## Problem Statement

Developers and lifelong learners face a fundamental challenge: **learning is fragmented across tools, time is poorly tracked, and progress is invisible**.

When learning technical subjects (or anything really), we:
- Jump between resources without a clear curriculum
- Lose track of what we've studied and for how long
- Have no mechanism to retain knowledge (flashcards scattered, if they exist)
- Can't visualize progress across multiple learning domains
- Struggle to maintain motivation without seeing cumulative effort

Existing solutions are either:
- **Too heavyweight**: LMS platforms, Anki with complex setup, notion databases
- **Wrong paradigm**: GUI-first tools that break developer flow
- **Single-domain**: Language apps, coding platforms, but nothing universal

## Vision

**Samedi is a learning operating system for the terminal.**

A CLI-native tool where:
- **LLMs are your tutors** - Claude/Codex design curricula, not rigid software
- **Terminal is your classroom** - Learn where you already work
- **Time is the unit** - Track effort, not just completion
- **Knowledge compounds** - Flashcards and stats accumulate across domains
- **Progress is visible** - Beautiful TUI dashboard of your learning journey

### The North Star

> "A developer can learn **anything** - Rust, French, music theory - with the same tool, see all their learning in one place, and feel inspired by their progress."

## Target Users

### Primary: Technical Learners
- Developers learning new languages/frameworks
- Engineers exploring adjacent fields (ML, systems, web3)
- Terminal-first workflow preference
- Already use LLM CLIs (Claude Code, Codex, llm)

### Secondary: Polymaths
- Technical professionals branching into music, languages, literature
- Self-directed learners who create their own curricula
- People who value time-tracking and quantified self

### Non-Users (Out of Scope for MVP)
- Students needing institutional LMS features
- Teams/organizations (solo learners only)
- People who need multimedia (video courses, audio lessons)

## Success Metrics

### Qualitative
- Users say: "I finally see all my learning in one place"
- Users learn across 3+ different domains with samedi
- Users maintain 30+ day streaks

### Quantitative (Post-MVP)
- 100 hours tracked per user in first 3 months
- 500+ flashcards accumulated per user
- 80% of started plans reach 50% completion

## Core Principles

### 1. **CLI-First, Always**
The TUI is for tracking and reminders. Learning happens in your editor, terminal, LLM chat.

### 2. **LLM as Curriculum Designer**
Samedi doesn't know how to teach French. Claude does. Samedi orchestrates.

### 3. **Time-Boxed Learning**
30-60 minute chunks. Manageable, measurable, motivating.

### 4. **Markdown Everywhere**
Plans, notes, exports - all human-readable, git-trackable, tool-agnostic.

### 5. **Offline-First, Sync-Optional**
Works on a plane. Cloudflare sync is convenience, not dependency.

### 6. **Boring Technology**
Go binary, SQLite, markdown files. No databases to manage, no servers to run (until phase 2).

## Non-Goals (What We Won't Build)

### MVP Phase
- ❌ Social features (sharing, leaderboards, communities)
- ❌ Multimedia support (video, audio lessons)
- ❌ Built-in content (courses, tutorials)
- ❌ Collaborative learning (group plans, shared decks)
- ❌ Mobile apps (terminal + web only)
- ❌ AI tutoring within samedi (use external LLMs)

### Philosophical
- ❌ We're not a course platform
- ❌ We're not a productivity tool (use existing task managers)
- ❌ We're not a social network
- ❌ We're not replacing your tools, we're coordinating them

## Product Phases

### Phase 1: Local MVP (Months 1-2)
- CLI for plan generation, session tracking, flashcard review
- TUI for stats and progress visualization
- Markdown plans, SQLite sessions, local storage
- LLM integration (configurable CLI calls)

### Phase 2: Cloud Sync (Months 3-4)
- Cloudflare Workers API
- Multi-device sync (laptop ↔ desktop)
- Web dashboard for mobile viewing
- Backup and export

### Phase 3: Intelligence (Months 5-6)
- Spaced repetition algorithms
- Adaptive quizzing via LLM
- Learning insights and recommendations
- Calendar integration (iCal export)

## Key Differentiators

| Feature | Samedi | Anki | Notion | Course Platforms |
|---------|--------|------|--------|------------------|
| Terminal-native | ✅ | ❌ | ❌ | ❌ |
| Multi-domain | ✅ | ⚠️ | ✅ | ❌ |
| LLM-powered curricula | ✅ | ❌ | ⚠️ | ❌ |
| Time tracking | ✅ | ❌ | ⚠️ | ⚠️ |
| Developer workflow | ✅ | ❌ | ⚠️ | ❌ |
| Markdown-based | ✅ | ❌ | ⚠️ | ❌ |

## Inspiration

- **Ledger CLI**: Plain text accounting, powerful queries
- **TaskWarrior**: Terminal task management with beautiful stats
- **Anki**: Spaced repetition, but make it terminal-native
- **GitHub Contributions**: Visual progress that motivates
- **Zettelkasten**: Interconnected knowledge, not siloed courses

## User Mantras

> "I track my learning the way I track my code commits"

> "My LLM designs the curriculum, samedi tracks my progress"

> "All my learning - Rust, French, guitar - in one dashboard"

> "I don't break flow to track learning, it just happens"
