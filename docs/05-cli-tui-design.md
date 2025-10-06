# CLI & TUI Design

## Design Principles

1. **Terminal-native**: Works beautifully in tmux, SSH, and local terminal
2. **Stateless commands**: Each command completes independently
3. **Pipe-friendly**: Outputs can be piped to other tools
4. **Progressive disclosure**: Simple by default, powerful with flags
5. **No surprises**: Destructive operations require confirmation

## Command Structure

### Top-Level Commands

```bash
samedi [command] [subcommand] [options]
```

### Command Categories

| Category | Commands | Purpose |
|----------|----------|---------|
| **Planning** | `init`, `plan` | Create and manage learning plans |
| **Sessions** | `start`, `stop`, `status` | Track learning time |
| **Review** | `review`, `cards` | Flashcard practice |
| **Stats** | `stats`, `report` | Progress visualization |
| **Management** | `config`, `sync`, `backup` | System operations |

## Command Reference

### 1. Plan Management

#### `samedi init <topic>`

Create a new learning plan with LLM assistance.

**Usage**:
```bash
samedi init "french b1"
samedi init rust-async --hours 20
samedi init "music theory basics" --model claude-opus
```

**Flow**:
1. Prompt user for details (if not provided):
   - Total hours (default: 40)
   - Learning style (beginner/intermediate/advanced)
   - Specific goals/focus areas
2. Generate prompt from template
3. Call configured LLM CLI
4. Save output to `~/.samedi/plans/{slug}.md`
5. Parse and index in SQLite
6. Generate initial flashcards
7. Print plan location and first chunk

**Options**:
- `--hours <n>`: Total estimated hours (default: 40)
- `--model <name>`: LLM model override
- `--template <path>`: Custom prompt template
- `--no-cards`: Skip flashcard generation
- `--edit`: Open plan in $EDITOR before saving

**Output**:
```
✓ Plan created: ~/.samedi/plans/french-b1.md
✓ Generated 50 chunks (50 hours total)
✓ Created 25 flashcards

Next: samedi start french-b1 chunk-001
```

#### `samedi plan list`

List all learning plans.

**Usage**:
```bash
samedi plan list
samedi plan list --status in-progress
samedi plan list --tag language
```

**Output**:
```
ID             TITLE                STATUS        PROGRESS    HOURS
french-b1      French B1 Mastery    in-progress   24% (12/50) 12/50h
rust-async     Rust Async/Await     completed     100% (20/20) 20/20h
music-theory   Music Theory Basics  not-started   0% (0/30)    0/30h
```

**Options**:
- `--status <status>`: Filter by status
- `--tag <tag>`: Filter by tag
- `--sort <field>`: Sort by created, updated, progress
- `--json`: Output as JSON

#### `samedi plan show <plan-id>`

Show plan details and progress.

**Usage**:
```bash
samedi plan show french-b1
samedi plan show french-b1 --chunks
```

**Output**:
```
French B1 Mastery
Status: in-progress | Progress: 24% (12/50 chunks)
Created: 2024-01-15 | Updated: 2024-01-20
Total: 50 hours | Spent: 12.5 hours | Remaining: 37.5 hours

Recent chunks:
✓ Chunk 1: Basic Greetings (1h) - completed
✓ Chunk 2: Present Tense Verbs (1.5h) - completed
→ Chunk 3: Past Tense (1h) - in-progress
  Chunk 4: Future Tense (1h) - not-started

Next: samedi start french-b1 chunk-003
```

**Options**:
- `--chunks`: Show all chunks
- `--sessions`: Show session history
- `--cards`: Show flashcard count

#### `samedi plan edit <plan-id>`

Open plan in $EDITOR.

**Usage**:
```bash
samedi plan edit french-b1
```

**Flow**:
1. Open `~/.samedi/plans/french-b1.md` in $EDITOR
2. On save, validate markdown structure
3. Update SQLite metadata
4. Regenerate flashcards if chunks changed

#### `samedi plan archive <plan-id>`

Archive a completed or abandoned plan.

**Usage**:
```bash
samedi plan archive french-b1
```

### 2. Session Tracking

#### `samedi start <plan-id> [chunk-id]`

