// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package plan

import (
	"context"
	"testing"
	"time"

	"github.com/pezware/samedi.dev/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*storage.SQLiteDB, func()) {
	t.Helper()

	// Create in-memory database
	db, err := storage.NewSQLiteDB(":memory:")
	require.NoError(t, err)

	// Run migrations
	migrator := storage.NewMigrator(db)
	err = migrator.Migrate()
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func TestSQLiteRepository_Upsert_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now()
	plan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		CreatedAt:  now,
		UpdatedAt:  now,
		TotalHours: 10.0,
		Status:     StatusNotStarted,
		Tags:       []string{"test", "example"},
	}

	record := ToRecord(plan, "/path/to/test-plan.md")
	err := repo.Upsert(ctx, record)
	require.NoError(t, err)

	// Verify it was created
	retrieved, err := repo.Get(ctx, "test-plan")
	require.NoError(t, err)
	assert.Equal(t, "test-plan", retrieved.ID)
	assert.Equal(t, "Test Plan", retrieved.Title)
	assert.Equal(t, 10.0, retrieved.TotalHours)
	assert.Equal(t, "not-started", retrieved.Status)
	assert.Equal(t, []string{"test", "example"}, retrieved.Tags)
	assert.Equal(t, "/path/to/test-plan.md", retrieved.FilePath)
}

func TestSQLiteRepository_Upsert_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now()
	plan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		CreatedAt:  now,
		UpdatedAt:  now,
		TotalHours: 10.0,
		Status:     StatusNotStarted,
		Tags:       []string{"test"},
	}

	// Create
	record := ToRecord(plan, "/path/to/test-plan.md")
	err := repo.Upsert(ctx, record)
	require.NoError(t, err)

	// Update
	plan.Title = "Updated Plan"
	plan.Status = StatusInProgress
	plan.TotalHours = 15.0
	plan.UpdatedAt = now.Add(time.Hour)

	record = ToRecord(plan, "/path/to/test-plan.md")
	err = repo.Upsert(ctx, record)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.Get(ctx, "test-plan")
	require.NoError(t, err)
	assert.Equal(t, "Updated Plan", retrieved.Title)
	assert.Equal(t, "in-progress", retrieved.Status)
	assert.Equal(t, 15.0, retrieved.TotalHours)
}

func TestSQLiteRepository_Get_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	_, err := repo.Get(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan not found")
}

func TestSQLiteRepository_List_All(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	// Create multiple plans
	now := time.Now()
	plans := []*Plan{
		{
			ID:         "plan-1",
			Title:      "Plan 1",
			CreatedAt:  now,
			UpdatedAt:  now,
			TotalHours: 10.0,
			Status:     StatusNotStarted,
			Tags:       []string{"tag1"},
		},
		{
			ID:         "plan-2",
			Title:      "Plan 2",
			CreatedAt:  now.Add(time.Hour),
			UpdatedAt:  now.Add(time.Hour),
			TotalHours: 20.0,
			Status:     StatusInProgress,
			Tags:       []string{"tag2"},
		},
		{
			ID:         "plan-3",
			Title:      "Plan 3",
			CreatedAt:  now.Add(2 * time.Hour),
			UpdatedAt:  now.Add(2 * time.Hour),
			TotalHours: 30.0,
			Status:     StatusCompleted,
			Tags:       []string{"tag1", "tag2"},
		},
	}

	for i, p := range plans {
		record := ToRecord(p, "/path/to/plan"+string(rune('1'+i))+".md")
		err := repo.Upsert(ctx, record)
		require.NoError(t, err)
	}

	// List all
	records, err := repo.List(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, records, 3)

	// Should be ordered by created_at DESC (newest first)
	assert.Equal(t, "plan-3", records[0].ID)
	assert.Equal(t, "plan-2", records[1].ID)
	assert.Equal(t, "plan-1", records[2].ID)
}

