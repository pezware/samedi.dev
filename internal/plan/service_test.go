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

// MockLLMProvider is a test double for LLM providers.
type MockLLMProvider struct {
	CallFunc func(ctx context.Context, prompt string) (string, error)
	Calls    []string // Track all prompts sent
}

func (m *MockLLMProvider) Call(ctx context.Context, prompt string) (string, error) {
	m.Calls = append(m.Calls, prompt)
	if m.CallFunc != nil {
		return m.CallFunc(ctx, prompt)
	}
	return "", fmt.Errorf("mock not configured")
}

func setupTestService(t *testing.T) (*Service, *MockLLMProvider, *storage.Paths, func()) {
	t.Helper()

	// Create temp directory
	tmpDir := t.TempDir()

	// Create paths
	paths := &storage.Paths{
		BaseDir:      tmpDir,
		PlansDir:     filepath.Join(tmpDir, "plans"),
		TemplatesDir: filepath.Join(tmpDir, "templates"),
		DatabasePath: filepath.Join(tmpDir, "test.db"),
	}

	// Initialize database
	db, err := storage.NewSQLiteDB(paths.DatabasePath)
	require.NoError(t, err)

	migrator := storage.NewMigrator(db)
	err = migrator.Migrate()
	require.NoError(t, err)

	// Initialize filesystem
	fs := storage.NewFilesystemStorage(paths)
	err = fs.Initialize()
	require.NoError(t, err)

	// Create mock template
	err = os.WriteFile(paths.TemplatePath("plan-generation"), []byte(mockTemplate), 0o600)
	require.NoError(t, err)

	// Create repositories
	sqliteRepo := NewSQLiteRepository(db)
	filesystemRepo := NewFilesystemRepository(fs, paths)

	// Create mock LLM
	mockLLM := &MockLLMProvider{}

	// Create service
	service := NewService(sqliteRepo, filesystemRepo, mockLLM, fs, paths)

	cleanup := func() {
		db.Close()
	}

	return service, mockLLM, paths, cleanup
}

const mockTemplate = `# Plan Generation Template

Topic: {{.Topic}}
Total Hours: {{.TotalHours}}
Level: {{.Level}}
Goals: {{.Goals}}
Slug: {{.Slug}}
Created: {{.Now}}
`

const validPlanMarkdown = `---
id: test-plan
title: Test Plan
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
total_hours: 10
status: not-started
tags: [test]
---

# Test Plan

## Chunk 1: First Chunk {#chunk-001}

**Duration**: 1 hour
**Status**: not-started
**Objectives**:
- Learn basics

**Resources**:
- Book chapter 1

**Deliverable**: Complete exercises
`

func TestService_Create_ValidRequest(t *testing.T) {
	service, mockLLM, paths, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Configure mock LLM to return valid plan
	mockLLM.CallFunc = func(_ context.Context, _ string) (string, error) {
		return validPlanMarkdown, nil
	}

	req := CreateRequest{
		Topic:      "Test Plan",
		TotalHours: 10.0,
		Level:      "beginner",
		Goals:      "Learn testing",
	}

	plan, err := service.Create(ctx, req)
	require.NoError(t, err)

	// Verify plan was created
	assert.Equal(t, "test-plan", plan.ID)
	assert.Equal(t, "Test Plan", plan.Title)
	assert.Equal(t, 10.0, plan.TotalHours)
	assert.Len(t, plan.Chunks, 1)

	// Verify file exists
	assert.FileExists(t, paths.PlanPath("test-plan"))

	// Verify LLM was called
	assert.Len(t, mockLLM.Calls, 1)
	assert.Contains(t, mockLLM.Calls[0], "Test Plan")
	assert.Contains(t, mockLLM.Calls[0], "10")

	// Verify SQLite record
	record, err := service.GetMetadata(ctx, "test-plan")
	require.NoError(t, err)
	assert.Equal(t, "test-plan", record.ID)
	assert.Equal(t, "Test Plan", record.Title)
}

func TestService_Create_EmptyTopic(t *testing.T) {
	service, _, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	req := CreateRequest{
		Topic:      "",
		TotalHours: 10.0,
	}

	_, err := service.Create(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "topic cannot be empty")
}

func TestService_Create_ZeroHours(t *testing.T) {
	service, _, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	req := CreateRequest{
		Topic:      "Test",
		TotalHours: 0,
	}

	_, err := service.Create(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "total hours must be positive")
}

func TestService_Create_NegativeHours(t *testing.T) {
	service, _, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	req := CreateRequest{
		Topic:      "Test",
		TotalHours: -10,
	}

	_, err := service.Create(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "total hours must be positive")
}

