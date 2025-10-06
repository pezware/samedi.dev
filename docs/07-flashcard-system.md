# Flashcard System

## Overview

Samedi uses **spaced repetition** to optimize long-term knowledge retention. The system combines:
- **SM-2 Algorithm**: Proven spaced repetition scheduling
- **LLM-Generated Cards**: Automatic extraction from learning content
- **Manual Creation**: User-added cards for custom needs
- **Cross-Platform Sync**: Desktop ↔ Mobile (Phase 2)

## SM-2 Algorithm

### Background

SuperMemo 2 (SM-2) is a simple, effective spaced repetition algorithm used by Anki and many others.

**Core Idea**: Cards you know well are shown less frequently; cards you struggle with are shown more often.

### Algorithm Details

**Variables per card**:
- `ease_factor` (EF): Difficulty multiplier (default: 2.5)
- `interval`: Days until next review (default: 1)
- `repetitions`: Consecutive successful reviews

**User Rating** (1-4):
- `1 - Again`: Completely forgot
- `2 - Hard`: Barely remembered
- `3 - Good`: Remembered with effort
- `4 - Easy`: Instantly recalled

**Update Rules**:

```
If rating >= 3 (Good or Easy):
    interval = interval * ease_factor
    repetitions += 1
Else (Again or Hard):
    interval = 1
    repetitions = 0

ease_factor = ease_factor + (0.1 - (3 - rating) * (0.08 + (3 - rating) * 0.02))
ease_factor = max(1.3, ease_factor)  // Floor at 1.3
```

**Special Cases**:
- First review: interval = 1 day
- Second review: interval = 6 days
- Third+ review: interval = previous_interval * ease_factor

### Implementation

```go
type Card struct {
    ID           string
    Question     string
    Answer       string
    EaseFactor   float64  // 1.3 - 2.5+
    Interval     int      // Days
    Repetitions  int
    NextReview   time.Time
    LastReview   time.Time
}

func (c *Card) Review(rating int) {
    if rating >= 3 {
        // Success
        if c.Repetitions == 0 {
            c.Interval = 1
        } else if c.Repetitions == 1 {
            c.Interval = 6
        } else {
            c.Interval = int(float64(c.Interval) * c.EaseFactor)
        }
        c.Repetitions++
    } else {
        // Failure
        c.Interval = 1
        c.Repetitions = 0
    }

    // Update ease factor
    c.EaseFactor = c.EaseFactor + (0.1 - float64(3-rating) * (0.08 + float64(3-rating) * 0.02))
    if c.EaseFactor < 1.3 {
        c.EaseFactor = 1.3
    }

    c.LastReview = time.Now()
    c.NextReview = c.LastReview.AddDate(0, 0, c.Interval)
}
```

### Example Progression

**Card**: "What is the French word for 'hello'?" → "Bonjour"

| Review | Rating | Interval | Next Review |
|--------|--------|----------|-------------|
| 1 | 3 (Good) | 1 day | Tomorrow |
| 2 | 3 (Good) | 6 days | 6 days from now |
| 3 | 4 (Easy) | 15 days | 15 days from now |
| 4 | 3 (Good) | 38 days | 38 days from now |
| 5 | 2 (Hard) | 1 day | Tomorrow |
| 6 | 3 (Good) | 6 days | 6 days from now |

## Card Types

### 1. Auto-Generated (from LLM)

**Source**: Learning plan chunks

**Generation Flow**:
```
User: samedi cards generate french-b1 chunk-003

1. Extract chunk content from plan markdown
2. Call LLM with flashcard-extraction template
3. Parse JSON response (Q&A pairs)
4. Preview cards to user
5. User approves/edits/deletes
6. Save to markdown + SQLite
```

**Example LLM Output**:
```json
[
  {
    "question": "What is the passé composé of 'avoir'?",
    "answer": "j'ai eu, tu as eu, il a eu",
    "tags": ["verb", "avoir", "passe-compose"]
  },
  {
    "question": "List 3 irregular past participles",
    "answer": "été (être), eu (avoir), fait (faire)",
    "tags": ["verb", "irregular", "past-participle"]
  }
]
```

