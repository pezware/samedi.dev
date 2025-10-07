// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package llm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "claude", cfg.Provider)
	assert.Equal(t, "claude", cfg.Command)
	assert.Equal(t, "claude-sonnet-4", cfg.Model)
	assert.Equal(t, 120*time.Second, cfg.Timeout)
	assert.Equal(t, 2, cfg.MaxRetries)
	assert.False(t, cfg.UseStdin)
}

func TestMockProvider_Call_DefaultResponse(t *testing.T) {
	mock := NewMockProvider()
	ctx := context.Background()

	response, err := mock.Call(ctx, "any prompt")

	require.NoError(t, err)
	assert.Contains(t, response, "id: test-plan")
	assert.Contains(t, response, "title: Test Learning Plan")
	assert.Equal(t, 1, mock.CallCount)
	assert.Equal(t, "any prompt", mock.LastPrompt)
}

func TestMockProvider_Call_PatternMatching(t *testing.T) {
	mock := NewMockProvider()
	mock.Responses["french"] = "French plan response"
	mock.Responses["rust"] = "Rust plan response"
	ctx := context.Background()

	tests := []struct {
		name           string
		prompt         string
		expectedSubstr string
	}{
		{
			name:           "matches french pattern",
			prompt:         "Create a plan for learning French",
			expectedSubstr: "French plan response",
		},
		{
			name:           "matches rust pattern",
			prompt:         "Rust async programming",
			expectedSubstr: "Rust plan response",
		},
		{
			name:           "case insensitive matching",
			prompt:         "FRENCH language",
			expectedSubstr: "French plan response",
		},
		{
			name:           "no match returns default",
			prompt:         "python programming",
			expectedSubstr: "Test Learning Plan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.Reset()
			response, err := mock.Call(ctx, tt.prompt)

			require.NoError(t, err)
			assert.Contains(t, response, tt.expectedSubstr)
		})
	}
}

func TestMockProvider_Call_Error(t *testing.T) {
	mock := NewMockProvider()
	mock.ShouldError = true
	mock.ErrorMessage = "simulated error"
	ctx := context.Background()

	response, err := mock.Call(ctx, "any prompt")

	require.Error(t, err)
	assert.Empty(t, response)
	assert.Contains(t, err.Error(), "simulated error")
	assert.Equal(t, 1, mock.CallCount)

	// Verify it's a ProviderError
	var providerErr *ProviderError
	require.ErrorAs(t, err, &providerErr)
	assert.Equal(t, "mock", providerErr.Provider)
	assert.False(t, providerErr.Retryable)
}

func TestMockProvider_Reset(t *testing.T) {
	mock := NewMockProvider()
	ctx := context.Background()

	// Make some calls
	mock.Call(ctx, "first prompt")
	mock.Call(ctx, "second prompt")
	mock.ShouldError = true

	assert.Equal(t, 2, mock.CallCount)
	assert.Equal(t, "second prompt", mock.LastPrompt)
	assert.True(t, mock.ShouldError)

	// Reset
	mock.Reset()

	assert.Equal(t, 0, mock.CallCount)
	assert.Empty(t, mock.LastPrompt)
	assert.False(t, mock.ShouldError)
}

func TestNewClaudeProvider_Defaults(t *testing.T) {
	provider := NewClaudeProvider(&Config{})

	assert.Equal(t, "claude", provider.config.Command)
	assert.Equal(t, "claude-sonnet-4", provider.config.Model)
	assert.Equal(t, 120*time.Second, provider.config.Timeout)
}

func TestNewClaudeProvider_CustomConfig(t *testing.T) {
	cfg := &Config{
		Command: "custom-claude",
		Model:   "claude-opus-4",
		Timeout: 60 * time.Second,
		Args:    []string{"--verbose"},
	}

	provider := NewClaudeProvider(cfg)

	assert.Equal(t, "custom-claude", provider.config.Command)
	assert.Equal(t, "claude-opus-4", provider.config.Model)
	assert.Equal(t, 60*time.Second, provider.config.Timeout)
	assert.Equal(t, []string{"--verbose"}, provider.config.Args)
}

func TestClaudeProvider_Call_CommandNotFound(t *testing.T) {
	// Use a command that definitely doesn't exist
	cfg := &Config{
		Command: "nonexistent-command-12345",
		Timeout: 5 * time.Second,
	}
	provider := NewClaudeProvider(cfg)
	ctx := context.Background()

	response, err := provider.Call(ctx, "test prompt")

	require.Error(t, err)
	assert.Empty(t, response)

	var providerErr *ProviderError
	require.ErrorAs(t, err, &providerErr)
	assert.Equal(t, "claude", providerErr.Provider)
	assert.False(t, providerErr.Retryable) // Command not found is not retryable
}

func TestClaudeProvider_Call_Timeout(t *testing.T) {
	t.Skip("Skipping timeout test - requires real CLI that accepts --prompt-file")
	// TODO(#6): Create mock CLI script for integration testing
	// This test would verify timeout behavior with a long-running command
}

func TestClaudeProvider_Call_Success(t *testing.T) {
	t.Skip("Skipping success test - requires real CLI that accepts --prompt-file")
	// TODO(#6): Create mock CLI script for integration testing
	// This test would verify successful execution with proper argument passing
}

func TestProviderError_Error(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	err := &ProviderError{
		Provider:  "test-provider",
		Err:       originalErr,
		Retryable: true,
	}

	assert.Contains(t, err.Error(), "test-provider")
	assert.Contains(t, err.Error(), "original error")
}

func TestProviderError_Unwrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	err := &ProviderError{
		Provider: "test",
		Err:      originalErr,
	}

	assert.Equal(t, originalErr, err.Unwrap())
}
