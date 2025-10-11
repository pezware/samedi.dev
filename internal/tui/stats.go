// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pezware/samedi.dev/internal/session"
	"github.com/pezware/samedi.dev/internal/stats"
	"github.com/pezware/samedi.dev/internal/tui/app"
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

// SessionProvider supplies session data for statistics.
type SessionProvider interface {
	ListAll(ctx context.Context) ([]*session.Session, error)
}

// StatsModel is the Bubble Tea module for the stats dashboard.
type StatsModel struct {
	service        *stats.Service
	sessionService SessionProvider
	ctx            context.Context
	timeRange      stats.TimeRange

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

	// Session history fields
	sessions             []*session.Session // All sessions for history view
	sessionHistoryCursor int                // Current cursor in session list

	// Export dialog fields
	exportType       string // "summary" or "full"
	exportMenuCursor int    // Cursor in export menu

	// Loading state
	loading    bool
	dataLoaded bool
	loadErr    error
}

// NewStatsModule constructs a stats module backed by the provided services.
func NewStatsModule(service *stats.Service, sessionService SessionProvider, timeRange stats.TimeRange) *StatsModel {
	return &StatsModel{
		service:        service,
		sessionService: sessionService,
		ctx:            context.Background(),
		timeRange:      timeRange,
		viewMode:       "total",
		width:          80,
		height:         24,
		currentView:    viewOverview,
		viewHistory:    []viewState{},
	}
}

// ID returns the module identifier.
func (m *StatsModel) ID() string {
	return "stats"
}

// Title is displayed in the navigation bar.
func (m *StatsModel) Title() string {
	return "Stats"
}

// Shortcuts exposes module-specific keyboard hints for the shell footer.
func (m *StatsModel) Shortcuts() []app.Shortcut {
	return []app.Shortcut{
		{Key: "p", Description: "plan list"},
		{Key: "s", Description: "sessions"},
		{Key: "e", Description: "export"},
	}
}