func TestService_Create_TooManyHours(t *testing.T) {
	service, _, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	req := CreateRequest{
		Topic:      "Test",
		TotalHours: 1500,
	}

	_, err := service.Create(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "total hours too large")
}

func TestService_Create_InvalidLevel(t *testing.T) {
	service, _, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	req := CreateRequest{
		Topic:      "Test",
		TotalHours: 10,
		Level:      "expert",
	}

	_, err := service.Create(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid level")
}

func TestService_Create_LLMFailure(t *testing.T) {
	service, mockLLM, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Configure mock LLM to return error
	mockLLM.CallFunc = func(_ context.Context, _ string) (string, error) {
		return "", fmt.Errorf("LLM service unavailable")
	}

	req := CreateRequest{
		Topic:      "Test",
		TotalHours: 10.0,
	}

	_, err := service.Create(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "LLM call failed")
}

func TestService_Create_InvalidLLMOutput(t *testing.T) {
	service, mockLLM, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Configure mock LLM to return invalid markdown
	mockLLM.CallFunc = func(_ context.Context, _ string) (string, error) {
		return "This is not valid plan markdown", nil
	}

	req := CreateRequest{
		Topic:      "Test",
		TotalHours: 10.0,
	}

	_, err := service.Create(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse LLM output")
}

func TestService_Create_DuplicatePlan(t *testing.T) {
	service, mockLLM, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Configure mock LLM
	mockLLM.CallFunc = func(_ context.Context, _ string) (string, error) {
		return validPlanMarkdown, nil
	}

	req := CreateRequest{
		Topic:      "Test Plan",
		TotalHours: 10.0,
	}

	// Create first plan
	_, err := service.Create(ctx, req)
	require.NoError(t, err)

	// Try to create duplicate
	_, err = service.Create(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan already exists")
}

func TestService_Get_ExistingPlan(t *testing.T) {
	service, mockLLM, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Configure mock and create plan
	mockLLM.CallFunc = func(_ context.Context, _ string) (string, error) {
		return validPlanMarkdown, nil
	}

	req := CreateRequest{
		Topic:      "Test Plan",
		TotalHours: 10.0,
	}

	createdPlan, err := service.Create(ctx, req)
	require.NoError(t, err)

	// Get the plan
	plan, err := service.Get(ctx, "test-plan")
	require.NoError(t, err)

	// Verify
	assert.Equal(t, createdPlan.ID, plan.ID)
	assert.Equal(t, createdPlan.Title, plan.Title)
	assert.Equal(t, createdPlan.TotalHours, plan.TotalHours)
	assert.Len(t, plan.Chunks, 1)
}

func TestService_Get_NonexistentPlan(t *testing.T) {
	service, _, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	_, err := service.Get(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan not found")
}

func TestService_Update_ExistingPlan(t *testing.T) {
	service, mockLLM, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Configure mock and create plan
	mockLLM.CallFunc = func(_ context.Context, _ string) (string, error) {
		return validPlanMarkdown, nil
	}

	req := CreateRequest{
		Topic:      "Test Plan",
		TotalHours: 10.0,
	}

	plan, err := service.Create(ctx, req)
	require.NoError(t, err)

	// Modify plan
	originalUpdatedAt := plan.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure timestamp changes
	plan.Title = "Updated Test Plan"
	plan.Status = StatusInProgress

	// Update
	err = service.Update(ctx, plan)
	require.NoError(t, err)

	// Verify changes
	updated, err := service.Get(ctx, "test-plan")
	require.NoError(t, err)
	assert.Equal(t, "Updated Test Plan", updated.Title)
	assert.Equal(t, StatusInProgress, updated.Status)
	assert.True(t, updated.UpdatedAt.After(originalUpdatedAt))

	// Verify SQLite was updated
	record, err := service.GetMetadata(ctx, "test-plan")
	require.NoError(t, err)
	assert.Equal(t, "Updated Test Plan", record.Title)
	assert.Equal(t, "in-progress", record.Status)
}

func TestService_Update_InvalidPlan(t *testing.T) {
	service, mockLLM, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Configure mock and create plan
	mockLLM.CallFunc = func(_ context.Context, _ string) (string, error) {
		return validPlanMarkdown, nil
	}

	req := CreateRequest{
		Topic:      "Test Plan",
		TotalHours: 10.0,
	}

	plan, err := service.Create(ctx, req)
	require.NoError(t, err)

	// Make plan invalid
	plan.TotalHours = -10

	// Update should fail
	err = service.Update(ctx, plan)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid plan")
}

func TestService_Update_NonexistentPlan(t *testing.T) {
	service, _, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	plan := &Plan{
		ID:         "nonexistent",
		Title:      "Test",
		CreatedAt:  now,
		UpdatedAt:  now,
		TotalHours: 10,
		Status:     StatusNotStarted,
		Chunks:     []Chunk{{ID: "chunk-001", Title: "Test", Duration: 60, Status: StatusNotStarted}},
	}

	err := service.Update(ctx, plan)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan not found")
}

func TestService_Delete_ExistingPlan(t *testing.T) {
	service, mockLLM, paths, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Configure mock and create plan
	mockLLM.CallFunc = func(_ context.Context, _ string) (string, error) {
		return validPlanMarkdown, nil
	}

	req := CreateRequest{
		Topic:      "Test Plan",
		TotalHours: 10.0,
	}

	_, err := service.Create(ctx, req)
	require.NoError(t, err)

	// Verify exists
	assert.True(t, service.Exists(ctx, "test-plan"))
	assert.FileExists(t, paths.PlanPath("test-plan"))

	// Delete
	err = service.Delete(ctx, "test-plan")
	require.NoError(t, err)

	// Verify deleted
	assert.False(t, service.Exists(ctx, "test-plan"))
	assert.NoFileExists(t, paths.PlanPath("test-plan"))

	// Verify SQLite record deleted
	_, err = service.GetMetadata(ctx, "test-plan")
	require.Error(t, err)
}

func TestService_Delete_NonexistentPlan(t *testing.T) {
	service, _, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	err := service.Delete(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan not found")
}

func TestService_List_EmptyResults(t *testing.T) {
	service, _, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	records, err := service.List(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestService_List_MultiplePlans(t *testing.T) {
	service, mockLLM, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Configure mock
	mockLLM.CallFunc = func(_ context.Context, _ string) (string, error) {
		return validPlanMarkdown, nil
	}

	// Create multiple plans
	topics := []string{"Plan Alpha", "Plan Beta", "Plan Gamma"}
	for _, topic := range topics {
		req := CreateRequest{
			Topic:      topic,
			TotalHours: 10.0,
		}
		_, err := service.Create(ctx, req)
		require.NoError(t, err)
	}

	// List all
	records, err := service.List(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, records, 3)

	// Verify IDs
	ids := make([]string, len(records))
	for i, r := range records {
		ids[i] = r.ID
	}
	assert.Contains(t, ids, "plan-alpha")
	assert.Contains(t, ids, "plan-beta")
	assert.Contains(t, ids, "plan-gamma")
}

func TestService_List_WithFilter(t *testing.T) {
	service, mockLLM, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Configure mock
	mockLLM.CallFunc = func(_ context.Context, _ string) (string, error) {
		return validPlanMarkdown, nil
	}

	// Create plans
	req := CreateRequest{
		Topic:      "Test Plan",
		TotalHours: 10.0,
	}
	plan, err := service.Create(ctx, req)
	require.NoError(t, err)

	// Update one to in-progress
	plan.Status = StatusInProgress
	err = service.Update(ctx, plan)
	require.NoError(t, err)

	// Filter by status
	filter := &storage.PlanFilter{
		Statuses: []string{"in-progress"},
	}

	records, err := service.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "test-plan", records[0].ID)
	assert.Equal(t, "in-progress", records[0].Status)
}

func TestService_Exists(t *testing.T) {
	service, mockLLM, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Configure mock
	mockLLM.CallFunc = func(_ context.Context, _ string) (string, error) {
		return validPlanMarkdown, nil
	}

	// Should not exist
	assert.False(t, service.Exists(ctx, "test-plan"))

	// Create plan
	req := CreateRequest{
		Topic:      "Test Plan",
		TotalHours: 10.0,
	}
	_, err := service.Create(ctx, req)
	require.NoError(t, err)

	// Should exist
	assert.True(t, service.Exists(ctx, "test-plan"))
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple case",
			input:    "Rust Async Programming",
			expected: "rust-async-programming",
		},
		{
			name:     "with numbers",
			input:    "French B1 Mastery",
			expected: "french-b1-mastery",
		},
		{
			name:     "with special characters",
			input:    "Music Theory (Basics)",
			expected: "music-theory-basics",
		},
		{
			name:     "multiple spaces",
			input:    "Learn   Go    Quickly",
			expected: "learn-go-quickly",
		},
		{
			name:     "trailing/leading spaces",
			input:    "  Test Plan  ",
			expected: "test-plan",
		},
		{
			name:     "already lowercase",
			input:    "already-lowercase",
			expected: "already-lowercase",
		},
		{
			name:     "unicode characters",
			input:    "Café français",
			expected: "caf-fran-ais",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slugify(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
