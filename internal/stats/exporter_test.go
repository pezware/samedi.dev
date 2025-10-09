// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package stats

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExporter(t *testing.T) {
	exporter := NewExporter()
	assert.NotNil(t, exporter)
}

func TestExporter_ExportTotalStats_Empty(t *testing.T) {
	exporter := NewExporter()
	stats := &TotalStats{}

	result, err := exporter.ExportTotalStats(stats)

	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "# Learning Statistics")
	assert.Contains(t, result, "No sessions recorded")
}

func TestExporter_ExportTotalStats_WithData(t *testing.T) {
	lastSession := time.Date(2025, 10, 9, 14, 30, 0, 0, time.UTC)
	stats := &TotalStats{
		TotalHours:      45.5,
		TotalSessions:   23,
		ActivePlans:     3,
		CompletedPlans:  2,
		CurrentStreak:   5,
		LongestStreak:   12,
		AverageSession:  118.7,
		LastSessionDate: &lastSession,
	}

	exporter := NewExporter()
	result, err := exporter.ExportTotalStats(stats)

	require.NoError(t, err)
	assert.Contains(t, result, "45.5 hours")
	assert.Contains(t, result, "23")
	assert.Contains(t, result, "**Active Plans:** 3")
	assert.Contains(t, result, "**Completed Plans:** 2")
	assert.Contains(t, result, "**Current Streak:** 5 days")
	assert.Contains(t, result, "**Longest Streak:** 12 days")
	assert.Contains(t, result, "118.7 minutes")
	assert.Contains(t, result, "2025-10-09")
}

func TestExporter_ExportTotalStats_NoSessions(t *testing.T) {
	stats := &TotalStats{
		TotalHours:     0,
		TotalSessions:  0,
		ActivePlans:    2,
		CompletedPlans: 0,
		CurrentStreak:  0,
		LongestStreak:  0,
	}

	exporter := NewExporter()
	result, err := exporter.ExportTotalStats(stats)

	require.NoError(t, err)
	assert.Contains(t, result, "No sessions recorded")
}

func TestExporter_ExportPlanStats_Basic(t *testing.T) {
	lastSession := time.Date(2025, 10, 8, 10, 0, 0, 0, time.UTC)
	stats := &PlanStats{
		PlanID:          "rust-async",
		PlanTitle:       "Rust Async Programming",
		TotalHours:      12.5,
		PlannedHours:    40.0,
		SessionCount:    8,
		CompletedChunks: 15,
		TotalChunks:     40,
		Progress:        0.375,
		Status:          "in-progress",
		LastSession:     &lastSession,
	}

	exporter := NewExporter()
	result, err := exporter.ExportPlanStats(stats)

	require.NoError(t, err)
	assert.Contains(t, result, "# Plan: Rust Async Programming")
	assert.Contains(t, result, "rust-async")
	assert.Contains(t, result, "12.5 hours")
	assert.Contains(t, result, "40.0 hours")
	assert.Contains(t, result, "8 sessions")
	assert.Contains(t, result, "15/40")
	assert.Contains(t, result, "37%")
	assert.Contains(t, result, "in-progress")
	assert.Contains(t, result, "2025-10-08")
}

func TestExporter_ExportPlanStats_NoSessions(t *testing.T) {
	stats := &PlanStats{
		PlanID:          "french-b1",
		PlanTitle:       "French B1 Mastery",
		TotalHours:      0,
		PlannedHours:    60.0,
		SessionCount:    0,
		CompletedChunks: 0,
		TotalChunks:     60,
		Progress:        0,
		Status:          "not-started",
	}

	exporter := NewExporter()
	result, err := exporter.ExportPlanStats(stats)

	require.NoError(t, err)
	assert.Contains(t, result, "No sessions recorded")
	assert.Contains(t, result, "not-started")
}

func TestExporter_ExportPlanStats_Completed(t *testing.T) {
	lastSession := time.Date(2025, 9, 30, 16, 0, 0, 0, time.UTC)
	stats := &PlanStats{
		PlanID:          "python-basics",
		PlanTitle:       "Python Basics",
		TotalHours:      25.0,
		PlannedHours:    20.0,
		SessionCount:    15,
		CompletedChunks: 20,
		TotalChunks:     20,
		Progress:        1.0,
		Status:          "completed",
		LastSession:     &lastSession,
	}

	exporter := NewExporter()
	result, err := exporter.ExportPlanStats(stats)

	require.NoError(t, err)
	assert.Contains(t, result, "100%")
	assert.Contains(t, result, "completed")
	assert.Contains(t, result, "20/20")
}

func TestExporter_ExportDailyStats_Empty(t *testing.T) {
	dailyStats := []DailyStats{}

	exporter := NewExporter()
	result, err := exporter.ExportDailyStats(dailyStats)

	require.NoError(t, err)
	assert.Contains(t, result, "# Daily Statistics")
	assert.Contains(t, result, "No daily statistics available")
}