Start a learning session.

**Usage**:
```bash
samedi start french-b1
samedi start french-b1 chunk-003
samedi start rust-async chunk-015 --note "Working on tokio tutorial"
```

**Flow**:
1. Check for active session → error if exists
2. Create session record with start_time
3. Display session info and timer hint
4. Remind user of chunk objectives (if chunk-id provided)

**Output**:
```
→ Session started: french-b1 (Chunk 3: Past Tense)

Objectives:
- Master passé composé with avoir
- Learn 15 irregular past participles
- Construct 20 example sentences

Timer running. Stop with: samedi stop
```

**Options**:
- `--note <text>`: Add initial note
- `--silent`: No output (for scripting)

#### `samedi stop [--note "text"]`

Stop the active session.

**Usage**:
```bash
samedi stop
samedi stop --note "Completed all exercises"
samedi stop --artifact "github.com/user/french-practice"
```

**Flow**:
1. Find active session
2. Calculate duration
3. Prompt for notes (if not provided)
4. Prompt for artifacts (optional)
5. Update session record
6. Show summary

**Output**:
```
✓ Session completed: french-b1 (Chunk 3)
Duration: 1h 15min
Notes: Completed all exercises

Add artifact URL/path? (optional): github.com/user/french-practice
Create flashcards from this session? (y/N):
```

**Options**:
- `--note <text>`: Session notes
- `--artifact <url>`: Add learning artifact
- `--no-cards`: Skip flashcard prompt
- `--auto`: Skip all prompts, use defaults

#### `samedi status`

Show active session status.

**Usage**:
```bash
samedi status
```

**Output (active)**:
```
→ Active session: french-b1 (Chunk 3: Past Tense)
Started: 2024-01-20 10:00
Elapsed: 45 minutes

Stop with: samedi stop
```

**Output (no active session)**:
```
No active session.

Recent:
- french-b1 (Chunk 2) - 1h 30min - 2 hours ago

Start: samedi start <plan-id>
```

#### `samedi pause` / `samedi resume`

Pause and resume active session (Phase 2).

### 3. Flashcard Review

#### `samedi review [plan-id]`

Review flashcards due today.

**Usage**:
```bash
samedi review                    # All cards due
samedi review french-b1          # Plan-specific
samedi review --new 5            # Include 5 new cards
```

**Flow**:
1. Query cards WHERE next_review <= TODAY
2. Launch TUI review interface
3. Show card, wait for user rating
4. Update SM-2 algorithm values
5. Save to markdown + SQLite
6. Show summary

**TUI Interface**:
```
┌─ Flashcard Review ─────────────────────────────────────────┐
│                                                             │
│  Q: How do you say "Good morning" formally in French?      │
│                                                             │
│  [Press SPACE to reveal answer]                            │
│                                                             │
│  Progress: 5/23 cards | Streak: 12 days                    │
└─────────────────────────────────────────────────────────────┘

[After reveal]
┌─ Flashcard Review ─────────────────────────────────────────┐
│                                                             │
│  Q: How do you say "Good morning" formally in French?      │
│                                                             │
│  A: Bonjour                                                │
│                                                             │
│  Rate difficulty:                                          │
│  [1] Again  [2] Hard  [3] Good  [4] Easy                   │
│                                                             │
│  Progress: 5/23 cards | Streak: 12 days                    │
└─────────────────────────────────────────────────────────────┘
```

**Output (summary)**:
```
✓ Review complete!

Reviewed: 23 cards in 8 minutes
Again: 2 | Hard: 5 | Good: 12 | Easy: 4

Next review: 15 cards tomorrow
```

**Options**:
- `--new <n>`: Include N new cards
- `--limit <n>`: Review max N cards
- `--tag <tag>`: Review specific tag

#### `samedi cards list [plan-id]`

List flashcards.

**Usage**:
```bash
samedi cards list
samedi cards list french-b1
samedi cards list --due
```

**Output**:
```
PLAN        TOTAL   DUE    NEW    LEARNING   MATURE
french-b1   125     23     15     45         65
rust-async  80      5      0      20         60

Total: 205 cards | 28 due today
```

#### `samedi cards add <plan-id> [chunk-id]`

