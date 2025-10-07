// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package llm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// ClaudeProvider executes the Claude CLI to generate LLM responses.
type ClaudeProvider struct {
	config Config
}

// NewClaudeProvider creates a new Claude CLI provider.
func NewClaudeProvider(config *Config) *ClaudeProvider {
	if config.Command == "" {
		config.Command = "claude"
	}
	if config.Model == "" {
		config.Model = "claude-sonnet-4"
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultConfig().Timeout
	}

	return &ClaudeProvider{
		config: *config,
	}
}

// Call sends a prompt to Claude CLI and returns the response.
func (c *ClaudeProvider) Call(ctx context.Context, prompt string) (string, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// Write prompt to temporary file
	tmpfile, err := os.CreateTemp("", "samedi-prompt-*.txt")
	if err != nil {
		return "", &ProviderError{
			Provider:  "claude",
			Err:       fmt.Errorf("failed to create temp file: %w", err),
			Retryable: true,
		}
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(prompt); err != nil {
		tmpfile.Close()
		return "", &ProviderError{
			Provider:  "claude",
			Err:       fmt.Errorf("failed to write prompt: %w", err),
			Retryable: true,
		}
	}
	tmpfile.Close()

	// Build command arguments
	args := []string{
		"--prompt-file", tmpfile.Name(),
	}

	// Add model if specified
	if c.config.Model != "" {
		args = append(args, "--model", c.config.Model)
	}

	// Add any custom arguments
	args = append(args, c.config.Args...)

	// Execute CLI command
	// #nosec G204 - command is user-configured in config, intentionally dynamic
	cmd := exec.CommandContext(ctx, c.config.Command, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if error is due to timeout
		if ctx.Err() == context.DeadlineExceeded {
			return "", &ProviderError{
				Provider:  "claude",
				Err:       fmt.Errorf("timeout after %v", c.config.Timeout),
				Retryable: true,
			}
		}

		// Check if command exited with error
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", &ProviderError{
				Provider:  "claude",
				Err:       fmt.Errorf("CLI error (exit code %d): %s", exitErr.ExitCode(), string(output)),
				Retryable: false,
			}
		}

		return "", &ProviderError{
			Provider:  "claude",
			Err:       fmt.Errorf("execution failed: %w", err),
			Retryable: false,
		}
	}

	return string(output), nil
}
