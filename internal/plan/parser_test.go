// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package plan

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFile_ValidPlan(t *testing.T) {
	plan, err := ParseFile("testdata/valid_plan.md")
	require.NoError(t, err)
	require.NotNil(t, plan)

	// Check frontmatter
	assert.Equal(t, "rust-async", plan.ID)
	assert.Equal(t, "Rust Async Programming", plan.Title)
	assert.Equal(t, 3.0, plan.TotalHours)
	assert.Equal(t, StatusNotStarted, plan.Status)
	assert.Equal(t, []string{"rust", "async", "programming"}, plan.Tags)

	// Check chunks
	require.Len(t, plan.Chunks, 3)

	// Chunk 1
	chunk1 := plan.Chunks[0]
	assert.Equal(t, "chunk-001", chunk1.ID)
	assert.Equal(t, "Async Basics", chunk1.Title)
	assert.Equal(t, 60, chunk1.Duration)
	assert.Equal(t, StatusNotStarted, chunk1.Status)
	assert.Len(t, chunk1.Objectives, 3)
	assert.Contains(t, chunk1.Objectives[0], "Future trait")
	assert.Len(t, chunk1.Resources, 2)
	assert.Equal(t, "Simple async function examples", chunk1.Deliverable)

	// Chunk 2
	chunk2 := plan.Chunks[1]
	assert.Equal(t, "chunk-002", chunk2.ID)
	assert.Equal(t, "Tokio Runtime", chunk2.Title)
	assert.Equal(t, 90, chunk2.Duration)
	assert.Equal(t, StatusInProgress, chunk2.Status)
	assert.Len(t, chunk2.Objectives, 3)
	assert.Len(t, chunk2.Resources, 2)

	// Chunk 3
	chunk3 := plan.Chunks[2]
	assert.Equal(t, "chunk-003", chunk3.ID)
	assert.Equal(t, "Advanced Patterns", chunk3.Title)
	assert.Equal(t, 30, chunk3.Duration)
	assert.Equal(t, StatusCompleted, chunk3.Status)
}

func TestParseFile_FileNotFound(t *testing.T) {
	_, err := ParseFile("testdata/nonexistent.md")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestParse_InvalidFrontmatter(t *testing.T) {
	plan, err := ParseFile("testdata/invalid_frontmatter.md")
	require.Error(t, err)
	assert.Nil(t, plan)
	assert.Contains(t, err.Error(), "missing frontmatter delimiter")
}

func TestParse_NoChunks(t *testing.T) {
	plan, err := ParseFile("testdata/no_chunks.md")
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, "empty-plan", plan.ID)
	assert.Empty(t, plan.Chunks)
}

func TestParse_MalformedChunk(t *testing.T) {
	// This should parse but validation will fail
	plan, err := ParseFile("testdata/malformed_chunk.md")
	require.NoError(t, err) // Parsing succeeds
	require.NotNil(t, plan)

	// But chunks will be missing required data
	if len(plan.Chunks) > 0 {
		// The malformed chunk won't have proper header so it won't be parsed
		// This tests that we handle gracefully
		assert.NotEmpty(t, plan.ID)
	}
}

func TestSplitFrontmatter_Valid(t *testing.T) {
	content := `---
id: test
title: Test Plan
---

Body content here`

	frontmatter, body, err := splitFrontmatter(content)
	require.NoError(t, err)

	assert.Contains(t, frontmatter, "id: test")
	assert.Contains(t, frontmatter, "title: Test Plan")
	assert.Contains(t, body, "Body content here")
}

func TestSplitFrontmatter_MissingStartDelimiter(t *testing.T) {
	content := `id: test
title: Test Plan
---

Body`

	_, _, err := splitFrontmatter(content)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing frontmatter delimiter at start")
}

