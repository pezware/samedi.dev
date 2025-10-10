// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package tui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pezware/samedi.dev/internal/session"
	"github.com/pezware/samedi.dev/internal/stats"
	"github.com/stretchr/testify/assert"
)

func TestStatsModel_Init(t *testing.T) {
	totalStats := &stats.TotalStats{
		TotalHours:     42.5,
		TotalSessions:  20,
		CurrentStreak:  5,
		LongestStreak:  10,
		ActivePlans:    3,
		CompletedPlans: 2,
	}

	model := NewStatsModel(totalStats, nil)

	// Init should return nil command
	cmd := model.Init()
	assert.Nil(t, cmd)
}

func TestStatsModel_Update_Quit(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Press 'q' to quit
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Should return quit command
	assert.NotNil(t, cmd)
	assert.IsType(t, &StatsModel{}, updatedModel)
}

func TestStatsModel_Update_CtrlC(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Press Ctrl+C to quit
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	// Should return quit command
	assert.NotNil(t, cmd)
	assert.IsType(t, &StatsModel{}, updatedModel)
}

func TestStatsModel_View_TotalStats(t *testing.T) {
	now := time.Now()
	totalStats := &stats.TotalStats{
		TotalHours:      42.5,
		TotalSessions:   20,
		AverageSession:  127.5,
		CurrentStreak:   5,
		LongestStreak:   10,
		ActivePlans:     3,
		CompletedPlans:  2,
		LastSessionDate: &now,
	}

	model := NewStatsModel(totalStats, nil)
	view := model.View()

	// Should contain key statistics
	assert.Contains(t, view, "Learning Statistics")
	assert.Contains(t, view, "42.5") // Total hours
	assert.Contains(t, view, "20")   // Total sessions
	assert.Contains(t, view, "5")    // Current streak
	assert.Contains(t, view, "10")   // Longest streak
	assert.Contains(t, view, "3")    // Active plans
	assert.Contains(t, view, "2")    // Completed plans
}

func TestStatsModel_View_PlanStats(t *testing.T) {
	now := time.Now()
	planStats := &stats.PlanStats{
		PlanTitle:       "Rust Async Programming",
		Status:          "in-progress",
		Progress:        0.5,
		TotalChunks:     20,
		CompletedChunks: 10,
		TotalHours:      21.0,
		PlannedHours:    40.0,
		SessionCount:    15,
		LastSession:     &now,
	}

	model := NewStatsModel(nil, planStats)
	view := model.View()

	// Should contain plan-specific statistics
	assert.Contains(t, view, "Rust Async Programming")
	assert.Contains(t, view, "50%")  // Progress
	assert.Contains(t, view, "10")   // Completed chunks
	assert.Contains(t, view, "20")   // Total chunks
	assert.Contains(t, view, "21.0") // Total hours
	assert.Contains(t, view, "40.0") // Planned hours
	assert.Contains(t, view, "15")   // Session count
}

func TestStatsModel_View_Help(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)
	view := model.View()

	// Should contain help text
	assert.Contains(t, view, "q") // Quit key
	assert.Contains(t, view, "quit")
}

func TestStatsModel_ViewMode_Total(t *testing.T) {
	totalStats := &stats.TotalStats{TotalHours: 100}
	model := NewStatsModel(totalStats, nil)

	// Should default to total stats view
	assert.Equal(t, "total", model.viewMode)

	view := model.View()
	assert.Contains(t, view, "100") // Total hours
}

func TestStatsModel_ViewMode_Plan(t *testing.T) {
	planStats := &stats.PlanStats{
		PlanTitle:  "Test Plan",
		TotalHours: 25.0,
	}
	model := NewStatsModel(nil, planStats)

	// Should use plan stats view
	assert.Equal(t, "plan", model.viewMode)

	view := model.View()
	assert.Contains(t, view, "Test Plan")
	assert.Contains(t, view, "25.0")
}

// Test new view state functionality for Stage 6

func TestStatsModel_InitialViewState(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Should start at overview
	assert.Equal(t, viewOverview, model.currentView)
}

func TestStatsModel_InitialViewHistory(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Should start with empty history
	assert.Empty(t, model.viewHistory)
}

