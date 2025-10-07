// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package config

import "fmt"

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	// Validate LLM provider
	validProviders := map[string]bool{
		"auto":    true, // Auto-detect available CLI (claude → codex → gemini → llm)
		"claude":  true, // Claude Code CLI
		"codex":   true, // Codex CLI
		"gemini":  true, // Gemini CLI
		"llm":     true, // Simon Willison's llm CLI (universal fallback)
		"stdin":   true, // Generic stdin-based provider
		"mock":    true, // Mock provider for testing
		"amazonq": true, // Amazon Q CLI
		"custom":  true, // Custom provider
	}
	if !validProviders[c.LLM.Provider] {
		return fmt.Errorf("invalid LLM provider: %s (must be one of: auto, claude, codex, gemini, llm, stdin, mock, amazonq, custom)", c.LLM.Provider)
	}

	// Validate timeout
	if c.LLM.TimeoutSeconds < 10 || c.LLM.TimeoutSeconds > 600 {
		return fmt.Errorf("LLM timeout must be between 10 and 600 seconds, got %d", c.LLM.TimeoutSeconds)
	}

	// Validate data directory exists or can be created
	if c.Storage.DataDir == "" {
		return fmt.Errorf("storage data_dir cannot be empty")
	}

	// Validate TUI theme
	validThemes := map[string]bool{
		"dracula": true,
		"monokai": true,
		"gruvbox": true,
	}
	if !validThemes[c.TUI.Theme] {
		return fmt.Errorf("invalid TUI theme: %s (must be one of: dracula, monokai, gruvbox)", c.TUI.Theme)
	}

	// Validate first day of week
	validFirstDays := map[string]bool{
		"monday": true,
		"sunday": true,
	}
	if !validFirstDays[c.TUI.FirstDayOfWeek] {
		return fmt.Errorf("invalid first_day_of_week: %s (must be monday or sunday)", c.TUI.FirstDayOfWeek)
	}

	return nil
}
