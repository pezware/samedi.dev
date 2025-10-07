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
}

// getConfig loads configuration from file or returns defaults.
func getConfig(_ *cobra.Command) (*config.Config, error) {
	return config.Load()
}

// getPlanService initializes the plan service with all dependencies.
// This includes: config, storage (SQLite + filesystem), LLM provider, and repositories.
func getPlanService(_ *cobra.Command) (*plan.Service, error) {
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
	var llmProvider llm.Provider
	switch strings.ToLower(cfg.LLM.Provider) {
	case "mock":
		llmProvider = llm.NewMockProvider()
	case "claude":
		llmConfig := &llm.Config{
			Provider: cfg.LLM.Provider,
			Command:  cfg.LLM.CLICommand,
			Model:    cfg.LLM.DefaultModel,
			Timeout:  time.Duration(cfg.LLM.TimeoutSeconds) * time.Second,
		}
		llmProvider = llm.NewClaudeProvider(llmConfig)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s (use 'claude' or 'mock')", cfg.LLM.Provider)
	}

	// Create repositories
	sqliteRepo := plan.NewSQLiteRepository(db)
	filesystemRepo := plan.NewFilesystemRepository(fs, paths)

	// Create and return service
	return plan.NewService(sqliteRepo, filesystemRepo, llmProvider, fs, paths), nil
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

// exitWithError prints an error and exits.
func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
