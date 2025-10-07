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

func TestNewClaudeCodeProvider_Defaults(t *testing.T) {
	provider := NewClaudeCodeProvider(&Config{})

	assert.Equal(t, "claude", provider.config.Command)
	assert.Equal(t, "sonnet", provider.config.Model)
	assert.Equal(t, 120*time.Second, provider.config.Timeout)
}

func TestNewClaudeCodeProvider_CustomConfig(t *testing.T) {
	cfg := &Config{
		Command: "custom-claude",
		Model:   "opus",
		Timeout: 60 * time.Second,
		Args:    []string{"--verbose"},
	}

	provider := NewClaudeCodeProvider(cfg)

	assert.Equal(t, "custom-claude", provider.config.Command)
	assert.Equal(t, "opus", provider.config.Model)
	assert.Equal(t, 60*time.Second, provider.config.Timeout)
	assert.Equal(t, []string{"--verbose"}, provider.config.Args)
}

func TestClaudeCodeProvider_Call_CommandNotFound(t *testing.T) {
	// Use a command that definitely doesn't exist
	cfg := &Config{
		Command: "nonexistent-claude-command-12345",
		Timeout: 5 * time.Second,
	}
	provider := NewClaudeCodeProvider(cfg)
	ctx := context.Background()

	response, err := provider.Call(ctx, "test prompt")

	require.Error(t, err)
	assert.Empty(t, response)

	var providerErr *ProviderError
	require.ErrorAs(t, err, &providerErr)
	assert.Equal(t, "claude", providerErr.Provider)
	assert.False(t, providerErr.Retryable) // Command not found is not retryable
}
