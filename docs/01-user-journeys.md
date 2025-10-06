# User Journeys

## Personas

### 1. Alex - Software Engineer Learning Rust
- **Background**: 5 years of Python, wants to learn systems programming
- **Goals**: Master Rust for backend services at work
- **Constraints**: 1-2 hours/day after work, prefers terminal workflow
- **Tools**: Uses claude-code, VS Code, tmux

### 2. Maria - Polyglot Learner
- **Background**: Fluent in English/Spanish, learning French
- **Goals**: B2 proficiency for job in Paris
- **Constraints**: Irregular schedule (freelancer), learns on laptop + phone
- **Tools**: Uses codex for curricula, Anki for flashcards

### 3. Jordan - Music Student
- **Background**: Hobbyist guitarist, wants to learn music theory
- **Goals**: Understand chord progressions, improvisation
- **Constraints**: Practices 30min/day before work
- **Tools**: Uses claude, watches YouTube, reads books

## Journey 1: Alex Learns Rust (Day 1)

### Morning: Create Learning Plan

**8:00 AM - First Use**

```bash
# Install samedi
brew install samedi

# Generate learning plan
samedi init "rust async programming" --hours 40

# Samedi calls claude-code...
```

**LLM Conversation** (Claude Code):
```
Claude: I'll create a 40-hour Rust async programming curriculum.
        What's your experience level with Rust?

Alex: I know basics, but async/await is new to me.

Claude: Got it. I'll focus on async fundamentals, tokio, and practical projects.
        Here's your learning plan...

[Generates 40 1-hour chunks]
```

**Output**:
```
✓ Plan created: ~/.samedi/plans/rust-async.md
✓ Generated 40 chunks (40 hours total)
✓ Created 30 flashcards

Next: samedi start rust-async chunk-001
```

**Alex reviews plan**:
```bash
samedi plan show rust-async
```

```
Rust Async Programming
Progress: ░░░░░░░░░░░░░░░░░░░░ 0% (0/40 chunks)
Total: 40 hours | Estimated: 8 weeks

First 5 chunks:
○ Chunk 1: Async Basics (1h) - Futures, async/await syntax
○ Chunk 2: Tokio Runtime (1h) - Runtime setup, spawning tasks
○ Chunk 3: Channels (1h) - mpsc, broadcast, oneshot
○ Chunk 4: Error Handling (1h) - Result, ? operator in async
○ Chunk 5: Mini Project (1h) - Async web scraper

Next: samedi start rust-async chunk-001
```

### Evening: First Learning Session

**6:30 PM - Start Session**

```bash
samedi start rust-async chunk-001
```

```
→ Session started: rust-async (Chunk 1: Async Basics)

Objectives:
- Understand the Future trait
- Learn async/await syntax
- Write basic async functions

Resources:
- Rust Async Book: Chapter 1
- Tokio Tutorial: Getting Started

Timer running. Stop with: samedi stop
```

**7:00 PM - Learning in Claude Code**

Alex opens VS Code, uses claude-code to work through async examples:

```rust
// Asks claude-code for explanations
async fn fetch_data() -> Result<String, Error> {
    // claude-code explains: This returns a Future<Output = Result<...>>
    let response = reqwest::get("https://api.example.com")
        .await?;
    Ok(response.text().await?)
}
```

**7:30 PM - Stop Session**

```bash
samedi stop
```

```
✓ Session completed: rust-async (Chunk 1)
Duration: 1h 00min

Notes: Struggled with lifetimes in async functions, need more practice
Artifacts (URLs/paths): github.com/alex/rust-async-practice

Create flashcards from this session? (y/N): y
```

**LLM generates flashcards**:
```
Analyzing session content...
✓ Created 5 flashcards:
  - What trait represents async computation? → Future
  - How to run async code? → Use async runtime (tokio, async-std)
  - Difference between async fn and fn() → Future? → async fn is syntax sugar
  ...

samedi review to practice!
```

## Journey 2: Alex's First Week (Days 1-7)

### Day 2-5: Building Momentum

**Daily Routine**:
```bash
# Morning: Review flashcards (5 min)
samedi review rust-async

# Evening: Learning session (1 hour)
samedi start rust-async chunk-002
# ... learn for 1 hour ...
samedi stop

# Check progress
samedi stats
```

