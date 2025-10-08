// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"fmt"

	"github.com/pezware/samedi.dev/internal/session"
	"github.com/spf13/cobra"
)

// stopCmd creates the `samedi stop` command for stopping an active session.
func stopCmd() *cobra.Command {
	var (
		notes     string
		artifacts []string
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
			// Initialize session service
			svc, err := getSessionService(cmd)
			if err != nil {
				exitWithError("Failed to initialize: %v", err)
			}

			// Prepare stop request
			req := session.StopRequest{
				Notes:     notes,
				Artifacts: artifacts,
			}

			// Stop session
			sess, err := svc.Stop(context.Background(), req)
			if err != nil {
				exitWithError("Failed to stop session: %v", err)
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
		},
	}

	// Flags
	cmd.Flags().StringVar(&notes, "note", "", "session notes")
	cmd.Flags().StringArrayVar(&artifacts, "artifact", []string{}, "learning artifacts (URLs or file paths)")

	return cmd
}
