// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/pezware/samedi.dev/internal/session"
	"github.com/spf13/cobra"
)

// statusCmd creates the `samedi status` command for checking active session.
func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check active session status",
		Long: `Display information about the currently active learning session.

If there is an active session, shows:
  - Plan and chunk being studied
  - Start time and elapsed duration
  - Current session notes (if any)

If there is no active session, shows recent sessions instead.

Examples:
  samedi status`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize session service
			svc, err := getSessionService(cmd)
			if err != nil {
				exitWithError("Failed to initialize: %v", err)
			}

			// Get status
			status, err := svc.GetStatus(context.Background())
			if err != nil {
				exitWithError("Failed to get status: %v", err)
			}

			// Check if there's an active session
			if status.Active != nil {
				displayActiveSession(cmd, status.Active)
			} else {
				displayNoActiveSession(status.Recent)
			}
		},
	}

	return cmd
}

// displayActiveSession shows information about the active session.
func displayActiveSession(cmd *cobra.Command, sess *session.Session) {
	fmt.Printf("→ Active session: %s", sess.PlanID)
	if sess.ChunkID != "" {
		fmt.Printf(" (%s)", sess.ChunkID)
	}
	fmt.Println()

	fmt.Printf("  Started: %s\n", sess.StartTime.Format("15:04"))
	fmt.Printf("  Elapsed: %s\n", sess.ElapsedTime())

	if sess.Notes != "" {
		fmt.Printf("  Notes: %s\n", sess.Notes)
	}

	// Display chunk details if this session has a chunk
	if sess.ChunkID != "" {
		info, err := getChunkDisplayInfo(cmd, sess.PlanID, sess.ChunkID)
		if err == nil {
			fmt.Println()
			displayChunkDetails(info)
		}
		// Silently ignore errors - chunk display is optional
	}

	// Check if session has been running for a very long time
	elapsed := time.Since(sess.StartTime)
	if elapsed > 8*time.Hour {
		fmt.Println("\n⚠ Warning: This session has been running for a long time.")
		fmt.Println("  Did you forget to stop it? Use: samedi stop")
	}

	fmt.Println("\nStop with: samedi stop")
}

// displayNoActiveSession shows information when there's no active session.
func displayNoActiveSession(recent []*session.Session) {
	fmt.Println("No active session.")

	if len(recent) > 0 {
		fmt.Println("\nRecent sessions:")
		for i, sess := range recent {
			if i >= 3 {
				break // Show max 3 recent sessions
			}

			status := "✓"
			if sess.IsActive() {
				status = "→"
			}

			fmt.Printf("  %s %s", status, sess.PlanID)
			if sess.ChunkID != "" {
				fmt.Printf(" (%s)", sess.ChunkID)
			}

			if !sess.IsActive() {
				fmt.Printf(" - %s", sess.ElapsedTime())

				// Show how long ago
				ago := time.Since(*sess.EndTime)
				switch {
				case ago < 1*time.Hour:
					fmt.Printf(" - %d min ago", int(ago.Minutes()))
				case ago < 24*time.Hour:
					fmt.Printf(" - %d hours ago", int(ago.Hours()))
				default:
					fmt.Printf(" - %s", sess.EndTime.Format("Jan 2"))
				}
			}
			fmt.Println()
		}
	}

	fmt.Println("\nStart a session: samedi start <plan-id>")
}
