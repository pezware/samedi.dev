// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"testing"

	"github.com/pezware/samedi.dev/internal/plan"
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
