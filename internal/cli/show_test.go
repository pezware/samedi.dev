// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowCmd_Structure(t *testing.T) {
	cmd := showCmd()

	assert.Equal(t, "show <plan-id> <chunk-id>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.Run)
}

func TestShowCmd_RequiresExactlyTwoArgs(t *testing.T) {
	cmd := showCmd()

	// Should reject 0 arguments
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 2 arguments")

	// Should reject 1 argument
	err = cmd.Args(cmd, []string{"plan-id"})
	assert.Error(t, err, "should require exactly 2 arguments")

	// Should accept 2 arguments
	err = cmd.Args(cmd, []string{"plan-id", "chunk-id"})
	assert.NoError(t, err)

	// Should reject 3+ arguments
	err = cmd.Args(cmd, []string{"plan-id", "chunk-id", "extra"})
	assert.Error(t, err, "should reject more than 2 arguments")
}

func TestShowCmd_UsageExamples(t *testing.T) {
	cmd := showCmd()

	// Verify usage examples are documented
	assert.Contains(t, cmd.Long, "rust-async")
	assert.Contains(t, cmd.Long, "chunk-001")
	assert.Contains(t, cmd.Long, "french-b1")
}
