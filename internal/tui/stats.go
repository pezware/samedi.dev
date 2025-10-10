// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pezware/samedi.dev/internal/stats"
	"github.com/pezware/samedi.dev/internal/tui/components"
)

// viewState represents the current view in the stats TUI.
type viewState string

const (
	viewOverview       viewState = "overview"        // Default stats overview
	viewPlanList       viewState = "plan-list"       // List of all plans
	viewPlanDetail     viewState = "plan-detail"     // Single plan drill-down
	viewSessionHistory viewState = "session-history" // Session list
	viewExport         viewState = "export-dialog"   // Export configuration
)

// StatsModel is the Bubble Tea model for the stats dashboard.
type StatsModel struct {
	totalStats *stats.TotalStats
	planStats  *stats.PlanStats
	viewMode   string // "total" or "plan" - kept for backward compatibility
	width      int
	height     int

	// New fields for multi-view navigation
	currentView    viewState   // Current active view
	viewHistory    []viewState // Stack for back navigation
	selectedPlanID string      // Plan ID for drill-down context
}

// NewStatsModel creates a new stats model.
// Either totalStats or planStats should be non-nil, but not both.
func NewStatsModel(totalStats *stats.TotalStats, planStats *stats.PlanStats) *StatsModel {
	viewMode := "total"
	if planStats != nil {
		viewMode = "plan"
	}

	return &StatsModel{
		totalStats:  totalStats,
		planStats:   planStats,
		viewMode:    viewMode,
		width:       80,
		height:      24,
		currentView: viewOverview,  // Start at overview
		viewHistory: []viewState{}, // Empty history stack
	}
}

// Init initializes the model.
func (m *StatsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m *StatsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyRunes:
			if len(msg.Runes) > 0 && msg.Runes[0] == 'q' {
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the TUI.
func (m *StatsModel) View() string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		PaddingBottom(1)

	if m.viewMode == "total" {
		content.WriteString(titleStyle.Render("📊 Learning Statistics"))
		content.WriteString("\n\n")
		content.WriteString(m.renderTotalStats())
	} else {
		content.WriteString(titleStyle.Render(fmt.Sprintf("📊 Statistics: %s", m.planStats.PlanTitle)))
		content.WriteString("\n\n")
		content.WriteString(m.renderPlanStats())
	}

	// Help
	content.WriteString("\n\n")
	content.WriteString(m.renderHelp())

	return content.String()
}

// renderTotalStats renders total statistics view.
func (m *StatsModel) renderTotalStats() string {
	if m.totalStats == nil {
		return "No statistics available"
	}

	var result strings.Builder

	// Learning time section
	result.WriteString(m.renderSection("⏱️  Learning Time", []string{
		fmt.Sprintf("Total hours:      %.1f hours", m.totalStats.TotalHours),
		fmt.Sprintf("Total sessions:   %d", m.totalStats.TotalSessions),
		fmt.Sprintf("Average session:  %.0f minutes", m.totalStats.AverageSession),
	}))

	result.WriteString("\n")

	// Streaks section
	result.WriteString(m.renderSection("🔥 Learning Streaks", []string{
		fmt.Sprintf("Current streak:   %d days", m.totalStats.CurrentStreak),
		fmt.Sprintf("Longest streak:   %d days", m.totalStats.LongestStreak),
	}))

	result.WriteString("\n")

	// Plans section
	result.WriteString(m.renderSection("📚 Learning Plans", []string{
		fmt.Sprintf("Active plans:     %d", m.totalStats.ActivePlans),
		fmt.Sprintf("Completed plans:  %d", m.totalStats.CompletedPlans),
		fmt.Sprintf("Total plans:      %d", m.totalStats.ActivePlans+m.totalStats.CompletedPlans),
	}))

	// Last session
	if m.totalStats.LastSessionDate != nil {
		result.WriteString("\n")
		result.WriteString(m.renderSection("📅 Last Session", []string{
			m.totalStats.LastSessionDate.Format("Monday, January 2, 2006 at 3:04 PM"),
		}))
	}

	return result.String()
}

// renderPlanStats renders plan-specific statistics view.
func (m *StatsModel) renderPlanStats() string {
	if m.planStats == nil {
		return "No plan statistics available"
	}

	var result strings.Builder

	// Progress section with visual progress bar
	progressBar := components.NewProgressBar(m.planStats.Progress, 30)
	result.WriteString(m.renderSection("📈 Progress", []string{
		progressBar.View(),
		fmt.Sprintf("Completed chunks: %d / %d", m.planStats.CompletedChunks, m.planStats.TotalChunks),
	}))

	result.WriteString("\n")

	// Time section
	avgMinutes := 0.0
	if m.planStats.SessionCount > 0 {
		avgMinutes = (m.planStats.TotalHours * 60) / float64(m.planStats.SessionCount)
	}

	result.WriteString(m.renderSection("⏱️  Time", []string{
		fmt.Sprintf("Total hours:      %.1f / %.1f hours", m.planStats.TotalHours, m.planStats.PlannedHours),
		fmt.Sprintf("Sessions:         %d", m.planStats.SessionCount),
		fmt.Sprintf("Average session:  %.0f minutes", avgMinutes),
	}))

	result.WriteString("\n")

	// Status section
	statusStr := formatPlanStatus(m.planStats.Status)
	result.WriteString(m.renderSection("📊 Status", []string{
		statusStr,
	}))

	// Last session
	if m.planStats.LastSession != nil {
		result.WriteString("\n")
		result.WriteString(m.renderSection("📅 Last Session", []string{
			m.planStats.LastSession.Format("Monday, January 2, 2006 at 3:04 PM"),
		}))
	}

	return result.String()
}

// renderSection renders a section with title and items.
func (m *StatsModel) renderSection(title string, items []string) string {
	var section strings.Builder

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")) // Cyan

	section.WriteString(sectionStyle.Render(title))
	section.WriteString("\n")

	for _, item := range items {
		section.WriteString("   ")
		section.WriteString(item)
		section.WriteString("\n")
	}

	return section.String()
}

// renderHelp renders help text.
func (m *StatsModel) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")) // Gray

	return helpStyle.Render("Press [q] to quit  |  Use --range flag to filter by time (e.g., samedi stats --range today --tui)")
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
