// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ProgressBar renders a styled progress bar with color-coding.
type ProgressBar struct {
	progress float64
	width    int
}

// NewProgressBar creates a new progress bar with the given progress (0.0-1.0) and width.
func NewProgressBar(progress float64, width int) *ProgressBar {
	// Clamp progress to [0.0, 1.0]
	if progress < 0.0 {
		progress = 0.0
	}
	if progress > 1.0 {
		progress = 1.0
	}

	return &ProgressBar{
		progress: progress,
		width:    width,
	}
}

// View renders the progress bar as a string with Lipgloss styling.
func (p *ProgressBar) View() string {
	// Calculate filled width
	filled := int(p.progress * float64(p.width))
	if filled > p.width {
		filled = p.width
	}
	if filled < 0 {
		filled = 0
	}

	// Build the bar string
	var bar strings.Builder
	for i := 0; i < p.width; i++ {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}

	// Choose color based on progress
	var barColor lipgloss.Color
	switch {
	case p.progress < 0.33:
		barColor = lipgloss.Color("9") // Red
	case p.progress < 0.66:
		barColor = lipgloss.Color("11") // Yellow
	default:
		barColor = lipgloss.Color("10") // Green
	}

	// Style the bar
	barStyle := lipgloss.NewStyle().Foreground(barColor)

	// Format percentage
	percentage := int(p.progress * 100)

	// Return styled string
	return fmt.Sprintf("[%s] %d%%",
		barStyle.Render(bar.String()),
		percentage,
	)
}
