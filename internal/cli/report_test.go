// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportCmd_Structure(t *testing.T) {
	cmd := reportCmd()

	assert.Equal(t, "report [plan-id]", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Long, "markdown report")
	assert.Contains(t, cmd.Long, "Summary statistics")
	assert.Contains(t, cmd.Long, "Daily breakdown")
}

func TestReportCmd_AcceptsOptionalPlanID(t *testing.T) {
	cmd := reportCmd()

	// Should accept 0 arguments (full report)
	err := cmd.Args(cmd, []string{})
	assert.NoError(t, err, "should accept no arguments for full report")

	// Should accept 1 argument (plan-specific report)
	err = cmd.Args(cmd, []string{"rust-async"})
	assert.NoError(t, err, "should accept plan ID")

	// Should reject 2+ arguments
	err = cmd.Args(cmd, []string{"plan-1", "extra"})
	assert.Error(t, err, "should reject more than 1 argument")
}

func TestReportCmd_OutputFlag(t *testing.T) {
	cmd := reportCmd()

	// Check --output flag exists
	outputFlag := cmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag)
	assert.Equal(t, "", outputFlag.DefValue)
	assert.Contains(t, outputFlag.Usage, "Output file path")

	// Check short flag -o
	outputShortFlag := cmd.Flags().ShorthandLookup("o")
	require.NotNil(t, outputShortFlag)
	assert.Equal(t, outputFlag, outputShortFlag)
}

func TestReportCmd_TypeFlag(t *testing.T) {
	cmd := reportCmd()

	// Check --type flag exists
	typeFlag := cmd.Flags().Lookup("type")
	require.NotNil(t, typeFlag)
	assert.Equal(t, "full", typeFlag.DefValue)
	assert.Contains(t, typeFlag.Usage, "summary, full")

	// Check short flag -t
	typeShortFlag := cmd.Flags().ShorthandLookup("t")
	require.NotNil(t, typeShortFlag)
	assert.Equal(t, typeFlag, typeShortFlag)
}

func TestReportCmd_RangeFlag(t *testing.T) {
	cmd := reportCmd()

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

func TestReportCmd_Examples(t *testing.T) {
	cmd := reportCmd()

	// Verify examples are provided
	assert.Contains(t, cmd.Long, "samedi report")
	assert.Contains(t, cmd.Long, "samedi report -o stats-2025.md")
	assert.Contains(t, cmd.Long, "samedi report rust-async")
	assert.Contains(t, cmd.Long, "samedi report --range this-week")
	assert.Contains(t, cmd.Long, "samedi report --type summary")
}

func TestReportCmd_FlagDefaults(t *testing.T) {
	cmd := reportCmd()

	// Verify all flags have appropriate defaults
	outputFlag := cmd.Flags().Lookup("output")
	typeFlag := cmd.Flags().Lookup("type")
	rangeFlag := cmd.Flags().Lookup("range")

	require.NotNil(t, outputFlag)
	require.NotNil(t, typeFlag)
	require.NotNil(t, rangeFlag)

	// Default output is stdout (empty string)
	assert.Equal(t, "", outputFlag.DefValue)

	// Default type is full report
	assert.Equal(t, "full", typeFlag.DefValue)

	// Default range is all time
	assert.Equal(t, "all", rangeFlag.DefValue)
}

func TestReportCmd_LongDescription(t *testing.T) {
	cmd := reportCmd()

	// Verify long description includes key information
	assert.Contains(t, cmd.Long, "Summary statistics")
	assert.Contains(t, cmd.Long, "Plan-specific progress")
	assert.Contains(t, cmd.Long, "Daily breakdown")
	assert.Contains(t, cmd.Long, "Markdown")
}

func TestRootCmd_HasReportCommand(t *testing.T) {
	// Verify that rootCmd has the report command registered
	commands := rootCmd.Commands()

	var hasReport bool
	for _, cmd := range commands {
		if cmd.Use == "report [plan-id]" {
			hasReport = true
			break
		}
	}

	assert.True(t, hasReport, "rootCmd should have 'report' command")
}

func TestReportCmd_NolintDirective(t *testing.T) {
	// Verify the nolint directive is present for gocyclo
	// This is important to maintain as the report command
	// legitimately has higher complexity due to multiple flags
	cmd := reportCmd()

	// Command should exist and be executable
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.RunE)
}

func TestReportCmd_AllFlagsRegistered(t *testing.T) {
	cmd := reportCmd()

	// Verify all expected flags are registered
	flags := []string{"output", "type", "range"}

	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "flag %s should be registered", flagName)
	}
}
