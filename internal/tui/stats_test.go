// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

	// Should render session history stub
	assert.Contains(t, view, "Session History View")
	assert.Contains(t, view, "Coming soon")
}

func TestStatsModel_View_RendersExportStub(t *testing.T) {
	totalStats := &stats.TotalStats{}
	model := NewStatsModel(totalStats, nil)

	// Switch to export
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m := updatedModel.(*StatsModel)

	view := m.View()

	// Should render export dialog stub
	assert.Contains(t, view, "Export Dialog")
	assert.Contains(t, view, "Coming soon")
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
