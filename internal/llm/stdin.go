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

// StdinProvider is a generic provider that passes prompts to any CLI via stdin.
// This is useful for custom LLM tools that accept input via stdin.
//
// Example configurations:
//   - aichat: echo "prompt" | aichat
//   - mods: echo "prompt" | mods
//   - ollama: echo "prompt" | ollama run llama2
//
// The provider executes:
//
//	echo "<prompt>" | <command> <args...>
type StdinProvider struct {
	config Config
}

// NewStdinProvider creates a new generic stdin provider.
func NewStdinProvider(config *Config) *StdinProvider {
	if config.Timeout == 0 {
		config.Timeout = DefaultConfig().Timeout
	}

	return &StdinProvider{
		config: *config,
	}
}

// Call sends a prompt to the configured command via stdin.
func (s *StdinProvider) Call(ctx context.Context, prompt string) (string, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	// Validate command
	if s.config.Command == "" {
		return "", &ProviderError{
			Provider:  "stdin",
			Err:       fmt.Errorf("command not configured"),
			Retryable: false,
		}
	}

	// Execute CLI command with prompt via stdin
	// #nosec G204 - command is user-configured in config, intentionally dynamic
	cmd := exec.CommandContext(ctx, s.config.Command, s.config.Args...)

	// Pass prompt via stdin
	cmd.Stdin = bytes.NewBufferString(prompt)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if error is due to timeout
		if ctx.Err() == context.DeadlineExceeded {
			return "", &ProviderError{
				Provider:  "stdin",
				Err:       fmt.Errorf("timeout after %v", s.config.Timeout),
				Retryable: true,
			}
		}

		// Check if command exited with error
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", &ProviderError{
				Provider:  "stdin",
				Err:       fmt.Errorf("CLI error (exit code %d): %s", exitErr.ExitCode(), string(output)),
				Retryable: false,
			}
		}

		return "", &ProviderError{
			Provider:  "stdin",
			Err:       fmt.Errorf("execution failed: %w", err),
			Retryable: false,
		}
	}

	return string(output), nil
}
