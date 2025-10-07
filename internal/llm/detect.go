// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package llm

import (
	"os/exec"
)

// CLIInfo holds information about a detected CLI tool.
type CLIInfo struct {
	// Name is the provider name (e.g., "claude", "codex", "gemini")
	Name string

	// Command is the executable name (e.g., "claude", "codex exec")
	Command string

	// Model is the default model for this CLI
	Model string

	// Found indicates whether the CLI was detected
	Found bool
}

// DetectCLI finds the first available LLM CLI tool on the system.
// It checks for tools in priority order: claude → codex → gemini → llm
// Returns CLIInfo with Found=false if no tool is available (will use mock).
//
// Priority rationale:
//   - Claude Code: Most integrated, best UX for Claude users
//   - Codex: OpenAI-focused tool with good model support
//   - Gemini: Google's CLI with strong model capabilities
//   - llm: Universal fallback supporting multiple providers
//
// Example:
//
//	info := DetectCLI()
//	if info.Found {
//	    fmt.Printf("Using %s CLI\n", info.Name)
//	}
func DetectCLI() CLIInfo {
	// Priority-ordered list of CLIs to check
	clis := []CLIInfo{
		{
			Name:    "claude",
			Command: "claude",
			Model:   "sonnet", // Claude Code uses short aliases
		},
		{
			Name:    "codex",
			Command: "codex",
			Model:   "o3",
		},
		{
			Name:    "gemini",
			Command: "gemini",
			Model:   "gemini-2.5-pro",
		},
		{
			Name:    "llm",
			Command: "llm",
			Model:   "claude-3-5-sonnet",
		},
	}

	// Check each CLI in order
	for _, cli := range clis {
		if _, err := exec.LookPath(cli.Command); err == nil {
			cli.Found = true
			return cli
		}
	}

	// No CLI found, will fall back to mock
	return CLIInfo{
		Name:  "mock",
		Found: false,
	}
}
