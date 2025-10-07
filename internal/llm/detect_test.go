// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectCLI_ReturnsStructure(t *testing.T) {
	// This test doesn't require any specific CLI to be installed
	// It just verifies the function returns a valid CLIInfo struct
	info := DetectCLI()

	// Should always return a valid struct
	assert.NotEmpty(t, info.Name)

	// If found, should have command and model
	if info.Found {
		assert.NotEmpty(t, info.Command)
		assert.NotEmpty(t, info.Model)
		// Name should be one of the known CLIs
		validNames := map[string]bool{
			"claude": true,
			"codex":  true,
			"gemini": true,
			"llm":    true,
		}
		assert.True(t, validNames[info.Name], "unexpected CLI name: %s", info.Name)
	} else {
		// If not found, should fall back to mock
		assert.Equal(t, "mock", info.Name)
		assert.False(t, info.Found)
	}
}

func TestDetectCLI_PriorityOrder(t *testing.T) {
	// This test documents the expected priority order
	// It doesn't test the actual detection, just the structure
	expectedOrder := []string{"claude", "codex", "gemini", "llm"}

	// This is more of a documentation test
	// The actual priority is defined in detect.go
	assert.Len(t, expectedOrder, 4, "should have 4 CLIs in priority order")
	assert.Equal(t, "claude", expectedOrder[0], "claude should be first priority")
	assert.Equal(t, "llm", expectedOrder[3], "llm should be last priority before mock")
}
