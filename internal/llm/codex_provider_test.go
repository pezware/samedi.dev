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

func TestNewCodexProvider_Defaults(t *testing.T) {
	provider := NewCodexProvider(&Config{})

	assert.Equal(t, "codex", provider.config.Command)
	assert.Equal(t, "o3", provider.config.Model)
	assert.Equal(t, 120*time.Second, provider.config.Timeout)
}

func TestNewCodexProvider_CustomConfig(t *testing.T) {
	cfg := &Config{
		Command: "custom-codex",
		Model:   "gpt-4",
		Timeout: 60 * time.Second,
		Args:    []string{"--verbose"},
	}

	provider := NewCodexProvider(cfg)

	assert.Equal(t, "custom-codex", provider.config.Command)
	assert.Equal(t, "gpt-4", provider.config.Model)
	assert.Equal(t, 60*time.Second, provider.config.Timeout)
	assert.Equal(t, []string{"--verbose"}, provider.config.Args)
}

func TestCodexProvider_Call_CommandNotFound(t *testing.T) {
	// Use a command that definitely doesn't exist
	cfg := &Config{
		Command: "nonexistent-codex-command-12345",
		Timeout: 5 * time.Second,
	}
	provider := NewCodexProvider(cfg)
	ctx := context.Background()

	response, err := provider.Call(ctx, "test prompt")

	require.Error(t, err)
	assert.Empty(t, response)

	var providerErr *ProviderError
	require.ErrorAs(t, err, &providerErr)
	assert.Equal(t, "codex", providerErr.Provider)
	assert.False(t, providerErr.Retryable) // Command not found is not retryable
}
