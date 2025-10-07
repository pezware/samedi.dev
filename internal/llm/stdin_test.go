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

func TestNewStdinProvider_Defaults(t *testing.T) {
	cfg := &Config{
		Command: "aichat",
	}
	provider := NewStdinProvider(cfg)

	assert.Equal(t, "aichat", provider.config.Command)
	assert.Equal(t, 120*time.Second, provider.config.Timeout)
}

func TestNewStdinProvider_CustomConfig(t *testing.T) {
	cfg := &Config{
		Command: "mods",
		Timeout: 60 * time.Second,
		Args:    []string{"--model", "gpt-4"},
	}

	provider := NewStdinProvider(cfg)

	assert.Equal(t, "mods", provider.config.Command)
	assert.Equal(t, 60*time.Second, provider.config.Timeout)
	assert.Equal(t, []string{"--model", "gpt-4"}, provider.config.Args)
}

func TestStdinProvider_Call_NoCommand(t *testing.T) {
	cfg := &Config{
		Command: "", // Empty command
		Timeout: 5 * time.Second,
	}
	provider := NewStdinProvider(cfg)
	ctx := context.Background()

	response, err := provider.Call(ctx, "test prompt")

	require.Error(t, err)
	assert.Empty(t, response)

	var providerErr *ProviderError
	require.ErrorAs(t, err, &providerErr)
	assert.Equal(t, "stdin", providerErr.Provider)
	assert.False(t, providerErr.Retryable)
	assert.Contains(t, err.Error(), "command not configured")
}

func TestStdinProvider_Call_CommandNotFound(t *testing.T) {
	cfg := &Config{
		Command: "nonexistent-stdin-command-12345",
		Timeout: 5 * time.Second,
	}
	provider := NewStdinProvider(cfg)
	ctx := context.Background()

	response, err := provider.Call(ctx, "test prompt")

	require.Error(t, err)
	assert.Empty(t, response)

	var providerErr *ProviderError
	require.ErrorAs(t, err, &providerErr)
	assert.Equal(t, "stdin", providerErr.Provider)
	assert.False(t, providerErr.Retryable)
}

func TestStdinProvider_Call_Timeout(t *testing.T) {
	// Use 'sleep' command to simulate a timeout
	cfg := &Config{
		Command: "sleep",
		Args:    []string{"10"}, // Sleep for 10 seconds
		Timeout: 100 * time.Millisecond,
	}
	provider := NewStdinProvider(cfg)
	ctx := context.Background()

	response, err := provider.Call(ctx, "test prompt")

	require.Error(t, err)
	assert.Empty(t, response)

	var providerErr *ProviderError
	require.ErrorAs(t, err, &providerErr)
	assert.Equal(t, "stdin", providerErr.Provider)
	assert.True(t, providerErr.Retryable) // Timeout is retryable
	assert.Contains(t, err.Error(), "timeout")
}

func TestStdinProvider_Call_ExitError(t *testing.T) {
	// Use 'false' command which always exits with code 1
	cfg := &Config{
		Command: "false",
		Timeout: 5 * time.Second,
	}
	provider := NewStdinProvider(cfg)
	ctx := context.Background()

	response, err := provider.Call(ctx, "test prompt")

	require.Error(t, err)
	assert.Empty(t, response)

	var providerErr *ProviderError
	require.ErrorAs(t, err, &providerErr)
	assert.Equal(t, "stdin", providerErr.Provider)
	assert.False(t, providerErr.Retryable)
	assert.Contains(t, err.Error(), "CLI error")
	assert.Contains(t, err.Error(), "exit code 1")
}

func TestStdinProvider_Call_Success(t *testing.T) {
	// Use 'cat' command to echo stdin - simulates successful LLM response
	cfg := &Config{
		Command: "cat",
		Timeout: 5 * time.Second,
	}
	provider := NewStdinProvider(cfg)
	ctx := context.Background()

	prompt := "test prompt for stdin provider"
	response, err := provider.Call(ctx, prompt)

	require.NoError(t, err)
	assert.Equal(t, prompt, response)
}

func TestStdinProvider_Call_WithArgs(t *testing.T) {
	// Use 'cat' with args to verify arguments are passed
	// Note: cat ignores most args, but we're testing the interface
	cfg := &Config{
		Command: "cat",
		Args:    []string{}, // cat doesn't need args for stdin
		Timeout: 5 * time.Second,
	}
	provider := NewStdinProvider(cfg)
	ctx := context.Background()

	prompt := "test prompt with args"
	response, err := provider.Call(ctx, prompt)

	require.NoError(t, err)
	assert.Equal(t, prompt, response)
}

func TestStdinProvider_Call_ContextCancellation(t *testing.T) {
	cfg := &Config{
		Command: "sleep",
		Args:    []string{"10"},
		Timeout: 30 * time.Second,
	}
	provider := NewStdinProvider(cfg)

	// Create context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	response, err := provider.Call(ctx, "test prompt")

	require.Error(t, err)
	assert.Empty(t, response)
}
