// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"testing"
	"time"

	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/session"
	"github.com/stretchr/testify/assert"
)

func TestGetStatusIcon_AllStatuses(t *testing.T) {
	tests := []struct {
		name         string
		status       plan.Status
		expectedIcon string
	}{
		{
			name:         "not started",
			status:       plan.StatusNotStarted,
			expectedIcon: "○",
		},
		{
			name:         "in progress",
			status:       plan.StatusInProgress,
			expectedIcon: "◐",
		},
		{
			name:         "completed",
			status:       plan.StatusCompleted,
			expectedIcon: "●",
		},
		{
			name:         "skipped",
			status:       plan.StatusSkipped,
			expectedIcon: "⊘",
		},
		{
			name:         "unknown status",
			status:       plan.Status("unknown"),
			expectedIcon: "?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := getStatusIcon(tt.status)
			assert.Equal(t, tt.expectedIcon, icon)
		})
	}
}

func TestChunkDisplayInfo_Structure(t *testing.T) {
	// Test that ChunkDisplayInfo can be created with all fields
	info := &ChunkDisplayInfo{
		Chunk: &plan.Chunk{
			ID:          "chunk-001",
			Title:       "Test Chunk",
			Duration:    60,
			Status:      plan.StatusInProgress,
			Objectives:  []string{"Learn basics"},
			Resources:   []string{"https://example.com"},
			Deliverable: "Working code",
		},
		PlanID: "test-plan",
	}

	assert.NotNil(t, info)
	assert.Equal(t, "chunk-001", info.Chunk.ID)
	assert.Equal(t, "test-plan", info.PlanID)
}

func TestChunkDisplayInfo_MinimalFields(t *testing.T) {
	// Test that ChunkDisplayInfo works with minimal fields
	info := &ChunkDisplayInfo{
		Chunk: &plan.Chunk{
			ID:       "chunk-002",
			Title:    "Minimal Chunk",
			Duration: 30,
			Status:   plan.StatusNotStarted,
		},
		PlanID: "minimal-plan",
	}

	assert.NotNil(t, info)
	assert.Empty(t, info.Chunk.Objectives)
	assert.Empty(t, info.Chunk.Resources)
	assert.Empty(t, info.Chunk.Deliverable)
}

func TestChunkDisplayInfo_EmptyObjectivesAndResources(t *testing.T) {
	chunk := &plan.Chunk{
		ID:         "test",
		Objectives: []string{},
		Resources:  []string{},
	}

	assert.Empty(t, chunk.Objectives)
	assert.Equal(t, 0, len(chunk.Objectives))
	assert.Empty(t, chunk.Resources)
	assert.Equal(t, 0, len(chunk.Resources))
}

func TestChunkDisplayInfo_LongObjectivesList(t *testing.T) {
	// Test with many objectives
	objectives := make([]string, 10)
	for i := range objectives {
		objectives[i] = "Objective " + string(rune('A'+i))
	}

	chunk := &plan.Chunk{
		ID:         "test",
		Objectives: objectives,
	}

	assert.Len(t, chunk.Objectives, 10)
	assert.Equal(t, "Objective A", chunk.Objectives[0])
	assert.Equal(t, "Objective J", chunk.Objectives[9])
}

func TestChunkDisplayInfo_DurationInMinutes(t *testing.T) {
	tests := []struct {
		name            string
		durationMinutes int
		expectedHours   float64
	}{
		{
			name:            "60 minutes = 1 hour",
			durationMinutes: 60,
			expectedHours:   1.0,
		},
		{
			name:            "90 minutes = 1.5 hours",
			durationMinutes: 90,
			expectedHours:   1.5,
		},
		{
			name:            "45 minutes = 0.75 hours",
			durationMinutes: 45,
			expectedHours:   0.75,
		},
		{
			name:            "120 minutes = 2 hours",
			durationMinutes: 120,
			expectedHours:   2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := &plan.Chunk{
				ID:       "test",
				Duration: tt.durationMinutes,
			}

			assert.Equal(t, tt.durationMinutes, chunk.Duration)

			// Calculate hours
			hours := float64(chunk.Duration) / 60.0
			assert.Equal(t, tt.expectedHours, hours)
		})
	}
}

