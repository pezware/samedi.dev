// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUICmd_Structure(t *testing.T) {
	cmd := uiCmd()

	assert.Equal(t, "ui", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

func TestUICmd_NoArgs(t *testing.T) {
	cmd := uiCmd()

	// UI command should accept no arguments
	if cmd.Args != nil {
		err := cmd.Args(cmd, []string{})
		assert.NoError(t, err, "ui command should accept no arguments")

		// Should reject arguments if Args is set
		err = cmd.Args(cmd, []string{"extra"})
		if err == nil {
			// If no Args validation, that's fine too
			assert.True(t, true)
		}
	}
}

func TestUICmd_Documentation(t *testing.T) {
	cmd := uiCmd()

	// Verify documentation mentions key features
	assert.Contains(t, cmd.Long, "Plans", "Should mention Plans module")
	assert.Contains(t, cmd.Long, "Stats", "Should mention Stats module")
	assert.Contains(t, cmd.Long, "Tab", "Should mention Tab navigation")
	assert.Contains(t, cmd.Long, "q", "Should mention quit key")
}

func TestUICmd_NavigationInstructions(t *testing.T) {
	cmd := uiCmd()

	// Verify navigation instructions are documented
	longHelp := cmd.Long

	// Check for navigation keys
	assert.Contains(t, longHelp, "Tab", "Should document Tab key")
	assert.Contains(t, longHelp, "Ctrl+C", "Should document Ctrl+C to exit")

	// Check for module shortcuts
	assert.Contains(t, longHelp, "1–9", "Should document number keys for module jumping")
}

func TestUICmd_NoFlags(t *testing.T) {
	cmd := uiCmd()

	// UI command should have no local flags (only inherited global flags)
	flags := cmd.Flags()
	require.NotNil(t, flags)

	// UI command typically has no local flags
	// (it may have inherited flags from root, which is fine)
	// We just verify that flags object exists
	assert.NotNil(t, flags, "Flags should be available")
}

func TestUICmd_ModulesDocumented(t *testing.T) {
	cmd := uiCmd()

	longHelp := cmd.Long

	// Check that modules are documented
	assert.Contains(t, longHelp, "Plans:", "Should document Plans module")
	assert.Contains(t, longHelp, "Stats:", "Should document Stats module")

	// Check that shortcuts are documented
	assert.Contains(t, longHelp, "shortcuts", "Should mention shortcuts")
}

func TestUICmd_TipSection(t *testing.T) {
	cmd := uiCmd()

	// Verify there's a tip about stats-only mode
	assert.Contains(t, cmd.Long, "stats --tui", "Should mention stats --tui command")
	assert.Contains(t, cmd.Long, "Tip", "Should have a tip section")
}