func TestStatsModel_InitialSelectedPlanID(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Should start with no selected plan
	assert.Equal(t, "", model.selectedPlanID)
}

func TestStatsModel_SwitchView_ToPlanList(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Press 'p' to switch to plan list
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*StatsModel)

	// Should be in plan list view
	assert.Equal(t, viewPlanList, m.currentView)
	// Should have overview in history
	assert.Equal(t, []viewState{viewOverview}, m.viewHistory)
}

func TestStatsModel_SwitchView_ToSessionHistory(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Press 's' to switch to session history
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m := updatedModel.(*StatsModel)

	// Should be in session history view
	assert.Equal(t, viewSessionHistory, m.currentView)
	assert.Equal(t, []viewState{viewOverview}, m.viewHistory)
}

func TestStatsModel_SwitchView_ToExport(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Press 'e' to switch to export
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m := updatedModel.(*StatsModel)

	// Should be in export view
	assert.Equal(t, viewExport, m.currentView)
	assert.Equal(t, []viewState{viewOverview}, m.viewHistory)
}

func TestStatsModel_GoBack_SingleLevel(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Switch to plan list
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*StatsModel)
	assert.Equal(t, viewPlanList, m.currentView)

	// Press Esc to go back
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(*StatsModel)

	// Should be back at overview
	assert.Equal(t, viewOverview, m.currentView)
	// History should be empty
	assert.Empty(t, m.viewHistory)
}

func TestStatsModel_GoBack_MultiLevel(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Switch to plan list
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*StatsModel)

	// Switch to session history
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, viewSessionHistory, m.currentView)
	assert.Equal(t, []viewState{viewOverview, viewPlanList}, m.viewHistory)

	// Go back once
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, viewPlanList, m.currentView)
	assert.Equal(t, []viewState{viewOverview}, m.viewHistory)

	// Go back again
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, viewOverview, m.currentView)
	assert.Empty(t, m.viewHistory)
}

func TestStatsModel_GoBack_EmptyHistory(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Press Esc at overview (no history)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m := updatedModel.(*StatsModel)

	// Should stay at overview
	assert.Equal(t, viewOverview, m.currentView)
	assert.Empty(t, m.viewHistory)
}

// Test view rendering and routing

func TestStatsModel_View_RendersOverviewByDefault(t *testing.T) {
	totalStats := &stats.TotalStats{TotalHours: 42.5}
	model := NewStatsModel(totalStats, nil)

	view := model.View()

	// Should render overview with stats
	assert.Contains(t, view, "Learning Statistics")
	assert.Contains(t, view, "42.5")
}

func TestStatsModel_View_RendersPlanListStub(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Switch to plan list
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*StatsModel)

	view := m.View()

	// Should render plan list view (now implemented!)
	assert.Contains(t, view, "Learning Plans")
	assert.Contains(t, view, "No plans found")
	assert.Contains(t, view, "Esc")
}

func TestStatsModel_View_RendersSessionHistoryStub(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Switch to session history
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m := updatedModel.(*StatsModel)

	view := m.View()

	// Should render session history view (now implemented!)
	assert.Contains(t, view, "Session History")
	// Should show empty state when no sessions
	assert.Contains(t, view, "No sessions found")
	// Should show help
	assert.Contains(t, view, "Esc")
}

func TestStatsModel_View_RendersExportStub(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Switch to export
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m := updatedModel.(*StatsModel)

	view := m.View()

	// Should render export dialog (now implemented!)
	assert.Contains(t, view, "Export Learning Report")
	// Should show export options
	assert.Contains(t, view, "Summary Report")
	assert.Contains(t, view, "Full Report")
	// Should show help
	assert.Contains(t, view, "Enter")
	assert.Contains(t, view, "Esc")
}

func TestStatsModel_View_SwitchingBetweenViews(t *testing.T) {
	totalStats := &stats.TotalStats{TotalHours: 10}
	model := NewStatsModel(totalStats, nil)

	// Start at overview
	view := model.View()
	assert.Contains(t, view, "Learning Statistics")

	// Switch to plan list
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*StatsModel)
	view = m.View()
	assert.Contains(t, view, "Learning Plans")

	// Go back to overview
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(*StatsModel)
	view = m.View()
	assert.Contains(t, view, "Learning Statistics")
}

