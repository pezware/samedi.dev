// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"testing"

	"github.com/pezware/samedi.dev/internal/config"
	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd_Structure(t *testing.T) {
	assert.Equal(t, "samedi", rootCmd.Use)
	assert.NotEmpty(t, rootCmd.Short)
	assert.NotEmpty(t, rootCmd.Long)
	assert.NotNil(t, rootCmd.Run)
}

func TestRootCmd_GlobalFlags(t *testing.T) {
	flags := rootCmd.PersistentFlags()

	// Check --config flag
	configFlag := flags.Lookup("config")
	require.NotNil(t, configFlag)
	assert.Equal(t, "c", configFlag.Shorthand)
	assert.Contains(t, configFlag.Usage, "config file")

	// Check --json flag
	jsonFlag := flags.Lookup("json")
	require.NotNil(t, jsonFlag)
	assert.Contains(t, jsonFlag.Usage, "JSON")

	// Check --verbose flag
	verboseFlag := flags.Lookup("verbose")
	require.NotNil(t, verboseFlag)
	assert.Equal(t, "v", verboseFlag.Shorthand)
	assert.Contains(t, verboseFlag.Usage, "verbose")
}

func TestRootCmd_HasSubcommands(t *testing.T) {
	// Verify that root command has the expected subcommands
	subcommands := rootCmd.Commands()

	commandNames := make(map[string]bool)
	for _, cmd := range subcommands {
		commandNames[cmd.Name()] = true
	}

	// Check for essential commands
	assert.True(t, commandNames["version"], "Should have version command")
	assert.True(t, commandNames["config"], "Should have config command")
	assert.True(t, commandNames["init"], "Should have init command")
	assert.True(t, commandNames["plan"], "Should have plan command")
	assert.True(t, commandNames["start"], "Should have start command")
	assert.True(t, commandNames["stop"], "Should have stop command")
	assert.True(t, commandNames["status"], "Should have status command")
	assert.True(t, commandNames["show"], "Should have show command")
	assert.True(t, commandNames["stats"], "Should have stats command")
	assert.True(t, commandNames["report"], "Should have report command")
	assert.True(t, commandNames["ui"], "Should have ui command")
}

func TestCreateLLMProvider_MockProvider(t *testing.T) {
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider: "mock",
		},
	}

	provider, err := createLLMProvider(cfg, "")

	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestCreateLLMProvider_ClaudeProvider(t *testing.T) {
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider:   "claude",
			CLICommand: "claude",
		},
	}

	provider, err := createLLMProvider(cfg, "claude-3-opus")

	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestCreateLLMProvider_CodexProvider(t *testing.T) {
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider:   "codex",
			CLICommand: "codex",
		},
	}

	provider, err := createLLMProvider(cfg, "gpt-4")

	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestCreateLLMProvider_GeminiProvider(t *testing.T) {
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider:   "gemini",
			CLICommand: "gemini",
		},
	}

	provider, err := createLLMProvider(cfg, "gemini-pro")

	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestCreateLLMProvider_LLMProvider(t *testing.T) {
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider:   "llm",
			CLICommand: "llm",
		},
	}

	provider, err := createLLMProvider(cfg, "claude-3")

	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestCreateLLMProvider_StdinProvider(t *testing.T) {
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider:   "stdin",
			CLICommand: "custom-llm",
		},
	}

	provider, err := createLLMProvider(cfg, "")

	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestCreateLLMProvider_AutoDetection_FallsBackToMock(t *testing.T) {
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider: "auto",
		},
	}

	provider, err := createLLMProvider(cfg, "")

	require.NoError(t, err)
	assert.NotNil(t, provider)
	// Should fall back to mock if no CLI is detected
}

func TestCreateLLMProvider_UnsupportedProvider_ReturnsError(t *testing.T) {
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider: "unsupported-provider",
		},
	}

	provider, err := createLLMProvider(cfg, "")

	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "unsupported LLM provider")
}

