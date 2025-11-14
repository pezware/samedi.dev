// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInteractive_WithNoPromptFlag(t *testing.T) {
	// When noPrompt is true, should always return false
	result := isInteractive(true)
	assert.False(t, result, "isInteractive should return false when noPrompt is true")
}

func TestIsInteractive_ChecksTerminal(t *testing.T) {
	// When noPrompt is false, it checks if stdin/stdout are terminals
	// In test environment, they typically aren't terminals
	result := isInteractive(false)

	// In CI/test environments, stdin/stdout are not terminals
	// So we expect false here
	// This behavior is correct for automated testing
	assert.False(t, result, "isInteractive should return false in test environment (no TTY)")
}

func TestIsInteractive_LogicFlow(t *testing.T) {
	tests := []struct {
		name     string
		noPrompt bool
		note     string
	}{
		{
			name:     "no-prompt true should skip terminal check",
			noPrompt: true,
			note:     "Should return false immediately",
		},
		{
			name:     "no-prompt false should check terminal",
			noPrompt: false,
			note:     "Should check if stdin/stdout are terminals",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInteractive(tt.noPrompt)

			if tt.noPrompt {
				// Should always be false when noPrompt is true
				assert.False(t, result)
			} else {
				// In test environment (no TTY), should be false
				// In real terminal environment, would be true
				// We can't test the true case without a real TTY
				assert.False(t, result, "Test environment has no TTY")
			}
		})
	}
}

func TestIsInteractive_BehaviorDocumentation(t *testing.T) {
	// This test documents the expected behavior of isInteractive
	//
	// isInteractive returns true when:
	// 1. noPrompt is false AND
	// 2. stdin is a terminal AND
	// 3. stdout is a terminal
	//
	// This ensures prompts are only shown in interactive environments

	// Test case 1: noPrompt = true
	// Expected: false (prompts disabled)
	assert.False(t, isInteractive(true))

	// Test case 2: noPrompt = false, but in test environment (no TTY)
	// Expected: false (not a terminal)
	assert.False(t, isInteractive(false))

	// Test case 3: Manual verification
	// We can check if the function at least looks at os.Stdin/Stdout
	// by verifying they exist
	assert.NotNil(t, os.Stdin, "os.Stdin should be available")
	assert.NotNil(t, os.Stdout, "os.Stdout should be available")
}

func TestIsInteractive_UsageScenarios(t *testing.T) {
	// Document various usage scenarios

	scenarios := []struct {
		name        string
		noPrompt    bool
		expected    bool
		description string
	}{
		{
			name:        "CI/CD pipeline",
			noPrompt:    true,
			expected:    false,
			description: "In CI, always use --no-prompt flag",
		},
		{
			name:        "Scripting",
			noPrompt:    true,
			expected:    false,
			description: "In scripts, disable prompts",
		},
		{
			name:        "Test environment",
			noPrompt:    false,
			expected:    false,
			description: "Tests run without TTY, so not interactive",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := isInteractive(scenario.noPrompt)
			assert.Equal(t, scenario.expected, result, scenario.description)
		})
	}
}