**Prompt Template** (`flashcard-extraction.md`):
```markdown
Extract 5-10 high-quality flashcards from this content.

## Content
{{.ChunkContent}}

## Guidelines
- Focus on key concepts, not trivia
- Use clear, concise language
- Include context in questions when needed
- Tag appropriately (concepts, difficulty, topic)

## Output Format
JSON array:
[
  {
    "question": "...",
    "answer": "...",
    "tags": ["tag1", "tag2"]
  }
]
```

### 2. Manual Creation

**Flow**:
```
User: samedi cards add french-b1

Question: What is the difference between "tu" and "vous"?
Answer: "Tu" is informal (friends, family), "vous" is formal (strangers, elders) or plural
Tags (comma-separated): pronoun, formality
✓ Card created (french-b1 #126)
```

**Use Cases**:
- Specific facts LLM missed
- Personal mnemonics
- Corrections to auto-generated cards
- Quick additions during review

### 3. Imported (Phase 2)

**From Anki**:
```
User: samedi import anki < french-deck.txt

✓ Imported 150 cards
- 120 new
- 30 duplicates (skipped)
```

**From Markdown**:
```
User: cat external-cards.md | samedi import cards --plan french-b1
```

## Card Storage

### Markdown Format

**File**: `~/.samedi/cards/{plan-id}.cards.md`

```markdown
# French B1 Flashcards

## Card 1 {#card-001}
**Q**: What is the passé composé of "avoir"?
**A**: j'ai eu, tu as eu, il a eu

**Tags**: verb, avoir, passe-compose
**Source**: Chunk 3
**Created**: 2024-01-16
**Ease**: 2.5
**Interval**: 6
**Repetitions**: 2
**Next Review**: 2024-01-22

---

## Card 2 {#card-002}
**Q**: List 3 irregular past participles
**A**: été (être), eu (avoir), fait (faire)

**Tags**: verb, irregular, past-participle
**Source**: Chunk 3
**Created**: 2024-01-16
**Ease**: 2.3
**Interval**: 1
**Repetitions**: 0
**Next Review**: 2024-01-17

---
```

**Why Markdown?**:
- Human-readable
- Git-trackable
- Portable (import to Anki, Notion, etc.)
- Editable in any text editor

### SQLite Schema

```sql
CREATE TABLE cards (
    id TEXT PRIMARY KEY,
    plan_id TEXT NOT NULL,
    chunk_id TEXT,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    tags TEXT,                        -- JSON array: ["verb", "avoir"]
    source TEXT,                      -- "Chunk 3" or "Manual"

    -- SM-2 fields
    ease_factor REAL DEFAULT 2.5,
    interval_days INTEGER DEFAULT 1,
    repetitions INTEGER DEFAULT 0,
    next_review DATE NOT NULL,
    last_review DATE,

    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE INDEX idx_cards_review ON cards(next_review);
CREATE INDEX idx_cards_plan ON cards(plan_id);
CREATE INDEX idx_cards_tags ON cards(tags);  -- JSON search
```

**Sync Strategy**:
- Markdown is source of truth
- SQLite for fast queries (due cards, stats)
- On edit, update both

## Review Interface (TUI)

### Review Mode

**Launch**:
```bash
samedi review                    # All cards due today
samedi review french-b1          # Plan-specific
samedi review --new 5            # Include 5 new cards
samedi review --tag verb         # Tag-specific
```

**UI Flow**:

