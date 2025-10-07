// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package llm

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

// GeminiCLIProvider executes the `gemini` CLI.
// See: https://github.com/google/gemini-cli
//
// The `gemini` CLI provides access to Google's Gemini models
// with agentic capabilities and multimodal support.
//
// Installation:
//
//	npm install -g @google/gemini-cli
//
// Usage:
//
//	gemini -p "prompt text"
//	gemini -p "prompt text" -m gemini-2.5-pro
type GeminiCLIProvider struct {
	config Config
}

// NewGeminiCLIProvider creates a new Gemini CLI provider.
func NewGeminiCLIProvider(config *Config) *GeminiCLIProvider {
	if config.Command == "" {
		config.Command = "gemini"
	}
	if config.Model == "" {
		config.Model = "gemini-2.5-pro"
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultConfig().Timeout
	}

	return &GeminiCLIProvider{
		config: *config,
	}
}

// Call sends a prompt to the gemini CLI and returns the response.
func (g *GeminiCLIProvider) Call(ctx context.Context, prompt string) (string, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, g.config.Timeout)
	defer cancel()

	// Build command arguments
	// gemini -p "prompt" -m <model>
	args := []string{"-p", prompt}

	// Add model if specified
	if g.config.Model != "" {
		args = append(args, "-m", g.config.Model)
	}

	// Add any custom arguments
	args = append(args, g.config.Args...)

	// Execute CLI command
	// #nosec G204 - command is user-configured in config, intentionally dynamic
	cmd := exec.CommandContext(ctx, g.config.Command, args...)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if error is due to timeout
		if ctx.Err() == context.DeadlineExceeded {
			return "", &ProviderError{
				Provider:  "gemini",
				Err:       fmt.Errorf("timeout after %v", g.config.Timeout),
				Retryable: true,
			}
		}

		// Check if command exited with error
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", &ProviderError{
				Provider:  "gemini",
				Err:       fmt.Errorf("CLI error (exit code %d): %s", exitErr.ExitCode(), string(output)),
				Retryable: false,
			}
		}

		return "", &ProviderError{
			Provider:  "gemini",
			Err:       fmt.Errorf("execution failed: %w", err),
			Retryable: false,
		}
	}

	return string(output), nil
}
