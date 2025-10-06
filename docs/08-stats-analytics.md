# Stats & Analytics

## Overview

Samedi tracks learning progress across multiple dimensions to:
- **Motivate**: Visual progress creates momentum
- **Optimize**: Identify patterns, improve efficiency
- **Celebrate**: Recognize milestones and streaks

## Core Metrics

### 1. Time Tracking

**Total Learning Time**:
```sql
SELECT SUM(duration_minutes) / 60.0 as total_hours
FROM sessions;
```

**Time by Plan**:
```sql
SELECT
    plan_id,
    SUM(duration_minutes) / 60.0 as hours,
    COUNT(*) as session_count,
    AVG(duration_minutes) as avg_session_minutes
FROM sessions
GROUP BY plan_id;
```

**Time by Date Range**:
```sql
SELECT
    DATE(start_time) as date,
    SUM(duration_minutes) / 60.0 as hours
FROM sessions
WHERE start_time >= DATE('now', '-30 days')
GROUP BY DATE(start_time)
ORDER BY date;
```

### 2. Progress Tracking

**Plan Completion**:
```sql
-- Stored in plan metadata
SELECT
    id,
    title,
    chunks_completed,
    chunks_total,
    (chunks_completed * 100.0 / chunks_total) as completion_pct
FROM plans;
```

**Chunk Status Distribution**:
```
not-started: 30 chunks
in-progress: 2 chunks
completed:   18 chunks
skipped:     0 chunks
```

### 3. Streak Tracking

**Current Streak**:
```sql
-- Days with at least one session
WITH daily_activity AS (
    SELECT DISTINCT DATE(start_time) as activity_date
    FROM sessions
    ORDER BY activity_date DESC
)
SELECT COUNT(*) as streak
FROM daily_activity
WHERE activity_date >= (
    SELECT MAX(activity_date) - (ROW_NUMBER() OVER (ORDER BY activity_date DESC) - 1)
    FROM daily_activity
)
AND activity_date = DATE('now') - (ROW_NUMBER() OVER (ORDER BY activity_date DESC) - 1);
```

**Longest Streak**:
```sql
-- Find longest consecutive sequence
WITH RECURSIVE streak_calc AS (
    SELECT
        activity_date,
        1 as streak_length,
        activity_date as streak_start
    FROM daily_activity
    WHERE activity_date = (SELECT MIN(activity_date) FROM daily_activity)

    UNION ALL

    SELECT
        da.activity_date,
        CASE
            WHEN da.activity_date = DATE(sc.activity_date, '+1 day')
            THEN sc.streak_length + 1
            ELSE 1
        END,
        CASE
            WHEN da.activity_date = DATE(sc.activity_date, '+1 day')
            THEN sc.streak_start
            ELSE da.activity_date
        END
    FROM daily_activity da
    JOIN streak_calc sc ON da.activity_date > sc.activity_date
)
SELECT MAX(streak_length) as longest_streak FROM streak_calc;
```

### 4. Flashcard Stats

**Review Performance**:
```sql
SELECT
    plan_id,
    COUNT(*) as total_cards,
    SUM(CASE WHEN next_review <= DATE('now') THEN 1 ELSE 0 END) as due_cards,
    SUM(CASE WHEN repetitions = 0 THEN 1 ELSE 0 END) as new_cards,
    SUM(CASE WHEN repetitions BETWEEN 1 AND 5 THEN 1 ELSE 0 END) as learning_cards,
    SUM(CASE WHEN repetitions > 5 THEN 1 ELSE 0 END) as mature_cards,
    AVG(ease_factor) as avg_ease,
    AVG(interval_days) as avg_interval
FROM cards
GROUP BY plan_id;
```

**Retention Rate** (Phase 2):
```sql
-- Track review ratings over time
CREATE TABLE review_history (
    id TEXT PRIMARY KEY,
    card_id TEXT,
    rating INTEGER,  -- 1-4
    reviewed_at DATETIME,
    FOREIGN KEY (card_id) REFERENCES cards(id)
);

-- Calculate success rate
SELECT
    plan_id,
    COUNT(*) as total_reviews,
    SUM(CASE WHEN rating >= 3 THEN 1 ELSE 0 END) as successful,
    (SUM(CASE WHEN rating >= 3 THEN 1 ELSE 0 END) * 100.0 / COUNT(*)) as success_rate
FROM review_history rh
JOIN cards c ON rh.card_id = c.id
WHERE reviewed_at >= DATE('now', '-30 days')
GROUP BY plan_id;
```

