// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"
	"os"

	"github.com/pezware/samedi.dev/internal/config"
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
}

// getConfig loads configuration from file or returns defaults.
func getConfig(_ *cobra.Command) (*config.Config, error) {
	return config.Load()
}

// getStorage initializes storage (database + filesystem).
// This will be used in Stage 2 for plan management.
//
//nolint:unused // Will be used in Stage 2 for plan management
func getStorage(_ *config.Config) (*storage.Storage, error) {
	paths, err := storage.DefaultPaths()
	if err != nil {
		return nil, err
	}

	return storage.NewStorage(paths.DatabasePath, paths)
}

// exitWithError prints an error and exits.
func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