func TestSQLiteRepository_List_FilterByStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	// Create plans with different statuses
	now := time.Now()
	plans := []*Plan{
		{ID: "plan-1", Title: "Plan 1", CreatedAt: now, UpdatedAt: now, TotalHours: 10.0, Status: StatusNotStarted},
		{ID: "plan-2", Title: "Plan 2", CreatedAt: now, UpdatedAt: now, TotalHours: 20.0, Status: StatusInProgress},
		{ID: "plan-3", Title: "Plan 3", CreatedAt: now, UpdatedAt: now, TotalHours: 30.0, Status: StatusCompleted},
	}

	for _, p := range plans {
		record := ToRecord(p, "/path/to/"+p.ID+".md")
		err := repo.Upsert(ctx, record)
		require.NoError(t, err)
	}

	// Filter by status
	filter := &storage.PlanFilter{
		Statuses: []string{"not-started", "in-progress"},
	}

	records, err := repo.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, records, 2)

	ids := []string{records[0].ID, records[1].ID}
	assert.Contains(t, ids, "plan-1")
	assert.Contains(t, ids, "plan-2")
}

func TestSQLiteRepository_List_FilterByTag(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now()
	plans := []*Plan{
		{ID: "plan-1", Title: "Plan 1", CreatedAt: now, UpdatedAt: now, TotalHours: 10.0, Status: StatusNotStarted, Tags: []string{"golang", "backend"}},
		{ID: "plan-2", Title: "Plan 2", CreatedAt: now, UpdatedAt: now, TotalHours: 20.0, Status: StatusInProgress, Tags: []string{"react", "frontend"}},
		{ID: "plan-3", Title: "Plan 3", CreatedAt: now, UpdatedAt: now, TotalHours: 30.0, Status: StatusCompleted, Tags: []string{"golang", "testing"}},
	}

	for _, p := range plans {
		record := ToRecord(p, "/path/to/"+p.ID+".md")
		err := repo.Upsert(ctx, record)
		require.NoError(t, err)
	}

	// Filter by tag
	filter := &storage.PlanFilter{
		Tag: "golang",
	}

	records, err := repo.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, records, 2)

	ids := []string{records[0].ID, records[1].ID}
	assert.Contains(t, ids, "plan-1")
	assert.Contains(t, ids, "plan-3")
}

func TestSQLiteRepository_List_FilterByIDs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now()
	plans := []*Plan{
		{ID: "plan-1", Title: "Plan 1", CreatedAt: now, UpdatedAt: now, TotalHours: 10.0, Status: StatusNotStarted},
		{ID: "plan-2", Title: "Plan 2", CreatedAt: now, UpdatedAt: now, TotalHours: 20.0, Status: StatusInProgress},
		{ID: "plan-3", Title: "Plan 3", CreatedAt: now, UpdatedAt: now, TotalHours: 30.0, Status: StatusCompleted},
	}

	for _, p := range plans {
		record := ToRecord(p, "/path/to/"+p.ID+".md")
		err := repo.Upsert(ctx, record)
		require.NoError(t, err)
	}

	// Filter by specific IDs
	filter := &storage.PlanFilter{
		IDs: []string{"plan-1", "plan-3"},
	}

	records, err := repo.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, records, 2)

	ids := []string{records[0].ID, records[1].ID}
	assert.Contains(t, ids, "plan-1")
	assert.Contains(t, ids, "plan-3")
}

func TestSQLiteRepository_Delete_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now()
	plan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		CreatedAt:  now,
		UpdatedAt:  now,
		TotalHours: 10.0,
		Status:     StatusNotStarted,
	}

	// Create
	record := ToRecord(plan, "/path/to/test-plan.md")
	err := repo.Upsert(ctx, record)
	require.NoError(t, err)

	// Delete
	err = repo.Delete(ctx, "test-plan")
	require.NoError(t, err)

	// Verify deleted
	_, err = repo.Get(ctx, "test-plan")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan not found")
}

func TestSQLiteRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan not found")
}
