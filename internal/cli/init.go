// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/storage"
	"github.com/spf13/cobra"
)

// initCmd creates the `samedi init` command for plan generation.
func initCmd() *cobra.Command {
	var (
		hours   float64
		level   string
		goals   string
		model   string
		edit    bool
		noCards bool
		debug   bool
	)

	cmd := &cobra.Command{
		Use:   "init <topic>",
		Short: "Create a new learning plan with LLM assistance",
		Long: `Generate a new learning plan for a specific topic using an LLM.

The LLM will create a structured curriculum broken into time-boxed chunks,
each with objectives, resources, and deliverables.

Examples:
  samedi init "french b1"
  samedi init "rust async programming" --hours 20
  samedi init "music theory basics" --level beginner --goals "read sheet music"`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			topic := args[0]

			// Get verbose flag
			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				verbose = false // Default to non-verbose on error
			}

			// Validate inputs
			if hours <= 0 {
				exitWithError("hours must be positive, got %.1f", hours)
			}
			if hours > 1000 {
				exitWithError("hours too large (max 1000), got %.1f", hours)
			}

			// Initialize plan service (with model override if provided)
			svc, err := getPlanService(cmd, model)
			if err != nil {
				exitWithError("Failed to initialize: %v", err)
			}

			// Show verbose initialization info
			if verbose {
				showVerboseInfo(cmd, model)
			}

			// Prepare create request
			req := plan.CreateRequest{
				Topic:      topic,
				TotalHours: hours,
				Level:      level,
				Goals:      goals,
				Debug:      debug,
			}

			// Create plan via LLM
			fmt.Printf("→ Generating learning plan for \"%s\" (%g hours)...\n", topic, hours)
			if level != "" {
				fmt.Printf("  Level: %s\n", level)
			}
			if verbose {
				fmt.Printf("→ Calling LLM...\n")
			}

			createdPlan, err := svc.Create(context.Background(), req)
			if err != nil {
				exitWithError("Failed to create plan: %v", err)
			}

			if verbose {
				fmt.Printf("→ Successfully parsed %d chunks\n", len(createdPlan.Chunks))
			}

			// Display success message
			fmt.Printf("\n✓ Plan created: %s\n", createdPlan.Title)
			fmt.Printf("✓ Location: ~/.samedi/plans/%s.md\n", createdPlan.ID)
			fmt.Printf("✓ Chunks: %d (%.1f hours total)\n", len(createdPlan.Chunks), createdPlan.TotalHours)

			// Open in editor if requested
			if edit {
				if err := openPlanInEditor(createdPlan.ID); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to open editor: %v\n", err)
				}
			}

			// Show next steps
			fmt.Printf("\nNext steps:\n")
			fmt.Printf("  View plan:  samedi plan show %s\n", createdPlan.ID)
			if !noCards {
				fmt.Printf("  Add cards:  samedi cards generate %s\n", createdPlan.ID)
			}
			if len(createdPlan.Chunks) > 0 {
				firstChunk := createdPlan.Chunks[0]
				fmt.Printf("  Start:      samedi start %s %s\n", createdPlan.ID, firstChunk.ID)
			}
		},
	}

	// Flags
	cmd.Flags().Float64Var(&hours, "hours", 40.0, "total estimated hours")
	cmd.Flags().StringVar(&level, "level", "", "learning level (beginner, intermediate, advanced)")
	cmd.Flags().StringVar(&goals, "goals", "", "specific learning goals or focus areas")
	cmd.Flags().StringVar(&model, "model", "", "LLM model override")
	cmd.Flags().BoolVar(&edit, "edit", false, "open plan in $EDITOR after creation")
	cmd.Flags().BoolVar(&noCards, "no-cards", false, "skip flashcard generation suggestion")
	cmd.Flags().BoolVar(&debug, "debug", false, "show full LLM prompt and response for debugging")

	return cmd
}

// showVerboseInfo displays verbose configuration information.
func showVerboseInfo(cmd *cobra.Command, modelOverride string) {
	cfg, err := getConfig(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config for verbose output: %v\n", err)
		return
	}

	actualModel := modelOverride
	if actualModel == "" {
		actualModel = cfg.LLM.DefaultModel
	}

	fmt.Printf("→ Verbose mode enabled\n")
	fmt.Printf("→ Provider: %s\n", cfg.LLM.Provider)
	fmt.Printf("→ Model: %s\n", actualModel)
	fmt.Printf("→ Timeout: %d seconds\n", cfg.LLM.TimeoutSeconds)
}

// openPlanInEditor opens a plan file in the configured editor.
func openPlanInEditor(planID string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Get plan path
	paths, err := storage.DefaultPaths()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	planPath := paths.PlanPath(planID)

	// Open in editor
	editorCmd := exec.Command(editor, planPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	return editorCmd.Run()
}
