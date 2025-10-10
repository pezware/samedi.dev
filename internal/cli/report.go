// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pezware/samedi.dev/internal/stats"
	"github.com/spf13/cobra"
)

// nolint:gocyclo // CLI command handlers can have higher complexity
func reportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report [plan-id]",
		Short: "Export learning statistics to markdown report",
		Long: `Export comprehensive learning statistics to a markdown report file.

The report includes:
  - Summary statistics (hours, sessions, streaks)
  - Plan-specific progress and completion
  - Daily breakdown with activity details

Output formats:
  - Markdown (default): formatted markdown file

Examples:
  samedi report                          # Generate full report
  samedi report -o stats-2025.md         # Save to specific file
  samedi report rust-async               # Generate plan-specific report
  samedi report --range this-week        # Report for current week
  samedi report --type summary           # Summary only (no daily breakdown)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Get flags
			outputFile, err := cmd.Flags().GetString("output")
			if err != nil {
				return fmt.Errorf("failed to get output flag: %w", err)
			}

			reportType, err := cmd.Flags().GetString("type")
			if err != nil {
				return fmt.Errorf("failed to get type flag: %w", err)
			}

			timeRange, err := cmd.Flags().GetString("range")
			if err != nil {
				return fmt.Errorf("failed to get range flag: %w", err)
			}

			// Get stats service
			statsService, err := getStatsService(cmd)
			if err != nil {
				return fmt.Errorf("failed to initialize stats service: %w", err)
			}

			// Parse time range
			var tr stats.TimeRange
			switch timeRange {
			case "all":
				tr = stats.NewTimeRangeAll()
			case "today":
				tr = stats.NewTimeRangeToday()
			case "this-week":
				tr = stats.NewTimeRangeThisWeek()
			case "this-month":
				tr = stats.NewTimeRangeThisMonth()
			default:
				return fmt.Errorf("invalid time range: %s (supported: all, today, this-week, this-month)", timeRange)
			}

			// Generate report based on type
			var report string
			exporter := stats.NewExporter()

			if len(args) > 0 {
				// Plan-specific report
				planID := args[0]
				planStats, err := statsService.GetPlanStats(ctx, planID, tr)
				if err != nil {
					return fmt.Errorf("failed to get plan stats: %w", err)
				}

				report, err = exporter.ExportPlanStats(planStats)
				if err != nil {
					return fmt.Errorf("failed to export plan stats: %w", err)
				}
			} else {
				// Full or summary report
				totalStats, err := statsService.GetTotalStats(ctx, tr)
				if err != nil {
					return fmt.Errorf("failed to get total stats: %w", err)
				}

				// Get global streak info (not scoped to time range)
				currentStreak, longestStreak, err := statsService.GetStreakInfo(ctx)
				if err != nil {
					return fmt.Errorf("failed to get streak info: %w", err)
				}
				totalStats.CurrentStreak = currentStreak
				totalStats.LongestStreak = longestStreak

				switch reportType {
				case "summary":
					// Summary only
					report, err = exporter.ExportTotalStats(totalStats)
					if err != nil {
						return fmt.Errorf("failed to export total stats: %w", err)
					}

				case "full":
					// Full report with plans and daily breakdown
					planStatsMap, err := statsService.GetAllPlanStats(ctx, tr)
					if err != nil {
						return fmt.Errorf("failed to get plan stats: %w", err)
					}

					// Convert map to slice for exporter
					planStats := make([]stats.PlanStats, 0, len(planStatsMap))
					for _, ps := range planStatsMap {
						planStats = append(planStats, ps)
					}

					dailyStats, err := statsService.GetDailyStats(ctx, tr)
					if err != nil {
						return fmt.Errorf("failed to get daily stats: %w", err)
					}

					report, err = exporter.ExportFullReport(totalStats, planStats, dailyStats)
					if err != nil {
						return fmt.Errorf("failed to export full report: %w", err)
					}

				default:
					return fmt.Errorf("invalid report type: %s (supported: summary, full)", reportType)
				}
			}

			// Output report
			if outputFile != "" {
				// Save to file
				absPath, err := filepath.Abs(outputFile)
				if err != nil {
					return fmt.Errorf("failed to resolve output path: %w", err)
				}

				if err := os.WriteFile(absPath, []byte(report), 0o600); err != nil {
					return fmt.Errorf("failed to write report: %w", err)
				}

				fmt.Printf("Report exported to: %s\n", absPath)
			} else {
				// Print to stdout
				fmt.Println(report)
			}

			return nil
		},
	}

	// Flags
	cmd.Flags().StringP("output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().StringP("type", "t", "full", "Report type: summary, full")
	cmd.Flags().StringP("range", "r", "all", "Time range: all, today, this-week, this-month")

	return cmd
}