### 5. Learning Velocity

**Hours per Week**:
```sql
SELECT
    STRFTIME('%Y-W%W', start_time) as week,
    SUM(duration_minutes) / 60.0 as hours
FROM sessions
GROUP BY week
ORDER BY week DESC
LIMIT 12;
```

**Sessions per Day of Week**:
```sql
SELECT
    CASE CAST(STRFTIME('%w', start_time) AS INTEGER)
        WHEN 0 THEN 'Sunday'
        WHEN 1 THEN 'Monday'
        WHEN 2 THEN 'Tuesday'
        WHEN 3 THEN 'Wednesday'
        WHEN 4 THEN 'Thursday'
        WHEN 5 THEN 'Friday'
        WHEN 6 THEN 'Saturday'
    END as day_of_week,
    COUNT(*) as session_count,
    SUM(duration_minutes) / 60.0 as total_hours
FROM sessions
GROUP BY STRFTIME('%w', start_time)
ORDER BY CAST(STRFTIME('%w', start_time) AS INTEGER);
```

## Dashboard Views

### 1. Main Dashboard (TUI)

**Command**: `samedi stats` or `samedi` (no args)

```
┌─ Learning Dashboard ───────────────────────────────────────┐
│                                                             │
│  Total Learning Time: 127.5 hours                          │
│  Active Plans: 3 | Completed: 1                            │
│  Current Streak: 12 days 🔥 (Longest: 18 days)             │
│                                                             │
│  This Week:                                                │
│  ████████████████████░░░░░░░░ 18.5 / 25 hours (74%)        │
│                                                             │
│  By Plan:                                                  │
│  french-b1       ████████░░  12h   24%  (12/50 chunks)     │
│  rust-async      ██████████  20h  100%  (20/20 chunks) ✓   │
│  music-theory    ░░░░░░░░░░   0h    0%  (0/30 chunks)      │
│                                                             │
│  Recent Sessions:                                          │
│  2h ago  french-b1 (Chunk 3: Past Tense)       1h 15min    │
│  1d ago  french-b1 (Chunk 2: Present Tense)    1h 30min    │
│  2d ago  rust-async (Chunk 20: Deployment)     2h 00min    │
│                                                             │
│  Flashcards: 205 total | 28 due today                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘

[p] Plans  [s] Sessions  [c] Cards  [w] Weekly  [m] Monthly  [e] Export  [q] Quit
```

**Key Interactions**:
- `p`: Drill into plan details
- `s`: View session history
- `c`: Card statistics
- `w`: Weekly breakdown
- `m`: Monthly summary
- `e`: Export report

### 2. Plan Details

**Command**: `samedi stats french-b1`

```
┌─ French B1 Mastery ────────────────────────────────────────┐
│                                                             │
│  Progress: ████████░░░░░░░░░░ 24% (12/50 chunks)           │
│  Time Spent: 12.5 / 50 hours (25%)                         │
│  Status: in-progress | Created: 2024-01-15                 │
│                                                             │
│  Learning Velocity:                                        │
│  ┌─────────────────────────────────────────────────┐      │
│  │  Week 1:  ███████░░░ 3.5h                       │      │
│  │  Week 2:  █████████░ 4.5h                       │      │
│  │  Week 3:  ████████░░ 4.0h  (current)            │      │
│  │  Week 4:  ░░░░░░░░░░ 0h    (planned)            │      │
│  └─────────────────────────────────────────────────┘      │
│                                                             │
│  Chunk Breakdown:                                          │
│  ✓ Chunk 1: Basic Greetings (1h)        completed          │
│  ✓ Chunk 2: Present Tense Verbs (1.5h)  completed          │
│  → Chunk 3: Past Tense (1h)              in-progress       │
│  ○ Chunk 4: Future Tense (1h)            not-started       │
│  ...                                                        │
│                                                             │
│  Flashcards: 125 cards | 23 due | 76% success rate         │
│                                                             │
│  Sessions: 15 total | Avg: 50 min                          │
│                                                             │
└─────────────────────────────────────────────────────────────┘

[b] Back  [c] Chunks  [s] Sessions  [f] Flashcards  [e] Export
```

