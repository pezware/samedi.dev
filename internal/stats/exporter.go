// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package stats

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"
)

// Exporter handles exporting statistics to various formats.
type Exporter struct {
	template *template.Template
}

// NewExporter creates a new statistics exporter with default templates.
func NewExporter() *Exporter {
	return &Exporter{}
}

// ExportTotalStats exports total statistics to markdown format.
func (e *Exporter) ExportTotalStats(stats *TotalStats) (string, error) {
	if err := stats.Validate(); err != nil {
		return "", fmt.Errorf("invalid stats: %w", err)
	}

	// Use custom template if set
	if e.template != nil {
		var buf bytes.Buffer
		if err := e.template.Execute(&buf, stats); err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}
		return buf.String(), nil
	}

	var buf bytes.Buffer

	buf.WriteString("# Learning Statistics\n\n")
	buf.WriteString("## Summary\n\n")

	if stats.TotalSessions == 0 {
		buf.WriteString("No sessions recorded yet.\n")
		return buf.String(), nil
	}

	buf.WriteString(fmt.Sprintf("**Total Hours:** %.1f hours\n", stats.TotalHours))
	buf.WriteString(fmt.Sprintf("**Total Sessions:** %d\n", stats.TotalSessions))
	buf.WriteString(fmt.Sprintf("**Average Session:** %.1f minutes\n", stats.AverageSession))
	buf.WriteString("\n")

	buf.WriteString("## Plans\n\n")
	buf.WriteString(fmt.Sprintf("**Active Plans:** %d\n", stats.ActivePlans))
	buf.WriteString(fmt.Sprintf("**Completed Plans:** %d\n", stats.CompletedPlans))
	buf.WriteString("\n")

	buf.WriteString("## Streaks\n\n")
	buf.WriteString(fmt.Sprintf("**Current Streak:** %d days\n", stats.CurrentStreak))
	buf.WriteString(fmt.Sprintf("**Longest Streak:** %d days\n", stats.LongestStreak))
	buf.WriteString("\n")

	if stats.LastSessionDate != nil {
		buf.WriteString(fmt.Sprintf("**Last Session:** %s\n", e.FormatDate(stats.LastSessionDate)))
	}

	return buf.String(), nil
}

// ExportPlanStats exports plan-specific statistics to markdown format.
func (e *Exporter) ExportPlanStats(stats *PlanStats) (string, error) {
	if err := stats.Validate(); err != nil {
		return "", fmt.Errorf("invalid plan stats: %w", err)
	}

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("# Plan: %s\n\n", stats.PlanTitle))
	buf.WriteString(fmt.Sprintf("**Plan ID:** %s\n", stats.PlanID))
	buf.WriteString(fmt.Sprintf("**Status:** %s\n", stats.Status))
	buf.WriteString("\n")

	buf.WriteString("## Progress\n\n")
	buf.WriteString(fmt.Sprintf("**Completion:** %s\n", e.FormatProgress(stats.Progress)))
	buf.WriteString(fmt.Sprintf("**Chunks:** %d/%d completed\n", stats.CompletedChunks, stats.TotalChunks))
	buf.WriteString(fmt.Sprintf("%s\n", e.GenerateProgressBar(stats.Progress, 30)))
	buf.WriteString("\n")

	buf.WriteString("## Time\n\n")
	buf.WriteString(fmt.Sprintf("**Actual Hours:** %.1f hours\n", stats.TotalHours))
	buf.WriteString(fmt.Sprintf("**Planned Hours:** %.1f hours\n", stats.PlannedHours))
	buf.WriteString(fmt.Sprintf("**Session Count:** %d sessions\n", stats.SessionCount))
	buf.WriteString("\n")

	if stats.SessionCount == 0 {
		buf.WriteString("No sessions recorded yet.\n")
	} else if stats.LastSession != nil {
		buf.WriteString(fmt.Sprintf("**Last Session:** %s\n", e.FormatDate(stats.LastSession)))
	}

	return buf.String(), nil
}

