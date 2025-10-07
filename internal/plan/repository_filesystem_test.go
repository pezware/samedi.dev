// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package plan

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pezware/samedi.dev/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestFilesystem(t *testing.T) (*FilesystemRepository, *storage.Paths, func()) {
	t.Helper()

	// Create temp directory
	tmpDir := t.TempDir()

	// Create paths
	paths := &storage.Paths{
		BaseDir:  tmpDir,
		PlansDir: filepath.Join(tmpDir, "plans"),
	}

	// Create filesystem storage
	fs := storage.NewFilesystemStorage(paths)
	err := fs.Initialize()
	require.NoError(t, err)

	// Create repository
	repo := NewFilesystemRepository(fs, paths)

	cleanup := func() {
		// TempDir will be automatically cleaned up by Go
	}

	return repo, paths, cleanup
}

func TestFilesystemRepository_Save_CreateNewPlan(t *testing.T) {
	repo, paths, cleanup := setupTestFilesystem(t)
	defer cleanup()

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
		Chunks: []Chunk{
			{
				ID:          "chunk-001",
				Title:       "First Chunk",
				Duration:    60,
				Status:      StatusNotStarted,
				Objectives:  []string{"Learn basics"},
				Resources:   []string{"Book chapter 1"},
				Deliverable: "Complete exercises",
			},
		},
	}

	err := repo.Save(ctx, plan)
	require.NoError(t, err)

	// Verify file exists
	filePath := paths.PlanPath("test-plan")
	assert.FileExists(t, filePath)

	// Verify file permissions (should be readable but not writable by others)
	info, err := os.Stat(filePath)
	require.NoError(t, err)
	mode := info.Mode()
	assert.Equal(t, os.FileMode(0o600), mode&0o777)
}

func TestFilesystemRepository_Save_UpdateExistingPlan(t *testing.T) {
	repo, _, cleanup := setupTestFilesystem(t)
	defer cleanup()

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
		Chunks: []Chunk{
			{
				ID:       "chunk-001",
				Title:    "First Chunk",
				Duration: 60,
				Status:   StatusNotStarted,
			},
		},
	}

	// Create initial plan
	err := repo.Save(ctx, plan)
	require.NoError(t, err)

	// Update plan
	plan.Title = "Updated Test Plan"
	plan.Status = StatusInProgress
	plan.UpdatedAt = now.Add(time.Hour)

	err = repo.Save(ctx, plan)
	require.NoError(t, err)

	// Load and verify
	loaded, err := repo.Load(ctx, "test-plan")
	require.NoError(t, err)
	assert.Equal(t, "Updated Test Plan", loaded.Title)
	assert.Equal(t, StatusInProgress, loaded.Status)
}

func TestFilesystemRepository_Save_InvalidPlan(t *testing.T) {
	repo, _, cleanup := setupTestFilesystem(t)
	defer cleanup()

	ctx := context.Background()

	// Plan with empty ID
	plan := &Plan{
		ID:         "",
		Title:      "Test Plan",
		TotalHours: 10.0,
		Status:     StatusNotStarted,
	}

	err := repo.Save(ctx, plan)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid plan")
}