// Test Phase 2.1: Plan List View

func TestStatsModel_SetAllPlanStats(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	planStats := []stats.PlanStats{
		{PlanID: "plan1", PlanTitle: "Rust", Progress: 0.5},
		{PlanID: "plan2", PlanTitle: "Go", Progress: 0.3},
	}

	model.SetAllPlanStats(planStats)

	// Should set plan stats
	assert.Equal(t, 2, len(model.allPlanStats))
	assert.Equal(t, "plan1", model.allPlanStats[0].PlanID)
	// Should reset cursor
	assert.Equal(t, 0, model.planListCursor)
}

func TestStatsModel_PlanList_EmptyState(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Switch to plan list without setting plans
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*StatsModel)

	view := m.View()

	// Should show empty state
	assert.Contains(t, view, "No plans found")
	assert.Contains(t, view, "Learning Plans")
}

func TestStatsModel_PlanList_NavigateDown(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	planStats := []stats.PlanStats{
		{PlanID: "plan1", PlanTitle: "Rust"},
		{PlanID: "plan2", PlanTitle: "Go"},
		{PlanID: "plan3", PlanTitle: "Python"},
	}
	model.SetAllPlanStats(planStats)

	// Switch to plan list
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*StatsModel)
	assert.Equal(t, 0, m.planListCursor)

	// Navigate down with 'j'
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 1, m.planListCursor)

	// Navigate down with arrow key
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 2, m.planListCursor)
}

func TestStatsModel_PlanList_NavigateUp(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	planStats := []stats.PlanStats{
		{PlanID: "plan1", PlanTitle: "Rust"},
		{PlanID: "plan2", PlanTitle: "Go"},
		{PlanID: "plan3", PlanTitle: "Python"},
	}
	model.SetAllPlanStats(planStats)

	// Switch to plan list and move to middle
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model.planListCursor = 2

	// Navigate up with 'k'
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m := updatedModel.(*StatsModel)
	assert.Equal(t, 1, m.planListCursor)

	// Navigate up with arrow key
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 0, m.planListCursor)
}

func TestStatsModel_PlanList_WrapAround(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	planStats := []stats.PlanStats{
		{PlanID: "plan1", PlanTitle: "Rust"},
		{PlanID: "plan2", PlanTitle: "Go"},
	}
	model.SetAllPlanStats(planStats)

	// Switch to plan list
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})

	// Navigate up from first item (should wrap to last)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
	m := updatedModel.(*StatsModel)
	assert.Equal(t, 1, m.planListCursor)

	// Navigate down from last item (should wrap to first)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 0, m.planListCursor)
}

func TestStatsModel_PlanList_SelectPlan(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	planStats := []stats.PlanStats{
		{PlanID: "plan1", PlanTitle: "Rust"},
		{PlanID: "plan2", PlanTitle: "Go"},
	}
	model.SetAllPlanStats(planStats)

	// Switch to plan list
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*StatsModel)

	// Move to second plan
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 1, m.planListCursor)

	// Press Enter to select
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(*StatsModel)

	// Should switch to plan detail view
	assert.Equal(t, viewPlanDetail, m.currentView)
	// Should set selected plan ID
	assert.Equal(t, "plan2", m.selectedPlanID)
}

func TestStatsModel_PlanList_RenderWithPlans(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	planStats := []stats.PlanStats{
		{PlanID: "plan1", PlanTitle: "Rust Async", Progress: 0.5, TotalHours: 10, PlannedHours: 20, Status: "in-progress"},
		{PlanID: "plan2", PlanTitle: "Go Concurrency", Progress: 0.3, TotalHours: 6, PlannedHours: 20, Status: "in-progress"},
	}
	model.SetAllPlanStats(planStats)

	// Switch to plan list
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*StatsModel)

	view := m.View()

	// Should contain plan titles
	assert.Contains(t, view, "Rust Async")
	assert.Contains(t, view, "Go Concurrency")
	// Should show progress
	assert.Contains(t, view, "50%")
	assert.Contains(t, view, "30%")
	// Should show hours
	assert.Contains(t, view, "10.0")
	assert.Contains(t, view, "6.0")
	// Should show help
	assert.Contains(t, view, "Enter")
	assert.Contains(t, view, "Esc")
	// Should show count
	assert.Contains(t, view, "Showing 2 plans")
}

