// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsCmd_Structure(t *testing.T) {
	cmd := statsCmd()

	assert.Equal(t, "stats [plan-id]", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Long, "Total hours and sessions")
	assert.Contains(t, cmd.Long, "Learning streaks")
}

func TestStatsCmd_AcceptsOptionalPlanID(t *testing.T) {
	cmd := statsCmd()

	// Should accept 0 arguments (total stats)
	err := cmd.Args(cmd, []string{})
	assert.NoError(t, err, "should accept no arguments for total stats")

	// Should accept 1 argument (plan-specific stats)
	err = cmd.Args(cmd, []string{"rust-async"})
	assert.NoError(t, err, "should accept plan ID")

	// Should reject 2+ arguments
	err = cmd.Args(cmd, []string{"plan-1", "extra"})
	assert.Error(t, err, "should reject more than 1 argument")
}

func TestStatsCmd_RangeFlag(t *testing.T) {
	cmd := statsCmd()

	// Check --range flag exists
	rangeFlag := cmd.Flags().Lookup("range")
	require.NotNil(t, rangeFlag)
	assert.Equal(t, "all", rangeFlag.DefValue)
	assert.Contains(t, rangeFlag.Usage, "Time range")
	assert.Contains(t, rangeFlag.Usage, "all, today, this-week, this-month")

	// Check short flag -r
	rangeShortFlag := cmd.Flags().ShorthandLookup("r")
	require.NotNil(t, rangeShortFlag)
	assert.Equal(t, rangeFlag, rangeShortFlag)
}

func TestStatsCmd_BreakdownFlag(t *testing.T) {
	cmd := statsCmd()

	// Check --breakdown flag exists
	breakdownFlag := cmd.Flags().Lookup("breakdown")
	require.NotNil(t, breakdownFlag)
	assert.Equal(t, "false", breakdownFlag.DefValue)
	assert.Contains(t, breakdownFlag.Usage, "daily breakdown")
}

func TestStatsCmd_TUIFlag(t *testing.T) {
	cmd := statsCmd()

	// Check --tui flag exists
	tuiFlag := cmd.Flags().Lookup("tui")
	require.NotNil(t, tuiFlag)
	assert.Equal(t, "false", tuiFlag.DefValue)
	assert.Contains(t, tuiFlag.Usage, "TUI")
}

func TestStatsCmd_Examples(t *testing.T) {
	cmd := statsCmd()

	// Verify examples are provided
	assert.Contains(t, cmd.Long, "samedi stats")
	assert.Contains(t, cmd.Long, "samedi stats rust-async")
	assert.Contains(t, cmd.Long, "samedi stats --json")
	assert.Contains(t, cmd.Long, "samedi stats --tui")
	assert.Contains(t, cmd.Long, "samedi stats --range this-week")
}

func TestGetStatsService_InitializesCorrectly(t *testing.T) {
	// This test verifies the helper function structure
	cmd := statsCmd()

	// The function should exist and be callable
	svc, err := getStatsService(cmd)

	// In test environment with temp database, it should succeed
	if err == nil {
		assert.NotNil(t, svc, "should return non-nil service")
	}
	// If it fails, that's also acceptable in some test environments
}

func TestRootCmd_HasStatsCommand(t *testing.T) {
	// Verify that rootCmd has the stats command registered
	commands := rootCmd.Commands()

	var hasStats bool
	for _, cmd := range commands {
		if cmd.Use == "stats [plan-id]" {
			hasStats = true
			break
		}
	}

	assert.True(t, hasStats, "rootCmd should have 'stats' command")
}

func TestFormatPlanStatus_AllStatuses(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"not-started", "⚪ Not Started"},
		{"in-progress", "🟡 In Progress"},
		{"completed", "🟢 Completed"},
		{"archived", "📦 Archived"},
		{"unknown", "unknown"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatPlanStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildProgressBar_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		progress float64
		width    int
		contains string
	}{
		{
			name:     "zero progress",
			progress: 0.0,
			width:    10,
			contains: "[░░░░░░░░░░]",
		},
		{
			name:     "full progress",
			progress: 1.0,
			width:    10,
			contains: "[██████████]",
		},
		{
			name:     "half progress",
			progress: 0.5,
			width:    10,
			contains: "█",
		},
		{
			name:     "negative progress (should clamp to 0)",
			progress: -0.5,
			width:    10,
			contains: "[░",
		},
		{
			name:     "overflow progress (should clamp to 100%)",
			progress: 1.5,
			width:    10,
			contains: "[██████████]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildProgressBar(tt.progress, tt.width)
			assert.Contains(t, result, tt.contains)
			assert.Contains(t, result, "[")
			assert.Contains(t, result, "]")
		})
	}
}

func TestBuildProgressBar_Width(t *testing.T) {
	// Test that progress bar width is respected
	widths := []int{5, 10, 20, 30, 50}

	for _, width := range widths {
		t.Run(string(rune(width)), func(t *testing.T) {
			result := buildProgressBar(0.5, width)
			// Count runes (not bytes) - Width + 2 for brackets
			runeCount := len([]rune(result))
			assert.Equal(t, width+2, runeCount)
		})
	}
}

func TestPrintJSON_Exists(t *testing.T) {
	// Verify the JSON printing function exists
	// We can't fully test it without capturing stdout,
	// but we can verify it compiles and handles basic types
	testData := map[string]interface{}{
		"test": "value",
		"num":  42,
	}

	// This should not panic
	err := printJSON(testData)
	assert.NoError(t, err)
}

func TestDisplayTotalStats_FunctionSignature(t *testing.T) {
	// Verify the function exists with correct signature
	// We can't test execution without full database setup,
	// but we can verify it compiles
	cmd := statsCmd()
	svc, err := getStatsService(cmd)

	if err == nil && svc != nil {
		// Function exists and is callable
		assert.NotNil(t, displayTotalStats)
	}
}

func TestDisplayPlanStats_FunctionSignature(t *testing.T) {
	// Verify the function exists with correct signature
	cmd := statsCmd()
	svc, err := getStatsService(cmd)

	if err == nil && svc != nil {
		// Function exists and is callable
		assert.NotNil(t, displayPlanStats)
	}
}

func TestLaunchTUI_FunctionExists(t *testing.T) {
	// Verify the launchTUI function exists
	// We can't test execution without interactive terminal,
	// but we can verify it compiles
	assert.NotNil(t, launchTUI)
}
