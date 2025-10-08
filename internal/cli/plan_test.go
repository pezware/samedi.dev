// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanCmd_Structure(t *testing.T) {
	cmd := planCmd()

	assert.Equal(t, "plan", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Should have subcommands
	assert.True(t, cmd.HasSubCommands(), "plan should have subcommands")
}

func TestPlanListCmd_Structure(t *testing.T) {
	cmd := planListCmd()

	assert.Equal(t, "list", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check flags
	status := cmd.Flags().Lookup("status")
	require.NotNil(t, status)

	tag := cmd.Flags().Lookup("tag")
	require.NotNil(t, tag)

	sort := cmd.Flags().Lookup("sort")
	require.NotNil(t, sort)
	assert.Equal(t, "", sort.DefValue) // Empty means use default (created_at DESC)
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"completed", "✓ completed"},
		{"in-progress", "→ in-progress"},
		{"not-started", "○ not-started"},
		{"archived", "archived"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"this is a very long title that should be truncated", 20, "this is a very lo..."},
		{"", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
			assert.LessOrEqual(t, len(result), tt.maxLen)
		})
	}
}

func TestPlanShowCmd_Structure(t *testing.T) {
	cmd := planShowCmd()

	assert.Equal(t, "show <plan-id>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check flags
	chunks := cmd.Flags().Lookup("chunks")
	require.NotNil(t, chunks)
	assert.Equal(t, "false", chunks.DefValue)
}

func TestPlanShowCmd_RequiresPlanID(t *testing.T) {
	cmd := planShowCmd()

	// Should require exactly 1 argument
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err)

	err = cmd.Args(cmd, []string{"plan-id"})
	assert.NoError(t, err)

	err = cmd.Args(cmd, []string{"plan-1", "plan-2"})
	assert.Error(t, err)
}

func TestPlanEditCmd_Structure(t *testing.T) {
	cmd := planEditCmd()

	assert.Equal(t, "edit <plan-id>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestPlanEditCmd_RequiresPlanID(t *testing.T) {
	cmd := planEditCmd()

	// Should require exactly 1 argument
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err)

	err = cmd.Args(cmd, []string{"plan-id"})
	assert.NoError(t, err)
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		minutes  int
		expected string
	}{
		{30, "30min"},
		{45, "45min"},
		{60, "1h"},
		{90, "1.5h"},
		{120, "2h"},
		{135, "2.2h"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.minutes)
			assert.Equal(t, tt.expected, result)
		})
	}
}