// Test Phase 2.2: Plan Detail View

func TestStatsModel_PlanDetail_Selection(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	now := time.Now()
	planStats := []stats.PlanStats{
		{
			PlanID:          "rust-async",
			PlanTitle:       "Rust Async Programming",
			Progress:        0.65,
			TotalHours:      26.0,
			PlannedHours:    40.0,
			SessionCount:    12,
			CompletedChunks: 13,
			TotalChunks:     20,
			Status:          "in-progress",
			LastSession:     &now,
		},
	}
	model.SetAllPlanStats(planStats)

	// Switch to plan list
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*StatsModel)

	// Press Enter to select plan
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(*StatsModel)

	// Should be in plan detail view
	assert.Equal(t, viewPlanDetail, m.currentView)
	// Should have selected plan ID
	assert.Equal(t, "rust-async", m.selectedPlanID)
	// Should have selected plan stats
	assert.NotNil(t, m.selectedPlan)
	assert.Equal(t, "Rust Async Programming", m.selectedPlan.PlanTitle)
}

func TestStatsModel_PlanDetail_Render(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	now := time.Now()
	planStats := []stats.PlanStats{
		{
			PlanID:          "rust-async",
			PlanTitle:       "Rust Async Programming",
			Progress:        0.65,
			TotalHours:      26.0,
			PlannedHours:    40.0,
			SessionCount:    12,
			CompletedChunks: 13,
			TotalChunks:     20,
			Status:          "in-progress",
			LastSession:     &now,
		},
	}
	model.SetAllPlanStats(planStats)

	// Navigate to plan detail
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(*StatsModel)

	view := m.View()

	// Should show plan title
	assert.Contains(t, view, "Rust Async Programming")
	// Should show status
	assert.Contains(t, view, "In Progress")
	// Should show progress
	assert.Contains(t, view, "65%")
	assert.Contains(t, view, "13 / 20 chunks")
	// Should show hours
	assert.Contains(t, view, "26.0 / 40.0 hours")
	// Should show session count
	assert.Contains(t, view, "12")
	// Should show help
	assert.Contains(t, view, "View Sessions")
	assert.Contains(t, view, "Esc")
}

func TestStatsModel_PlanDetail_EmptyState(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Manually set to plan detail view without selecting a plan
	model.currentView = viewPlanDetail
	model.selectedPlan = nil

	view := model.View()

	// Should show empty state
	assert.Contains(t, view, "No plan selected")
}

func TestStatsModel_PlanDetail_NavigateToSessions(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	planStats := []stats.PlanStats{
		{PlanID: "rust-async", PlanTitle: "Rust Async", Progress: 0.5},
	}
	model.SetAllPlanStats(planStats)

	// Navigate to plan detail
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(*StatsModel)

	// Press 's' to view sessions
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updatedModel.(*StatsModel)

	// Should switch to session history view
	assert.Equal(t, viewSessionHistory, m.currentView)
	// Should maintain selected plan ID for filtering
	assert.Equal(t, "rust-async", m.selectedPlanID)
}

func TestStatsModel_PlanDetail_GoBack(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	planStats := []stats.PlanStats{
		{PlanID: "rust-async", PlanTitle: "Rust Async", Progress: 0.5},
	}
	model.SetAllPlanStats(planStats)

	// Navigate: overview -> plan list -> plan detail
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(*StatsModel)
	assert.Equal(t, viewPlanDetail, m.currentView)

	// Press Esc to go back
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(*StatsModel)

	// Should be back at plan list
	assert.Equal(t, viewPlanList, m.currentView)
}

