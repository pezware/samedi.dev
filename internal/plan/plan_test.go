// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package plan

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		valid  bool
	}{
		{"not-started is valid", StatusNotStarted, true},
		{"in-progress is valid", StatusInProgress, true},
		{"completed is valid", StatusCompleted, true},
		{"skipped is valid", StatusSkipped, true},
		{"archived is valid", StatusArchived, true},
		{"invalid status", Status("invalid"), false},
		{"empty status", Status(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.status.IsValid())
		})
	}
}

func TestPlan_Validate_ValidPlan(t *testing.T) {
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

	err := plan.Validate()
	assert.NoError(t, err)
}

func TestPlan_Validate_EmptyID(t *testing.T) {
	plan := &Plan{
		Title:      "Test Plan",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		TotalHours: 10.0,
		Status:     StatusNotStarted,
	}

	err := plan.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")
}

func TestPlan_Validate_EmptyTitle(t *testing.T) {
	plan := &Plan{
		ID:         "test-plan",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		TotalHours: 10.0,
		Status:     StatusNotStarted,
	}

	err := plan.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title cannot be empty")
}

func TestPlan_Validate_NegativeTotalHours(t *testing.T) {
	plan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		TotalHours: -5.0,
		Status:     StatusNotStarted,
	}

	err := plan.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "total hours must be positive")
}

func TestPlan_Validate_ZeroTotalHours(t *testing.T) {
	plan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		TotalHours: 0,
		Status:     StatusNotStarted,
	}

	err := plan.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "total hours must be positive")
}

func TestPlan_Validate_InvalidStatus(t *testing.T) {
	plan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		TotalHours: 10.0,
		Status:     Status("invalid"),
	}

	err := plan.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestPlan_Validate_ZeroCreatedAt(t *testing.T) {
	plan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		UpdatedAt:  time.Now(),
		TotalHours: 10.0,
		Status:     StatusNotStarted,
	}

	err := plan.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "created_at cannot be zero")
}

func TestPlan_Validate_ZeroUpdatedAt(t *testing.T) {
	plan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		CreatedAt:  time.Now(),
		TotalHours: 10.0,
		Status:     StatusNotStarted,
	}

	err := plan.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "updated_at cannot be zero")
}

func TestPlan_Validate_UpdatedBeforeCreated(t *testing.T) {
	now := time.Now()
	plan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		CreatedAt:  now,
		UpdatedAt:  now.Add(-1 * time.Hour),
		TotalHours: 10.0,
		Status:     StatusNotStarted,
	}

	err := plan.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "updated_at cannot be before created_at")
}

func TestPlan_Validate_InvalidChunk(t *testing.T) {
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
				ID:       "", // Invalid: empty ID
				Title:    "Chunk",
				Duration: 60,
				Status:   StatusNotStarted,
			},
		},
	}

	err := plan.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chunk 0")
	assert.Contains(t, err.Error(), "ID cannot be empty")
}

func TestPlan_Validate_DuplicateChunkIDs(t *testing.T) {
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
			{
				ID:       "chunk-001", // Duplicate
				Title:    "Second Chunk",
				Duration: 60,
				Status:   StatusNotStarted,
			},
		},
	}

	err := plan.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate chunk ID")
}

func TestChunk_Validate_ValidChunk(t *testing.T) {
	chunk := &Chunk{
		ID:       "chunk-001",
		Title:    "Test Chunk",
		Duration: 60,
		Status:   StatusNotStarted,
	}

	err := chunk.Validate()
	assert.NoError(t, err)
}

func TestChunk_Validate_EmptyID(t *testing.T) {
	chunk := &Chunk{
		Title:    "Test Chunk",
		Duration: 60,
		Status:   StatusNotStarted,
	}

	err := chunk.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")
}

func TestChunk_Validate_EmptyTitle(t *testing.T) {
	chunk := &Chunk{
		ID:       "chunk-001",
		Duration: 60,
		Status:   StatusNotStarted,
	}

	err := chunk.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title cannot be empty")
}

func TestChunk_Validate_NegativeDuration(t *testing.T) {
	chunk := &Chunk{
		ID:       "chunk-001",
		Title:    "Test Chunk",
		Duration: -10,
		Status:   StatusNotStarted,
	}

	err := chunk.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duration must be positive")
}

func TestChunk_Validate_ZeroDuration(t *testing.T) {
	chunk := &Chunk{
		ID:       "chunk-001",
		Title:    "Test Chunk",
		Duration: 0,
		Status:   StatusNotStarted,
	}

	err := chunk.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duration must be positive")
}

