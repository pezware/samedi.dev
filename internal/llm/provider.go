// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package llm

import (
	"context"
	"time"
)

// Provider defines the interface for LLM interactions.
// Implementations should handle API calls, CLI execution, or other
// methods of communicating with language models.
type Provider interface {
	// Call sends a prompt to the LLM and returns the response.
	// Returns an error if the call fails, times out, or produces invalid output.
	Call(ctx context.Context, prompt string) (string, error)
}

// Config holds configuration for an LLM provider.
type Config struct {
	// Provider type (e.g., "claude", "codex", "gemini", "custom")
	Provider string

	// CLI command to execute (e.g., "claude", "llm")
	Command string

	// Model identifier (e.g., "claude-sonnet-4", "gpt-4")
	Model string

	// Timeout for LLM calls
	Timeout time.Duration

	// Maximum number of retries on transient failures
	MaxRetries int

	// Custom arguments for CLI providers
	Args []string

	// Whether to pass prompt via stdin (vs command line arg)
	UseStdin bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Provider:   "claude",
		Command:    "claude",
		Model:      "claude-sonnet-4",
		Timeout:    120 * time.Second,
		MaxRetries: 2,
		UseStdin:   false,
	}
}

// ProviderError represents an error from an LLM provider.
type ProviderError struct {
	// Provider that generated the error
	Provider string

	// Original error
	Err error

	// Whether this error is retryable
	Retryable bool
}

func (e *ProviderError) Error() string {
	return e.Provider + " error: " + e.Err.Error()
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}
