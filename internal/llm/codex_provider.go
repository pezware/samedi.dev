// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package llm

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

// CodexProvider executes the `codex` CLI.
// See: https://codex.dev
//
// The `codex` CLI provides access to OpenAI and other models
// with agentic capabilities and code execution.
//
// Installation:
//
//	npm install -g @codex/cli
//
// Usage:
//
//	codex exec "prompt text"
//	codex exec -m o3 "prompt text"
type CodexProvider struct {
	config Config
}

// NewCodexProvider creates a new Codex CLI provider.
func NewCodexProvider(config *Config) *CodexProvider {
	if config.Command == "" {
		config.Command = "codex"
	}
	// Note: We intentionally don't set a default model.
	// If config.Model is empty, Codex CLI will use its own default (currently GPT-5-Codex).
	// This allows the CLI to always use the latest/best available model.
	if config.Timeout == 0 {
		config.Timeout = DefaultConfig().Timeout
	}

	return &CodexProvider{
		config: *config,
	}
}

// Call sends a prompt to the codex CLI and returns the response.
// Uses `codex exec` for non-interactive execution.
func (c *CodexProvider) Call(ctx context.Context, prompt string) (string, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// Build command arguments
	// codex exec -m <model> "prompt"
	args := []string{"exec"}

	// Add model if specified
	if c.config.Model != "" {
		args = append(args, "-m", c.config.Model)
	}

	// Add any custom arguments
	args = append(args, c.config.Args...)

	// Add prompt as final argument
	args = append(args, prompt)

	// Execute CLI command
	// #nosec G204 - command is user-configured in config, intentionally dynamic
	cmd := exec.CommandContext(ctx, c.config.Command, args...)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if error is due to timeout
		if ctx.Err() == context.DeadlineExceeded {
			return "", &ProviderError{
				Provider:  "codex",
				Err:       fmt.Errorf("timeout after %v", c.config.Timeout),
				Retryable: true,
			}
		}

		// Check if command exited with error
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", &ProviderError{
				Provider:  "codex",
				Err:       fmt.Errorf("CLI error (exit code %d): %s", exitErr.ExitCode(), string(output)),
				Retryable: false,
			}
		}

		return "", &ProviderError{
			Provider:  "codex",
			Err:       fmt.Errorf("execution failed: %w", err),
			Retryable: false,
		}
	}

	return string(output), nil
}