func TestChunk_Validate_InvalidStatus(t *testing.T) {
	chunk := &Chunk{
		ID:       "chunk-001",
		Title:    "Test Chunk",
		Duration: 60,
		Status:   Status("invalid"),
	}

	err := chunk.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestPlan_Progress(t *testing.T) {
	tests := []struct {
		name     string
		chunks   []Chunk
		expected float64
	}{
		{
			name:     "no chunks",
			chunks:   []Chunk{},
			expected: 0.0,
		},
		{
			name: "no completed chunks",
			chunks: []Chunk{
				{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusNotStarted},
				{ID: "chunk-002", Title: "Chunk 2", Duration: 60, Status: StatusInProgress},
			},
			expected: 0.0,
		},
		{
			name: "some completed chunks",
			chunks: []Chunk{
				{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusCompleted},
				{ID: "chunk-002", Title: "Chunk 2", Duration: 60, Status: StatusNotStarted},
			},
			expected: 0.5,
		},
		{
			name: "all completed chunks",
			chunks: []Chunk{
				{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusCompleted},
				{ID: "chunk-002", Title: "Chunk 2", Duration: 60, Status: StatusCompleted},
			},
			expected: 1.0,
		},
		{
			name: "skipped chunks not counted as completed",
			chunks: []Chunk{
				{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusCompleted},
				{ID: "chunk-002", Title: "Chunk 2", Duration: 60, Status: StatusSkipped},
				{ID: "chunk-003", Title: "Chunk 3", Duration: 60, Status: StatusNotStarted},
			},
			expected: 1.0 / 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &Plan{Chunks: tt.chunks}
			assert.InDelta(t, tt.expected, plan.Progress(), 0.001)
		})
	}
}

func TestPlan_ProgressPercent(t *testing.T) {
	plan := &Plan{
		Chunks: []Chunk{
			{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusCompleted},
			{ID: "chunk-002", Title: "Chunk 2", Duration: 60, Status: StatusNotStarted},
		},
	}

	assert.Equal(t, 50, plan.ProgressPercent())
}

func TestPlan_TotalMinutes(t *testing.T) {
	plan := &Plan{
		Chunks: []Chunk{
			{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusNotStarted},
			{ID: "chunk-002", Title: "Chunk 2", Duration: 90, Status: StatusNotStarted},
			{ID: "chunk-003", Title: "Chunk 3", Duration: 45, Status: StatusNotStarted},
		},
	}

	assert.Equal(t, 195, plan.TotalMinutes())
}

func TestPlan_CompletedHours(t *testing.T) {
	plan := &Plan{
		Chunks: []Chunk{
			{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusCompleted},
			{ID: "chunk-002", Title: "Chunk 2", Duration: 90, Status: StatusCompleted},
			{ID: "chunk-003", Title: "Chunk 3", Duration: 45, Status: StatusNotStarted},
		},
	}

	assert.InDelta(t, 2.5, plan.CompletedHours(), 0.01)
}

func TestPlan_RemainingHours(t *testing.T) {
	plan := &Plan{
		Chunks: []Chunk{
			{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusCompleted},
			{ID: "chunk-002", Title: "Chunk 2", Duration: 90, Status: StatusNotStarted},
			{ID: "chunk-003", Title: "Chunk 3", Duration: 45, Status: StatusInProgress},
			{ID: "chunk-004", Title: "Chunk 4", Duration: 30, Status: StatusSkipped},
		},
	}

	// Should count not-started (90) + in-progress (45), but not completed or skipped
	assert.InDelta(t, 2.25, plan.RemainingHours(), 0.01)
}

func TestPlan_NextChunk(t *testing.T) {
	tests := []struct {
		name     string
		chunks   []Chunk
		expected *string // nil if no next chunk
	}{
		{
			name:     "no chunks",
			chunks:   []Chunk{},
			expected: nil,
		},
		{
			name: "in-progress chunk has priority",
			chunks: []Chunk{
				{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusNotStarted},
				{ID: "chunk-002", Title: "Chunk 2", Duration: 60, Status: StatusInProgress},
				{ID: "chunk-003", Title: "Chunk 3", Duration: 60, Status: StatusNotStarted},
			},
			expected: strPtr("chunk-002"),
		},
		{
			name: "first not-started chunk if no in-progress",
			chunks: []Chunk{
				{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusCompleted},
				{ID: "chunk-002", Title: "Chunk 2", Duration: 60, Status: StatusNotStarted},
				{ID: "chunk-003", Title: "Chunk 3", Duration: 60, Status: StatusNotStarted},
			},
			expected: strPtr("chunk-002"),
		},
		{
			name: "nil if all completed or skipped",
			chunks: []Chunk{
				{ID: "chunk-001", Title: "Chunk 1", Duration: 60, Status: StatusCompleted},
				{ID: "chunk-002", Title: "Chunk 2", Duration: 60, Status: StatusSkipped},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &Plan{Chunks: tt.chunks}
			next := plan.NextChunk()

			if tt.expected == nil {
				assert.Nil(t, next)
			} else {
				require.NotNil(t, next)
				assert.Equal(t, *tt.expected, next.ID)
			}
		})
	}
}

// Helper function for test readability
func strPtr(s string) *string {
	return &s
}