func TestStatsModel_PlanDetail_ProgressBar(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	planStats := []stats.PlanStats{
		{
			PlanID:          "test-plan",
			PlanTitle:       "Test Plan",
			Progress:        0.75,
			CompletedChunks: 15,
			TotalChunks:     20,
		},
	}
	model.SetAllPlanStats(planStats)

	// Navigate to plan detail
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(*StatsModel)

	view := m.View()

	// Should show progress bar (check for progress bar characters)
	assert.Contains(t, view, "█") // Progress bar filled character
	// Should show 75%
	assert.Contains(t, view, "75%")
}

// Test comprehensive session history and export features (addressing code review gaps)

func TestStatsModel_SessionHistory_WithData(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Create test sessions
	now := time.Now()
	sessions := []*session.Session{
		{
			ID:        "sess1",
			PlanID:    "plan1",
			StartTime: now.Add(-2 * time.Hour),
			EndTime:   &[]time.Time{now.Add(-1 * time.Hour)}[0],
			Duration:  60,
			Notes:     "First session",
		},
		{
			ID:        "sess2",
			PlanID:    "plan2",
			StartTime: now.Add(-4 * time.Hour),
			EndTime:   &[]time.Time{now.Add(-3 * time.Hour)}[0],
			Duration:  60,
			Notes:     "Second session",
		},
	}
	model.SetSessions(sessions)

	// Switch to session history
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m := updatedModel.(*StatsModel)

	view := m.View()

	// Should show session data
	assert.Contains(t, view, "Session History")
	assert.Contains(t, view, "plan1")
	assert.Contains(t, view, "plan2")
	assert.Contains(t, view, "First session")
	assert.Contains(t, view, "Showing 2 sessions")
}

func TestStatsModel_SessionHistory_FilteredByPlan(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Create test sessions for multiple plans
	now := time.Now()
	sessions := []*session.Session{
		{ID: "sess1", PlanID: "rust-async", StartTime: now.Add(-1 * time.Hour), Duration: 60, Notes: "Rust session 1"},
		{ID: "sess2", PlanID: "rust-async", StartTime: now.Add(-2 * time.Hour), Duration: 60, Notes: "Rust session 2"},
		{ID: "sess3", PlanID: "go-concurrency", StartTime: now.Add(-3 * time.Hour), Duration: 60, Notes: "Go session 1"},
	}
	model.SetSessions(sessions)

	// Set selected plan
	model.selectedPlanID = "rust-async"
	model.selectedPlan = &stats.PlanStats{PlanID: "rust-async", PlanTitle: "Rust Async"}

	// Switch to session history
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m := updatedModel.(*StatsModel)

	view := m.View()

	// Should show filtered sessions
	assert.Contains(t, view, "Session History: Rust Async")
	assert.Contains(t, view, "Rust session 1")
	assert.Contains(t, view, "Rust session 2")
	// Should not show Go sessions
	assert.NotContains(t, view, "Go session 1")
	// Should show filtered count
	assert.Contains(t, view, "Showing 2 sessions")
}

func TestStatsModel_SessionHistory_Pagination(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Create 25 sessions (more than maxDisplay of 20)
	now := time.Now()
	sessions := make([]*session.Session, 25)
	for i := 0; i < 25; i++ {
		sessions[i] = &session.Session{
			ID:        fmt.Sprintf("sess%d", i),
			PlanID:    "test-plan",
			StartTime: now.Add(-time.Duration(i) * time.Hour),
			Duration:  60,
			Notes:     fmt.Sprintf("Session %d", i),
		}
	}
	model.SetSessions(sessions)

	// Switch to session history
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	// Should only display 20 sessions (pagination)
	// Get the filtered and paginated sessions
	filteredSessions := model.filterSessionsByPlan()
	displaySessions, _ := model.paginateSessions(filteredSessions, 20)

	// Should paginate to 20
	assert.Equal(t, 20, len(displaySessions))

	// Total count should still show all 25
	view := model.View()
	assert.Contains(t, view, "Showing 25 sessions")
}