**Day 5 Stats**:
```
┌─ Learning Dashboard ───────────────────────────────────────┐
│                                                             │
│  Total Learning Time: 5.5 hours                            │
│  Active Plans: 1                                           │
│  Current Streak: 5 days 🔥                                 │
│                                                             │
│  This Week:                                                │
│  ████████████░░░░░░░░ 5.5 / 10 hours (55%)                 │
│                                                             │
│  rust-async:                                               │
│  ████░░░░░░░░░░░░░░░░ 5.5h  14% (5/40 chunks)              │
│                                                             │
│  Flashcards: 45 total | 12 due today                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Day 6: Weekend Deep Dive

**10:00 AM - Extended Session**

```bash
samedi start rust-async chunk-006
# 2.5 hour session on building async web server
samedi stop
```

```
✓ Session completed: rust-async (Chunk 6)
Duration: 2h 30min

Notes: Built complete async API server with routes, middleware, DB
Artifacts: github.com/alex/rust-api-server
```

### Day 7: Review & Export

**Morning: Stats Review**

```bash
samedi stats rust-async --this-week
```

```
┌─ This Week (Rust Async) ──────────────────────────────────┐
│                                                             │
│  Total: 10.5 hours | Goal: 10 hours (105%) ✓               │
│                                                             │
│  Daily:                                                    │
│  Mon  ██████░░░░ 1.0h                                      │
│  Tue  ██████░░░░ 1.0h                                      │
│  Wed  ██████░░░░ 1.0h                                      │
│  Thu  ████████░░ 1.5h                                      │
│  Fri  ██████░░░░ 1.0h                                      │
│  Sat  ████████████████████████ 5.0h  🔥                    │
│  Sun  ░░░░░░░░░░ 0h (today)                                │
│                                                             │
│  Progress: 15% (6/40 chunks)                               │
│  On track to finish in: 7 weeks                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Export for Sharing**

```bash
samedi report markdown > weekly-report.md
cat weekly-report.md
```

```markdown
# Rust Async Learning - Week 1
Progress: 15% (6/40 chunks) | Time: 10.5 hours

## Completed Chunks
1. Async Basics (1h) ✓
2. Tokio Runtime (1h) ✓
3. Channels (1.5h) ✓
4. Error Handling (1h) ✓
5. Mini Project (1h) ✓
6. Web Server (2.5h) ✓

## Key Learnings
- Future trait is core to async
- Tokio handles runtime, tasks, I/O
- Built working API server!

## Next Week
- Chunks 7-13: Advanced patterns, streams, async iterators
- Goal: 10 hours
```

Alex shares on LinkedIn/Twitter to celebrate progress 🎉

## Journey 3: Maria Learns French (Multi-Device)

### Phase 1: Local Learning (Weeks 1-2)

**Laptop Workflow**:
```bash
# Generate plan
samedi init "french b1" --hours 50

# Daily: 1-hour sessions
samedi start french-b1 chunk-001
# ... study with codex ...
samedi stop

# Review flashcards
samedi review french-b1
```

**After 2 Weeks**:
```
Total: 12 hours
Progress: 24% (12/50 chunks)
Flashcards: 125 cards, 78% retention
```

### Phase 2: Enable Cloud Sync (Week 3)

**One-Time Setup**:
```bash
samedi login maria@example.com
```

```
✉️  Magic link sent to maria@example.com
Check your email and paste the token here: abc123def456

✓ Logged in successfully
```

```bash
samedi sync push --all
```

```
Uploading all data to cloud...
✓ 12 plans synced
✓ 18 sessions synced
✓ 125 flashcards synced

Sync enabled. Will auto-sync every 30min.
```

### Phase 3: Mobile Access (Week 4+)

**Morning - Desktop**:
```bash
samedi start french-b1 chunk-015
# ... 1 hour session ...
samedi stop

# Auto-syncs to cloud
```

**Lunch - Phone (Web Dashboard)**:

Opens `https://samedi.dev/dashboard` on phone:

