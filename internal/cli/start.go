// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"fmt"

	"github.com/pezware/samedi.dev/internal/session"
	"github.com/spf13/cobra"
)

// startCmd creates the `samedi start` command for starting a learning session.
func startCmd() *cobra.Command {
	var (
		notes string
	)

	cmd := &cobra.Command{
		Use:   "start <plan-id> [chunk-id]",
		Short: "Start a learning session",
		Long: `Start a new learning session for a specific plan and optional chunk.

The session timer begins immediately and tracks your learning time.
Only one session can be active at a time.

Examples:
  samedi start french-b1
  samedi start french-b1 chunk-003
  samedi start rust-async chunk-015 --note "Working on tokio tutorial"`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			planID := args[0]
			chunkID := ""
			if len(args) > 1 {
				chunkID = args[1]
			}

			// Initialize session service
			svc, err := getSessionService(cmd)
			if err != nil {
				exitWithError("Failed to initialize: %v", err)
			}

			// Prepare start request
			req := session.StartRequest{
				PlanID:  planID,
				ChunkID: chunkID,
				Notes:   notes,
			}

			// Start session
			sess, err := svc.Start(context.Background(), req)
			if err != nil {
				exitWithError("Failed to start session: %v", err)
			}

			// Display session started message
			fmt.Printf("→ Session started: %s", planID)
			if chunkID != "" {
				fmt.Printf(" (%s)", chunkID)
			}
			fmt.Println()

			// Show session details
			fmt.Printf("  Started at: %s\n", sess.StartTime.Format("15:04"))
			if sess.Notes != "" {
				fmt.Printf("  Notes: %s\n", sess.Notes)
			}

			// If we have a plan service, try to get chunk objectives
			// For now, we'll skip this since it requires loading the plan

			fmt.Println("\nTimer running. Stop with: samedi stop")
		},
	}

	// Flags
	cmd.Flags().StringVar(&notes, "note", "", "initial notes for the session")

	return cmd
}
