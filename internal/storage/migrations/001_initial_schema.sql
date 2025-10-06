-- Initial schema for samedi database
-- Creates tables for plans, sessions, and flashcards

-- Plans table (metadata only, full plan is in markdown)
CREATE TABLE IF NOT EXISTS plans (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    total_hours REAL,
    status TEXT NOT NULL CHECK(status IN ('not-started', 'in-progress', 'completed', 'archived')),
    tags TEXT, -- JSON array
    file_path TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_plans_status ON plans(status);
CREATE INDEX IF NOT EXISTS idx_plans_created ON plans(created_at);

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    plan_id TEXT NOT NULL,
    chunk_id TEXT,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    duration_minutes INTEGER,
    notes TEXT,
    artifacts TEXT, -- JSON array of URLs/paths
    cards_created INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_plan ON sessions(plan_id);
CREATE INDEX IF NOT EXISTS idx_sessions_start ON sessions(start_time);
CREATE INDEX IF NOT EXISTS idx_sessions_chunk ON sessions(chunk_id);

-- Flashcards table (also stored in markdown, this is for scheduling)
CREATE TABLE IF NOT EXISTS cards (
    id TEXT PRIMARY KEY,
    plan_id TEXT NOT NULL,
    chunk_id TEXT,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    tags TEXT, -- JSON array
    created_at DATETIME NOT NULL,

    -- Spaced repetition (SM-2 algorithm)
    ease_factor REAL DEFAULT 2.5,
    interval_days INTEGER DEFAULT 1,
    repetitions INTEGER DEFAULT 0,
    next_review DATE NOT NULL,
    last_review DATE,

    FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_cards_review ON cards(next_review);
CREATE INDEX IF NOT EXISTS idx_cards_plan ON cards(plan_id);

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO schema_migrations (version) VALUES (1);
