// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package session

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
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

func createTestPlan(t *testing.T, db *storage.SQLiteDB, planID string) {
	t.Helper()

	ctx := context.Background()
	now := time.Now()

	// Insert a test plan (required for foreign key)
	// Use unique file path per plan ID
	query := `
		INSERT INTO plans (id, title, created_at, updated_at, total_hours, status, tags, file_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.DB().ExecContext(ctx, query,
		planID, "Test Plan", now, now, 10.0, "not-started", "[]", "/test/"+planID+".md",
	)
	require.NoError(t, err)
}

func TestSQLiteRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	createTestPlan(t, db, "test-plan")

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now()
	session := &Session{
		ID:        uuid.New().String(),
		PlanID:    "test-plan",
		ChunkID:   "chunk-001",
		StartTime: now,
		CreatedAt: now,
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Verify it was created
	retrieved, err := repo.Get(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, retrieved.ID)
	assert.Equal(t, "test-plan", retrieved.PlanID)
	assert.Equal(t, "chunk-001", retrieved.ChunkID)
	assert.True(t, retrieved.IsActive())
}

func TestSQLiteRepository_Create_WithoutChunkID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	createTestPlan(t, db, "test-plan")

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now()
	session := &Session{
		ID:        uuid.New().String(),
		PlanID:    "test-plan",
		ChunkID:   "", // No chunk
		StartTime: now,
		CreatedAt: now,
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	retrieved, err := repo.Get(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, "", retrieved.ChunkID)
}

func TestSQLiteRepository_Create_CompletedSession(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	createTestPlan(t, db, "test-plan")

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	start := time.Now()
	end := start.Add(1 * time.Hour)

	session := &Session{
		ID:        uuid.New().String(),
		PlanID:    "test-plan",
		StartTime: start,
		EndTime:   &end,
		Duration:  60,
		Notes:     "Test notes",
		Artifacts: []string{"https://github.com/user/repo"},
		CreatedAt: start,
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	retrieved, err := repo.Get(ctx, session.ID)
	require.NoError(t, err)
	assert.False(t, retrieved.IsActive())
	assert.Equal(t, 60, retrieved.Duration)
	assert.Equal(t, "Test notes", retrieved.Notes)
	assert.Len(t, retrieved.Artifacts, 1)
	assert.Equal(t, "https://github.com/user/repo", retrieved.Artifacts[0])
}

func TestSQLiteRepository_Get_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	_, err := repo.Get(ctx, "non-existent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestSQLiteRepository_GetActive_NoActiveSession(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	active, err := repo.GetActive(ctx)
	require.NoError(t, err)
	assert.Nil(t, active)
}

func TestSQLiteRepository_GetActive_WithActiveSession(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	createTestPlan(t, db, "test-plan")

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now()
	session := &Session{
		ID:        uuid.New().String(),
		PlanID:    "test-plan",
		StartTime: now,
		CreatedAt: now,
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	active, err := repo.GetActive(ctx)
	require.NoError(t, err)
	require.NotNil(t, active)
	assert.Equal(t, session.ID, active.ID)
	assert.True(t, active.IsActive())
}

func TestSQLiteRepository_GetActive_IgnoresCompletedSessions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	createTestPlan(t, db, "test-plan")

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	start := time.Now()
	end := start.Add(1 * time.Hour)

	// Create completed session
	completedSession := &Session{
		ID:        uuid.New().String(),
		PlanID:    "test-plan",
		StartTime: start,
		EndTime:   &end,
		Duration:  60,
		CreatedAt: start,
	}

	err := repo.Create(ctx, completedSession)
	require.NoError(t, err)

	// Should return nil - no active sessions
	active, err := repo.GetActive(ctx)
	require.NoError(t, err)
	assert.Nil(t, active)
}

func TestSQLiteRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	createTestPlan(t, db, "test-plan")

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now()
	session := &Session{
		ID:        uuid.New().String(),
		PlanID:    "test-plan",
		StartTime: now,
		CreatedAt: now,
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Complete the session
	end := now.Add(1 * time.Hour)
	session.EndTime = &end
	session.Duration = 60
	session.Notes = "Completed work"
	session.Artifacts = []string{"https://example.com"}

	err = repo.Update(ctx, session)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.Get(ctx, session.ID)
	require.NoError(t, err)
	assert.False(t, retrieved.IsActive())
	assert.Equal(t, 60, retrieved.Duration)
	assert.Equal(t, "Completed work", retrieved.Notes)
	assert.Len(t, retrieved.Artifacts, 1)
}

func TestSQLiteRepository_Update_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	session := &Session{
		ID:        "non-existent",
		PlanID:    "test-plan",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
	}

	err := repo.Update(ctx, session)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestSQLiteRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	createTestPlan(t, db, "test-plan")

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	// Create multiple sessions
	now := time.Now()
	for i := 0; i < 5; i++ {
		session := &Session{
			ID:        uuid.New().String(),
			PlanID:    "test-plan",
			StartTime: now.Add(time.Duration(i) * time.Hour),
			CreatedAt: now,
		}
		err := repo.Create(ctx, session)
		require.NoError(t, err)
	}

	// List with limit
	sessions, err := repo.List(ctx, "test-plan", 3)
	require.NoError(t, err)
	assert.Len(t, sessions, 3)

	// Should be ordered by start_time DESC (most recent first)
	assert.True(t, sessions[0].StartTime.After(sessions[1].StartTime))
	assert.True(t, sessions[1].StartTime.After(sessions[2].StartTime))
}

func TestSQLiteRepository_List_EmptyResult(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	createTestPlan(t, db, "test-plan")

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	sessions, err := repo.List(ctx, "test-plan", 10)
	require.NoError(t, err)
	assert.Len(t, sessions, 0)
}

func TestSQLiteRepository_GetByPlan(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	createTestPlan(t, db, "plan-1")
	createTestPlan(t, db, "plan-2")

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create sessions for plan-1
	for i := 0; i < 3; i++ {
		session := &Session{
			ID:        uuid.New().String(),
			PlanID:    "plan-1",
			StartTime: now,
			CreatedAt: now,
		}
		err := repo.Create(ctx, session)
		require.NoError(t, err)
	}

	// Create sessions for plan-2
	for i := 0; i < 2; i++ {
		session := &Session{
			ID:        uuid.New().String(),
			PlanID:    "plan-2",
			StartTime: now,
			CreatedAt: now,
		}
		err := repo.Create(ctx, session)
		require.NoError(t, err)
	}

	// Get sessions for plan-1
	sessions, err := repo.GetByPlan(ctx, "plan-1")
	require.NoError(t, err)
	assert.Len(t, sessions, 3)

	// Verify all belong to plan-1
	for _, s := range sessions {
		assert.Equal(t, "plan-1", s.PlanID)
	}
}

func TestSQLiteRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	createTestPlan(t, db, "test-plan")

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now()
	session := &Session{
		ID:        uuid.New().String(),
		PlanID:    "test-plan",
		StartTime: now,
		CreatedAt: now,
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Delete
	err = repo.Delete(ctx, session.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = repo.Get(ctx, session.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestSQLiteRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "non-existent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestSQLiteRepository_Artifacts_EmptyArray(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	createTestPlan(t, db, "test-plan")

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now()
	session := &Session{
		ID:        uuid.New().String(),
		PlanID:    "test-plan",
		StartTime: now,
		Artifacts: []string{}, // Empty array
		CreatedAt: now,
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	retrieved, err := repo.Get(ctx, session.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved.Artifacts)
	assert.Len(t, retrieved.Artifacts, 0)
}

func TestSQLiteRepository_Artifacts_MultipleValues(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	createTestPlan(t, db, "test-plan")

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now()
	session := &Session{
		ID:        uuid.New().String(),
		PlanID:    "test-plan",
		StartTime: now,
		Artifacts: []string{
			"https://github.com/user/repo",
			"/path/to/file.md",
			"https://example.com/resource",
		},
		CreatedAt: now,
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	retrieved, err := repo.Get(ctx, session.ID)
	require.NoError(t, err)
	assert.Len(t, retrieved.Artifacts, 3)
	assert.Equal(t, "https://github.com/user/repo", retrieved.Artifacts[0])
	assert.Equal(t, "/path/to/file.md", retrieved.Artifacts[1])
	assert.Equal(t, "https://example.com/resource", retrieved.Artifacts[2])
}