Manually add flashcard.

**Usage**:
```bash
samedi cards add french-b1
samedi cards add french-b1 chunk-003 --tag verb
```

**Flow**:
1. Prompt for question
2. Prompt for answer
3. Prompt for tags (optional)
4. Save to markdown + SQLite
5. Set initial SM-2 values

**Interactive**:
```
Question: What is the past participle of "avoir"?
Answer: eu
Tags (comma-separated): verb, irregular, avoir
✓ Card created (french-b1 #126)
```

#### `samedi cards generate <plan-id> [chunk-id]`

Generate flashcards from plan with LLM.

**Usage**:
```bash
samedi cards generate french-b1 chunk-003
```

**Flow**:
1. Extract chunk content from plan markdown
2. Call LLM with flashcard-extraction template
3. Parse LLM output (Q&A pairs)
4. Preview cards, allow editing
5. Save approved cards

### 4. Stats & Progress

#### `samedi stats [plan-id]`

Show learning statistics.

**Usage**:
```bash
samedi stats                     # All plans
samedi stats french-b1           # Specific plan
samedi stats --this-week         # Time filter
```

**TUI Dashboard**:
```
┌─ Learning Stats ───────────────────────────────────────────┐
│                                                             │
│  Total Learning Time: 127 hours                            │
│  Active Plans: 3 | Completed: 1                            │
│  Current Streak: 12 days 🔥                                │
│                                                             │
│  This Week:                                                │
│  ████████████████████░░░░░░░░ 18.5 / 25 hours              │
│                                                             │
│  By Plan:                                                  │
│  french-b1     ████████░░  12h  24%                        │
│  rust-async    ██████████  20h  100% ✓                     │
│  music-theory  ░░░░░░░░░░   0h  0%                         │
│                                                             │
│  Recent Sessions:                                          │
│  2024-01-20  french-b1 (Chunk 3)    1h 15min               │
│  2024-01-19  french-b1 (Chunk 2)    1h 30min               │
│  2024-01-18  rust-async (Chunk 20)  2h 00min               │
│                                                             │
└─────────────────────────────────────────────────────────────┘

[q] Quit  [p] Plans  [w] Weekly  [m] Monthly  [e] Export
```

**Options**:
- `--today`: Today only
- `--this-week`: Current week
- `--this-month`: Current month
- `--since <date>`: From date
- `--json`: JSON output

#### `samedi report <format>`

Generate learning report.

**Usage**:
```bash
samedi report markdown > report.md
samedi report json > stats.json
samedi report ical > learning.ics     # Phase 2
```

**Markdown Output**:
```markdown
# Learning Report
Generated: 2024-01-20

## Summary
- Total time: 127 hours
- Active plans: 3
- Flashcards: 205 (28 due)
- Streak: 12 days

## Plans
### French B1 Mastery (24% complete)
- Time spent: 12 hours / 50 hours
- Sessions: 15
- Chunks completed: 12/50
...
```

### 5. System Management

#### `samedi config`

Manage configuration.

**Usage**:
```bash
samedi config list
samedi config set llm.provider claude
samedi config get llm.provider
samedi config edit                  # Open in $EDITOR
```

#### `samedi sync`

Sync with Cloudflare (Phase 2).

**Usage**:
```bash
samedi sync                         # Two-way sync
samedi sync push                    # Upload only
samedi sync pull                    # Download only
```

#### `samedi backup`

Create local backup.

**Usage**:
```bash
samedi backup
samedi backup --to ~/Dropbox/samedi-backup.tar.gz
```

**Output**:
```
✓ Backup created: ~/samedi-backups/samedi-2024-01-20.tar.gz
Size: 2.3 MB
Contains: 3 plans, 205 cards, 127 sessions
```

#### `samedi check`

Validate data integrity.

**Usage**:
```bash
samedi check
samedi check --fix                  # Auto-repair if possible
```

**Output**:
```
Checking data integrity...
✓ Plans: 3 valid
✓ Sessions: 127 valid
✓ Cards: 205 valid
✗ Orphaned cards: 2 (no matching plan)

Run with --fix to repair.
```

### 6. Quick Access

