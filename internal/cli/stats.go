// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/session"
	"github.com/pezware/samedi.dev/internal/stats"
	"github.com/pezware/samedi.dev/internal/storage"
	"github.com/pezware/samedi.dev/internal/tui"
	"github.com/spf13/cobra"
)

func statsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats [plan-id]",
		Short: "Display learning statistics",
		Long: `Display comprehensive learning statistics including:
  - Total hours and sessions
  - Learning streaks (current and longest)
  - Active and completed plans
  - Average session duration
  - Per-plan statistics (if plan ID provided)

Examples:
  samedi stats                    # Show overall statistics
  samedi stats rust-async         # Show stats for specific plan
  samedi stats --json             # Output in JSON format
  samedi stats --tui              # Interactive TUI dashboard
  samedi stats --range this-week  # Stats for current week`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Get flags
			jsonOutput, err := cmd.Flags().GetBool("json")
			if err != nil {
				return fmt.Errorf("failed to get json flag: %w", err)
			}

			tuiMode, err := cmd.Flags().GetBool("tui")
			if err != nil {
				return fmt.Errorf("failed to get tui flag: %w", err)
			}

			timeRangeStr, err := cmd.Flags().GetString("range")
			if err != nil {
				return fmt.Errorf("failed to get range flag: %w", err)
			}

			breakdown, err := cmd.Flags().GetBool("breakdown")
			if err != nil {
				return fmt.Errorf("failed to get breakdown flag: %w", err)
			}

			// Parse time range
			var tr stats.TimeRange
			switch timeRangeStr {
			case "all":
				tr = stats.NewTimeRangeAll()
			case "today":
				tr = stats.NewTimeRangeToday()
			case "this-week":
				tr = stats.NewTimeRangeThisWeek()
			case "this-month":
				tr = stats.NewTimeRangeThisMonth()
			default:
				return fmt.Errorf("invalid time range: %s (supported: all, today, this-week, this-month)", timeRangeStr)
			}

			// Initialize stats service
			statsService, err := getStatsService(cmd)
			if err != nil {
				return fmt.Errorf("failed to initialize stats service: %w", err)
			}

			// If plan ID provided, show plan stats
			if len(args) > 0 {
				planID := args[0]
				return displayPlanStats(ctx, statsService, planID, tr, jsonOutput, tuiMode, breakdown)
			}

			// Otherwise show total stats
			return displayTotalStats(ctx, statsService, tr, jsonOutput, tuiMode, breakdown)
		},
	}

	// Add flags
	cmd.Flags().StringP("range", "r", "all", "Time range: all, today, this-week, this-month")
	cmd.Flags().Bool("breakdown", false, "Show daily breakdown")
	cmd.Flags().Bool("tui", false, "Launch interactive TUI dashboard")

	return cmd
}

// displayTotalStats shows aggregate statistics across all learning.
func displayTotalStats(ctx context.Context, service *stats.Service, timeRange stats.TimeRange, jsonOutput, tuiMode, breakdown bool) error {
	// Get total stats with time range filtering
	totalStats, err := service.GetTotalStats(ctx, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get total stats: %w", err)
	}

	// Get streak info
	currentStreak, longestStreak, err := service.GetStreakInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get streak info: %w", err)
	}

	// Update stats with streak info (redundant but for display)
	totalStats.CurrentStreak = currentStreak
	totalStats.LongestStreak = longestStreak

	if jsonOutput {
		// If breakdown requested, include daily stats in JSON output
		if breakdown {
			dailyStats, err := service.GetDailyStats(ctx, timeRange)
			if err != nil {
				return fmt.Errorf("failed to get daily stats: %w", err)
			}
			output := map[string]interface{}{
				"total": totalStats,
				"daily": dailyStats,
			}
			return printJSON(output)
		}
		return printJSON(totalStats)
	}

	if tuiMode {
		return launchTUI(totalStats, nil)
	}

	// Print total stats
	if err := printTotalStatsText(totalStats); err != nil {
		return err
	}

	// If breakdown requested, print daily stats
	if breakdown {
		fmt.Println("\n📅 Daily Breakdown")
		fmt.Println(strings.Repeat("─", 50))

		dailyStats, err := service.GetDailyStats(ctx, timeRange)
		if err != nil {
			return fmt.Errorf("failed to get daily stats: %w", err)
		}

		if len(dailyStats) == 0 {
			fmt.Println("No activity in selected time range.")
		} else {
			for _, ds := range dailyStats {
				fmt.Printf("\n%s:\n", ds.Date.Format("Monday, January 2, 2006"))
				fmt.Printf("  ⏱️  Duration: %.1f hours (%d minutes)\n", ds.Hours(), ds.Duration)
				fmt.Printf("  📊 Sessions: %d\n", ds.SessionCount)
				if len(ds.Plans) > 0 {
					fmt.Printf("  📚 Plans: %s\n", strings.Join(ds.Plans, ", "))
				}
			}
		}
		fmt.Println()
	}

	return nil
}

