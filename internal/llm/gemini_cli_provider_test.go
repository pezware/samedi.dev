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

func TestNewGeminiCLIProvider_Defaults(t *testing.T) {
	provider := NewGeminiCLIProvider(&Config{})

	assert.Equal(t, "gemini", provider.config.Command)
	assert.Equal(t, "gemini-2.5-pro", provider.config.Model)
	assert.Equal(t, 120*time.Second, provider.config.Timeout)
}

func TestNewGeminiCLIProvider_CustomConfig(t *testing.T) {
	cfg := &Config{
		Command: "custom-gemini",
		Model:   "gemini-ultra",
		Timeout: 60 * time.Second,
		Args:    []string{"--verbose"},
	}

	provider := NewGeminiCLIProvider(cfg)

	assert.Equal(t, "custom-gemini", provider.config.Command)
	assert.Equal(t, "gemini-ultra", provider.config.Model)
	assert.Equal(t, 60*time.Second, provider.config.Timeout)
	assert.Equal(t, []string{"--verbose"}, provider.config.Args)
}

func TestGeminiCLIProvider_Call_CommandNotFound(t *testing.T) {
	// Use a command that definitely doesn't exist
	cfg := &Config{
		Command: "nonexistent-gemini-command-12345",
		Timeout: 5 * time.Second,
	}
	provider := NewGeminiCLIProvider(cfg)
	ctx := context.Background()

	response, err := provider.Call(ctx, "test prompt")

	require.Error(t, err)
	assert.Empty(t, response)

	var providerErr *ProviderError
	require.ErrorAs(t, err, &providerErr)
	assert.Equal(t, "gemini", providerErr.Provider)
	assert.False(t, providerErr.Retryable) // Command not found is not retryable
}
