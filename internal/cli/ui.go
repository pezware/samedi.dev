// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pezware/samedi.dev/internal/stats"
	"github.com/pezware/samedi.dev/internal/tui"
	"github.com/pezware/samedi.dev/internal/tui/app"
	"github.com/spf13/cobra"
)

func uiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Launch the interactive Samedi dashboard",
		Long: `Launch the Bubble Tea dashboard for Samedi.

Modules:
  - Plans: browse plans, inspect chunks, create or edit plans, toggle chunk status.
  - Stats: review streaks, drill into plan metrics, inspect session history, export summaries.

Navigation:
  - Tab / Shift+Tab cycle modules, 1–9 jump directly, q or Ctrl+C exits.
  - Plans shortcuts: Enter view plan, n new plan, space toggle chunk, e edit metadata, d delete.
  - Stats shortcuts: p plan list, s session history, e export dialog.

Tip: open the stats module on its own with 'samedi stats --tui'.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			planService, err := getPlanService(cmd, "")
			if err != nil {
				return fmt.Errorf("failed to initialize plan service: %w", err)
			}

			sessionService, err := getSessionService(cmd)
			if err != nil {
				return fmt.Errorf("failed to initialize session service: %w", err)
			}

			statsService := stats.NewService(planService, sessionService)

			modules := []app.Module{
				tui.NewPlanModule(planService),
				tui.NewStatsModule(statsService, sessionService, stats.NewTimeRangeAll()),
			}

			shell, err := app.New(modules)
			if err != nil {
				return fmt.Errorf("failed to initialize TUI: %w", err)
			}

			program := tea.NewProgram(shell)
			if _, err := program.Run(); err != nil {
				return fmt.Errorf("failed to run TUI: %w", err)
			}

			return nil
		},
	}
}
