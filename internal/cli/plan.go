// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/storage"
	"github.com/spf13/cobra"
)

// planCmd creates the parent `samedi plan` command.
func planCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Manage learning plans",
		Long: `View, edit, and manage your learning plans.

Plans are stored as markdown files in ~/.samedi/plans/ and indexed
in SQLite for fast queries.

Examples:
  samedi plan list                    # List all plans
  samedi plan list --status in-progress
  samedi plan show rust-async         # Show plan details
  samedi plan edit rust-async         # Edit in $EDITOR
  samedi plan archive french-b1       # Archive completed plan`,
	}

	// Add subcommands
	cmd.AddCommand(planListCmd())
	cmd.AddCommand(planShowCmd())
	cmd.AddCommand(planEditCmd())
	cmd.AddCommand(planArchiveCmd())

	return cmd
}

// planListCmd creates the `samedi plan list` subcommand.
func planListCmd() *cobra.Command {
	var (
		statusFilter string
		tagFilter    string
		sortBy       string
		showAll      bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all learning plans",
		Long: `List all learning plans with optional filtering.

By default, archived plans are hidden. Use --all to show all plans
including archived ones, or --status archived to show only archived plans.

Examples:
  samedi plan list                     # Active plans only
  samedi plan list --all               # Include archived plans
  samedi plan list --status archived   # Only archived plans
  samedi plan list --status in-progress
  samedi plan list --tag language
  samedi plan list --json`,
		Run: func(cmd *cobra.Command, _ []string) {
			svc, err := getPlanService(cmd, "")
			if err != nil {
				exitWithError("Failed to initialize: %v", err)
			}

			// Build filter
			filter := &storage.PlanFilter{}
			if statusFilter != "" {
				filter.Statuses = []string{statusFilter}
			} else if !showAll {
				// By default, exclude archived plans unless --all is specified
				filter.Statuses = []string{
					string(plan.StatusNotStarted),
					string(plan.StatusInProgress),
					string(plan.StatusCompleted),
				}
			}
			if tagFilter != "" {
				filter.Tag = tagFilter
			}
			if sortBy != "" {
				filter.SortBy = sortBy
			}

			// Get plans
			plans, err := svc.List(context.Background(), filter)
			if err != nil {
				exitWithError("Failed to list plans: %v", err)
			}

			// Check for JSON output
			jsonOutput, err := cmd.Flags().GetBool("json")
			if err != nil {
				exitWithError("Failed to get json flag: %v", err)
			}
			if jsonOutput {
				data, err := json.MarshalIndent(plans, "", "  ")
				if err != nil {
					exitWithError("Failed to marshal JSON: %v", err)
				}
				fmt.Println(string(data))
				return
			}

			// Table output
			if len(plans) == 0 {
				fmt.Println("No plans found.")
				fmt.Println("\nCreate a plan: samedi init <topic>")
				return
			}

			// Print table
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tPROGRESS\tHOURS")

			for _, record := range plans {
				// Calculate progress by loading full plan
				progress := calculateProgress(svc, record.ID)

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.1fh\n",
					record.ID,
					truncate(record.Title, 40),
					formatStatus(record.Status),
					progress,
					record.TotalHours,
				)
			}

			w.Flush()
		},
	}

	cmd.Flags().StringVar(&statusFilter, "status", "", "filter by status (not-started, in-progress, completed, archived)")
	cmd.Flags().StringVar(&tagFilter, "tag", "", "filter by tag")
	cmd.Flags().StringVar(&sortBy, "sort", "", "sort by field (created, updated, title, status, hours)")
	cmd.Flags().BoolVar(&showAll, "all", false, "show all plans including archived")

	return cmd
}

// formatStatus returns a human-friendly status string with indicators.
func formatStatus(status string) string {
	switch status {
	case "completed":
		return "✓ completed"
	case "in-progress":
		return "→ in-progress"
	case "not-started":
		return "○ not-started"
	case "archived":
		return "archived"
	default:
		return status
	}
}

// truncate truncates a string to a maximum length with ellipsis.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// calculateProgress computes progress as "percentage (completed/total)".
// Returns "-" if the plan cannot be loaded or has no chunks.
func calculateProgress(svc *plan.Service, planID string) string {
	// Load full plan to get chunks
	p, err := svc.Get(context.Background(), planID)
	if err != nil {
		return "-"
	}

	totalChunks := len(p.Chunks)
	if totalChunks == 0 {
		return "0% (0/0)"
	}

	completedChunks := 0
	for _, chunk := range p.Chunks {
		if chunk.Status == plan.StatusCompleted {
			completedChunks++
		}
	}

	percentage := int(float64(completedChunks) / float64(totalChunks) * 100)
	return fmt.Sprintf("%d%% (%d/%d)", percentage, completedChunks, totalChunks)
}