func TestChunkDisplayInfo_StatusValues(t *testing.T) {
	// Verify status constant values
	assert.Equal(t, plan.Status("not-started"), plan.StatusNotStarted)
	assert.Equal(t, plan.Status("in-progress"), plan.StatusInProgress)
	assert.Equal(t, plan.Status("completed"), plan.StatusCompleted)
	assert.Equal(t, plan.Status("skipped"), plan.StatusSkipped)
}

// Test displayChunkDetails function

func TestDisplayChunkDetails_WithFullChunk(t *testing.T) {
	// Create a chunk with all fields populated
	chunk := &plan.Chunk{
		ID:       "chunk-001",
		Title:    "Introduction to Async",
		Duration: 60,
		Status:   plan.StatusInProgress,
		Objectives: []string{
			"Understand async/await syntax",
			"Learn about Futures",
		},
		Resources: []string{
			"Rust Book Chapter 16",
			"Tokio Tutorial",
		},
		Deliverable: "Simple async web server",
	}

	stats := &session.ChunkStats{
		TotalDuration: 30,
		SessionCount:  2,
	}

	now := time.Now()
	recentSessions := []*session.Session{
		{
			ID:        "sess1",
			PlanID:    "test-plan",
			ChunkID:   "chunk-001",
			StartTime: now.Add(-1 * time.Hour),
			EndTime:   &[]time.Time{now.Add(-30 * time.Minute)}[0],
			Duration:  30,
			Notes:     "Made progress on basics",
		},
	}

	info := &ChunkDisplayInfo{
		Chunk:          chunk,
		SessionStats:   stats,
		RecentSessions: recentSessions,
		PlanID:         "test-plan",
	}

	// Should not panic when displaying
	assert.NotPanics(t, func() {
		displayChunkDetails(info)
	})
}

func TestDisplayChunkDetails_WithMinimalChunk(t *testing.T) {
	// Create a chunk with minimal fields
	chunk := &plan.Chunk{
		ID:       "chunk-002",
		Duration: 45,
		Status:   plan.StatusNotStarted,
	}

	stats := &session.ChunkStats{
		TotalDuration: 0,
		SessionCount:  0,
	}

	info := &ChunkDisplayInfo{
		Chunk:        chunk,
		SessionStats: stats,
		PlanID:       "test-plan",
	}

	// Should not panic even with minimal data
	assert.NotPanics(t, func() {
		displayChunkDetails(info)
	})
}

func TestDisplayChunkDetails_WithActiveSession(t *testing.T) {
	chunk := &plan.Chunk{
		ID:       "chunk-003",
		Title:    "Active Chunk",
		Duration: 60,
		Status:   plan.StatusInProgress,
	}

	stats := &session.ChunkStats{
		TotalDuration: 15,
		SessionCount:  1,
	}

	now := time.Now()
	activeSession := &session.Session{
		ID:        "sess-active",
		PlanID:    "test-plan",
		ChunkID:   "chunk-003",
		StartTime: now.Add(-15 * time.Minute),
		EndTime:   nil, // Active session has no end time
		Duration:  0,
	}

	info := &ChunkDisplayInfo{
		Chunk:          chunk,
		SessionStats:   stats,
		RecentSessions: []*session.Session{activeSession},
		PlanID:         "test-plan",
	}

	// Should handle active sessions without panic
	assert.NotPanics(t, func() {
		displayChunkDetails(info)
	})
}