type statsDataLoadedMsg struct {
	totalStats   *stats.TotalStats
	allPlanStats []stats.PlanStats
	sessions     []*session.Session
	err          error
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

// SetSessions sets the list of sessions for the session history view.
func (m *StatsModel) SetSessions(sessions []*session.Session) {
	m.sessions = sessions
	m.sessionHistoryCursor = 0 // Reset cursor
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

	// Reset cursors when switching to certain views to handle filter changes
	if newView == viewSessionHistory {
		m.sessionHistoryCursor = 0
	}

	return m, nil
}

func (m *StatsModel) refreshData() tea.Cmd {
	if m.service == nil || m.sessionService == nil {
		return func() tea.Msg {
			return app.StatusMsg{
				Message: "Stats services unavailable",
				IsError: true,
			}
		}
	}

	m.loading = true
	m.loadErr = nil

	return func() tea.Msg {
		totalStats, err := m.service.GetTotalStats(m.ctx, m.timeRange)
		if err != nil {
			return statsDataLoadedMsg{err: err}
		}

		currentStreak, longestStreak, err := m.service.GetStreakInfo(m.ctx)
		if err == nil {
			totalStats.CurrentStreak = currentStreak
			totalStats.LongestStreak = longestStreak
		}

		allPlanStatsMap, err := m.service.GetAllPlanStats(m.ctx, m.timeRange)
		if err != nil {
			return statsDataLoadedMsg{err: err}
		}

		allPlanStats := make([]stats.PlanStats, 0, len(allPlanStatsMap))
		for _, ps := range allPlanStatsMap {
			allPlanStats = append(allPlanStats, ps)
		}

		sessions, err := m.sessionService.ListAll(m.ctx)
		if err != nil {
			return statsDataLoadedMsg{err: err}
		}

		return statsDataLoadedMsg{
			totalStats:   totalStats,
			allPlanStats: allPlanStats,
			sessions:     sessions,
		}
	}
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
		if m.loading {
			// Ignore input while loading to prevent inconsistent state
			return m, nil
		}
		return m.handleKeyMsg(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case app.ModuleActivatedMsg:
		if msg.ID == m.ID() {
			cmd := m.refreshData()
			return m, cmd
		}
	case app.BroadcastMsg:
		if msg.Topic == app.TopicPlansChanged {
			cmd := m.refreshData()
			return m, cmd
		}
	case statsDataLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err
			return m, func() tea.Msg {
				return app.StatusMsg{
					Message: fmt.Sprintf("Failed to load stats: %v", msg.err),
					IsError: true,
				}
			}
		}

		m.loadErr = nil
		m.dataLoaded = true
		m.totalStats = msg.totalStats
		m.planStats = nil // Reset any plan-specific view
		m.viewMode = "total"
		m.currentView = viewOverview
		m.viewHistory = m.viewHistory[:0]
		m.SetAllPlanStats(msg.allPlanStats)
		m.SetSessions(msg.sessions)

		return m, func() tea.Msg {
			return app.StatusMsg{Message: "Stats updated"}
		}
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

	if m.currentView == viewExport {
		// Store export type based on selection
		if m.exportMenuCursor == 0 {
			m.exportType = "summary"
		} else {
			m.exportType = "full"
		}
		// Note: Actual export happens via CLI commands (samedi report)
		// TUI mode is for viewing, CLI for exporting
		// Return to previous view (showing the selection was registered)
		return m.goBack()
	}

	return m, nil
}

// handleRuneKey handles character key presses.
func (m *StatsModel) handleRuneKey(r rune) (tea.Model, tea.Cmd) {
	switch r {
	case 'q':
		return m, tea.Quit
	case 'p':
		// Don't switch if already on plan list view
		if m.currentView == viewPlanList {
			return m, nil
		}
		return m.switchView(viewPlanList)
	case 's':
		// Don't switch if already on session history view
		if m.currentView == viewSessionHistory {
			return m, nil
		}
		// If in plan detail view, switch to session history filtered by this plan
		// Otherwise, switch to session history (all sessions)
		return m.switchView(viewSessionHistory)
	case 'e':
		// Don't switch if already on export dialog view
		if m.currentView == viewExport {
			return m, nil
		}
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

	// Handle session history navigation
	if m.currentView == viewSessionHistory {
		// Get filtered sessions to navigate within the correct bounds
		filteredSessions := m.filterSessionsByPlan()
		if len(filteredSessions) > 0 {
			m.sessionHistoryCursor += direction
			// Wrap around using filtered list length
			if m.sessionHistoryCursor < 0 {
				m.sessionHistoryCursor = len(filteredSessions) - 1
			} else if m.sessionHistoryCursor >= len(filteredSessions) {
				m.sessionHistoryCursor = 0
			}
		}
	}

	// Handle export menu navigation
	if m.currentView == viewExport {
		m.exportMenuCursor += direction
		// Wrap around (2 options: summary and full)
		if m.exportMenuCursor < 0 {
			m.exportMenuCursor = 1
		} else if m.exportMenuCursor > 1 {
			m.exportMenuCursor = 0
		}
	}

	return m, nil
}

// View renders the TUI based on current view state.
func (m *StatsModel) View() string {
	if m.loading {
		return "Loading statistics…"
	}

	if m.loadErr != nil {
		return fmt.Sprintf("Failed to load stats: %v\n\nPress Tab to retry.", m.loadErr)
	}

	if !m.dataLoaded || m.totalStats == nil {
		return "No statistics available yet.\nStart a session to generate learning data."
	}

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
		content.WriteString(titleStyle.Render("Learning Statistics"))
		content.WriteString("\n\n")
		content.WriteString(m.renderTotalStats())
	} else {
		content.WriteString(titleStyle.Render(fmt.Sprintf("Statistics: %s", m.planStats.PlanTitle)))
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

	content.WriteString(titleStyle.Render("Learning Plans"))
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

	content.WriteString(titleStyle.Render(m.selectedPlan.PlanTitle))
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
	content.WriteString(m.renderSection("Progress", []string{
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

	content.WriteString(m.renderSection("Time Investment", []string{
		fmt.Sprintf("Total hours:      %.1f / %.1f hours", m.selectedPlan.TotalHours, m.selectedPlan.PlannedHours),
		fmt.Sprintf("Sessions:         %d", m.selectedPlan.SessionCount),
		fmt.Sprintf("Average session:  %.0f minutes", avgMinutes),
	}))

	// Last session info
	if m.selectedPlan.LastSession != nil {
		content.WriteString("\n")
		content.WriteString(m.renderSection("Last Session", []string{
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

// renderSessionHistory renders the session history view with filtering and navigation.
func (m *StatsModel) renderSessionHistory() string {
	var content strings.Builder

	// Title
	title := m.getSessionHistoryTitle()
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		PaddingBottom(1)
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Filter sessions
	filteredSessions := m.filterSessionsByPlan()

	// Empty state
	if len(filteredSessions) == 0 {
		content.WriteString(m.renderSessionHistoryEmpty())
		return content.String()
	}

	// Build table
	table := m.buildSessionTable(filteredSessions)
	content.WriteString(table)
	content.WriteString("\n\n")

	// Footer
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	content.WriteString(footerStyle.Render(fmt.Sprintf("Showing %d sessions", len(filteredSessions))))
	content.WriteString("\n\n")

	// Help
	content.WriteString(m.renderSessionHistoryHelp())

	return content.String()
}

// getSessionHistoryTitle returns the title for session history view with optional plan filter.
func (m *StatsModel) getSessionHistoryTitle() string {
	if m.selectedPlanID != "" && m.selectedPlan != nil {
		return fmt.Sprintf("Session History: %s", m.selectedPlan.PlanTitle)
	}
	return "Session History"
}

// filterSessionsByPlan filters sessions by selected plan if applicable.
func (m *StatsModel) filterSessionsByPlan() []*session.Session {
	if m.selectedPlanID == "" {
		return m.sessions
	}

	filtered := make([]*session.Session, 0)
	for _, s := range m.sessions {
		if s.PlanID == m.selectedPlanID {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// renderSessionHistoryEmpty renders empty state for session history.
func (m *StatsModel) renderSessionHistoryEmpty() string {
	var content strings.Builder
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	content.WriteString(emptyStyle.Render("No sessions found."))
	content.WriteString("\n\n")
	content.WriteString(m.renderSessionHistoryHelp())
	return content.String()
}

// buildSessionTable builds the session table with pagination.
func (m *StatsModel) buildSessionTable(filteredSessions []*session.Session) string {
	table := components.NewTable([]string{"Date", "Plan", "Duration", "Notes"})

	// Paginate sessions (max 20 visible)
	displaySessions, startOffset := m.paginateSessions(filteredSessions, 20)

	// Add rows with absolute index for highlight comparison
	for i, sess := range displaySessions {
		absoluteIndex := startOffset + i
		row := m.formatSessionRow(sess, absoluteIndex == m.sessionHistoryCursor)
		table.AddRow(row)
	}

	return table.View()
}

// paginateSessions returns a slice of sessions to display based on cursor position
// and the starting offset in the original list.
func (m *StatsModel) paginateSessions(sessions []*session.Session, maxDisplay int) ([]*session.Session, int) {
	if len(sessions) <= maxDisplay {
		return sessions, 0
	}

	// Center window around cursor
	start := m.sessionHistoryCursor - maxDisplay/2
	if start < 0 {
		start = 0
	}

	end := start + maxDisplay
	if end > len(sessions) {
		end = len(sessions)
		start = end - maxDisplay
		if start < 0 {
			start = 0
		}
	}

	return sessions[start:end], start
}

// formatSessionRow formats a single session as a table row.
func (m *StatsModel) formatSessionRow(sess *session.Session, isSelected bool) []string {
	// Format values
	dateStr := sess.StartTime.Format("Jan 2, 2006 15:04")
	planID := truncateString(sess.PlanID, 15)
	durationStr := sess.ElapsedTime()
	notesPreview := formatNotes(sess.Notes, 30)

	// Apply highlighting if selected
	if isSelected {
		highlightStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("12")).
			Bold(true)
		dateStr = highlightStyle.Render(dateStr)
		planID = highlightStyle.Render(planID)
		durationStr = highlightStyle.Render(durationStr)
		notesPreview = highlightStyle.Render(notesPreview)
	}

	return []string{dateStr, planID, durationStr, notesPreview}
}

// truncateString truncates a string to maxLen with ellipsis.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatNotes formats notes for display in table.
func formatNotes(notes string, maxLen int) string {
	// Remove newlines
	notes = strings.ReplaceAll(notes, "\n", " ")

	// Truncate if needed
	if len(notes) > maxLen {
		notes = notes[:maxLen-3] + "..."
	}

	// Return dash if empty
	if notes == "" {
		return "-"
	}

	return notes
}

// renderSessionHistoryHelp renders help text for the session history view.
func (m *StatsModel) renderSessionHistoryHelp() string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	return helpStyle.Render("[↑/k] Up  |  [↓/j] Down  |  [Esc] Back")
}

// renderExportDialog renders the export dialog with options for quick export.
func (m *StatsModel) renderExportDialog() string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		PaddingBottom(1)

	content.WriteString(titleStyle.Render("Export Learning Report"))
	content.WriteString("\n\n")

	// Info text
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	content.WriteString(infoStyle.Render("Select export type:"))
	content.WriteString("\n\n")

	// Export options
	exportOptions := []struct {
		name        string
		description string
	}{
		{"Summary Report", "Quick overview of your learning progress"},
		{"Full Report", "Detailed report with daily breakdowns"},
	}

	for i, option := range exportOptions {
		optionStyle := lipgloss.NewStyle()

		// Highlight selected option
		if i == m.exportMenuCursor {
			optionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("12")).
				Bold(true).
				Width(50)
		}

		nameText := fmt.Sprintf("  [%d] %s", i+1, option.name)
		content.WriteString(optionStyle.Render(nameText))
		content.WriteString("\n")

		if i == m.exportMenuCursor {
			descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).PaddingLeft(6)
			content.WriteString(descStyle.Render(option.description))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Note about output
	noteStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Italic(true)
	content.WriteString(noteStyle.Render("Note: Report will be printed to terminal. Use shell redirection to save to file."))
	content.WriteString("\n")
	content.WriteString(noteStyle.Render("      Example: samedi stats --tui (then press 'e' and Enter) > report.md"))
	content.WriteString("\n\n")

	// Help
	content.WriteString(m.renderExportHelp())

	return content.String()
}

// renderExportHelp renders help text for the export dialog.
func (m *StatsModel) renderExportHelp() string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	return helpStyle.Render("[↑/k] Up  |  [↓/j] Down  |  [Enter] Export  |  [Esc] Cancel")
}

// renderTotalStats renders total statistics view.
func (m *StatsModel) renderTotalStats() string {
	if m.totalStats == nil {
		return "No statistics available"
	}

	var result strings.Builder

	// Learning time section
	result.WriteString(m.renderSection("Learning Time", []string{
		fmt.Sprintf("Total hours:      %.1f hours", m.totalStats.TotalHours),
		fmt.Sprintf("Total sessions:   %d", m.totalStats.TotalSessions),
		fmt.Sprintf("Average session:  %.0f minutes", m.totalStats.AverageSession),
	}))

	result.WriteString("\n")

	// Streaks section
	result.WriteString(m.renderSection("Learning Streaks", []string{
		fmt.Sprintf("Current streak:   %d days", m.totalStats.CurrentStreak),
		fmt.Sprintf("Longest streak:   %d days", m.totalStats.LongestStreak),
	}))

	result.WriteString("\n")

	// Plans section
	result.WriteString(m.renderSection("Learning Plans", []string{
		fmt.Sprintf("Active plans:     %d", m.totalStats.ActivePlans),
		fmt.Sprintf("Completed plans:  %d", m.totalStats.CompletedPlans),
		fmt.Sprintf("Total plans:      %d", m.totalStats.ActivePlans+m.totalStats.CompletedPlans),
	}))

	// Last session
	if m.totalStats.LastSessionDate != nil {
		result.WriteString("\n")
		result.WriteString(m.renderSection("Last Session", []string{
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
	result.WriteString(m.renderSection("Progress", []string{
		progressBar.View(),
		fmt.Sprintf("Completed chunks: %d / %d", m.planStats.CompletedChunks, m.planStats.TotalChunks),
	}))

	result.WriteString("\n")

	// Time section
	avgMinutes := 0.0
	if m.planStats.SessionCount > 0 {
		avgMinutes = (m.planStats.TotalHours * 60) / float64(m.planStats.SessionCount)
	}

	result.WriteString(m.renderSection("Time", []string{
		fmt.Sprintf("Total hours:      %.1f / %.1f hours", m.planStats.TotalHours, m.planStats.PlannedHours),
		fmt.Sprintf("Sessions:         %d", m.planStats.SessionCount),
		fmt.Sprintf("Average session:  %.0f minutes", avgMinutes),
	}))

	result.WriteString("\n")

	// Status section
	statusStr := formatPlanStatus(m.planStats.Status)
	result.WriteString(m.renderSection("Status", []string{
		statusStr,
	}))

	// Last session
	if m.planStats.LastSession != nil {
		result.WriteString("\n")
		result.WriteString(m.renderSection("Last Session", []string{
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

	helpText := "[q] quit  |  [p] plan list  |  [s] sessions  |  [e] export\n" +
		"[↑/k] up  |  [↓/j] down  |  [Enter] select  |  [Esc] back"

	return helpStyle.Render(helpText)
}

// formatPlanStatus formats a status string with text indicators.
func formatPlanStatus(status string) string {
	switch status {
	case "not-started":
		return "[ ] Not Started"
	case "in-progress":
		return "[→] In Progress"
	case "completed":
		return "[✓] Completed"
	case "archived":
		return "[*] Archived"
	default:
		return status
	}
}