func TestFilesystemRepository_Load_ExistingPlan(t *testing.T) {
	repo, _, cleanup := setupTestFilesystem(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	originalPlan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		CreatedAt:  now,
		UpdatedAt:  now,
		TotalHours: 10.0,
		Status:     StatusNotStarted,
		Tags:       []string{"golang", "testing"},
		Chunks: []Chunk{
			{
				ID:          "chunk-001",
				Title:       "First Chunk",
				Duration:    60,
				Status:      StatusNotStarted,
				Objectives:  []string{"Learn basics", "Practice coding"},
				Resources:   []string{"Book chapter 1", "Video tutorial"},
				Deliverable: "Complete exercises",
			},
			{
				ID:       "chunk-002",
				Title:    "Second Chunk",
				Duration: 90,
				Status:   StatusNotStarted,
			},
		},
	}

	// Save plan
	err := repo.Save(ctx, originalPlan)
	require.NoError(t, err)

	// Load plan
	loaded, err := repo.Load(ctx, "test-plan")
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, originalPlan.ID, loaded.ID)
	assert.Equal(t, originalPlan.Title, loaded.Title)
	assert.Equal(t, originalPlan.TotalHours, loaded.TotalHours)
	assert.Equal(t, originalPlan.Status, loaded.Status)
	assert.Equal(t, originalPlan.Tags, loaded.Tags)
	assert.Len(t, loaded.Chunks, 2)

	// Verify first chunk
	assert.Equal(t, "chunk-001", loaded.Chunks[0].ID)
	assert.Equal(t, "First Chunk", loaded.Chunks[0].Title)
	assert.Equal(t, 60, loaded.Chunks[0].Duration)
	assert.Equal(t, []string{"Learn basics", "Practice coding"}, loaded.Chunks[0].Objectives)
	assert.Equal(t, []string{"Book chapter 1", "Video tutorial"}, loaded.Chunks[0].Resources)
	assert.Equal(t, "Complete exercises", loaded.Chunks[0].Deliverable)

	// Verify second chunk
	assert.Equal(t, "chunk-002", loaded.Chunks[1].ID)
	assert.Equal(t, 90, loaded.Chunks[1].Duration)
}