func TestSplitFrontmatter_MissingEndDelimiter(t *testing.T) {
	content := `---
id: test
title: Test Plan

Body`

	_, _, err := splitFrontmatter(content)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing closing frontmatter delimiter")
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{"1 hour", "1 hour", 60, false},
		{"2 hours", "2 hours", 120, false},
		{"1.5 hours", "1.5 hours", 90, false},
		{"90 minutes", "90 minutes", 90, false},
		{"30 mins", "30 mins", 30, false},
		{"1 hr", "1 hr", 60, false},
		{"45 m", "45 m", 45, false},
		{"uppercase", "1 HOUR", 60, false},
		{"invalid format", "one hour", 0, true},
		{"invalid unit", "1 day", 0, true},
		{"no unit", "60", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDuration(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	plan := &Plan{
		ID:         "test-plan",
		Title:      "Test Plan",
		CreatedAt:  now,
		UpdatedAt:  now,
		TotalHours: 2.5,
		Status:     StatusNotStarted,
		Tags:       []string{"test"},
		Chunks: []Chunk{
			{
				ID:       "chunk-001",
				Title:    "First Chunk",
				Duration: 60,
				Status:   StatusNotStarted,
				Objectives: []string{
					"Objective 1",
					"Objective 2",
				},
				Resources: []string{
					"Resource 1",
				},
				Deliverable: "Test deliverable",
			},
			{
				ID:       "chunk-002",
				Title:    "Second Chunk",
				Duration: 90,
				Status:   StatusInProgress,
			},
		},
	}

	markdown, err := Format(plan)
	require.NoError(t, err)

	// Verify frontmatter is present
	assert.Contains(t, markdown, "---")
	assert.Contains(t, markdown, "id: test-plan")
	assert.Contains(t, markdown, "title: Test Plan")

	// Verify title
	assert.Contains(t, markdown, "# Test Plan")

	// Verify chunks
	assert.Contains(t, markdown, "## Chunk 1: First Chunk {#chunk-001}")
	assert.Contains(t, markdown, "**Duration**: 1 hour")
	assert.Contains(t, markdown, "**Status**: not-started")
	assert.Contains(t, markdown, "**Objectives**:")
	assert.Contains(t, markdown, "- Objective 1")
	assert.Contains(t, markdown, "**Resources**:")
	assert.Contains(t, markdown, "**Deliverable**: Test deliverable")

	assert.Contains(t, markdown, "## Chunk 2: Second Chunk {#chunk-002}")
	assert.Contains(t, markdown, "**Duration**: 1.5 hours")
	assert.Contains(t, markdown, "**Status**: in-progress")
}

func TestFormat_RoundTrip(t *testing.T) {
	// Test that we can parse what we format
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	original := &Plan{
		ID:         "test-roundtrip",
		Title:      "Round Trip Test",
		CreatedAt:  now,
		UpdatedAt:  now,
		TotalHours: 3.0,
		Status:     StatusNotStarted,
		Tags:       []string{"test", "roundtrip"},
		Chunks: []Chunk{
			{
				ID:          "chunk-001",
				Title:       "Test Chunk",
				Duration:    120,
				Status:      StatusNotStarted,
				Objectives:  []string{"Obj 1", "Obj 2"},
				Resources:   []string{"Res 1"},
				Deliverable: "Deliverable",
			},
		},
	}

	// Format to markdown
	markdown, err := Format(original)
	require.NoError(t, err)

	// Parse back
	parsed, err := Parse(markdown)
	require.NoError(t, err)

	// Compare (Note: YAML might have slight differences in formatting)
	assert.Equal(t, original.ID, parsed.ID)
	assert.Equal(t, original.Title, parsed.Title)
	assert.Equal(t, original.Status, parsed.Status)
	assert.Equal(t, original.Tags, parsed.Tags)

	require.Len(t, parsed.Chunks, len(original.Chunks))
	assert.Equal(t, original.Chunks[0].ID, parsed.Chunks[0].ID)
	assert.Equal(t, original.Chunks[0].Title, parsed.Chunks[0].Title)
	assert.Equal(t, original.Chunks[0].Duration, parsed.Chunks[0].Duration)
	assert.Equal(t, original.Chunks[0].Status, parsed.Chunks[0].Status)
}

func TestFormat_RoundTrip_SubHourChunks(t *testing.T) {
	// Test round-trip with sub-hour durations (regression test for hourss bug)
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	original := &Plan{
		ID:         "test-subhour",
		Title:      "Sub-Hour Test",
		CreatedAt:  now,
		UpdatedAt:  now,
		TotalHours: 2.0,
		Status:     StatusNotStarted,
		Tags:       []string{"test"},
		Chunks: []Chunk{
			{
				ID:       "chunk-001",
				Title:    "30 Minutes",
				Duration: 30, // 0.5 hours
				Status:   StatusNotStarted,
			},
			{
				ID:       "chunk-002",
				Title:    "90 Minutes",
				Duration: 90, // 1.5 hours
				Status:   StatusNotStarted,
			},
		},
	}

	// Format to markdown
	markdown, err := Format(original)
	require.NoError(t, err)

	// Verify no "hourss" bug
	assert.NotContains(t, markdown, "hourss")
	assert.Contains(t, markdown, "0.5 hours")
	assert.Contains(t, markdown, "1.5 hours")

	// Parse back
	parsed, err := Parse(markdown)
	require.NoError(t, err)

	// Verify durations survived round-trip
	require.Len(t, parsed.Chunks, 2)
	assert.Equal(t, 30, parsed.Chunks[0].Duration, "30 min chunk should survive round-trip")
	assert.Equal(t, 90, parsed.Chunks[1].Duration, "90 min chunk should survive round-trip")

	// Format again to ensure double round-trip works
	markdown2, err := Format(parsed)
	require.NoError(t, err)
	assert.NotContains(t, markdown2, "hourss")
}

func TestParseChunks_VariousFormats(t *testing.T) {
	body := `
## Chunk 1: Basic Test {#chunk-001}
**Duration**: 1 hour
**Status**: not-started
**Objectives**:
- First objective
- Second objective

**Resources**:
- Resource one

**Deliverable**: Something useful

---

## Chunk 2: No Details {#chunk-002}
**Duration**: 30 minutes
**Status**: completed
`

	chunks, err := parseChunks(body)
	require.NoError(t, err)
	require.Len(t, chunks, 2)

	assert.Equal(t, "chunk-001", chunks[0].ID)
	assert.Len(t, chunks[0].Objectives, 2)
	assert.Len(t, chunks[0].Resources, 1)
	assert.NotEmpty(t, chunks[0].Deliverable)

	assert.Equal(t, "chunk-002", chunks[1].ID)
	assert.Empty(t, chunks[1].Objectives)
	assert.Empty(t, chunks[1].Resources)
	assert.Empty(t, chunks[1].Deliverable)
}

func TestChunkHeaderRegex(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		matches bool
		title   string
		id      string
	}{
		{
			"valid format",
			"## Chunk 1: Test Title {#chunk-001}",
			true,
			"Test Title",
			"chunk-001",
		},
		{
			"with spaces",
			"## Chunk 10:   Spaced Title   {#chunk-010}  ",
			true,
			"Spaced Title",
			"chunk-010",
		},
		{
			"missing ID",
			"## Chunk 1: Test Title",
			false,
			"",
			"",
		},
		{
			"wrong format",
			"### Chunk 1: Test {#chunk-001}",
			false,
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := chunkHeaderRegex.FindStringSubmatch(tt.line)

			if tt.matches {
				require.NotNil(t, matches)
				assert.Equal(t, tt.title, matches[1])
				assert.Equal(t, tt.id, matches[2])
			} else {
				assert.Nil(t, matches)
			}
		})
	}
}

func TestFormat_DurationFormatting(t *testing.T) {
	tests := []struct {
		minutes  int
		expected string
	}{
		{60, "1 hour"},
		{120, "2 hours"},
		{90, "1.5 hours"},
		{30, "0.5 hours"},
		{75, "1.2 hours"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			plan := &Plan{
				ID:         "test",
				Title:      "Test",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				TotalHours: 1,
				Status:     StatusNotStarted,
				Chunks: []Chunk{
					{
						ID:       "chunk-001",
						Title:    "Test",
						Duration: tt.minutes,
						Status:   StatusNotStarted,
					},
				},
			}

			markdown, err := Format(plan)
			require.NoError(t, err)

			// Check that the duration formatting is correct
			assert.Contains(t, markdown, "**Duration**: "+tt.expected)
		})
	}
}