**1. Card List (before review)**:
```
┌─ Review ───────────────────────────────────────────────────┐
│                                                             │
│  23 cards due today                                        │
│                                                             │
│  New: 5                                                    │
│  Learning: 8                                               │
│  Review: 10                                                │
│                                                             │
│  Press ENTER to start                                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**2. Question Phase**:
```
┌─ Review (5/23) ────────────────────────────────────────────┐
│                                                             │
│  Q: What is the passé composé of "avoir"?                  │
│                                                             │
│                                                             │
│  [Press SPACE to reveal answer]                            │
│                                                             │
│                                                             │
│  Tags: verb, avoir, passe-compose                          │
│  Plan: french-b1 (Chunk 3)                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**3. Answer Phase**:
```
┌─ Review (5/23) ────────────────────────────────────────────┐
│                                                             │
│  Q: What is the passé composé of "avoir"?                  │
│                                                             │
│  A: j'ai eu, tu as eu, il a eu                             │
│                                                             │
│  How well did you know this?                               │
│                                                             │
│  [1] Again  [2] Hard  [3] Good  [4] Easy                   │
│                                                             │
│  Next: 1d | 6d | 15d | 38d                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**4. Summary**:
```
┌─ Review Complete ──────────────────────────────────────────┐
│                                                             │
│  ✓ Reviewed 23 cards in 8 minutes                          │
│                                                             │
│  Again: 2 (9%)                                             │
│  Hard:  5 (22%)                                            │
│  Good: 12 (52%)                                            │
│  Easy:  4 (17%)                                            │
│                                                             │
│  Next review: 15 cards tomorrow                            │
│                                                             │
│  [q] Quit  [r] Retry failed cards                          │
└─────────────────────────────────────────────────────────────┘
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Space` | Reveal answer |
| `1` | Rate: Again |
| `2` | Rate: Hard |
| `3` | Rate: Good |
| `4` | Rate: Easy |
| `e` | Edit card |
| `d` | Delete card |
| `s` | Skip card |
| `?` | Help |
| `q` | Quit |

### Edit During Review

```
[User presses 'e' during review]

Edit Card #42:

Question: What is the passé composé of "avoir"?
[Editable field]

Answer: j'ai eu, tu as eu, il a eu
[Editable field]

Tags: verb, avoir, passe-compose
[Editable field]

[Save] [Cancel]
```

## Card Management

### List Cards

```bash
samedi cards list
samedi cards list french-b1
samedi cards list --due
samedi cards list --tag verb
```

**Output**:
```
PLAN        TOTAL   DUE    NEW    LEARNING   MATURE
french-b1   125     23     15     45         65
rust-async  80      5      0      20         60
music       50      10     5      15         30

Total: 255 cards | 38 due today
```

### Search Cards

```bash
samedi cards search "past participle"
samedi cards search --tag verb --plan french-b1
```

**Output**:
```
Found 8 cards:

#42: What is the past participle of "avoir"?
     Plan: french-b1 | Tags: verb, avoir | Due: today

#87: List 3 irregular past participles
     Plan: french-b1 | Tags: verb, irregular | Due: 2024-01-22

...
```

### Export Cards

#### To Anki

```bash
samedi export anki french-b1 > french-deck.txt
```

**Format** (Anki tab-separated):
```
What is the passé composé of "avoir"?	j'ai eu, tu as eu, il a eu	verb avoir passe-compose
List 3 irregular past participles	été (être), eu (avoir), fait (faire)	verb irregular
```

#### To Markdown

```bash
samedi export cards french-b1 > backup.md
```

**Format**: Same as storage format (human-readable)

#### To CSV

```bash
samedi export cards --format csv french-b1 > cards.csv
```

```csv
id,question,answer,tags,ease_factor,interval,next_review
card-001,"What is...",  "j'ai eu...", "verb,avoir", 2.5, 6, 2024-01-22
...
```

## Card Statistics

### Per-Plan Stats

```bash
samedi stats cards french-b1
```

**Output**:
```
French B1 Flashcards

Total: 125 cards
Due: 23 (18%)
New: 15 (12%)
Learning: 45 (36%)
Mature: 65 (52%)

Average ease: 2.4
Average interval: 18 days

Success rate (last 30 days): 76%

Tags:
  verb: 45 cards
  pronoun: 20 cards
  tense: 30 cards
```

### Review Heatmap (Phase 2)

```
Reviews Last 30 Days:

Mon Tue Wed Thu Fri Sat Sun
 5   12  8   15  20  3   0   ← Week 1
 10  18  12  9   22  5   2   ← Week 2
 ...