func TestStatsModel_SessionHistory_CursorNavigation(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Create test sessions
	now := time.Now()
	sessions := []*session.Session{
		{ID: "sess1", PlanID: "plan1", StartTime: now, Duration: 60},
		{ID: "sess2", PlanID: "plan1", StartTime: now.Add(-1 * time.Hour), Duration: 60},
		{ID: "sess3", PlanID: "plan1", StartTime: now.Add(-2 * time.Hour), Duration: 60},
	}
	model.SetSessions(sessions)

	// Switch to session history
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m := updatedModel.(*StatsModel)
	assert.Equal(t, 0, m.sessionHistoryCursor)

	// Navigate down with 'j'
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 1, m.sessionHistoryCursor)

	// Navigate down with arrow key
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 2, m.sessionHistoryCursor)

	// Navigate up with 'k'
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 1, m.sessionHistoryCursor)
}

func TestStatsModel_SessionHistory_CursorWrapAround(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Create test sessions
	now := time.Now()
	sessions := []*session.Session{
		{ID: "sess1", PlanID: "plan1", StartTime: now, Duration: 60},
		{ID: "sess2", PlanID: "plan1", StartTime: now.Add(-1 * time.Hour), Duration: 60},
	}
	model.SetSessions(sessions)

	// Switch to session history
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	// Navigate up from first (should wrap to last)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
	m := updatedModel.(*StatsModel)
	assert.Equal(t, 1, m.sessionHistoryCursor)

	// Navigate down from last (should wrap to first)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 0, m.sessionHistoryCursor)
}

func TestStatsModel_SessionHistory_CursorResetOnViewSwitch(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Create test sessions
	now := time.Now()
	sessions := []*session.Session{
		{ID: "sess1", PlanID: "plan1", StartTime: now, Duration: 60},
		{ID: "sess2", PlanID: "plan1", StartTime: now, Duration: 60},
		{ID: "sess3", PlanID: "plan1", StartTime: now, Duration: 60},
	}
	model.SetSessions(sessions)

	// Switch to session history and navigate
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, model.sessionHistoryCursor)

	// Go back and return to session history
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m := updatedModel.(*StatsModel)

	// Cursor should be reset
	assert.Equal(t, 0, m.sessionHistoryCursor)
}

func TestStatsModel_SessionHistory_FilteredCursorBounds(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Create sessions for multiple plans
	now := time.Now()
	sessions := []*session.Session{
		{ID: "sess1", PlanID: "plan1", StartTime: now, Duration: 60},
		{ID: "sess2", PlanID: "plan1", StartTime: now, Duration: 60},
		{ID: "sess3", PlanID: "plan2", StartTime: now, Duration: 60},
		{ID: "sess4", PlanID: "plan2", StartTime: now, Duration: 60},
		{ID: "sess5", PlanID: "plan2", StartTime: now, Duration: 60},
	}
	model.SetSessions(sessions)

	// Set filter to plan1 (2 sessions)
	model.selectedPlanID = "plan1"
	model.selectedPlan = &stats.PlanStats{PlanID: "plan1", PlanTitle: "Plan 1"}

	// Switch to session history
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	// Navigate to last filtered session
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, model.sessionHistoryCursor)

	// Try to go beyond filtered list (should wrap to first)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updatedModel.(*StatsModel)
	assert.Equal(t, 0, m.sessionHistoryCursor)

	// Navigate up (should wrap to last filtered session, which is index 1)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 1, m.sessionHistoryCursor)
}

func TestStatsModel_Export_NavigateOptions(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Switch to export
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m := updatedModel.(*StatsModel)
	assert.Equal(t, 0, m.exportMenuCursor)

	// Navigate down
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 1, m.exportMenuCursor)

	// Navigate down again (should wrap to first)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 0, m.exportMenuCursor)

	// Navigate up (should wrap to last)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 1, m.exportMenuCursor)
}

func TestStatsModel_Export_SelectOption(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Switch to export
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m := updatedModel.(*StatsModel)

	// Press Enter to select first option (summary)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, "summary", m.exportType)

	// Should go back to previous view
	assert.Equal(t, viewOverview, m.currentView)
}

func TestStatsModel_Export_SelectFullReport(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Switch to export
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m := updatedModel.(*StatsModel)

	// Navigate to second option (full report)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, 1, m.exportMenuCursor)

	// Press Enter to select
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(*StatsModel)
	assert.Equal(t, "full", m.exportType)

	// Should go back to previous view
	assert.Equal(t, viewOverview, m.currentView)
}
