// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"testing"

	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherStartInputs_ValidInputs_Success(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		opts           startOptions
		expectedPlanID string
		expectedChunk  string
		expectedNote   string
	}{
		{
			name:           "plan ID only",
			args:           []string{"rust-async"},
			opts:           startOptions{noPrompt: true},
			expectedPlanID: "rust-async",
			expectedChunk:  "",
			expectedNote:   "",
		},
		{
			name:           "plan ID and chunk ID",
			args:           []string{"rust-async", "chunk-001"},
			opts:           startOptions{noPrompt: true},
			expectedPlanID: "rust-async",
			expectedChunk:  "chunk-001",
			expectedNote:   "",
		},
		{
			name: "with note flag",
			args: []string{"rust-async"},
			opts: startOptions{
				noPrompt:    true,
				note:        stringPtr("Working on tokio"),
				noteFlagSet: true,
			},
			expectedPlanID: "rust-async",
			expectedChunk:  "",
			expectedNote:   "Working on tokio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planID, chunkID, note, err := gatherStartInputs(nil, tt.args, tt.opts)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedPlanID, planID)
			assert.Equal(t, tt.expectedChunk, chunkID)
			assert.Equal(t, tt.expectedNote, note)
		})
	}
}

func TestGatherStartInputs_EmptyPlanID_ReturnsError(t *testing.T) {
	_, _, _, err := gatherStartInputs(nil, []string{}, startOptions{noPrompt: true})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan ID is required")
}

func TestPromptForInitialNote_ValidInput_ReturnsNote(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedNote string
		expectError  bool
	}{
		{
			name:         "simple note",
			input:        "Working on chapter 3\n",
			expectedNote: "Working on chapter 3",
			expectError:  false,
		},
		{
			name:         "empty note",
			input:        "\n",
			expectedNote: "",
			expectError:  false,
		},
		{
			name:         "EOF without newline",
			input:        "",
			expectedNote: "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := testutil.MockReader(tt.input)
			writer := testutil.MockWriter()

			note, err := promptForInitialNote(reader, writer)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNote, note)
			}

			output := writer.String()
			assert.Contains(t, output, "Initial note (optional):")
		})
	}
}

func TestNextActiveChunk_ReturnsFirstNonCompleted(t *testing.T) {
	tests := []struct {
		name        string
		plan        *plan.Plan
		expectedID  string
		shouldBeNil bool
	}{
		{
			name: "returns first not-started chunk",
			plan: &plan.Plan{
				Chunks: []plan.Chunk{
					{ID: "chunk-001", Status: plan.StatusNotStarted},
					{ID: "chunk-002", Status: plan.StatusNotStarted},
				},
			},
			expectedID:  "chunk-001",
			shouldBeNil: false,
		},
		{
			name: "skips completed chunks",
			plan: &plan.Plan{
				Chunks: []plan.Chunk{
					{ID: "chunk-001", Status: plan.StatusCompleted},
					{ID: "chunk-002", Status: plan.StatusInProgress},
					{ID: "chunk-003", Status: plan.StatusNotStarted},
				},
			},
			expectedID:  "chunk-002",
			shouldBeNil: false,
		},
		{
			name: "returns nil when all completed",
			plan: &plan.Plan{
				Chunks: []plan.Chunk{
					{ID: "chunk-001", Status: plan.StatusCompleted},
					{ID: "chunk-002", Status: plan.StatusCompleted},
				},
			},
			shouldBeNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := nextActiveChunk(tt.plan)

			if tt.shouldBeNil {
				assert.Nil(t, chunk)
			} else {
				require.NotNil(t, chunk)
				assert.Equal(t, tt.expectedID, chunk.ID)
			}
		})
	}
}

func TestJoinSample_LimitsOutput(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		limit    int
		expected string
	}{
		{
			name:     "empty slice",
			values:   []string{},
			limit:    3,
			expected: "-",
		},
		{
			name:     "within limit",
			values:   []string{"a", "b", "c"},
			limit:    5,
			expected: "a, b, c",
		},
		{
			name:     "exceeds limit",
			values:   []string{"a", "b", "c", "d", "e"},
			limit:    3,
			expected: "a, b, c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinSample(tt.values, tt.limit)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintAllChunkDetails_FormatsCorrectly(t *testing.T) {
	plan := testutil.NewTestPlan(t)
	writer := testutil.MockWriter()

	printAllChunkDetails(writer, plan)

	output := writer.String()

	// Check that all chunks are printed
	assert.Contains(t, output, "chunk-001")
	assert.Contains(t, output, "Test Chunk 1")
	assert.Contains(t, output, "60 min")
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}
