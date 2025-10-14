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
	"strings"

	"github.com/pezware/samedi.dev/internal/session"
	"github.com/spf13/cobra"
)

// stopCmd creates the `samedi stop` command for stopping an active session.
func stopCmd() *cobra.Command {
	var (
		notes     string
		artifacts []string
		auto      bool
	)

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the active learning session",
		Long: `Stop the currently active learning session and record notes.

This calculates the total session duration and optionally records notes
and learning artifacts (URLs, file paths, etc.).

Examples:
  samedi stop
  samedi stop --note "Completed chapter 3"
  samedi stop --note "Built API server" --artifact "github.com/user/rust-api"
  samedi stop --artifact "file.md" --artifact "notes.txt"`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			noteFlagSet := cmd.Flags().Changed("note")
			artifactFlagSet := cmd.Flags().Changed("artifact")
			if err := executeStop(cmd, stopOptions{
				notes:           &notes,
				noteFlagSet:     noteFlagSet,
				artifacts:       &artifacts,
				artifactFlagSet: artifactFlagSet,
				noPrompt:        auto,
			}); err != nil {
				exitWithError("%v", err)
			}
		},
	}

	// Flags
	cmd.Flags().StringVar(&notes, "note", "", "session notes")
	cmd.Flags().StringArrayVar(&artifacts, "artifact", []string{}, "learning artifacts (URLs or file paths)")
	cmd.Flags().BoolVar(&auto, "auto", false, "skip interactive prompts and use defaults")

	return cmd
}

type stopOptions struct {
	notes           *string
	noteFlagSet     bool
	artifacts       *[]string
	artifactFlagSet bool
	noPrompt        bool
}

func executeStop(cmd *cobra.Command, opts stopOptions) error {
	note, artifacts, err := collectStopInputs(opts)
	if err != nil {
		return err
	}

	// Initialize session service
	svc, err := getSessionService(cmd)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Prepare stop request
	req := session.StopRequest{
		Notes:     note,
		Artifacts: artifacts,
	}

	// Stop session
	sess, err := svc.Stop(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to stop session: %w", err)
	}

	// Display session summary
	fmt.Printf("✓ Session completed: %s", sess.PlanID)
	if sess.ChunkID != "" {
		fmt.Printf(" (%s)", sess.ChunkID)
	}
	fmt.Println()

	fmt.Printf("  Duration: %s\n", sess.ElapsedTime())
	fmt.Printf("  Started: %s\n", sess.StartTime.Format("15:04"))
	fmt.Printf("  Stopped: %s\n", sess.EndTime.Format("15:04"))

	if sess.Notes != "" {
		fmt.Printf("  Notes: %s\n", sess.Notes)
	}

	if len(sess.Artifacts) > 0 {
		fmt.Println("  Artifacts:")
		for _, artifact := range sess.Artifacts {
			fmt.Printf("    - %s\n", artifact)
		}
	}

	// Show next steps
	fmt.Println("\nNext steps:")
	fmt.Printf("  View history:  samedi plan show %s --sessions\n", sess.PlanID)
	fmt.Printf("  Start new:     samedi start %s\n", sess.PlanID)

	return nil
}

func collectStopInputs(opts stopOptions) (string, []string, error) {
	note := ""
	if opts.notes != nil {
		note = *opts.notes
	}

	artifacts := []string{}
	if opts.artifacts != nil {
		artifacts = append(artifacts, (*opts.artifacts)...)
	}

	if !isInteractive(opts.noPrompt) {
		return note, artifacts, nil
	}

	reader := bufio.NewReader(os.Stdin)
	writer := os.Stdout

	if !opts.noteFlagSet && note == "" {
		entered, err := promptForStopNote(reader, writer)
		if err != nil {
			return "", nil, fmt.Errorf("failed to read notes: %w", err)
		}
		note = entered
	}

	if !opts.artifactFlagSet {
		added, err := promptForArtifacts(reader, writer)
		if err != nil {
			return "", nil, fmt.Errorf("failed to capture artifacts: %w", err)
		}
		artifacts = append(artifacts, added...)
	}

	if opts.notes != nil {
		*opts.notes = note
	}
	if opts.artifacts != nil {
		*opts.artifacts = artifacts
	}

	return note, artifacts, nil
}

func promptForStopNote(reader *bufio.Reader, writer io.Writer) (string, error) {
	fmt.Fprint(writer, "Session notes (optional): ")
	line, err := reader.ReadString('\n')
	if errors.Is(err, io.EOF) && line == "" {
		return "", nil
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptForArtifacts(reader *bufio.Reader, writer io.Writer) ([]string, error) {
	var artifacts []string

	for {
		fmt.Fprint(writer, "Add artifact URL/path (leave blank to finish): ")
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		value := strings.TrimSpace(line)
		if value == "" {
			return artifacts, nil
		}
		artifacts = append(artifacts, value)

		if errors.Is(err, io.EOF) {
			return artifacts, nil
		}
	}
}