// displayPlanStats shows statistics for a specific plan.
func displayPlanStats(ctx context.Context, service *stats.Service, planID string, timeRange stats.TimeRange, jsonOutput, tuiMode, breakdown bool) error {
	// Get plan stats with time range filtering
	planStats, err := service.GetPlanStats(ctx, planID, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get plan stats: %w", err)
	}

	if jsonOutput {
		return printPlanStatsJSON(ctx, service, planID, timeRange, planStats, breakdown)
	}

	if tuiMode {
		return launchTUI(nil, planStats)
	}

	// Print plan stats
	if err := printPlanStatsText(planStats); err != nil {
		return err
	}

	// If breakdown requested, print daily stats for this plan
	if breakdown {
		return printPlanBreakdown(ctx, service, planID, timeRange)
	}

	return nil
}

// printPlanStatsJSON outputs plan stats in JSON format with optional breakdown.
func printPlanStatsJSON(ctx context.Context, service *stats.Service, planID string, timeRange stats.TimeRange, planStats *stats.PlanStats, breakdown bool) error {
	if !breakdown {
		return printJSON(planStats)
	}

	dailyStats, err := service.GetDailyStats(ctx, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get daily stats: %w", err)
	}

	// Filter daily stats for this plan only
	planDaily := []stats.DailyStats{}
	for _, ds := range dailyStats {
		for _, pid := range ds.Plans {
			if pid == planID {
				planDaily = append(planDaily, ds)
				break
			}
		}
	}

	output := map[string]interface{}{
		"plan":  planStats,
		"daily": planDaily,
	}
	return printJSON(output)
}

// printPlanBreakdown prints daily breakdown for a specific plan.
func printPlanBreakdown(ctx context.Context, service *stats.Service, planID string, timeRange stats.TimeRange) error {
	fmt.Println("\n📅 Daily Breakdown")
	fmt.Println(strings.Repeat("─", 50))

	dailyStats, err := service.GetDailyStats(ctx, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get daily stats: %w", err)
	}

	// Filter for this plan
	found := false
	for _, ds := range dailyStats {
		for _, pid := range ds.Plans {
			if pid != planID {
				continue
			}
			fmt.Printf("\n%s:\n", ds.Date.Format("Monday, January 2, 2006"))
			fmt.Printf("  ⏱️  Duration: %.1f hours (%d minutes)\n", ds.Hours(), ds.Duration)
			fmt.Printf("  📊 Sessions: %d\n", ds.SessionCount)
			found = true
			break
		}
	}

	if !found {
		fmt.Println("No activity in selected time range.")
	}
	fmt.Println()

	return nil
}

// printTotalStatsText formats total stats as human-readable text.
func printTotalStatsText(s *stats.TotalStats) error {
	fmt.Println("📊 Learning Statistics")
	fmt.Println(strings.Repeat("─", 50))

	// Learning time
	fmt.Printf("\n⏱️  Learning Time:\n")
	fmt.Printf("   Total hours:      %.1f hours\n", s.TotalHours)
	fmt.Printf("   Total sessions:   %d\n", s.TotalSessions)
	if s.TotalSessions > 0 {
		fmt.Printf("   Average session:  %.0f minutes\n", s.AverageSession)
	}

	// Streaks
	fmt.Printf("\n🔥 Learning Streaks:\n")
	fmt.Printf("   Current streak:   %d days\n", s.CurrentStreak)
	fmt.Printf("   Longest streak:   %d days\n", s.LongestStreak)

	// Plans
	fmt.Printf("\n📚 Learning Plans:\n")
	fmt.Printf("   Active plans:     %d\n", s.ActivePlans)
	fmt.Printf("   Completed plans:  %d\n", s.CompletedPlans)
	fmt.Printf("   Total plans:      %d\n", s.ActivePlans+s.CompletedPlans)

	// Last session
	if s.LastSessionDate != nil {
		fmt.Printf("\n📅 Last Session:\n")
		fmt.Printf("   %s\n", s.LastSessionDate.Format("Monday, January 2, 2006 at 3:04 PM"))
	}

	fmt.Println()
	return nil
}