### 3. Weekly View

**Command**: `samedi stats --this-week`

```
┌─ This Week (Jan 15-21, 2024) ─────────────────────────────┐
│                                                             │
│  Total: 18.5 hours | Goal: 25 hours (74%)                  │
│                                                             │
│  Daily Breakdown:                                          │
│  ┌─────────────────────────────────────────────────┐      │
│  │  Mon  ████████░░ 2.5h  (3 sessions)             │      │
│  │  Tue  ██████████ 3.0h  (2 sessions)             │      │
│  │  Wed  ████░░░░░░ 1.5h  (1 session)              │      │
│  │  Thu  ████████████ 3.5h  (4 sessions)           │      │
│  │  Fri  ████████████████ 5.0h  (5 sessions) 🔥    │      │
│  │  Sat  ██████████ 3.0h  (2 sessions)             │      │
│  │  Sun  ░░░░░░░░░░ 0h    (0 sessions)             │      │
│  └─────────────────────────────────────────────────┘      │
│                                                             │
│  By Plan:                                                  │
│  french-b1:    10.5h (57%)                                 │
│  music-theory:  5.0h (27%)                                 │
│  rust-async:    3.0h (16%)                                 │
│                                                             │
│  Most Productive Day: Friday (5h)                          │
│  Longest Session: Thursday 3pm (2h 30min)                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘

[t] Today  [w] This Week  [m] This Month  [a] All Time  [b] Back
```

### 4. Monthly Summary

**Command**: `samedi stats --this-month`

```
┌─ January 2024 ────────────────────────────────────────────┐
│                                                             │
│  Total: 52.5 hours | Avg: 1.75 hours/day                   │
│  Sessions: 63 | Avg: 50 minutes/session                    │
│  Active Days: 22/31 (71%) | Streak: 12 days                │
│                                                             │
│  Heatmap:                                                  │
│       M  T  W  T  F  S  S                                  │
│  W1   ░  ░  █  █  █  ░  ░                                  │
│  W2   █  █  █  ░  █  █  ░                                  │
│  W3   █  █  ░  █  █  █  █                                  │
│  W4   █  █  █  █  █  █  ░                                  │
│  W5   █  ░  ░  -  -  -  -                                  │
│                                                             │
│  Legend: ░ 0h  █ >1h  - future                             │
│                                                             │
│  Top Plans:                                                │
│  1. french-b1     25.5h (49%)  +12 chunks                  │
│  2. rust-async    20.0h (38%)  ✓ completed                 │
│  3. music-theory   7.0h (13%)  +3 chunks                   │
│                                                             │
│  Flashcards: 205 total | 1,250 reviews | 78% success       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 5. Flashcard Analytics

**Command**: `samedi stats cards`

```
┌─ Flashcard Statistics ─────────────────────────────────────┐
│                                                             │
│  Total Cards: 205                                          │
│  Due Today: 28 (14%)                                       │
│                                                             │
│  By Status:                                                │
│  ████████████░░░░░░░░ New:      45 (22%)                   │
│  ██████████████░░░░░░ Learning: 80 (39%)                   │
│  ████████████████████ Mature:   80 (39%)                   │
│                                                             │
│  Review Performance (Last 30 Days):                        │
│  Total Reviews: 1,250                                      │
│  Success Rate: 78% (rating >= 3)                           │
│                                                             │
│  Rating Distribution:                                      │
│  Again: 120 (10%)  ██                                      │
│  Hard:  150 (12%)  ███                                     │
│  Good:  750 (60%)  ████████████                            │
│  Easy:  230 (18%)  ████                                    │
│                                                             │
│  Average Ease: 2.42                                        │
│  Average Interval: 15 days                                 │
│                                                             │
│  Top Tags:                                                 │
│  1. verb:      85 cards                                    │
│  2. tense:     60 cards                                    │
│  3. pronoun:   30 cards                                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Export Formats

