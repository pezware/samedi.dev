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
	currentView    viewState         // Current active view
	viewHistory    []viewState       // Stack for back navigation
	selectedPlanID string            // Plan ID for drill-down context
	selectedPlan   *stats.PlanStats  // Detailed stats for selected plan
	allPlanStats   []stats.PlanStats // All plan statistics for list view
	planListCursor int               // Current cursor position in plan list (0-indexed)
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

// SetAllPlanStats sets the list of all plan statistics for the plan list view.
func (m *StatsModel) SetAllPlanStats(planStats []stats.PlanStats) {
	m.allPlanStats = planStats
	m.planListCursor = 0 // Reset cursor
}

// switchView transitions to a new view and updates history stack.
// Data loading commands for specific views will be added in Phase 2+.
//
//nolint:unparam // tea.Cmd will be used when data loading is implemented
func (m *StatsModel) switchView(newView viewState) (*StatsModel, tea.Cmd) {
	// Push current view to history stack
	m.viewHistory = append(m.viewHistory, m.currentView)

	// Update current view
	m.currentView = newView

	return m, nil
}

// goBack returns to the previous view from history stack.
//
//nolint:unparam // tea.Cmd return kept for consistency with Bubble Tea patterns
func (m *StatsModel) goBack() (*StatsModel, tea.Cmd) {
	// If history is empty, stay at current view
	if len(m.viewHistory) == 0 {
		return m, nil
	}

	// Pop last view from history
	lastIndex := len(m.viewHistory) - 1
	previousView := m.viewHistory[lastIndex]
	m.viewHistory = m.viewHistory[:lastIndex]

	// Update current view
	m.currentView = previousView

	return m, nil
}

// Update handles messages and updates the model.
func (m *StatsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// handleKeyMsg handles keyboard input messages.
func (m *StatsModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		return m.goBack()
	case tea.KeyEnter:
		return m.handleEnterKey()
	case tea.KeyUp:
		return m.handleArrowKey(-1)
	case tea.KeyDown:
		return m.handleArrowKey(1)
	case tea.KeyRunes:
		if len(msg.Runes) > 0 {
			return m.handleRuneKey(msg.Runes[0])
		}
	}

	return m, nil
}

// handleEnterKey handles the Enter key based on current view.
func (m *StatsModel) handleEnterKey() (tea.Model, tea.Cmd) {
	if m.currentView == viewPlanList && len(m.allPlanStats) > 0 {
		// Select plan and switch to detail view
		selectedStat := m.allPlanStats[m.planListCursor]
		m.selectedPlanID = selectedStat.PlanID
		m.selectedPlan = &selectedStat
		return m.switchView(viewPlanDetail)
	}
	return m, nil
}

// handleRuneKey handles character key presses.
func (m *StatsModel) handleRuneKey(r rune) (tea.Model, tea.Cmd) {
	switch r {
	case 'q':
		return m, tea.Quit
	case 'p':
		return m.switchView(viewPlanList)
	case 's':
		// If in plan detail view, switch to session history filtered by this plan
		// Otherwise, switch to session history (all sessions)
		return m.switchView(viewSessionHistory)
	case 'e':
		return m.switchView(viewExport)
	case 'j':
		return m.handleArrowKey(1)
	case 'k':
		return m.handleArrowKey(-1)
	}

	return m, nil
}

// handleArrowKey handles up/down navigation in list views.
//
//nolint:unparam // tea.Cmd return kept for consistency with Bubble Tea patterns
func (m *StatsModel) handleArrowKey(direction int) (*StatsModel, tea.Cmd) {
	// Handle plan list navigation
	if m.currentView == viewPlanList && len(m.allPlanStats) > 0 {
		m.planListCursor += direction
		// Wrap around
		if m.planListCursor < 0 {
			m.planListCursor = len(m.allPlanStats) - 1
		} else if m.planListCursor >= len(m.allPlanStats) {
			m.planListCursor = 0
		}
	}
	// Add other list views here in future phases

	return m, nil
}

// View renders the TUI based on current view state.
func (m *StatsModel) View() string {
	switch m.currentView {
	case viewPlanList:
		return m.renderPlanList()
	case viewPlanDetail:
		return m.renderPlanDetail()
	case viewSessionHistory:
		return m.renderSessionHistory()
	case viewExport:
		return m.renderExportDialog()
	default: // viewOverview
		return m.renderOverview()
	}
}

