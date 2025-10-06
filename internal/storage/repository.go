// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package storage

import (
	"context"
	"time"
)

// PlanRecord represents the metadata row for a learning plan stored in SQLite.
type PlanRecord struct {
	ID         string
	Title      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	TotalHours float64
	Status     string
	Tags       []string
	FilePath   string
}

// PlanFilter provides optional filtering when listing plans.
type PlanFilter struct {
	IDs      []string
	Statuses []string
	Tag      string
	SortBy   string
}

// PlanRepository defines storage operations for plan metadata.
type PlanRepository interface {
	Upsert(ctx context.Context, plan *PlanRecord) error
	Get(ctx context.Context, id string) (*PlanRecord, error)
	List(ctx context.Context, filter *PlanFilter) ([]*PlanRecord, error)
	Delete(ctx context.Context, id string) error
}

// SessionRecord mirrors the sessions table schema.
type SessionRecord struct {
	ID              string
	PlanID          string
	ChunkID         string
	StartTime       time.Time
	EndTime         *time.Time
	DurationMinutes int
	Notes           string
	Artifacts       []string
	CardsCreated    int
	CreatedAt       time.Time
}

// SessionFilter provides options for querying sessions.
type SessionFilter struct {
	PlanID    string
	ChunkID   string
	Active    bool
	Limit     int
	SinceTime *time.Time
}

// SessionRepository defines storage operations for learning sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *SessionRecord) error
	Get(ctx context.Context, id string) (*SessionRecord, error)
	List(ctx context.Context, filter *SessionFilter) ([]*SessionRecord, error)
	Update(ctx context.Context, session *SessionRecord) error
	Delete(ctx context.Context, id string) error
}

// CardRecord mirrors the cards table schema for spaced repetition data.
type CardRecord struct {
	ID           string
	PlanID       string
	ChunkID      string
	Question     string
	Answer       string
	Tags         []string
	CreatedAt    time.Time
	EaseFactor   float64
	IntervalDays int
	Repetitions  int
	NextReview   time.Time
	LastReview   *time.Time
}

// CardFilter provides options for selecting cards to review.
type CardFilter struct {
	PlanID    string
	DueBefore *time.Time
	Tags      []string
	Limit     int
}

// CardRepository defines storage operations for flashcards.
type CardRepository interface {
	Upsert(ctx context.Context, card *CardRecord) error
	Get(ctx context.Context, id string) (*CardRecord, error)
	List(ctx context.Context, filter *CardFilter) ([]*CardRecord, error)
	Delete(ctx context.Context, id string) error
}

// Storage combines database and filesystem storage.
type Storage struct {
	DB         *SQLiteDB
	Filesystem *FilesystemStorage
	Paths      *Paths
}

// NewStorage creates a new storage instance.
func NewStorage(dbPath string, paths *Paths) (*Storage, error) {
	// Initialize database
	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		return nil, err
	}

	// Run migrations
	migrator := NewMigrator(db)
	if err := migrator.Migrate(); err != nil {
		db.Close()
		return nil, err
	}

	// Initialize filesystem
	fs := NewFilesystemStorage(paths)
	if err := fs.Initialize(); err != nil {
		db.Close()
		return nil, err
	}

	return &Storage{
		DB:         db,
		Filesystem: fs,
		Paths:      paths,
	}, nil
}

// Close closes all storage connections.
func (s *Storage) Close() error {
	return s.DB.Close()
}