func TestFilesystemRepository_Load_NonexistentPlan(t *testing.T) {
	repo, _, cleanup := setupTestFilesystem(t)
	defer cleanup()

	ctx := context.Background()

	_, err := repo.Load(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan not found")
}

func TestFilesystemRepository_Load_CorruptedFile(t *testing.T) {
	repo, paths, cleanup := setupTestFilesystem(t)
	defer cleanup()

	ctx := context.Background()

	// Create a corrupted markdown file
	filePath := paths.PlanPath("corrupted")
	content := []byte("This is not valid YAML frontmatter\nNo delimiters here")
	err := os.WriteFile(filePath, content, 0o600)
	require.NoError(t, err)

	_, err = repo.Load(ctx, "corrupted")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse plan")
}

func TestFilesystemRepository_Delete_ExistingPlan(t *testing.T) {
	repo, paths, cleanup := setupTestFilesystem(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	plan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		CreatedAt:  now,
		UpdatedAt:  now,
		TotalHours: 10.0,
		Status:     StatusNotStarted,
		Chunks: []Chunk{
			{
				ID:       "chunk-001",
				Title:    "First Chunk",
				Duration: 60,
				Status:   StatusNotStarted,
			},
		},
	}

	// Create plan
	err := repo.Save(ctx, plan)
	require.NoError(t, err)

	// Verify exists
	filePath := paths.PlanPath("test-plan")
	assert.FileExists(t, filePath)

	// Delete plan
	err = repo.Delete(ctx, "test-plan")
	require.NoError(t, err)

	// Verify deleted
	assert.NoFileExists(t, filePath)
}

func TestFilesystemRepository_Delete_NonexistentPlan(t *testing.T) {
	repo, _, cleanup := setupTestFilesystem(t)
	defer cleanup()

	ctx := context.Background()

	err := repo.Delete(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan not found")
}

func TestFilesystemRepository_Exists(t *testing.T) {
	repo, _, cleanup := setupTestFilesystem(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	plan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		CreatedAt:  now,
		UpdatedAt:  now,
		TotalHours: 10.0,
		Status:     StatusNotStarted,
		Chunks: []Chunk{
			{
				ID:       "chunk-001",
				Title:    "First Chunk",
				Duration: 60,
				Status:   StatusNotStarted,
			},
		},
	}

	// Should not exist initially
	assert.False(t, repo.Exists(ctx, "test-plan"))

	// Create plan
	err := repo.Save(ctx, plan)
	require.NoError(t, err)

	// Should exist now
	assert.True(t, repo.Exists(ctx, "test-plan"))

	// Delete plan
	err = repo.Delete(ctx, "test-plan")
	require.NoError(t, err)

	// Should not exist after deletion
	assert.False(t, repo.Exists(ctx, "test-plan"))
}

func TestFilesystemRepository_List_EmptyDirectory(t *testing.T) {
	repo, _, cleanup := setupTestFilesystem(t)
	defer cleanup()

	ctx := context.Background()

	ids, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestFilesystemRepository_List_MultiplePlans(t *testing.T) {
	repo, _, cleanup := setupTestFilesystem(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	// Create multiple plans
	plans := []*Plan{
		{
			ID:         "plan-1",
			Title:      "Plan 1",
			CreatedAt:  now,
			UpdatedAt:  now,
			TotalHours: 10.0,
			Status:     StatusNotStarted,
			Chunks:     []Chunk{{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusNotStarted}},
		},
		{
			ID:         "plan-2",
			Title:      "Plan 2",
			CreatedAt:  now,
			UpdatedAt:  now,
			TotalHours: 20.0,
			Status:     StatusInProgress,
			Chunks:     []Chunk{{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusNotStarted}},
		},
		{
			ID:         "plan-3",
			Title:      "Plan 3",
			CreatedAt:  now,
			UpdatedAt:  now,
			TotalHours: 30.0,
			Status:     StatusCompleted,
			Chunks:     []Chunk{{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusNotStarted}},
		},
	}

	for _, p := range plans {
		err := repo.Save(ctx, p)
		require.NoError(t, err)
	}

	// List all plans
	ids, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, ids, 3)
	assert.Contains(t, ids, "plan-1")
	assert.Contains(t, ids, "plan-2")
	assert.Contains(t, ids, "plan-3")
}

func TestFilesystemRepository_List_IgnoresNonMarkdownFiles(t *testing.T) {
	repo, paths, cleanup := setupTestFilesystem(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	// Create a valid plan
	plan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		CreatedAt:  now,
		UpdatedAt:  now,
		TotalHours: 10.0,
		Status:     StatusNotStarted,
		Chunks:     []Chunk{{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusNotStarted}},
	}
	err := repo.Save(ctx, plan)
	require.NoError(t, err)

	// Create non-markdown files
	err = os.WriteFile(filepath.Join(paths.PlansDir, "readme.txt"), []byte("test"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(paths.PlansDir, "config.json"), []byte("{}"), 0o600)
	require.NoError(t, err)

	// Create subdirectory (should be ignored)
	err = os.Mkdir(filepath.Join(paths.PlansDir, "archive"), 0o755)
	require.NoError(t, err)

	// List should only return the markdown plan
	ids, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Equal(t, "test-plan", ids[0])
}

func TestFilesystemRepository_LoadAll(t *testing.T) {
	repo, _, cleanup := setupTestFilesystem(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	// Create multiple plans
	for i := 1; i <= 3; i++ {
		plan := &Plan{
			ID:         fmt.Sprintf("plan-%d", i),
			Title:      fmt.Sprintf("Plan %d", i),
			CreatedAt:  now,
			UpdatedAt:  now,
			TotalHours: float64(i * 10),
			Status:     StatusNotStarted,
			Chunks:     []Chunk{{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusNotStarted}},
		}
		err := repo.Save(ctx, plan)
		require.NoError(t, err)
	}

	// Load all plans
	plans, err := repo.LoadAll(ctx)
	require.NoError(t, err)
	assert.Len(t, plans, 3)

	// Verify plans are loaded correctly
	planIDs := make([]string, len(plans))
	for i, p := range plans {
		planIDs[i] = p.ID
	}
	assert.Contains(t, planIDs, "plan-1")
	assert.Contains(t, planIDs, "plan-2")
	assert.Contains(t, planIDs, "plan-3")
}

func TestFilesystemRepository_Path(t *testing.T) {
	repo, paths, cleanup := setupTestFilesystem(t)
	defer cleanup()

	path := repo.Path("test-plan")
	expectedPath := filepath.Join(paths.PlansDir, "test-plan.md")
	assert.Equal(t, expectedPath, path)
}