// planShowCmd creates the `samedi plan show` subcommand.
func planShowCmd() *cobra.Command {
	var (
		showChunks   bool
		showSessions bool
		showCards    bool
	)

	cmd := &cobra.Command{
		Use:   "show <plan-id>",
		Short: "Show plan details and progress",
		Long: `Display detailed information about a specific plan.

Shows plan metadata, progress, recent chunks, session history, and flashcard count.
Use --chunks to display all chunks, --sessions for full session history,
or --cards for detailed card statistics.

Examples:
  samedi plan show rust-async
  samedi plan show french-b1 --chunks
  samedi plan show french-b1 --sessions
  samedi plan show french-b1 --cards`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			planID := args[0]

			svc, err := getPlanService(cmd, "")
			if err != nil {
				exitWithError("Failed to initialize: %v", err)
			}

			// Get plan
			plan, err := svc.Get(context.Background(), planID)
			if err != nil {
				exitWithError("Failed to get plan: %v", err)
			}

			// Display plan details
			displayPlanSummary(plan)
			displayPlanExtras(svc, planID, showSessions, showCards)
			displayPlanChunks(plan, showChunks)
			displayNextSteps(plan, planID)
		},
	}

	cmd.Flags().BoolVar(&showChunks, "chunks", false, "show all chunks")
	cmd.Flags().BoolVar(&showSessions, "sessions", false, "show full session history")
	cmd.Flags().BoolVar(&showCards, "cards", false, "show detailed flashcard statistics")

	return cmd
}

// planEditCmd creates the `samedi plan edit` subcommand.
func planEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <plan-id>",
		Short: "Edit plan in $EDITOR",
		Long: `Open a plan file in your configured editor.

After editing, the plan is validated and SQLite metadata is updated.
If validation fails, you'll be prompted to fix the errors.

Examples:
  samedi plan edit rust-async
  EDITOR=nano samedi plan edit french-b1`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			planID := args[0]

			svc, err := getPlanService(cmd, "")
			if err != nil {
				exitWithError("Failed to initialize: %v", err)
			}

			// Verify plan exists
			exists := svc.Exists(context.Background(), planID)
			if !exists {
				exitWithError("Plan not found: %s", planID)
			}

			// Open in editor
			if err := openPlanInEditor(planID); err != nil {
				exitWithError("Failed to edit plan: %v", err)
			}

			// Reload and validate
			plan, err := svc.Get(context.Background(), planID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to reload plan: %v\n", err)
				fmt.Fprintf(os.Stderr, "Please check the plan file for errors.\n")
				return
			}

			// Update metadata
			if err := svc.Update(context.Background(), plan); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to update plan metadata: %v\n", err)
			} else {
				fmt.Printf("✓ Plan updated: %s\n", plan.Title)
			}
		},
	}

	return cmd
}

// formatProgress returns a progress string like "24% (12/50)".
func formatProgress(p *plan.Plan) string {
	if len(p.Chunks) == 0 {
		return "0%"
	}

	completed := 0
	for _, chunk := range p.Chunks {
		if chunk.Status == plan.StatusCompleted {
			completed++
		}
	}

	percent := float64(completed) / float64(len(p.Chunks)) * 100
	return fmt.Sprintf("%.0f%% (%d/%d chunks)", percent, completed, len(p.Chunks))
}

// chunkStatusIcon returns an icon for chunk status.
func chunkStatusIcon(status plan.Status) string {
	switch status {
	case plan.StatusCompleted:
		return "✓"
	case plan.StatusInProgress:
		return "→"
	case plan.StatusNotStarted:
		return "○"
	default:
		return " "
	}
}

// formatDuration formats minutes into a human-readable string.
func formatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dmin", minutes)
	}
	hours := float64(minutes) / 60.0
	if minutes%60 == 0 {
		return fmt.Sprintf("%.0fh", hours)
	}
	return fmt.Sprintf("%.1fh", hours)
}

// displayPlanSummary shows the plan metadata (title, status, dates, etc).
func displayPlanSummary(plan *plan.Plan) {
	fmt.Printf("%s\n", plan.Title)
	fmt.Printf("Status: %s | Progress: %s\n",
		formatStatus(string(plan.Status)),
		formatProgress(plan),
	)
	fmt.Printf("Created: %s | Updated: %s\n",
		plan.CreatedAt.Format("2006-01-02"),
		plan.UpdatedAt.Format("2006-01-02"),
	)
	fmt.Printf("Total: %.1f hours", plan.TotalHours)
	if len(plan.Tags) > 0 {
		fmt.Printf(" | Tags: %v", plan.Tags)
	}
	fmt.Println()
}

