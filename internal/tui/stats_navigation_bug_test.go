// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pezware/samedi.dev/internal/stats"
	"github.com/stretchr/testify/assert"
)

// TestStatsModel_NavigationBugFix_PressingSameKeyTwice verifies that pressing
// p/s/e while already on those views doesn't push duplicates onto viewHistory.
func TestStatsModel_NavigationBugFix_PressingSameKeyTwice(t *testing.T) {
	t.Run("pressing 'p' while on plan list", func(t *testing.T) {
		totalStats := &stats.TotalStats{}
		model := newTestStatsModuleWithTotals(totalStats)

		// Switch to plan list
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
		m1 := updatedModel.(*StatsModel)
		assert.Equal(t, viewPlanList, m1.currentView)
		assert.Equal(t, 1, len(m1.viewHistory)) // overview in history

		// Press 'p' again while already on plan list
		updatedModel, _ = m1.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
		m2 := updatedModel.(*StatsModel)

		// Should stay on plan list view
		assert.Equal(t, viewPlanList, m2.currentView)
		// History should NOT grow (bug fix)
		assert.Equal(t, 1, len(m2.viewHistory), "History should not grow when pressing 'p' while already on plan list")
	})

	t.Run("pressing 's' while on session history", func(t *testing.T) {
		totalStats := &stats.TotalStats{}
		model := newTestStatsModuleWithTotals(totalStats)

		// Switch to session history
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
		m1 := updatedModel.(*StatsModel)
		assert.Equal(t, viewSessionHistory, m1.currentView)
		assert.Equal(t, 1, len(m1.viewHistory))

		// Press 's' again while already on session history
		updatedModel, _ = m1.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
		m2 := updatedModel.(*StatsModel)

		// Should stay on session history view
		assert.Equal(t, viewSessionHistory, m2.currentView)
		// History should NOT grow (bug fix)
		assert.Equal(t, 1, len(m2.viewHistory), "History should not grow when pressing 's' while already on session history")
	})

	t.Run("pressing 'e' while on export dialog", func(t *testing.T) {
		totalStats := &stats.TotalStats{}
		model := newTestStatsModuleWithTotals(totalStats)

		// Switch to export
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
		m1 := updatedModel.(*StatsModel)
		assert.Equal(t, viewExport, m1.currentView)
		assert.Equal(t, 1, len(m1.viewHistory))

		// Press 'e' again while already on export
		updatedModel, _ = m1.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
		m2 := updatedModel.(*StatsModel)

		// Should stay on export view
		assert.Equal(t, viewExport, m2.currentView)
		// History should NOT grow (bug fix)
		assert.Equal(t, 1, len(m2.viewHistory), "History should not grow when pressing 'e' while already on export")
	})
}