```

## Advanced Features (Phase 2+)

### Cloze Deletions

**Input**:
```
Question: The capital of France is {{c1::Paris}}.
Answer: [Cloze card - answer hidden in question]
```

**Review**:
```
Q: The capital of France is _______.
A: Paris
```

### Image Cards

**Input**:
```markdown
## Card 42 {#card-042}
**Q**: What is this chord?
![](./images/c-major-chord.png)

**A**: C Major
```

**Review**: Display image in TUI (using kitty/iTerm2 inline images)

### Audio Cards (Music Learning)

```markdown
## Card 15 {#card-015}
**Q**: What interval is this?
🔊 [interval-perfect-fifth.mp3]

**A**: Perfect fifth
```

### Reverse Cards

**Auto-generate reverse**:
```
Original: Q: French for "hello"? A: Bonjour
Reverse:  Q: English for "bonjour"? A: Hello
```

**Config**:
```toml
[flashcards]
auto_reverse = ["language", "vocabulary"]  # Tags that get reversed
```

## Data Migration

### Anki Import

```bash
samedi import anki < my-deck.txt --plan imported-anki
```

**Mapping**:
- Front → Question
- Back → Answer
- Anki tags → Samedi tags
- Anki due date → next_review
- Anki ease → ease_factor

### Export to Anki

```bash
samedi export anki french-b1 > deck.txt
```

Import to Anki: File → Import → deck.txt

## Testing Strategy

### Unit Tests

**SM-2 Algorithm**:
```go
func TestSM2_GoodRating(t *testing.T) {
    card := &Card{
        EaseFactor: 2.5,
        Interval:   1,
        Repetitions: 1,
    }

    card.Review(3) // Good

    assert.Equal(t, 6, card.Interval)
    assert.Equal(t, 2, card.Repetitions)
    assert.Equal(t, 2.5, card.EaseFactor)
}

func TestSM2_FailureResets(t *testing.T) {
    card := &Card{
        EaseFactor: 2.3,
        Interval:   15,
        Repetitions: 3,
    }

    card.Review(1) // Again

    assert.Equal(t, 1, card.Interval)
    assert.Equal(t, 0, card.Repetitions)
    assert.Less(t, card.EaseFactor, 2.3) // Decreased
}
```

### Integration Tests

**Full Review Flow**:
```go
func TestReviewFlow(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Create cards
    cards := []*Card{
        {Question: "Q1", Answer: "A1", NextReview: today},
        {Question: "Q2", Answer: "A2", NextReview: today},
    }
    for _, c := range cards {
        db.CreateCard(c)
    }

    // Run review
    svc := NewReviewService(db)
    due := svc.GetDueCards()
    assert.Len(t, due, 2)

    // Rate cards
    svc.RateCard(cards[0].ID, 3) // Good
    svc.RateCard(cards[1].ID, 1) // Again

    // Verify updates
    c1, _ := db.GetCard(cards[0].ID)
    assert.Equal(t, 6, c1.Interval)

    c2, _ := db.GetCard(cards[1].ID)
    assert.Equal(t, 1, c2.Interval)
}
```

## Performance

### Query Optimization

**Due cards query**:
```sql
-- Fast: Uses index on next_review
SELECT * FROM cards
WHERE next_review <= DATE('now')
ORDER BY next_review
LIMIT 100;

-- Slow: Full table scan
SELECT * FROM cards
WHERE created_at > DATE('now', '-30 days')
ORDER BY created_at;
```

### Caching

**In-memory cache for active review**:
```go
type ReviewSession struct {
    cards  []*Card
    index  int
    cache  map[string]*Card  // Fast lookup
}

func (s *ReviewSession) RateCard(id string, rating int) {
    card := s.cache[id]  // O(1) lookup
    card.Review(rating)
    s.db.UpdateCard(card)
}
```

## Future Enhancements

### Phase 3+

1. **Adaptive Algorithm**: FSRS (Free Spaced Repetition Scheduler) for better predictions
2. **Collaborative Decks**: Share card decks with other users
3. **AI Hints**: LLM-generated hints for failed cards
4. **Voice Review**: Speak answers, AI verifies correctness
5. **Gamification**: Streaks, achievements, leaderboards