```
📱 Samedi Dashboard

Total: 28.5 hours | Streak: 18 days 🔥

french-b1
████████████░░░░░░░░ 60% (30/50 chunks)

Recent:
2h ago - Chunk 15: Subjunctive Mood (1h)

Review 15 flashcards due →
```

Taps "Review 15 flashcards", completes on phone.

**Evening - Desktop**:
```bash
samedi stats
```

```
# Shows flashcard review from phone!
Last review: 15 cards (12:30 PM) - 87% success
```

## Journey 4: Jordan's Music Theory (Focused Learning)

### Week 1: Intensive Start

**Monday - Plan Creation**:
```bash
samedi init "music theory for guitarists" --hours 30
```

LLM (claude) generates focused 30-hour plan:
- Chunks 1-10: Fundamentals (scales, intervals, chords)
- Chunks 11-20: Harmony (progressions, voice leading)
- Chunks 21-30: Application (analysis, improvisation)

**Daily Routine** (6:30-7:00 AM before work):

```bash
samedi start music-theory chunk-001
# 30min: Read chapter, practice on guitar
samedi stop
```

**Compact Sessions**:
```
Day 1: Chunk 1 - Major Scale (30min)
Day 2: Chunk 2 - Intervals (30min)
Day 3: Chunk 3 - Triads (30min)
...
```

### Week 2: Integration with Practice

**TUI Dashboard** (launched from tmux):

```bash
samedi
```

```
┌─ Samedi ───────────────────────────────────────────────────┐
│                                                             │
│  🎸 Music Theory for Guitarists                            │
│                                                             │
│  Progress: ████████░░░░░░░░░░ 40% (12/30 chunks)           │
│  Time: 6.5 / 30 hours                                      │
│  Streak: 12 days 🔥                                        │
│                                                             │
│  Today's Focus:                                            │
│  → Chunk 13: ii-V-I Progressions                           │
│                                                             │
│  [1] Start Session                                         │
│  [2] Review 8 flashcards                                   │
│  [3] View Stats                                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

Presses `1` → starts session, tmux splits to claude for help.

### Month 1: Completion

**Final Stats**:
```bash
samedi stats music-theory --all-time
```

```
Music Theory for Guitarists - Complete ✓

Total Time: 30 hours (4 weeks)
Sessions: 60 (30min avg)
Flashcards: 180 cards, 82% retention

Consistency:
- 28/30 days active (93%)
- Longest streak: 18 days

Knowledge Retained:
- 180 concepts mastered
- 95% success on final review

Export certificate: samedi export certificate music-theory
```

**Share Achievement**:
```bash
samedi export certificate music-theory
```

Generates shareable image:
```
┌─────────────────────────────────────────┐
│  🎵 Certificate of Completion           │
│                                         │
│  Jordan completed:                      │
│  Music Theory for Guitarists            │
│                                         │
│  30 hours | 30 chunks | 180 flashcards │
│  Completed: Feb 1, 2024                │
│                                         │
│  samedi.dev/verify/abc123              │
└─────────────────────────────────────────┘
```

## Journey 5: Cross-Domain Learning (Alex - 6 Months Later)

### Learning Multiple Topics

**Active Plans**:
```bash
samedi plan list
```

```
ID              TITLE                    STATUS        PROGRESS
rust-async      Rust Async/Await         completed     100% ✓
french-b1       French B1                in-progress   68%
music-guitar    Guitar Improvisation     in-progress   25%
ml-basics       Machine Learning Basics  not-started   0%
```

### Unified Dashboard

```bash
samedi stats --all
```

```
┌─ All Learning ─────────────────────────────────────────────┐
│                                                             │
│  Total Time: 187.5 hours (6 months)                        │
│  Plans: 4 (1 completed, 2 active, 1 planned)               │
│  Longest Streak: 42 days 🔥                                │
│                                                             │
│  By Domain:                                                │
│  Tech:     ████████████████░░ 120h  (64%)                  │
│  Language: ████████░░░░░░░░░░  60h  (32%)                  │
│  Music:    ██░░░░░░░░░░░░░░░░   7.5h (4%)                  │
│                                                             │
│  This Month:                                               │
│  french-b1:    12h (weekly goal: 3h) ✓                     │
│  music-guitar:  6h (weekly goal: 2h) ✓                     │
│                                                             │
│  Flashcards: 450 total | 52 due today                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Tag-Based Review