### 1. Markdown Report

**Command**: `samedi report markdown > report.md`

**Output**:
```markdown
# Learning Report
Generated: 2024-01-20 14:30

## Summary
- **Total Time**: 127.5 hours
- **Active Plans**: 3 (1 completed)
- **Flashcards**: 205 (28 due)
- **Current Streak**: 12 days

## Plans

### French B1 Mastery
- **Status**: in-progress (24% complete)
- **Time**: 12.5 / 50 hours
- **Sessions**: 15 (avg 50 min)
- **Chunks**: 12/50 completed
- **Flashcards**: 125 cards, 76% success rate

#### Recent Sessions
- 2024-01-20 10:00 - Chunk 3: Past Tense (1h 15min)
- 2024-01-19 14:30 - Chunk 2: Present Tense (1h 30min)
- 2024-01-18 16:00 - Chunk 1: Basic Greetings (1h 00min)

### Rust Async/Await
- **Status**: completed ✓
- **Time**: 20 / 20 hours
- **Sessions**: 20 (avg 60 min)
- **Chunks**: 20/20 completed

...

## This Week
| Day | Hours | Sessions | Plans |
|-----|-------|----------|-------|
| Mon | 2.5   | 3        | french-b1, music |
| Tue | 3.0   | 2        | french-b1 |
| ... | ...   | ...      | ... |

## Flashcards
- Total: 205
- Due Today: 28
- Success Rate: 78%

### By Plan
| Plan | Total | Due | Success Rate |
|------|-------|-----|--------------|
| french-b1 | 125 | 23 | 76% |
| rust-async | 80 | 5 | 85% |
```

### 2. JSON Export

**Command**: `samedi report json > stats.json`

**Output**:
```json
{
  "generated_at": "2024-01-20T14:30:00Z",
  "summary": {
    "total_hours": 127.5,
    "active_plans": 3,
    "completed_plans": 1,
    "total_sessions": 63,
    "current_streak": 12,
    "longest_streak": 18,
    "flashcards": {
      "total": 205,
      "due": 28,
      "success_rate": 0.78
    }
  },
  "plans": [
    {
      "id": "french-b1",
      "title": "French B1 Mastery",
      "status": "in-progress",
      "progress": 0.24,
      "hours_spent": 12.5,
      "hours_total": 50,
      "chunks_completed": 12,
      "chunks_total": 50,
      "sessions": 15,
      "avg_session_minutes": 50,
      "flashcards": {
        "total": 125,
        "due": 23,
        "success_rate": 0.76
      }
    }
  ],
  "this_week": {
    "total_hours": 18.5,
    "daily": [
      {"day": "Monday", "hours": 2.5, "sessions": 3},
      {"day": "Tuesday", "hours": 3.0, "sessions": 2}
    ]
  }
}
```

### 3. iCal Export (Phase 2)

**Command**: `samedi export ical > learning.ics`

**Output**:
```ical
BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Samedi//Learning Tracker//EN

BEGIN:VEVENT
UID:session-550e8400@samedi.dev
DTSTART:20240120T100000Z
DTEND:20240120T111500Z
SUMMARY:Learning: French B1 (Chunk 3)
DESCRIPTION:Past Tense\nDuration: 1h 15min\nNotes: Completed exercises
CATEGORIES:Learning,french-b1
END:VEVENT

BEGIN:VEVENT
UID:session-661f9511@samedi.dev
DTSTART:20240119T143000Z
DTEND:20240119T160000Z
SUMMARY:Learning: French B1 (Chunk 2)
DESCRIPTION:Present Tense Verbs\nDuration: 1h 30min
CATEGORIES:Learning,french-b1
END:VEVENT

END:VCALENDAR
```

### 4. CSV Export

**Command**: `samedi export csv sessions > sessions.csv`