// ExportDailyStats exports daily statistics to markdown format.
func (e *Exporter) ExportDailyStats(dailyStats []DailyStats) (string, error) {
	var buf bytes.Buffer

	buf.WriteString("# Daily Statistics\n\n")

	if len(dailyStats) == 0 {
		buf.WriteString("No daily statistics available.\n")
		return buf.String(), nil
	}

	// Calculate totals
	totalDuration := 0
	totalSessions := 0
	for _, ds := range dailyStats {
		totalDuration += ds.Duration
		totalSessions += ds.SessionCount
	}

	buf.WriteString(fmt.Sprintf("**Total:** %.1f hours across %d sessions\n\n", float64(totalDuration)/60.0, totalSessions))

	buf.WriteString("## Breakdown\n\n")

	for _, ds := range dailyStats {
		buf.WriteString(fmt.Sprintf("### %s\n\n", ds.Date.Format("2006-01-02")))
		buf.WriteString(fmt.Sprintf("- **Duration:** %.1f hours\n", ds.Hours()))
		buf.WriteString(fmt.Sprintf("- **Sessions:** %d sessions\n", ds.SessionCount))
		if len(ds.Plans) > 0 {
			buf.WriteString(fmt.Sprintf("- **Plans:** %s\n", strings.Join(ds.Plans, ", ")))
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// ExportFullReport generates a comprehensive markdown report with all statistics.
func (e *Exporter) ExportFullReport(totalStats *TotalStats, planStats []PlanStats, dailyStats []DailyStats) (string, error) {
	var buf bytes.Buffer

	buf.WriteString("# Learning Statistics Report\n\n")
	buf.WriteString(fmt.Sprintf("*Generated: %s*\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Check if we have any data
	if totalStats.TotalSessions == 0 && len(planStats) == 0 && len(dailyStats) == 0 {
		buf.WriteString("No data available.\n")
		return buf.String(), nil
	}

	// Summary section
	buf.WriteString("## Summary\n\n")
	if totalStats.TotalSessions > 0 {
		buf.WriteString(fmt.Sprintf("- **Total Hours:** %.1f hours\n", totalStats.TotalHours))
		buf.WriteString(fmt.Sprintf("- **Total Sessions:** %d\n", totalStats.TotalSessions))
		buf.WriteString(fmt.Sprintf("- **Average Session:** %.1f minutes\n", totalStats.AverageSession))
		buf.WriteString(fmt.Sprintf("- **Active Plans:** %d\n", totalStats.ActivePlans))
		buf.WriteString(fmt.Sprintf("- **Completed Plans:** %d\n", totalStats.CompletedPlans))
		buf.WriteString(fmt.Sprintf("- **Current Streak:** %d days\n", totalStats.CurrentStreak))
		buf.WriteString(fmt.Sprintf("- **Longest Streak:** %d days\n", totalStats.LongestStreak))
		if totalStats.LastSessionDate != nil {
			buf.WriteString(fmt.Sprintf("- **Last Session:** %s\n", e.FormatDate(totalStats.LastSessionDate)))
		}
	} else {
		buf.WriteString("No sessions recorded.\n")
	}
	buf.WriteString("\n")

	// Plans section
	if len(planStats) > 0 {
		buf.WriteString("## Plans\n\n")
		buf.WriteString(e.GenerateMarkdownTable(planStats))
		buf.WriteString("\n")
	}

	// Daily breakdown section
	if len(dailyStats) > 0 {
		buf.WriteString("## Daily Breakdown\n\n")
		for _, ds := range dailyStats {
			buf.WriteString(fmt.Sprintf("- **%s:** %.1f hours (%d sessions)\n",
				ds.Date.Format("2006-01-02"),
				ds.Hours(),
				ds.SessionCount))
		}
		buf.WriteString("\n")
	}

	buf.WriteString("---\n")
	buf.WriteString("*Report generated by Samedi*\n")

	return buf.String(), nil
}

// WithTemplate sets a custom template for exports.
func (e *Exporter) WithTemplate(templateStr string) error {
	tmpl, err := template.New("custom").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	e.template = tmpl
	return nil
}

// FormatDuration formats a duration in minutes to a human-readable string.
func (e *Exporter) FormatDuration(minutes float64) string {
	if minutes < 60 {
		return fmt.Sprintf("%.0f minutes", minutes)
	}
	return fmt.Sprintf("%.1f hours", minutes/60.0)
}

// FormatDate formats a date pointer to a string, returning "N/A" if nil.
func (e *Exporter) FormatDate(date *time.Time) string {
	if date == nil {
		return "N/A"
	}
	return date.Format("2006-01-02")
}

// FormatProgress formats a progress float (0.0-1.0) to a percentage string.
func (e *Exporter) FormatProgress(progress float64) string {
	return fmt.Sprintf("%d%%", int(progress*100))
}

// ExportToFile exports statistics to a markdown file.
func (e *Exporter) ExportToFile(stats *TotalStats, filepath string) error {
	content, err := e.ExportTotalStats(stats)
	if err != nil {
		return fmt.Errorf("failed to export stats: %w", err)
	}

	if err := os.WriteFile(filepath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ReadFile reads and returns the contents of a file.
func (e *Exporter) ReadFile(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(data), nil
}

// ValidateStats validates statistics before export.
func (e *Exporter) ValidateStats(stats *TotalStats) error {
	return stats.Validate()
}

// GenerateMarkdownTable generates a markdown table from plan statistics.
func (e *Exporter) GenerateMarkdownTable(planStats []PlanStats) string {
	var buf bytes.Buffer

	// Table header
	buf.WriteString("| Plan | Hours | Sessions | Progress | Status |\n")
	buf.WriteString("|------|-------|----------|----------|--------|\n")

	// Table rows
	for _, ps := range planStats {
		buf.WriteString(fmt.Sprintf("| %s | %.1f | %d | %s | %s |\n",
			ps.PlanTitle,
			ps.TotalHours,
			ps.SessionCount,
			e.FormatProgress(ps.Progress),
			ps.Status))
	}

	return buf.String()
}

// GenerateProgressBar generates an ASCII progress bar.
func (e *Exporter) GenerateProgressBar(progress float64, width int) string {
	// Clamp progress
	if progress < 0 {
		progress = 0
	}
	if progress > 1.0 {
		progress = 1.0
	}

	filled := int(progress * float64(width))
	var bar strings.Builder

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
