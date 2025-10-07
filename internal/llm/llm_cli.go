// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package llm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

// CLIProvider executes the `llm` CLI tool (by Simon Willison).
// See: https://llm.datasette.io/
//
// The `llm` tool provides a unified interface to multiple LLM providers
// including Claude, GPT-4, Gemini, and more.
//
// Installation:
//
//	pip install llm
//	llm install llm-claude-3  # For Claude support
//
// Usage:
//
//	echo "prompt" | llm -m claude-3-5-sonnet
//	llm "prompt text" -m gpt-4
type CLIProvider struct {
	config Config
}

// NewCLIProvider creates a new llm CLI provider.
func NewCLIProvider(config *Config) *CLIProvider {
	if config.Command == "" {
		config.Command = "llm"
	}
	if config.Model == "" {
		config.Model = "claude-3-5-sonnet"
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultConfig().Timeout
	}

	return &CLIProvider{
		config: *config,
	}
}

// Call sends a prompt to the llm CLI via stdin and returns the response.
func (l *CLIProvider) Call(ctx context.Context, prompt string) (string, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, l.config.Timeout)
	defer cancel()

	// Build command arguments
	args := []string{}

	// Add model if specified
	if l.config.Model != "" {
		args = append(args, "-m", l.config.Model)
	}

	// Add any custom arguments
	args = append(args, l.config.Args...)

	// Execute CLI command with prompt via stdin
	// #nosec G204 - command is user-configured in config, intentionally dynamic
	cmd := exec.CommandContext(ctx, l.config.Command, args...)

	// Pass prompt via stdin
	cmd.Stdin = bytes.NewBufferString(prompt)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if error is due to timeout
		if ctx.Err() == context.DeadlineExceeded {
			return "", &ProviderError{
				Provider:  "llm",
				Err:       fmt.Errorf("timeout after %v", l.config.Timeout),
				Retryable: true,
			}
		}

		// Check if command exited with error
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", &ProviderError{
				Provider:  "llm",
				Err:       fmt.Errorf("CLI error (exit code %d): %s", exitErr.ExitCode(), string(output)),
				Retryable: false,
			}
		}

		return "", &ProviderError{
			Provider:  "llm",
			Err:       fmt.Errorf("execution failed: %w", err),
			Retryable: false,
		}
	}

	return string(output), nil
}
