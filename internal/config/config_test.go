// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	require.NotNil(t, cfg)

	// Check LLM defaults
	assert.Equal(t, "auto", cfg.LLM.Provider)
	assert.Equal(t, "", cfg.LLM.CLICommand)
	assert.Equal(t, 300, cfg.LLM.TimeoutSeconds)

	// Check storage defaults
	assert.NotEmpty(t, cfg.Storage.DataDir)
	assert.True(t, cfg.Storage.BackupEnabled)

	// Check TUI defaults
	assert.Equal(t, "dracula", cfg.TUI.Theme)

	// Check learning defaults
	assert.Equal(t, 60, cfg.Learning.DefaultChunkMinutes)
	assert.True(t, cfg.Learning.StreakTracking)
}

func TestConfig_Validate_ValidConfig(t *testing.T) {
	cfg := DefaultConfig()
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_InvalidProvider(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LLM.Provider = "invalid"

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid LLM provider")
}

func TestConfig_Validate_InvalidTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout int
	}{
		{"too small", 5},
		{"too large", 700},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.LLM.TimeoutSeconds = tt.timeout

			err := cfg.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "timeout must be between")
		})
	}
}

func TestConfig_Validate_EmptyDataDir(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Storage.DataDir = ""

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "data_dir cannot be empty")
}

func TestConfig_Validate_InvalidTheme(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TUI.Theme = "invalid"

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid TUI theme")
}

func TestConfig_Validate_InvalidFirstDayOfWeek(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TUI.FirstDayOfWeek = "wednesday"

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid first_day_of_week")
}

func TestConfig_Validate_ProviderCommandMismatch(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		command     string
		shouldError bool
	}{
		{"claude with llm command", "claude", "llm", true},
		{"claude with empty command", "claude", "", false},
		{"claude with correct command", "claude", "claude", false},
		{"llm with claude command", "llm", "claude", true},
		{"auto with any command", "auto", "anything", false},
		{"mock with any command", "mock", "anything", false},
		{"gemini with llm command", "gemini", "llm", true},
		{"gemini with correct command", "gemini", "gemini", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.LLM.Provider = tt.provider
			cfg.LLM.CLICommand = tt.command

			err := cfg.Validate()
			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "provider/command mismatch")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
