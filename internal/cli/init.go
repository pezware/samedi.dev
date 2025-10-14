// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/storage"
	"github.com/spf13/cobra"
)

// initCmd creates the `samedi init` command for plan generation.
func initCmd() *cobra.Command {
	var (
		hours    float64
		level    string
		goals    string
		model    string
		edit     bool
		noCards  bool
		debug    bool
		noPrompt bool
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
			if err := runInit(cmd, args, initOptions{
				hours:    &hours,
				level:    &level,
				goals:    &goals,
				model:    model,
				edit:     edit,
				noCards:  noCards,
				debug:    debug,
				noPrompt: noPrompt,
			}); err != nil {
				exitWithError("%v", err)
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
	cmd.Flags().BoolVar(&noPrompt, "no-prompt", false, "skip interactive prompts and use flag values")

	return cmd
}

type initOptions struct {
	hours    *float64
	level    *string
	goals    *string
	model    string
	edit     bool
	noCards  bool
	debug    bool
	noPrompt bool
}

func runInit(cmd *cobra.Command, args []string, opts initOptions) error {
	topic := args[0]

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		verbose = false
	}

	inputs := initInputs{
		hours: *opts.hours,
		level: *opts.level,
		goals: *opts.goals,
	}

	if err := collectInitInputs(cmd, &inputs, opts.noPrompt); err != nil {
		return fmt.Errorf("failed to collect inputs: %w", err)
	}

	if err := validateInitInputs(inputs.hours); err != nil {
		return err
	}

	svc, err := getPlanService(cmd, opts.model)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	if verbose {
		showVerboseInfo(cmd, opts.model)
	}

	req := plan.CreateRequest{
		Topic:      topic,
		TotalHours: inputs.hours,
		Level:      inputs.level,
		Goals:      inputs.goals,
		Debug:      opts.debug,
	}

	fmt.Printf("→ Generating learning plan for \"%s\" (%g hours)...\n", topic, inputs.hours)
	if inputs.level != "" {
		fmt.Printf("  Level: %s\n", inputs.level)
	}
	if verbose {
		fmt.Printf("→ Calling LLM...\n")
	}

	createdPlan, err := svc.Create(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	if verbose {
		fmt.Printf("→ Successfully parsed %d chunks\n", len(createdPlan.Chunks))
	}

	fmt.Printf("\n✓ Plan created: %s\n", createdPlan.Title)
	fmt.Printf("✓ Location: ~/.samedi/plans/%s.md\n", createdPlan.ID)
	fmt.Printf("✓ Chunks: %d (%.1f hours total)\n", len(createdPlan.Chunks), createdPlan.TotalHours)

	if opts.edit {
		if err := openPlanInEditor(createdPlan.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to open editor: %v\n", err)
		}
	}

	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  View plan:  samedi plan show %s\n", createdPlan.ID)
	if !opts.noCards {
		fmt.Printf("  Add cards:  samedi cards generate %s\n", createdPlan.ID)
	}
	if len(createdPlan.Chunks) > 0 {
		firstChunk := createdPlan.Chunks[0]
		fmt.Printf("  Start:      samedi start %s %s\n", createdPlan.ID, firstChunk.ID)
	}

	return nil
}

type initInputs struct {
	hours float64
	level string
	goals string
}

func collectInitInputs(cmd *cobra.Command, inputs *initInputs, noPrompt bool) error {
	if !isInteractive(noPrompt) {
		return nil
	}

	flags := cmd.Flags()
	if flags.Changed("hours") && flags.Changed("level") && flags.Changed("goals") {
		return nil
	}

	reader := bufio.NewReader(os.Stdin)
	writer := os.Stdout

	fmt.Println("Let's tailor your plan. Press Enter to accept the suggested value.")

	if !flags.Changed("hours") {
		hours, err := promptForHours(reader, writer, inputs.hours)
		if err != nil {
			return err
		}
		inputs.hours = hours
	}

	if !flags.Changed("level") {
		level, err := promptForLevel(reader, writer)
		if err != nil {
			return err
		}
		inputs.level = level
	}

	if !flags.Changed("goals") {
		goals, err := promptForGoals(reader, writer)
		if err != nil {
			return err
		}
		inputs.goals = goals
	}

	return nil
}

func validateInitInputs(hours float64) error {
	if hours <= 0 {
		return fmt.Errorf("hours must be positive, got %.1f", hours)
	}
	if hours > 1000 {
		return fmt.Errorf("hours too large (max 1000), got %.1f", hours)
	}
	return nil
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

// promptForHours asks the user for total hours, enforcing validation rules.
func promptForHours(reader *bufio.Reader, writer io.Writer, defaultHours float64) (float64, error) {
	const maxHours = 1000.0

	for {
		fmt.Fprintf(writer, "Total hours [%g]: ", defaultHours)

		line, err := reader.ReadString('\n')
		if errors.Is(err, io.EOF) && line == "" {
			// Treat EOF with no input as accepting the default.
			return defaultHours, nil
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			return defaultHours, nil
		}

		value, parseErr := strconv.ParseFloat(line, 64)
		if parseErr != nil || value <= 0 || value > maxHours {
			fmt.Fprintln(writer, "Please enter a number between 1 and 1000.")
			if errors.Is(err, io.EOF) {
				// If we hit EOF after an invalid entry, fall back to default.
				return defaultHours, nil
			}
			continue
		}

		return value, nil
	}
}

// promptForLevel asks the user for their learning level, ensuring a known option or blank.
func promptForLevel(reader *bufio.Reader, writer io.Writer) (string, error) {
	valid := map[string]struct{}{
		"beginner":     {},
		"intermediate": {},
		"advanced":     {},
	}

	for {
		fmt.Fprint(writer, "Learning level [beginner/intermediate/advanced]: ")
		line, err := reader.ReadString('\n')
		if errors.Is(err, io.EOF) && line == "" {
			return "", nil
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}

		line = strings.TrimSpace(strings.ToLower(line))
		if line == "" {
			return "", nil
		}

		if _, ok := valid[line]; ok {
			return line, nil
		}

		fmt.Fprintln(writer, "Please choose beginner, intermediate, advanced or leave blank.")
		if errors.Is(err, io.EOF) {
			return "", nil
		}
	}
}

// promptForGoals collects optional goals text from the user.
func promptForGoals(reader *bufio.Reader, writer io.Writer) (string, error) {
	fmt.Fprint(writer, "Specific goals or focus areas (optional): ")
	line, err := reader.ReadString('\n')
	if errors.Is(err, io.EOF) && line == "" {
		return "", nil
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	return strings.TrimSpace(line), nil
}