// renderOverview renders the overview/stats view.
func (m *StatsModel) renderOverview() string {
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

// renderPlanList renders the plan list view with navigation.
func (m *StatsModel) renderPlanList() string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		PaddingBottom(1)

	content.WriteString(titleStyle.Render("📚 Learning Plans"))
	content.WriteString("\n\n")

	// If no plans, show empty state
	if len(m.allPlanStats) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		content.WriteString(emptyStyle.Render("No plans found. Create a plan to get started!"))
		content.WriteString("\n\n")
		content.WriteString(m.renderPlanListHelp())
		return content.String()
	}

	// Create table with headers
	table := components.NewTable([]string{"Title", "Progress", "Hours", "Status"})

	// Add rows for each plan
	for i, planStat := range m.allPlanStats {
		// Format values
		title := planStat.PlanTitle
		progress := fmt.Sprintf("%d%%", planStat.ProgressPercent())
		hours := fmt.Sprintf("%.1f / %.1f", planStat.TotalHours, planStat.PlannedHours)
		status := formatPlanStatus(planStat.Status)

		// Highlight selected row
		if i == m.planListCursor {
			highlightStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("12")).
				Bold(true)
			title = highlightStyle.Render(title)
			progress = highlightStyle.Render(progress)
			hours = highlightStyle.Render(hours)
			status = highlightStyle.Render(status)
		}

		table.AddRow([]string{title, progress, hours, status})
	}

	content.WriteString(table.View())
	content.WriteString("\n\n")

	// Footer info
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	content.WriteString(footerStyle.Render(fmt.Sprintf("Showing %d plans", len(m.allPlanStats))))
	content.WriteString("\n\n")

	// Help
	content.WriteString(m.renderPlanListHelp())

	return content.String()
}

// renderPlanListHelp renders help text for the plan list view.
func (m *StatsModel) renderPlanListHelp() string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	return helpStyle.Render("[↑/k] Up  |  [↓/j] Down  |  [Enter] View Details  |  [Esc] Back")
}

// renderPlanDetail renders the plan detail view with comprehensive plan information.
func (m *StatsModel) renderPlanDetail() string {
	if m.selectedPlan == nil {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		return lipgloss.NewStyle().Padding(2).Render(
			emptyStyle.Render("No plan selected"),
		)
	}

	var content strings.Builder

	// Title with plan name
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		PaddingBottom(1)

	content.WriteString(titleStyle.Render(fmt.Sprintf("📊 %s", m.selectedPlan.PlanTitle)))
	content.WriteString("\n\n")

	// Status badge
	statusStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14"))
	content.WriteString(statusStyle.Render("Status: "))
	content.WriteString(formatPlanStatus(m.selectedPlan.Status))
	content.WriteString("\n\n")

	// Progress section with visual progress bar
	progressBar := components.NewProgressBar(m.selectedPlan.Progress, 40)
	content.WriteString(m.renderSection("📈 Progress", []string{
		progressBar.View(),
		fmt.Sprintf("Completed: %d / %d chunks (%.0f%%)",
			m.selectedPlan.CompletedChunks,
			m.selectedPlan.TotalChunks,
			m.selectedPlan.Progress*100),
	}))

	content.WriteString("\n")

	// Time statistics
	avgMinutes := 0.0
	if m.selectedPlan.SessionCount > 0 {
		avgMinutes = (m.selectedPlan.TotalHours * 60) / float64(m.selectedPlan.SessionCount)
	}

	content.WriteString(m.renderSection("⏱️  Time Investment", []string{
		fmt.Sprintf("Total hours:      %.1f / %.1f hours", m.selectedPlan.TotalHours, m.selectedPlan.PlannedHours),
		fmt.Sprintf("Sessions:         %d", m.selectedPlan.SessionCount),
		fmt.Sprintf("Average session:  %.0f minutes", avgMinutes),
	}))

	// Last session info
	if m.selectedPlan.LastSession != nil {
		content.WriteString("\n")
		content.WriteString(m.renderSection("📅 Last Session", []string{
			m.selectedPlan.LastSession.Format("Monday, January 2, 2006 at 3:04 PM"),
		}))
	}

	// Help text
	content.WriteString("\n\n")
	content.WriteString(m.renderPlanDetailHelp())

	return content.String()
}

// renderPlanDetailHelp renders help text for the plan detail view.
func (m *StatsModel) renderPlanDetailHelp() string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	return helpStyle.Render("[s] View Sessions  |  [Esc] Back to Plan List")
}

// renderSessionHistory renders the session history view (stub for Phase 3.1).
func (m *StatsModel) renderSessionHistory() string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	return lipgloss.NewStyle().Padding(2).Render(
		"📅 Session History View\n\n" +
			"Coming soon in Phase 3.1...\n\n" +
			helpStyle.Render("[Esc] Back to Overview"),
	)
}

// renderExportDialog renders the export dialog view (stub for Phase 4.1).
func (m *StatsModel) renderExportDialog() string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	return lipgloss.NewStyle().Padding(2).Render(
		"📤 Export Dialog\n\n" +
			"Coming soon in Phase 4.1...\n\n" +
			helpStyle.Render("[Esc] Cancel and go back"),
	)
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