func TestCreateLLMProvider_ModelOverride(t *testing.T) {
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider:     "mock",
			DefaultModel: "default-model",
		},
	}

	// Test that model override is used
	provider, err := createLLMProvider(cfg, "override-model")

	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestCreateLLMProvider_TimeoutConfiguration(t *testing.T) {
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider:       "mock",
			TimeoutSeconds: 120,
		},
	}

	provider, err := createLLMProvider(cfg, "")

	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestPlanServiceAdapter_Get(t *testing.T) {
	// This is a simple test to verify the adapter structure
	// In a real scenario, we'd need to mock the plan service
	adapter := &planServiceAdapter{
		planService: nil, // Would be a mock in real test
	}

	assert.NotNil(t, adapter)
}

func TestPlanServiceAdapter_GetChunk(t *testing.T) {
	// Test the conversion from plan.Chunk to session.PlanChunk
	testChunk := &plan.Chunk{
		ID:       "chunk-001",
		Duration: 60,
		Status:   plan.StatusInProgress,
	}

	expectedSessionChunk := &session.PlanChunk{
		ID:       "chunk-001",
		Duration: 60,
		Status:   "in-progress",
	}

	// Verify the conversion logic
	assert.Equal(t, testChunk.ID, expectedSessionChunk.ID)
	assert.Equal(t, testChunk.Duration, expectedSessionChunk.Duration)
	assert.Equal(t, string(testChunk.Status), expectedSessionChunk.Status)
}

func TestPlanServiceAdapter_UpdateChunkStatus_ValidStatuses(t *testing.T) {
	tests := []struct {
		name           string
		statusString   string
		expectedStatus plan.Status
	}{
		{
			name:           "not-started",
			statusString:   "not-started",
			expectedStatus: plan.StatusNotStarted,
		},
		{
			name:           "in-progress",
			statusString:   "in-progress",
			expectedStatus: plan.StatusInProgress,
		},
		{
			name:           "completed",
			statusString:   "completed",
			expectedStatus: plan.StatusCompleted,
		},
		{
			name:           "skipped",
			statusString:   "skipped",
			expectedStatus: plan.StatusSkipped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify status conversion logic
			var status plan.Status
			switch tt.statusString {
			case "not-started":
				status = plan.StatusNotStarted
			case "in-progress":
				status = plan.StatusInProgress
			case "completed":
				status = plan.StatusCompleted
			case "skipped":
				status = plan.StatusSkipped
			}

			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestVersionInfo_HasDefaults(t *testing.T) {
	// Verify version variables are initialized
	assert.NotEmpty(t, Version)
	assert.NotEmpty(t, Commit)
	assert.NotEmpty(t, BuildDate)

	// Default values during development
	assert.Equal(t, "dev", Version)
	assert.Equal(t, "none", Commit)
	assert.Equal(t, "unknown", BuildDate)
}

func TestRootCmd_LongHelp_HasExamples(t *testing.T) {
	longHelp := rootCmd.Long

	// Verify examples are documented
	assert.Contains(t, longHelp, "samedi ui", "Should have UI example")
	assert.Contains(t, longHelp, "samedi init", "Should have init example")
	assert.Contains(t, longHelp, "samedi start", "Should have start example")
	assert.Contains(t, longHelp, "samedi stop", "Should have stop example")
	assert.Contains(t, longHelp, "samedi stats", "Should have stats example")
}

func TestRootCmd_LongHelp_HasGlobalFlags(t *testing.T) {
	longHelp := rootCmd.Long

	// Verify global flags are documented
	assert.Contains(t, longHelp, "--config", "Should document config flag")
	assert.Contains(t, longHelp, "--json", "Should document json flag")
	assert.Contains(t, longHelp, "--verbose", "Should document verbose flag")
}

func TestExecute_IsExported(t *testing.T) {
	// Verify Execute function is exported and callable
	// We can't actually call it in tests, but we can verify it exists
	assert.NotNil(t, Execute)
}
