// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package llm

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCLIProvider_Defaults(t *testing.T) {
	provider := NewCLIProvider(&Config{})

	assert.Equal(t, "llm", provider.config.Command)
	assert.Equal(t, "claude-3-5-sonnet", provider.config.Model)
	assert.Equal(t, 120*time.Second, provider.config.Timeout)
}

func TestNewCLIProvider_CustomConfig(t *testing.T) {
	cfg := &Config{
		Command: "custom-llm",
		Model:   "gpt-4",
		Timeout: 60 * time.Second,
		Args:    []string{"--verbose"},
	}

	provider := NewCLIProvider(cfg)

	assert.Equal(t, "custom-llm", provider.config.Command)
	assert.Equal(t, "gpt-4", provider.config.Model)
	assert.Equal(t, 60*time.Second, provider.config.Timeout)
	assert.Equal(t, []string{"--verbose"}, provider.config.Args)
}

func TestCLIProvider_Call_CommandNotFound(t *testing.T) {
	// Use a command that definitely doesn't exist
	cfg := &Config{
		Command: "nonexistent-llm-command-12345",
		Timeout: 5 * time.Second,
	}
	provider := NewCLIProvider(cfg)
	ctx := context.Background()

	response, err := provider.Call(ctx, "test prompt")

	require.Error(t, err)
	assert.Empty(t, response)

	var providerErr *ProviderError
	require.ErrorAs(t, err, &providerErr)
	assert.Equal(t, "llm", providerErr.Provider)
	assert.False(t, providerErr.Retryable) // Command not found is not retryable
}

func TestCLIProvider_Call_Timeout(t *testing.T) {
	t.Skip("Skipping timeout test - requires real llm CLI for proper testing")
	// NOTE: Add integration test with mock llm script
	// This test would verify timeout behavior with a long-running command
}

func TestCLIProvider_Call_ExitError(t *testing.T) {
	t.Skip("Skipping exit error test - requires real llm CLI for proper testing")
	// NOTE: Add integration test with mock llm script
	// This test would verify error handling when CLI exits with non-zero code
}

func TestCLIProvider_Call_Success(t *testing.T) {
	t.Skip("Skipping success test - requires real llm CLI for proper testing")
	// NOTE: Add integration test with real llm installation
	// This test would verify successful execution with proper stdin handling
}

func TestCLIProvider_Call_WithModelFlag(t *testing.T) {
	t.Skip("Skipping model flag test - requires real llm CLI for proper testing")
	// NOTE: Add integration test with real llm installation
	// This test would verify that -m flag is properly passed to llm CLI
}