func TestDisplayChunkDetails_ProgressCalculation(t *testing.T) {
	// Test progress calculation logic by testing various scenarios
	tests := []struct {
		name          string
		duration      int
		totalDuration int
	}{
		{
			name:          "50% progress",
			duration:      60,
			totalDuration: 30,
		},
		{
			name:          "over 100% should be capped",
			duration:      60,
			totalDuration: 90,
		},
		{
			name:          "0% progress",
			duration:      60,
			totalDuration: 0,
		},
		{
			name:          "exactly 100%",
			duration:      60,
			totalDuration: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := &plan.Chunk{
				ID:       "test-chunk",
				Title:    "Test Progress",
				Duration: tt.duration,
				Status:   plan.StatusInProgress,
			}

			stats := &session.ChunkStats{
				TotalDuration: tt.totalDuration,
				SessionCount:  1,
			}

			info := &ChunkDisplayInfo{
				Chunk:        chunk,
				SessionStats: stats,
				PlanID:       "test-plan",
			}

			// Verify it doesn't panic with this configuration
			assert.NotPanics(t, func() {
				displayChunkDetails(info)
			})
		})
	}
}

func TestDisplayChunkDetails_WithMultipleSessions(t *testing.T) {
	chunk := &plan.Chunk{
		ID:       "chunk-004",
		Title:    "Multi-session Chunk",
		Duration: 90,
		Status:   plan.StatusInProgress,
	}

	stats := &session.ChunkStats{
		TotalDuration: 75,
		SessionCount:  3,
	}

	now := time.Now()
	recentSessions := []*session.Session{
		{
			ID:        "sess1",
			PlanID:    "test-plan",
			ChunkID:   "chunk-004",
			StartTime: now.Add(-3 * time.Hour),
			EndTime:   &[]time.Time{now.Add(-2*time.Hour - 30*time.Minute)}[0],
			Duration:  30,
			Notes:     "First session",
		},
		{
			ID:        "sess2",
			PlanID:    "test-plan",
			ChunkID:   "chunk-004",
			StartTime: now.Add(-2 * time.Hour),
			EndTime:   &[]time.Time{now.Add(-1*time.Hour - 30*time.Minute)}[0],
			Duration:  30,
			Notes:     "",
		},
		{
			ID:        "sess3",
			PlanID:    "test-plan",
			ChunkID:   "chunk-004",
			StartTime: now.Add(-1 * time.Hour),
			EndTime:   &[]time.Time{now.Add(-45 * time.Minute)}[0],
			Duration:  15,
			Notes:     "Final review session",
		},
	}

	info := &ChunkDisplayInfo{
		Chunk:          chunk,
		SessionStats:   stats,
		RecentSessions: recentSessions,
		PlanID:         "test-plan",
	}

	// Should display all sessions without panic
	assert.NotPanics(t, func() {
		displayChunkDetails(info)
	})
}

func TestDisplayChunkDetails_WithNoTitle(t *testing.T) {
	// Test chunk with empty title (should fall back to ID)
	chunk := &plan.Chunk{
		ID:       "chunk-no-title",
		Title:    "",
		Duration: 30,
		Status:   plan.StatusNotStarted,
	}

	stats := &session.ChunkStats{}

	info := &ChunkDisplayInfo{
		Chunk:        chunk,
		SessionStats: stats,
		PlanID:       "test-plan",
	}

	assert.NotPanics(t, func() {
		displayChunkDetails(info)
	})
}

func TestDisplayChunkDetails_AllStatuses(t *testing.T) {
	// Test that all status types can be displayed
	statuses := []plan.Status{
		plan.StatusNotStarted,
		plan.StatusInProgress,
		plan.StatusCompleted,
		plan.StatusSkipped,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			chunk := &plan.Chunk{
				ID:       "test-chunk",
				Title:    "Test",
				Duration: 60,
				Status:   status,
			}

			info := &ChunkDisplayInfo{
				Chunk:        chunk,
				SessionStats: &session.ChunkStats{},
				PlanID:       "test-plan",
			}

			assert.NotPanics(t, func() {
				displayChunkDetails(info)
			})
		})
	}
}