**Output**:
```csv
session_id,plan_id,chunk_id,start_time,end_time,duration_minutes,notes
550e8400,french-b1,chunk-003,2024-01-20T10:00:00Z,2024-01-20T11:15:00Z,75,"Completed exercises"
661f9511,french-b1,chunk-002,2024-01-19T14:30:00Z,2024-01-19T16:00:00Z,90,""
```

## Visualization Components

### Progress Bars

**Implementation** (Go + Bubble Tea):
```go
func renderProgressBar(current, total int, width int) string {
    pct := float64(current) / float64(total)
    filled := int(pct * float64(width))

    bar := strings.Repeat("█", filled)
    empty := strings.Repeat("░", width-filled)

    return fmt.Sprintf("%s%s %d%%", bar, empty, int(pct*100))
}

// Usage
renderProgressBar(12, 50, 20)
// Output: ████████░░░░░░░░░░░░ 24%
```

### Heatmap Calendar

```go
func renderHeatmap(sessions []Session) string {
    // Group by date
    byDate := make(map[string]float64)
    for _, s := range sessions {
        date := s.StartTime.Format("2006-01-02")
        byDate[date] += float64(s.DurationMinutes) / 60.0
    }

    // Render grid
    var buf strings.Builder
    buf.WriteString("  M  T  W  T  F  S  S\n")

    for week := 0; week < 4; week++ {
        buf.WriteString(fmt.Sprintf("W%d ", week+1))
        for day := 0; day < 7; day++ {
            date := getDateForWeekDay(week, day)
            hours := byDate[date]

            var cell string
            if hours == 0 {
                cell = "░"
            } else if hours < 1 {
                cell = "▒"
            } else {
                cell = "█"
            }
            buf.WriteString(cell + " ")
        }
        buf.WriteString("\n")
    }

    return buf.String()
}
```

### Sparklines (Trend)

```go
func renderSparkline(values []float64) string {
    max := maxFloat64(values)

    chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

    var buf strings.Builder
    for _, v := range values {
        idx := int((v / max) * float64(len(chars)-1))
        buf.WriteRune(chars[idx])
    }

    return buf.String()
}

// Usage
hoursByWeek := []float64{3.5, 4.5, 4.0, 5.5}
renderSparkline(hoursByWeek)
// Output: ▅▆▅█
```

## Gamification (Phase 3)

### Achievements

```go
type Achievement struct {
    ID          string
    Title       string
    Description string
    Icon        string
    Unlocked    bool
    UnlockedAt  time.Time
}

var achievements = []Achievement{
    {
        ID:    "first-session",
        Title: "First Steps",
        Description: "Complete your first learning session",
        Icon:  "🎯",
    },
    {
        ID:    "week-streak",
        Title: "Week Warrior",
        Description: "Maintain a 7-day learning streak",
        Icon:  "🔥",
    },
    {
        ID:    "hundred-hours",
        Title: "Century",
        Description: "Log 100 hours of learning",
        Icon:  "💯",
    },
    {
        ID:    "polyglot",
        Title: "Polyglot",
        Description: "Learn 3 different languages",
        Icon:  "🌍",
    },
}
```

**Check on Each Session**:
```go
func (s *StatsService) CheckAchievements(userID string) []Achievement {
    unlocked := []Achievement{}

    // Check each achievement
    if s.GetTotalHours() >= 100 && !s.IsUnlocked("hundred-hours") {
        unlocked = append(unlocked, achievements["hundred-hours"])
    }

    if s.GetCurrentStreak() >= 7 && !s.IsUnlocked("week-streak") {
        unlocked = append(unlocked, achievements["week-streak"])
    }

    return unlocked
}
```

### Leaderboards (Optional, Phase 3)

```
┌─ Leaderboards (This Month) ───────────────────────────────┐
│                                                             │
│  Top Learners by Hours:                                    │
│  1. 🥇 alice         52.5h  (85 sessions)                  │
│  2. 🥈 bob           48.0h  (72 sessions)                  │
│  3. 🥉 carol         45.5h  (68 sessions)                  │
│  4.    you (dave)    42.0h  (63 sessions)                  │
│  5.    eve           38.5h  (55 sessions)                  │
│                                                             │
│  Top Plans:                                                │
│  1. Python for Data Science  (125 learners)                │
│  2. French B1                (98 learners)                 │
│  3. Rust Fundamentals        (87 learners)                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Performance Optimization

### Caching

**In-Memory Stats Cache**:
```go
type StatsCache struct {
    mu          sync.RWMutex
    totalHours  float64
    streak      int
    lastUpdated time.Time
    ttl         time.Duration
}

