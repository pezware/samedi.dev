// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/pezware/samedi.dev/internal/storage"
)

// Repository defines the interface for session persistence.
type Repository interface {
	// Create inserts a new session.
	Create(ctx context.Context, session *Session) error

	// Get retrieves a session by ID.
	Get(ctx context.Context, id string) (*Session, error)

	// GetActive retrieves the currently active session (end_time IS NULL).
	// Returns nil if no active session exists.
	GetActive(ctx context.Context) (*Session, error)

	// Update updates an existing session.
	Update(ctx context.Context, session *Session) error

	// List retrieves sessions for a plan, ordered by start time descending.
	List(ctx context.Context, planID string, limit int) ([]*Session, error)

	// GetByPlan retrieves all sessions for a specific plan.
	GetByPlan(ctx context.Context, planID string) ([]*Session, error)

	// Delete removes a session by ID.
	Delete(ctx context.Context, id string) error
}

// SQLiteRepository implements session storage using SQLite.
type SQLiteRepository struct {
	db *storage.SQLiteDB
}

// NewSQLiteRepository creates a new SQLite-backed session repository.
func NewSQLiteRepository(db *storage.SQLiteDB) Repository {
	return &SQLiteRepository{db: db}
}

// Create inserts a new session into the database.
func (r *SQLiteRepository) Create(ctx context.Context, session *Session) error {
	artifactsJSON, err := json.Marshal(session.Artifacts)
	if err != nil {
		return fmt.Errorf("failed to marshal artifacts: %w", err)
	}

	query := `
		INSERT INTO sessions (
			id, plan_id, chunk_id, start_time, end_time, duration_minutes,
			notes, artifacts, cards_created, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.DB().ExecContext(ctx, query,
		session.ID,
		session.PlanID,
		nullString(session.ChunkID),
		session.StartTime,
		nullTime(session.EndTime),
		session.Duration,
		session.Notes,
		string(artifactsJSON),
		session.CardsCreated,
		session.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// Get retrieves a session by ID.
func (r *SQLiteRepository) Get(ctx context.Context, id string) (*Session, error) {
	query := `
		SELECT id, plan_id, chunk_id, start_time, end_time, duration_minutes,
			notes, artifacts, cards_created, created_at
		FROM sessions
		WHERE id = ?
	`

	session, err := r.scanSession(r.db.DB().QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// GetActive retrieves the currently active session (end_time IS NULL).
func (r *SQLiteRepository) GetActive(ctx context.Context) (*Session, error) {
	query := `
		SELECT id, plan_id, chunk_id, start_time, end_time, duration_minutes,
			notes, artifacts, cards_created, created_at
		FROM sessions
		WHERE end_time IS NULL
		ORDER BY start_time DESC
		LIMIT 1
	`

	session, err := r.scanSession(r.db.DB().QueryRowContext(ctx, query))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // No active session is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}

	return session, nil
}

// Update updates an existing session.
func (r *SQLiteRepository) Update(ctx context.Context, session *Session) error {
	artifactsJSON, err := json.Marshal(session.Artifacts)
	if err != nil {
		return fmt.Errorf("failed to marshal artifacts: %w", err)
	}

	query := `
		UPDATE sessions
		SET plan_id = ?, chunk_id = ?, start_time = ?, end_time = ?,
			duration_minutes = ?, notes = ?, artifacts = ?, cards_created = ?
		WHERE id = ?
	`

	result, err := r.db.DB().ExecContext(ctx, query,
		session.PlanID,
		nullString(session.ChunkID),
		session.StartTime,
		nullTime(session.EndTime),
		session.Duration,
		session.Notes,
		string(artifactsJSON),
		session.CardsCreated,
		session.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	return nil
}

// List retrieves sessions for a plan, ordered by start time descending.
func (r *SQLiteRepository) List(ctx context.Context, planID string, limit int) ([]*Session, error) {
	var query string
	var rows *sql.Rows
	var err error

	if limit > 0 {
		query = `
			SELECT id, plan_id, chunk_id, start_time, end_time, duration_minutes,
				notes, artifacts, cards_created, created_at
			FROM sessions
			WHERE plan_id = ?
			ORDER BY start_time DESC
			LIMIT ?
		`
		rows, err = r.db.DB().QueryContext(ctx, query, planID, limit)
	} else {
		// No limit - return all sessions for the plan
		query = `
			SELECT id, plan_id, chunk_id, start_time, end_time, duration_minutes,
				notes, artifacts, cards_created, created_at
			FROM sessions
			WHERE plan_id = ?
			ORDER BY start_time DESC
		`
		rows, err = r.db.DB().QueryContext(ctx, query, planID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	return r.scanSessions(rows)
}

// GetByPlan retrieves all sessions for a specific plan.
func (r *SQLiteRepository) GetByPlan(ctx context.Context, planID string) ([]*Session, error) {
	query := `
		SELECT id, plan_id, chunk_id, start_time, end_time, duration_minutes,
			notes, artifacts, cards_created, created_at
		FROM sessions
		WHERE plan_id = ?
		ORDER BY start_time DESC
	`

	rows, err := r.db.DB().QueryContext(ctx, query, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by plan: %w", err)
	}
	defer rows.Close()

	return r.scanSessions(rows)
}

// Delete removes a session by ID.
func (r *SQLiteRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM sessions WHERE id = ?"

	result, err := r.db.DB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", id)
	}

	return nil
}

// scanSession scans a single session from a database row.
func (r *SQLiteRepository) scanSession(row *sql.Row) (*Session, error) {
	var session Session
	var chunkID sql.NullString
	var endTime sql.NullTime
	var artifactsJSON string

	err := row.Scan(
		&session.ID,
		&session.PlanID,
		&chunkID,
		&session.StartTime,
		&endTime,
		&session.Duration,
		&session.Notes,
		&artifactsJSON,
		&session.CardsCreated,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if chunkID.Valid {
		session.ChunkID = chunkID.String
	}

	if endTime.Valid {
		t := endTime.Time
		session.EndTime = &t
	}

	// Unmarshal artifacts
	if artifactsJSON != "" && artifactsJSON != "null" {
		if err := json.Unmarshal([]byte(artifactsJSON), &session.Artifacts); err != nil {
			return nil, fmt.Errorf("failed to unmarshal artifacts: %w", err)
		}
	}

	return &session, nil
}

// scanSessions scans multiple sessions from database rows.
func (r *SQLiteRepository) scanSessions(rows *sql.Rows) ([]*Session, error) {
	sessions := make([]*Session, 0)

	for rows.Next() {
		var session Session
		var chunkID sql.NullString
		var endTime sql.NullTime
		var artifactsJSON string

		err := rows.Scan(
			&session.ID,
			&session.PlanID,
			&chunkID,
			&session.StartTime,
			&endTime,
			&session.Duration,
			&session.Notes,
			&artifactsJSON,
			&session.CardsCreated,
			&session.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// Handle nullable fields
		if chunkID.Valid {
			session.ChunkID = chunkID.String
		}

		if endTime.Valid {
			t := endTime.Time
			session.EndTime = &t
		}

		// Unmarshal artifacts
		if artifactsJSON != "" && artifactsJSON != "null" {
			if err := json.Unmarshal([]byte(artifactsJSON), &session.Artifacts); err != nil {
				return nil, fmt.Errorf("failed to unmarshal artifacts: %w", err)
			}
		}

		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// nullString converts a string to sql.NullString.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// nullTime converts a time pointer to sql.NullTime.
func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