func TestExporter_ExportDailyStats_SingleDay(t *testing.T) {
	dailyStats := []DailyStats{
		{
			Date:         time.Date(2025, 10, 9, 0, 0, 0, 0, time.UTC),
			Duration:     120,
			SessionCount: 2,
			Plans:        []string{"rust-async", "french-b1"},
		},
	}

	exporter := NewExporter()
	result, err := exporter.ExportDailyStats(dailyStats)

	require.NoError(t, err)
	assert.Contains(t, result, "2025-10-09")
	assert.Contains(t, result, "2.0 hours")
	assert.Contains(t, result, "2 sessions")
	assert.Contains(t, result, "rust-async")
	assert.Contains(t, result, "french-b1")
}

func TestExporter_ExportDailyStats_MultipleDays(t *testing.T) {
	dailyStats := []DailyStats{
		{
			Date:         time.Date(2025, 10, 7, 0, 0, 0, 0, time.UTC),
			Duration:     60,
			SessionCount: 1,
			Plans:        []string{"rust-async"},
		},
		{
			Date:         time.Date(2025, 10, 8, 0, 0, 0, 0, time.UTC),
			Duration:     90,
			SessionCount: 2,
			Plans:        []string{"rust-async", "python-basics"},
		},
		{
			Date:         time.Date(2025, 10, 9, 0, 0, 0, 0, time.UTC),
			Duration:     120,
			SessionCount: 2,
			Plans:        []string{"french-b1"},
		},
	}

	exporter := NewExporter()
	result, err := exporter.ExportDailyStats(dailyStats)

	require.NoError(t, err)
	assert.Contains(t, result, "2025-10-07")
	assert.Contains(t, result, "2025-10-08")
	assert.Contains(t, result, "2025-10-09")
	assert.Contains(t, result, "**Total:** 4.5 hours")
}

func TestExporter_ExportFullReport_Complete(t *testing.T) {
	lastSession := time.Date(2025, 10, 9, 14, 30, 0, 0, time.UTC)
	totalStats := &TotalStats{
		TotalHours:      45.5,
		TotalSessions:   23,
		ActivePlans:     3,
		CompletedPlans:  2,
		CurrentStreak:   5,
		LongestStreak:   12,
		AverageSession:  118.7,
		LastSessionDate: &lastSession,
	}

	planStats := []PlanStats{
		{
			PlanID:          "rust-async",
			PlanTitle:       "Rust Async Programming",
			TotalHours:      12.5,
			PlannedHours:    40.0,
			SessionCount:    8,
			CompletedChunks: 15,
			TotalChunks:     40,
			Progress:        0.375,
			Status:          "in-progress",
			LastSession:     &lastSession,
		},
		{
			PlanID:          "french-b1",
			PlanTitle:       "French B1 Mastery",
			TotalHours:      20.0,
			PlannedHours:    60.0,
			SessionCount:    10,
			CompletedChunks: 20,
			TotalChunks:     60,
			Progress:        0.333,
			Status:          "in-progress",
			LastSession:     &lastSession,
		},
	}

	dailyStats := []DailyStats{
		{
			Date:         time.Date(2025, 10, 9, 0, 0, 0, 0, time.UTC),
			Duration:     120,
			SessionCount: 2,
			Plans:        []string{"rust-async", "french-b1"},
		},
	}

	exporter := NewExporter()
	result, err := exporter.ExportFullReport(totalStats, planStats, dailyStats)

	require.NoError(t, err)
	assert.Contains(t, result, "# Learning Statistics Report")
	assert.Contains(t, result, "## Summary")
	assert.Contains(t, result, "## Plans")
	assert.Contains(t, result, "## Daily Breakdown")
	assert.Contains(t, result, "Rust Async Programming")
	assert.Contains(t, result, "French B1 Mastery")
}

func TestExporter_ExportFullReport_NoData(t *testing.T) {
	totalStats := &TotalStats{}
	planStats := []PlanStats{}
	dailyStats := []DailyStats{}

	exporter := NewExporter()
	result, err := exporter.ExportFullReport(totalStats, planStats, dailyStats)

	require.NoError(t, err)
	assert.Contains(t, result, "# Learning Statistics Report")
	assert.Contains(t, result, "No data available")
}

func TestExporter_WithTemplate_Custom(t *testing.T) {
	template := `# Custom Report
Total: {{.TotalHours}} hours
Sessions: {{.TotalSessions}}`

	exporter := NewExporter()
	exporter.WithTemplate(template)

	stats := &TotalStats{
		TotalHours:     10.5,
		TotalSessions:  5,
		AverageSession: 126.0, // 10.5 * 60 / 5 = 126 minutes
	}

	result, err := exporter.ExportTotalStats(stats)

	require.NoError(t, err)
	assert.Contains(t, result, "# Custom Report")
	assert.Contains(t, result, "Total: 10.5 hours")
	assert.Contains(t, result, "Sessions: 5")
}