func (c *StatsCache) GetTotalHours() (float64, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    if time.Since(c.lastUpdated) > c.ttl {
        return 0, false  // Cache miss
    }
    return c.totalHours, true
}
```

### Pre-computed Aggregations

**Daily rollup table** (for faster historical queries):
```sql
CREATE TABLE daily_stats (
    date DATE PRIMARY KEY,
    total_hours REAL,
    session_count INTEGER,
    plans_active INTEGER,
    cards_reviewed INTEGER
);

-- Update nightly via cron or on-demand
INSERT OR REPLACE INTO daily_stats
SELECT
    DATE(start_time) as date,
    SUM(duration_minutes) / 60.0 as total_hours,
    COUNT(*) as session_count,
    COUNT(DISTINCT plan_id) as plans_active,
    (SELECT COUNT(*) FROM review_history WHERE DATE(reviewed_at) = DATE(start_time)) as cards_reviewed
FROM sessions
WHERE DATE(start_time) = DATE('now', '-1 day')
GROUP BY DATE(start_time);
```

## Future Analytics (Phase 3+)

### LLM-Powered Insights

**Command**: `samedi insights`

**LLM Prompt**:
```markdown
Analyze this user's learning data and provide insights:

## Stats
- Total: 127.5 hours over 63 sessions
- Active plans: French B1 (24%), Rust Async (100%), Music Theory (0%)
- Streak: 12 days (longest: 18)
- Best day: Friday (avg 5h)
- Worst day: Sunday (avg 0.5h)
- Flashcards: 78% success rate

## Recent Sessions
[Session data...]

Provide:
1. Strengths (what's working well)
2. Weaknesses (what needs improvement)
3. Recommendations (3-5 actionable suggestions)
4. Pace assessment (on track to complete plans?)
```

**Example Output**:
```
🤖 Learning Insights

Strengths:
✓ Strong consistency (12-day streak!)
✓ Excellent Friday productivity (5h avg)
✓ High flashcard success rate (78%)

Weaknesses:
⚠ Music theory plan has 0 progress
⚠ Sunday learning drops to near-zero
⚠ French B1 pace is slower than planned (24% in 4 weeks)

Recommendations:
1. Block 30min Sunday morning for light review to maintain streak
2. Dedicate 1 Friday session to music theory kickstart
3. Increase French B1 to 2h/day to finish in 10 weeks
4. Consider pairing French study with music (French songs?)
5. Use Sunday for flashcard review only (low energy, high value)

Pace: You're trending to complete French B1 in ~16 weeks vs. planned 10.
Action: Add 30min/day or accept extended timeline.
```

### Predictive Analytics

**Forecast completion date**:
```go
func (s *StatsService) ForecastCompletion(planID string) time.Time {
    hoursRemaining := s.GetRemainingHours(planID)
    avgHoursPerWeek := s.GetAvgHoursPerWeek(planID, 4) // Last 4 weeks

    weeksNeeded := hoursRemaining / avgHoursPerWeek
    return time.Now().AddDate(0, 0, int(weeksNeeded*7))
}
```

**Suggest optimal schedule**:
```go
func (s *StatsService) SuggestSchedule(planID string, targetDate time.Time) Schedule {
    hoursRemaining := s.GetRemainingHours(planID)
    daysUntilTarget := int(targetDate.Sub(time.Now()).Hours() / 24)

    hoursPerDay := hoursRemaining / float64(daysUntilTarget)

    // Consider user's productive days
    productiveDays := s.GetProductiveDays(planID)

    return Schedule{
        HoursPerDay: hoursPerDay,
        Recommendation: fmt.Sprintf("Study %.1fh/day on %s to finish by %s",
            hoursPerDay, strings.Join(productiveDays, ", "), targetDate.Format("Jan 2")),
    }
}
```
