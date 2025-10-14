// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCmd_Structure(t *testing.T) {
	cmd := initCmd()

	assert.Equal(t, "init <topic>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check flags are defined
	hours := cmd.Flags().Lookup("hours")
	require.NotNil(t, hours)
	assert.Equal(t, "40", hours.DefValue)

	level := cmd.Flags().Lookup("level")
	require.NotNil(t, level)

	goals := cmd.Flags().Lookup("goals")
	require.NotNil(t, goals)

	edit := cmd.Flags().Lookup("edit")
	require.NotNil(t, edit)
	assert.Equal(t, "false", edit.DefValue)

	noCards := cmd.Flags().Lookup("no-cards")
	require.NotNil(t, noCards)
	assert.Equal(t, "false", noCards.DefValue)

	noPrompt := cmd.Flags().Lookup("no-prompt")
	require.NotNil(t, noPrompt)
	assert.Equal(t, "false", noPrompt.DefValue)
}

func TestInitCmd_RequiresTopicArg(t *testing.T) {
	cmd := initCmd()

	// Should require exactly 1 argument
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require topic argument")

	err = cmd.Args(cmd, []string{"topic"})
	assert.NoError(t, err, "should accept one argument")

	err = cmd.Args(cmd, []string{"topic1", "topic2"})
	assert.Error(t, err, "should reject multiple arguments")
}

// Note: Full integration tests with actual plan creation will be in
// integration test suite to avoid complex mocking of the service layer

func TestPromptForHours_Default(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\n"))
	var output bytes.Buffer

	value, err := promptForHours(reader, &output, 40)
	require.NoError(t, err)
	assert.Equal(t, 40.0, value)
}

func TestPromptForHours_InvalidThenValid(t *testing.T) {
	input := "abc\n1001\n80\n"
	reader := bufio.NewReader(strings.NewReader(input))
	var output bytes.Buffer

	value, err := promptForHours(reader, &output, 40)
	require.NoError(t, err)
	assert.Equal(t, 80.0, value)
	assert.Contains(t, output.String(), "Please enter a number between 1 and 1000.")
}

func TestPromptForLevel(t *testing.T) {
	input := "expert\nIntermediate\n"
	reader := bufio.NewReader(strings.NewReader(input))
	var output bytes.Buffer

	level, err := promptForLevel(reader, &output)
	require.NoError(t, err)
	assert.Equal(t, "intermediate", level)
	assert.Contains(t, output.String(), "Please choose beginner, intermediate, advanced or leave blank.")
}

func TestPromptForGoals(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("Focus on conversation\n"))
	var output bytes.Buffer

	goals, err := promptForGoals(reader, &output)
	require.NoError(t, err)
	assert.Equal(t, "Focus on conversation", goals)
}