func TestExporter_WithTemplate_InvalidTemplate(t *testing.T) {
	template := `{{.TotalHours` // Invalid template syntax

	exporter := NewExporter()
	err := exporter.WithTemplate(template)

	assert.Error(t, err)
}

func TestExporter_FormatDuration_Minutes(t *testing.T) {
	tests := []struct {
		name     string
		minutes  float64
		expected string
	}{
		{"zero", 0, "0 minutes"},
		{"less than hour", 45, "45 minutes"},
		{"exactly one hour", 60, "1.0 hours"},
		{"more than hour", 125, "2.1 hours"},
		{"decimal hours", 90.5, "1.5 hours"},
	}

	exporter := NewExporter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.FormatDuration(tt.minutes)
			assert.Contains(t, result, tt.expected)
		})
	}
}

func TestExporter_FormatDate_Various(t *testing.T) {
	tests := []struct {
		name     string
		date     *time.Time
		expected string
	}{
		{
			name:     "nil date",
			date:     nil,
			expected: "N/A",
		},
		{
			name: "valid date",
			date: func() *time.Time {
				d := time.Date(2025, 10, 9, 14, 30, 0, 0, time.UTC)
				return &d
			}(),
			expected: "2025-10-09",
		},
	}

	exporter := NewExporter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.FormatDate(tt.date)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExporter_FormatProgress_Percentage(t *testing.T) {
	tests := []struct {
		name     string
		progress float64
		expected string
	}{
		{"zero progress", 0.0, "0%"},
		{"partial progress", 0.375, "37%"},
		{"half progress", 0.5, "50%"},
		{"near complete", 0.95, "95%"},
		{"complete", 1.0, "100%"},
	}

	exporter := NewExporter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.FormatProgress(tt.progress)
			assert.Contains(t, result, tt.expected)
		})
	}
}

func TestExporter_ExportToFile_Success(t *testing.T) {
	stats := &TotalStats{
		TotalHours:     10.0,
		TotalSessions:  5,
		AverageSession: 120.0, // 10.0 * 60 / 5 = 120 minutes
	}

	exporter := NewExporter()

	// Use temp file
	tmpFile := t.TempDir() + "/report.md"

	err := exporter.ExportToFile(stats, tmpFile)

	require.NoError(t, err)

	// Verify file exists and contains expected content
	content, err := exporter.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Contains(t, content, "Learning Statistics")
	assert.Contains(t, content, "10.0 hours")
}

func TestExporter_ExportToFile_InvalidPath(t *testing.T) {
	stats := &TotalStats{}
	exporter := NewExporter()

	err := exporter.ExportToFile(stats, "/invalid/path/report.md")

	assert.Error(t, err)
}

func TestExporter_ValidateStats_TotalStats(t *testing.T) {
	tests := []struct {
		name    string
		stats   *TotalStats
		wantErr bool
	}{
		{
			name: "valid stats",
			stats: &TotalStats{
				TotalHours:     10.0,
				TotalSessions:  5,
				AverageSession: 120.0, // 10.0 * 60 / 5 = 120 minutes
				CurrentStreak:  3,
				LongestStreak:  7,
			},
			wantErr: false,
		},
		{
			name: "invalid negative hours",
			stats: &TotalStats{
				TotalHours: -5.0,
			},
			wantErr: true,
		},
	}

	exporter := NewExporter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := exporter.ValidateStats(tt.stats)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExporter_GenerateMarkdownTable_Plans(t *testing.T) {
	planStats := []PlanStats{
		{
			PlanID:          "rust-async",
			PlanTitle:       "Rust Async",
			TotalHours:      12.5,
			SessionCount:    8,
			Progress:        0.375,
			Status:          "in-progress",
			CompletedChunks: 15,
			TotalChunks:     40,
		},
		{
			PlanID:          "french-b1",
			PlanTitle:       "French B1",
			TotalHours:      20.0,
			SessionCount:    10,
			Progress:        0.333,
			Status:          "in-progress",
			CompletedChunks: 20,
			TotalChunks:     60,
		},
	}

	exporter := NewExporter()
	result := exporter.GenerateMarkdownTable(planStats)

	assert.Contains(t, result, "| Plan | Hours | Sessions | Progress | Status |")
	assert.Contains(t, result, "Rust Async")
	assert.Contains(t, result, "French B1")
	assert.Contains(t, result, "12.5")
	assert.Contains(t, result, "20.0")
}

func TestExporter_GenerateProgressBar_ASCII(t *testing.T) {
	tests := []struct {
		name     string
		progress float64
		width    int
	}{
		{"zero progress", 0.0, 20},
		{"half progress", 0.5, 20},
		{"full progress", 1.0, 20},
	}

	exporter := NewExporter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.GenerateProgressBar(tt.progress, tt.width)
			assert.NotEmpty(t, result)
			// Should contain progress indicator characters
			assert.True(t, strings.Contains(result, "█") || strings.Contains(result, "░"))
		})
	}
}
