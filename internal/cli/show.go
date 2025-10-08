// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"github.com/spf13/cobra"
)

// showCmd creates the `samedi show` command for displaying chunk details.
func showCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <plan-id> <chunk-id>",
		Short: "Show detailed information about a chunk",
		Long: `Display comprehensive information about a specific chunk within a plan.

Shows:
  - Chunk title, status, and progress
  - Learning objectives
  - Resources and deliverables
  - Session history and time spent

Examples:
  samedi show rust-async chunk-001
  samedi show french-b1 chunk-015`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			planID := args[0]
			chunkID := args[1]

			// Get chunk display information
			info, err := getChunkDisplayInfo(cmd, planID, chunkID)
			if err != nil {
				exitWithError("Failed to get chunk information: %v", err)
			}

			// Display comprehensive chunk details
			displayChunkDetails(info)
		},
	}

	return cmd
}
