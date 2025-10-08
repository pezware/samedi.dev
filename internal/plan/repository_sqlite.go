// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package plan

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/pezware/samedi.dev/internal/storage"
)

// SQLiteRepository implements plan storage using SQLite.
type SQLiteRepository struct {
	db *storage.SQLiteDB
}

// NewSQLiteRepository creates a new SQLite-backed plan repository.
func NewSQLiteRepository(db *storage.SQLiteDB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// ToRecord converts a Plan domain model to a storage PlanRecord.
func ToRecord(plan *Plan, filePath string) *storage.PlanRecord {
	return &storage.PlanRecord{
		ID:         plan.ID,
		Title:      plan.Title,
		CreatedAt:  plan.CreatedAt,
		UpdatedAt:  plan.UpdatedAt,
		TotalHours: plan.TotalHours,
		Status:     string(plan.Status),
		Tags:       plan.Tags,
		FilePath:   filePath,
	}
}

// RecordToPlan converts a storage PlanRecord to a Plan domain model.
// Note: This only converts metadata; chunks must be loaded from filesystem.
func RecordToPlan(record *storage.PlanRecord) *Plan {
	return &Plan{
		ID:         record.ID,
		Title:      record.Title,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  record.UpdatedAt,
		TotalHours: record.TotalHours,
		Status:     Status(record.Status),
		Tags:       record.Tags,
		Chunks:     []Chunk{}, // Chunks must be loaded separately
	}
}

// Upsert creates or updates a plan's metadata in SQLite.
func (r *SQLiteRepository) Upsert(ctx context.Context, record *storage.PlanRecord) error {
	tagsJSON, err := json.Marshal(record.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO plans (id, title, created_at, updated_at, total_hours, status, tags, file_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			updated_at = excluded.updated_at,
			total_hours = excluded.total_hours,
			status = excluded.status,
			tags = excluded.tags,
			file_path = excluded.file_path
	`

	_, err = r.db.DB().ExecContext(ctx, query,
		record.ID,
		record.Title,
		record.CreatedAt,
		record.UpdatedAt,
		record.TotalHours,
		record.Status,
		string(tagsJSON),
		record.FilePath,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert plan: %w", err)
	}

	return nil
}

// Get retrieves a plan's metadata by ID.
func (r *SQLiteRepository) Get(ctx context.Context, id string) (*storage.PlanRecord, error) {
	query := `
		SELECT id, title, created_at, updated_at, total_hours, status, tags, file_path
		FROM plans
		WHERE id = ?
	`

	var record storage.PlanRecord
	var tagsJSON string

	err := r.db.DB().QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.Title,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.TotalHours,
		&record.Status,
		&tagsJSON,
		&record.FilePath,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plan not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	// Unmarshal tags
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &record.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	}

	return &record, nil
}

// List retrieves plans with optional filtering.
func (r *SQLiteRepository) List(ctx context.Context, filter *storage.PlanFilter) ([]*storage.PlanRecord, error) {
	query := "SELECT id, title, created_at, updated_at, total_hours, status, tags, file_path FROM plans"
	conditions, args := r.buildWhereClause(filter)

	if conditions != "" {
		query += " WHERE " + conditions
	}

	query += r.buildOrderClause(filter)

	rows, err := r.db.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// buildWhereClause constructs the WHERE clause and arguments for filtering.
func (r *SQLiteRepository) buildWhereClause(filter *storage.PlanFilter) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}

	var conditions []string
	var args []interface{}

	if len(filter.IDs) > 0 {
		placeholders := r.buildPlaceholders(len(filter.IDs))
		conditions = append(conditions, fmt.Sprintf("id IN (%s)", placeholders))
		for _, id := range filter.IDs {
			args = append(args, id)
		}
	}

	if len(filter.Statuses) > 0 {
		placeholders := r.buildPlaceholders(len(filter.Statuses))
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", placeholders))
		for _, status := range filter.Statuses {
			args = append(args, status)
		}
	}

	if filter.Tag != "" {
		conditions = append(conditions, "tags LIKE ?")
		args = append(args, "%"+filter.Tag+"%")
	}

	whereClause := ""
	for i, condition := range conditions {
		if i > 0 {
			whereClause += " AND "
		}
		whereClause += condition
	}

	return whereClause, args
}

// buildPlaceholders creates a string of SQL placeholders (?, ?, ?).
func (r *SQLiteRepository) buildPlaceholders(count int) string {
	if count == 0 {
		return ""
	}

	placeholders := "?"
	for i := 1; i < count; i++ {
		placeholders += ", ?"
	}
	return placeholders
}

// buildOrderClause constructs the ORDER BY clause.
func (r *SQLiteRepository) buildOrderClause(filter *storage.PlanFilter) string {
	if filter == nil || filter.SortBy == "" {
		return " ORDER BY created_at DESC"
	}

	// Map user-friendly names to SQL columns (prevent SQL injection)
	sortField := map[string]string{
		"created": "created_at DESC",
		"updated": "updated_at DESC",
		"title":   "title ASC",
		"status":  "status ASC",
		"hours":   "total_hours DESC",
	}

	if sqlOrder, ok := sortField[filter.SortBy]; ok {
		return " ORDER BY " + sqlOrder
	}

	// Default if invalid sort field
	return " ORDER BY created_at DESC"
}

// scanRows scans all rows into PlanRecords.
func (r *SQLiteRepository) scanRows(rows *sql.Rows) ([]*storage.PlanRecord, error) {
	var records []*storage.PlanRecord

	for rows.Next() {
		record, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plan rows: %w", err)
	}

	return records, nil
}

// scanRow scans a single row into a PlanRecord.
func (r *SQLiteRepository) scanRow(rows *sql.Rows) (*storage.PlanRecord, error) {
	var record storage.PlanRecord
	var tagsJSON string

	err := rows.Scan(
		&record.ID,
		&record.Title,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.TotalHours,
		&record.Status,
		&tagsJSON,
		&record.FilePath,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan plan row: %w", err)
	}

	// Unmarshal tags
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &record.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	}

	return &record, nil
}

// Delete removes a plan's metadata from SQLite.
func (r *SQLiteRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM plans WHERE id = ?"

	result, err := r.db.DB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("plan not found: %s", id)
	}

	return nil
}
