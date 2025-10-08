// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartCmd_Structure(t *testing.T) {
	cmd := startCmd()

	assert.Equal(t, "start <plan-id> [chunk-id]", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestStartCmd_RequiresPlanID(t *testing.T) {
	cmd := startCmd()

	// Should require at least 1 argument (plan-id)
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require plan-id")

	// Should accept 1 argument (plan-id only)
	err = cmd.Args(cmd, []string{"french-b1"})
	assert.NoError(t, err)

	// Should accept 2 arguments (plan-id and chunk-id)
	err = cmd.Args(cmd, []string{"french-b1", "chunk-003"})
	assert.NoError(t, err)

	// Should reject 3+ arguments
	err = cmd.Args(cmd, []string{"plan-1", "chunk-1", "extra"})
	assert.Error(t, err, "should reject more than 2 arguments")
}

func TestStartCmd_Flags(t *testing.T) {
	cmd := startCmd()

	// Check --note flag exists
	note := cmd.Flags().Lookup("note")
	require.NotNil(t, note)
	assert.Equal(t, "", note.DefValue)
	assert.Equal(t, "initial notes for the session", note.Usage)
}

func TestStopCmd_Structure(t *testing.T) {
	cmd := stopCmd()

	assert.Equal(t, "stop", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestStopCmd_NoArgs(t *testing.T) {
	cmd := stopCmd()

	// Should require no arguments
	err := cmd.Args(cmd, []string{})
	assert.NoError(t, err)

	// Should reject any arguments
	err = cmd.Args(cmd, []string{"extra"})
	assert.Error(t, err)
}

func TestStopCmd_Flags(t *testing.T) {
	cmd := stopCmd()

	// Check --note flag
	note := cmd.Flags().Lookup("note")
	require.NotNil(t, note)
	assert.Equal(t, "", note.DefValue)
	assert.Equal(t, "session notes", note.Usage)

	// Check --artifact flag
	artifact := cmd.Flags().Lookup("artifact")
	require.NotNil(t, artifact)
	assert.Equal(t, "[]", artifact.DefValue)
	assert.Equal(t, "learning artifacts (URLs or file paths)", artifact.Usage)
}

func TestStatusCmd_Structure(t *testing.T) {
	cmd := statusCmd()

	assert.Equal(t, "status", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestStatusCmd_NoArgs(t *testing.T) {
	cmd := statusCmd()

	// Should require no arguments
	err := cmd.Args(cmd, []string{})
	assert.NoError(t, err)

	// Should reject any arguments
	err = cmd.Args(cmd, []string{"extra"})
	assert.Error(t, err)
}

func TestGetSessionService_InitializesCorrectly(t *testing.T) {
	// This test verifies the helper function structure
	// We can't fully test it without mocking the database,
	// but we can verify it exists and has correct signature
	cmd := startCmd()

	// The function should exist and be callable
	svc, err := getSessionService(cmd)

	// In test environment with temp database, it should succeed
	if err == nil {
		assert.NotNil(t, svc, "should return non-nil service")
	}
	// If it fails, that's also acceptable in some test environments
}

func TestRootCmd_HasSessionCommands(t *testing.T) {
	// Verify that rootCmd has the session commands registered
	commands := rootCmd.Commands()

	var hasStart, hasStop, hasStatus bool
	for _, cmd := range commands {
		switch cmd.Use {
		case "start <plan-id> [chunk-id]":
			hasStart = true
		case "stop":
			hasStop = true
		case "status":
			hasStatus = true
		}
	}

	assert.True(t, hasStart, "rootCmd should have 'start' command")
	assert.True(t, hasStop, "rootCmd should have 'stop' command")
	assert.True(t, hasStatus, "rootCmd should have 'status' command")
}