// displaySessionSummary formats and displays a single session from the session map.
func displaySessionSummary(sess map[string]interface{}) {
	// Get chunk ID if present
	chunkID := ""
	if cid, ok := sess["chunk_id"].(string); ok && cid != "" {
		chunkID = fmt.Sprintf(" (%s)", cid)
	}

	// Format start time
	var startTime string
	if st, ok := sess["start_time"].(time.Time); ok {
		startTime = st.Format("Jan 2 15:04")
	}

	// Check if active or completed
	isActive := false
	if active, ok := sess["is_active"].(bool); ok {
		isActive = active
	}

	if isActive {
		// Active session - show running indicator and elapsed time
		fmt.Printf("  → Active%s - started %s\n", chunkID, startTime)
	} else {
		// Completed session - show duration
		duration := 0
		if d, ok := sess["duration"].(int); ok {
			duration = d
		}
		fmt.Printf("  ✓ %s%s - %s\n", formatDuration(duration), chunkID, startTime)
	}

	// Show notes if present
	if notes, ok := sess["notes"].(string); ok && notes != "" {
		fmt.Printf("    Notes: %s\n", notes)
	}
}

// displayPlanExtras shows session history and flashcard count.
// The showSessions and showCards flags will control detail level in future stages.
func displayPlanExtras(svc *plan.Service, planID string, _ /* showSessions */, _ /* showCards */ bool) {
	ctx := context.Background()

	// Get session and card data (Stage 3/4 - currently returns empty/zero)
	sessions, err := svc.GetRecentSessions(ctx, planID, 5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to get sessions: %v\n", err)
		sessions = []map[string]interface{}{}
	}

	cardCount, err := svc.GetCardCount(ctx, planID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to get card count: %v\n", err)
		cardCount = 0
	}

	// Display flashcard count (always show per FR-010)
	fmt.Printf("Flashcards: %d cards", cardCount)
	if cardCount > 0 {
		fmt.Printf(" (use --cards for details)")
	}
	fmt.Println()

	// Display session history (always show per FR-010)
	fmt.Printf("\nRecent Sessions:\n")
	if len(sessions) == 0 {
		fmt.Println("  No sessions recorded yet")
	} else {
		for _, sess := range sessions {
			displaySessionSummary(sess)
		}
	}
}

// displayPlanChunks shows either all chunks or a summary of recent chunks.
func displayPlanChunks(p *plan.Plan, showAll bool) {
	if showAll {
		fmt.Println("\nChunks:")
		for i, chunk := range p.Chunks {
			fmt.Printf("%d. %s (%s) - %s\n",
				i+1,
				chunk.Title,
				formatDuration(chunk.Duration),
				formatStatus(string(chunk.Status)),
			)
		}
	} else {
		// Show recent chunks (first 5)
		fmt.Println("\nRecent chunks:")
		maxChunks := 5
		if len(p.Chunks) < maxChunks {
			maxChunks = len(p.Chunks)
		}
		for i := 0; i < maxChunks; i++ {
			chunk := p.Chunks[i]
			fmt.Printf("  %s %s (%s)\n",
				chunkStatusIcon(chunk.Status),
				chunk.Title,
				formatDuration(chunk.Duration),
			)
		}
		if len(p.Chunks) > maxChunks {
			fmt.Printf("  ... and %d more chunks\n", len(p.Chunks)-maxChunks)
		}
	}
}

// displayNextSteps shows the next recommended action for the plan.
func displayNextSteps(p *plan.Plan, planID string) {
	fmt.Println()
	nextChunk := p.NextChunk()
	if nextChunk != nil {
		fmt.Printf("Next: samedi start %s %s\n", planID, nextChunk.ID)
	} else {
		fmt.Println("All chunks completed!")
	}
}

// planArchiveCmd creates the `samedi plan archive` subcommand.
func planArchiveCmd() *cobra.Command {
	var skipConfirm bool

	cmd := &cobra.Command{
		Use:   "archive <plan-id>",
		Short: "Archive a completed or abandoned plan",
		Long: `Archive a learning plan by changing its status to 'archived'.

Archived plans are hidden from default listings but can still be viewed
with 'samedi plan list --status archived'.

Examples:
  samedi plan archive french-b1
  samedi plan archive rust-async
  samedi plan archive french-b1 --yes  # Skip confirmation`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			planID := args[0]

			svc, err := getPlanService(cmd, "")
			if err != nil {
				exitWithError("Failed to initialize: %v", err)
			}

			// Get plan
			p, err := svc.Get(context.Background(), planID)
			if err != nil {
				exitWithError("Failed to get plan: %v", err)
			}

			// Confirmation prompt (unless --yes flag)
			if !skipConfirm {
				fmt.Printf("⚠ Archive plan '%s'?\n", p.Title)
				fmt.Printf("  This will hide it from default views.\n")
				fmt.Printf("  Type plan ID to confirm: ")

				var input string
				if _, err := fmt.Scanln(&input); err != nil || input != planID {
					fmt.Println("✗ Archive canceled")
					os.Exit(0)
				}
			}

			// Update status to archived
			p.Status = plan.StatusArchived
			if err := svc.Update(context.Background(), p); err != nil {
				exitWithError("Failed to archive plan: %v", err)
			}

			fmt.Printf("✓ Plan archived: %s\n", p.Title)
			fmt.Printf("  View archived plans: samedi plan list --status archived\n")
		},
	}

	cmd.Flags().BoolVar(&skipConfirm, "yes", false, "skip confirmation prompt")

	return cmd
}
