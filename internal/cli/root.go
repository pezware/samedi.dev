// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pezware/samedi.dev/internal/config"
	"github.com/pezware/samedi.dev/internal/llm"
	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/session"
	"github.com/pezware/samedi.dev/internal/storage"
	"github.com/spf13/cobra"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "samedi",
	Short: "A learning operating system for the terminal",
	Long: `Samedi helps you track and manage your learning journey across any domain.

Features:
  - LLM-powered learning plan generation
  - Session time tracking
  - Spaced repetition flashcards
  - Progress visualization
  - Markdown-based, git-trackable plans

Example:
  samedi init "rust async programming" --hours 40
  samedi start rust-async chunk-001
  samedi stop
  samedi review
  samedi stats`,
	Run: func(cmd *cobra.Command, _ []string) {
		// If no subcommand, show help or launch TUI (future)
		if err := cmd.Help(); err != nil {
			fmt.Fprintf(os.Stderr, "Error displaying help: %v\n", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.samedi/config.toml)")
	rootCmd.PersistentFlags().Bool("json", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("samedi version %s\n", Version)
			fmt.Printf("commit: %s\n", Commit)
			fmt.Printf("built: %s\n", BuildDate)
		},
	})

	// Add subcommands
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(planCmd())
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(stopCmd())
	rootCmd.AddCommand(statusCmd())
}

// getConfig loads configuration from file or returns defaults.
func getConfig(_ *cobra.Command) (*config.Config, error) {
	return config.Load()
}

// getPlanService initializes the plan service with all dependencies.
// This includes: config, storage (SQLite + filesystem), LLM provider, and repositories.
// modelOverride, if non-empty, overrides the configured default model.
func getPlanService(_ *cobra.Command, modelOverride string) (*plan.Service, error) {
	// Get configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Get default paths
	paths, err := storage.DefaultPaths()
	if err != nil {
		return nil, fmt.Errorf("failed to get paths: %w", err)
	}

	// Ensure directories exist
	if err := paths.EnsureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Initialize SQLite database
	db, err := storage.NewSQLiteDB(paths.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	migrator := storage.NewMigrator(db)
	if err := migrator.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize filesystem storage
	fs := storage.NewFilesystemStorage(paths)

	// Ensure template exists
	if err := ensureTemplate(fs, paths); err != nil {
		return nil, fmt.Errorf("failed to ensure template: %w", err)
	}

	// Create LLM provider based on config
	modelToUse := cfg.LLM.DefaultModel
	if modelOverride != "" {
		modelToUse = modelOverride
	}
	llmProvider, err := createLLMProvider(cfg, modelToUse)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	// Create repositories
	sqliteRepo := plan.NewSQLiteRepository(db)
	filesystemRepo := plan.NewFilesystemRepository(fs, paths)

	// Create plan service
	planService := plan.NewService(sqliteRepo, filesystemRepo, llmProvider, fs, paths)

	// Optionally integrate session service for plan history
	sessionRepo := session.NewSQLiteRepository(db)
	sessionService := session.NewService(sessionRepo, nil)
	planService.SetSessionService(sessionService)

	return planService, nil
}

// createLLMProvider creates an LLM provider based on configuration.
func createLLMProvider(cfg *config.Config, model string) (llm.Provider, error) {
	llmConfig := &llm.Config{
		Provider: cfg.LLM.Provider,
		Command:  cfg.LLM.CLICommand,
		Model:    model,
		Timeout:  time.Duration(cfg.LLM.TimeoutSeconds) * time.Second,
	}

	providerName := strings.ToLower(cfg.LLM.Provider)

	// Auto-detect if provider is "auto"
	if providerName == "auto" {
		detected := llm.DetectCLI()
		if detected.Found {
			providerName = detected.Name
			llmConfig.Command = detected.Command
			// Only use detected model if user didn't provide override
			if model == "" {
				llmConfig.Model = detected.Model
			}
			// NOTE: Could add logging here if verbose mode is enabled
			// fmt.Printf("Auto-detected %s CLI\n", detected.Name)
		} else {
			// No CLI found, fall back to mock
			providerName = "mock"
		}
	}

	switch providerName {
	case "mock":
		return llm.NewMockProvider(), nil
	case "claude":
		// Claude Code CLI (https://claude.com/claude-code)
		// Installation: npm install -g @anthropic/claude-code
		return llm.NewClaudeCodeProvider(llmConfig), nil
	case "codex":
		// Codex CLI (https://codex.dev)
		// Installation: npm install -g @codex/cli
		return llm.NewCodexProvider(llmConfig), nil
	case "gemini":
		// Gemini CLI (https://github.com/google/gemini-cli)
		// Installation: npm install -g @google/gemini-cli
		return llm.NewGeminiCLIProvider(llmConfig), nil
	case "llm":
		// Simon Willison's llm CLI tool (universal fallback)
		// Installation: uv pip install llm && llm install llm-claude-3
		return llm.NewCLIProvider(llmConfig), nil
	case "stdin":
		// Generic stdin-based provider for custom CLIs
		// Requires llm.cli_command to be set in config
		return llm.NewStdinProvider(llmConfig), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s (supported: auto, claude, codex, gemini, llm, stdin, mock)", cfg.LLM.Provider)
	}
}

// ensureTemplate copies the plan generation template to ~/.samedi/templates if it doesn't exist.
func ensureTemplate(fs *storage.FilesystemStorage, paths *storage.Paths) error {
	templatePath := paths.TemplatePath("plan-generation")

	// Check if template already exists
	if _, err := os.Stat(templatePath); err == nil {
		return nil // Template exists
	}

	// Read template from embedded location (templates/plan-generation.md in repo)
	// For now, we'll assume it exists in the repo root during development
	// In production, this would be embedded in the binary
	repoTemplatePath := "templates/plan-generation.md"
	content, err := os.ReadFile(repoTemplatePath)
	if err != nil {
		// If we can't find the repo template, try current directory
		content, err = os.ReadFile("../../templates/plan-generation.md")
		if err != nil {
			return fmt.Errorf("template not found (run from repo root or ensure templates are installed): %w", err)
		}
	}

	// Write template to user directory
	if err := fs.WriteFile(templatePath, content); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	return nil
}

// getSessionService initializes the session service with all dependencies.
// This includes: database, session repository, and optional plan service.
func getSessionService(_ *cobra.Command) (*session.Service, error) {
	// Get default paths
	paths, err := storage.DefaultPaths()
	if err != nil {
		return nil, fmt.Errorf("failed to get paths: %w", err)
	}

	// Ensure directories exist
	if err := paths.EnsureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Initialize SQLite database
	db, err := storage.NewSQLiteDB(paths.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	migrator := storage.NewMigrator(db)
	if err := migrator.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create session repository
	sessionRepo := session.NewSQLiteRepository(db)

	// Create session service without plan service for now
	return session.NewService(sessionRepo, nil), nil
}

// exitWithError prints an error and exits.
func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