// printPlanStatsText formats plan stats as human-readable text.
func printPlanStatsText(s *stats.PlanStats) error {
	fmt.Printf("📊 Statistics for: %s\n", s.PlanTitle)
	fmt.Println(strings.Repeat("─", 50))

	// Progress
	progressBar := buildProgressBar(s.Progress, 30)
	fmt.Printf("\n📈 Progress:\n")
	fmt.Printf("   %s %.0f%%\n", progressBar, s.Progress*100)
	fmt.Printf("   Completed chunks: %d / %d\n", s.CompletedChunks, s.TotalChunks)

	// Time
	fmt.Printf("\n⏱️  Time:\n")
	fmt.Printf("   Total hours:      %.1f / %.1f hours\n", s.TotalHours, s.PlannedHours)
	fmt.Printf("   Sessions:         %d\n", s.SessionCount)
	if s.SessionCount > 0 {
		avgMinutes := (s.TotalHours * 60) / float64(s.SessionCount)
		fmt.Printf("   Average session:  %.0f minutes\n", avgMinutes)
	}

	// Status
	fmt.Printf("\n📊 Status:\n")
	fmt.Printf("   %s\n", formatPlanStatus(s.Status))

	// Last session
	if s.LastSession != nil {
		fmt.Printf("\n📅 Last Session:\n")
		fmt.Printf("   %s\n", s.LastSession.Format("Monday, January 2, 2006 at 3:04 PM"))
	}

	fmt.Println()
	return nil
}

// buildProgressBar creates a visual progress bar.
func buildProgressBar(progress float64, width int) string {
	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := strings.Builder{}
	bar.WriteString("[")

	for i := 0; i < width; i++ {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}

	bar.WriteString("]")
	return bar.String()
}

// formatPlanStatus formats a status string with emoji.
func formatPlanStatus(status string) string {
	switch status {
	case "not-started":
		return "⚪ Not Started"
	case "in-progress":
		return "🟡 In Progress"
	case "completed":
		return "🟢 Completed"
	case "archived":
		return "📦 Archived"
	default:
		return status
	}
}

// printJSON outputs data as formatted JSON.
func printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// getStatsService initializes the stats service with all dependencies.
func getStatsService(_ *cobra.Command) (*stats.Service, error) {
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

	// Create plan repositories
	planSQLiteRepo := plan.NewSQLiteRepository(db)
	planFilesystemRepo := plan.NewFilesystemRepository(fs, paths)

	// Create plan service (nil LLM provider - we only read plans)
	planService := plan.NewService(planSQLiteRepo, planFilesystemRepo, nil, fs, paths)

	// Create session repository
	sessionRepo := session.NewSQLiteRepository(db)

	// Create session service adapter
	sessionService := &statsSessionServiceAdapter{
		repo: sessionRepo,
	}

	// Create stats service
	return stats.NewService(planService, sessionService), nil
}

// statsSessionServiceAdapter adapts session.Repository to stats.SessionService interface.
type statsSessionServiceAdapter struct {
	repo session.Repository
}

func (a *statsSessionServiceAdapter) List(ctx context.Context, planID string, limit int) ([]*session.Session, error) {
	return a.repo.List(ctx, planID, limit)
}

func (a *statsSessionServiceAdapter) ListAll(ctx context.Context) ([]*session.Session, error) {
	// Empty planID means all sessions
	return a.repo.List(ctx, "", 0)
}

// launchTUI starts the Bubble Tea program with the stats model.
func launchTUI(totalStats *stats.TotalStats, planStats *stats.PlanStats) error {
	model := tui.NewStatsModel(totalStats, planStats)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