#### `samedi` (no args)

Launch TUI dashboard.

**Usage**:
```bash
samedi
```

**TUI Main Menu**:
```
┌─ Samedi ───────────────────────────────────────────────────┐
│                                                             │
│  What did you learn today?                                 │
│                                                             │
│  [1] Start Session                                         │
│  [2] Review Flashcards (23 due)                            │
│  [3] View Stats                                            │
│  [4] Manage Plans                                          │
│  [5] Sync & Backup                                         │
│                                                             │
│  Recent:                                                   │
│  → french-b1 (Chunk 3) - 1h 15min - 2 hours ago            │
│                                                             │
│  Streak: 12 days 🔥                                        │
└─────────────────────────────────────────────────────────────┘

[q] Quit  [?] Help
```

#### Aliases & Shortcuts

```bash
samedi s french-b1      # Alias: samedi start french-b1
samedi r                # Alias: samedi review
samedi st               # Alias: samedi stats
```

## TUI Design Patterns

### Color Scheme

**Dracula (default)**:
- Background: `#282a36`
- Foreground: `#f8f8f2`
- Primary: `#bd93f9` (purple)
- Success: `#50fa7b` (green)
- Warning: `#ffb86c` (orange)
- Error: `#ff5555` (red)
- Accent: `#8be9fd` (cyan)

### Components

**Progress Bar**:
```
████████████████░░░░░░░░ 60%
```

**Table**:
```
┌────────┬─────────┬────────┐
│ Plan   │ Time    │ Status │
├────────┼─────────┼────────┤
│ french │ 12h/50h │ 24%    │
└────────┴─────────┴────────┘
```

**Status Indicators**:
- `✓` Completed
- `→` In progress
- `○` Not started
- `✗` Error
- `🔥` Streak

### Keyboard Navigation

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate |
| `Enter` | Select |
| `q` or `Esc` | Back/Quit |
| `?` | Help |
| `r` | Refresh |
| `/` | Search |
| `n/p` | Next/Previous page |

## Error Handling

### Graceful Failures

**No active session**:
```
✗ No active session to stop.

Start: samedi start <plan-id>
```

**LLM CLI error**:
```
✗ Failed to generate plan: claude CLI not found

Install: brew install claude
Configure: samedi config set llm.cli_command claude
```

**Invalid plan format**:
```
✗ Plan format error: french-b1.md
Line 15: Missing '**Duration**: ...' in chunk header

Fix manually: samedi plan edit french-b1
Or validate: samedi check --fix
```

### User Confirmations

**Destructive operations**:
```bash
$ samedi plan delete french-b1
⚠ Delete plan 'French B1 Mastery'?
  This will also delete 125 flashcards and 15 sessions.

  Type plan ID to confirm: french-b1
✓ Plan deleted
```

## Shell Integration

### Completions

**Bash/Zsh**:
```bash
samedi completion bash > /usr/local/etc/bash_completion.d/samedi
```

**Fish**:
```bash
samedi completion fish > ~/.config/fish/completions/samedi.fish
```

### Prompt Integration

**Zsh** (show active session):
```zsh
# .zshrc
RPROMPT='$(samedi status --format prompt)'
# Output: 🎓 french-b1 (45m)
```

### Tmux Integration

**Status line**:
```tmux
# .tmux.conf
set -g status-right '#(samedi status --format tmux)'
# Output: 🎓 45m | Streak: 12🔥
```

## API for Scripting

### JSON Output

All commands support `--json` flag:

```bash
samedi stats --json | jq '.total_hours'
# Output: 127

samedi plan list --json | jq -r '.[] | select(.status=="in-progress") | .id'
# Output: french-b1
#         music-theory
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | LLM CLI error |
| 4 | Data validation error |
| 5 | Active session conflict |
| 6 | Not found (plan/session) |

## Future Commands (Phase 2+)

- `samedi insights` - LLM-powered learning insights
- `samedi quiz <plan-id>` - Adaptive quizzing
- `samedi export anki` - Export to Anki format
- `samedi import anki` - Import Anki decks
- `samedi share <plan-id>` - Generate shareable link
- `samedi streak freeze` - Protect streak (vacation mode)