```bash
# Review only verb flashcards across all languages
samedi review --tag verb
```

```
Reviewing 'verb' cards (35 cards):
- french-b1: 20 cards
- spanish: 15 cards (imported from Anki)

[Starts mixed review]
```

### Export Portfolio

```bash
samedi report markdown --all > learning-portfolio.md
```

Uses in resume/LinkedIn:
```markdown
# Continuous Learning Portfolio

## 2024 Progress
- **187.5 hours** of structured learning
- **4 domains**: Programming, Language, Music, ML
- **450 concepts** mastered via spaced repetition

### Completed Courses
1. **Rust Async Programming** (40h)
   - Built production-ready async services
   - [GitHub Projects](github.com/alex/rust-projects)

2. [In Progress] **French B1** (68% complete)
   - Conversational fluency achieved
   - 200+ flashcards mastered

...
```

## Journey 6: Team Learning (WONTDO - Phase 3)

### Shared Plan (Engineering Team)

**Lead creates plan**:
```bash
samedi init "kubernetes deep dive" --hours 60
samedi plan share k8s-deep-dive --team engineering
```

**Team members join**:
```bash
samedi plan join k8s-deep-dive
```

**Collaborative Progress**:
```
┌─ Kubernetes Deep Dive (Team Plan) ────────────────────────┐
│                                                             │
│  Team Progress: ████████░░░░░░░░░░ 42% (25/60 chunks)      │
│                                                             │
│  Top Contributors:                                         │
│  1. alice    12h  (8 chunks)                               │
│  2. bob      10h  (7 chunks)                               │
│  3. you       8h  (5 chunks)                               │
│  4. carol     6h  (3 chunks)                               │
│                                                             │
│  Shared Flashcards: 180 cards                              │
│  Team Success Rate: 81%                                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Common Workflows

### Morning Review Routine

```bash
# Quick stats check
samedi

# Review due cards (5-10 min)
samedi review

# Check what's next
samedi plan show <plan-id>
```

### Evening Learning Session

```bash
# Start session
samedi start <plan> <chunk>

# Learn in preferred environment:
# - claude-code for coding
# - codex for explanations
# - Browser for videos/articles
# - Books + notes

# Stop and reflect
samedi stop
# → Prompted for notes, artifacts
# → Optional flashcard generation
```

### Weekly Review

```bash
# Export weekly report
samedi report markdown --this-week > week-$(date +%V).md

# Check progress vs. goals
samedi stats --this-week

# Plan next week
samedi plan show <plan-id>
```

### Backup & Sync

```bash
# Local backup (weekly)
samedi backup

# Cloud sync (automatic if enabled)
samedi sync

# Manual sync
samedi sync push   # Upload changes
samedi sync pull   # Download changes
```

## Edge Cases & Error Handling

### Forgot to Stop Session

```bash
samedi status
```

```
→ Active session: french-b1 (Chunk 3)
Started: Yesterday 10:00 PM (22 hours ago!)

⚠ This seems unusually long. Did you forget to stop?

Options:
[1] Stop now (22h duration)
[2] Set custom end time
[3] Delete session
```

### Lost Progress (No Backup)

```bash
samedi check
```

```
Checking data integrity...
✓ Plans: 3 valid
✓ Sessions: 125 valid
✗ Orphaned cards: 5 (plan 'spanish' not found)

Recommendations:
1. Restore plan from cloud: samedi sync pull
2. Or create new plan: samedi init spanish
3. Reassign cards: samedi cards reassign spanish → spanish-a1
```

### Sync Conflict

```bash
samedi sync
```

```
⚠ Sync conflict detected:

Local:  french-b1 updated 2 hours ago (desktop)
Remote: french-b1 updated 1 hour ago (laptop)

Both modified Chunk 15 objectives.

Resolution:
[1] Keep local (overwrite remote)
[2] Keep remote (overwrite local)
[3] Merge (manual)
[4] Keep both as separate plans

Choice: 3

Opening merge editor...
```
