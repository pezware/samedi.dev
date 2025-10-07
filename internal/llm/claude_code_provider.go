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

// ClaudeCodeProvider executes the `claude` CLI from Claude Code.
// See: https://claude.com/claude-code
//
// The `claude` CLI provides access to Anthropic's Claude models
// with integrated file context and tool usage capabilities.
//
// Installation:
//
//	npm install -g @anthropic/claude-code
//
// Usage:
//
//	echo "prompt" | claude -p --model sonnet
//	claude -p "prompt text" --model opus
type ClaudeCodeProvider struct {
	config Config
}

// NewClaudeCodeProvider creates a new Claude Code CLI provider.
func NewClaudeCodeProvider(config *Config) *ClaudeCodeProvider {
	if config.Command == "" {
		config.Command = "claude"
	}
	if config.Model == "" {
		config.Model = "sonnet" // Claude Code uses short aliases
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultConfig().Timeout
	}

	return &ClaudeCodeProvider{
		config: *config,
	}
}

// Call sends a prompt to the claude CLI via stdin and returns the response.
func (c *ClaudeCodeProvider) Call(ctx context.Context, prompt string) (string, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// Build command arguments
	// claude -p --model <model>
	args := []string{"-p"} // Print mode for non-interactive output

	// Add model if specified
	if c.config.Model != "" {
		args = append(args, "--model", c.config.Model)
	}

	// Add any custom arguments
	args = append(args, c.config.Args...)

	// Execute CLI command with prompt via stdin
	// #nosec G204 - command is user-configured in config, intentionally dynamic
	cmd := exec.CommandContext(ctx, c.config.Command, args...)

	// Pass prompt via stdin
	cmd.Stdin = bytes.NewBufferString(prompt)

	// Capture output
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
