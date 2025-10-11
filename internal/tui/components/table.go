// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Table renders a styled table with headers and rows.
type Table struct {
	headers         []string
	rows            [][]string
	border          bool
	highlightedRows map[int]bool
}

// NewTable creates a new table with the given headers.
func NewTable(headers []string) *Table {
	return &Table{
		headers:         headers,
		rows:            make([][]string, 0),
		border:          false,
		highlightedRows: map[int]bool{},
	}
}

// AddRow adds a new row to the table.
// If the row has fewer values than headers, it will be padded with empty strings.
// If the row has more values than headers, extra values will be truncated.
func (t *Table) AddRow(values []string) {
	// Ensure row matches column count
	row := make([]string, len(t.headers))
	for i := range t.headers {
		if i < len(values) {
			row[i] = values[i]
		} else {
			row[i] = ""
		}
	}
	t.rows = append(t.rows, row)
}

// AddHighlightedRow adds a row and marks it as highlighted.
func (t *Table) AddHighlightedRow(values []string) {
	initialCount := len(t.rows)
	t.AddRow(values)
	if t.highlightedRows == nil {
		t.highlightedRows = map[int]bool{}
	}
	t.highlightedRows[initialCount] = true
}

// SetBorder enables or disables table borders.
func (t *Table) SetBorder(enabled bool) {
	t.border = enabled
}

// View renders the table as a string with Lipgloss styling.
func (t *Table) View() string {
	if len(t.headers) == 0 {
		return ""
	}

	// Calculate column widths
	colWidths := make([]int, len(t.headers))
	for i, header := range t.headers {
		colWidths[i] = len(header)
	}
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")) // Bright blue

	var result strings.Builder

	// Render top border
	if t.border {
		result.WriteString(t.renderBorder(colWidths, "┌", "─", "┬", "┐"))
		result.WriteString("\n")
	}

	// Render headers
	result.WriteString(t.renderRow(t.headers, colWidths, &headerStyle))
	result.WriteString("\n")

	// Render separator (only if border is enabled)
	if t.border {
		result.WriteString(t.renderBorder(colWidths, "├", "─", "┼", "┤"))
		result.WriteString("\n")
	}

	// Render rows
	normalStyle := lipgloss.NewStyle()
	highlightStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213"))
	for i, row := range t.rows {
		style := normalStyle
		if t.highlightedRows[i] {
			style = highlightStyle
		}
		result.WriteString(t.renderRow(row, colWidths, &style))
		result.WriteString("\n")
	}

	// Render bottom border
	if t.border {
		result.WriteString(t.renderBorder(colWidths, "└", "─", "┴", "┘"))
		result.WriteString("\n")
	}

	return strings.TrimSuffix(result.String(), "\n")
}

// renderRow renders a single row with proper alignment and styling.
func (t *Table) renderRow(cells []string, widths []int, style *lipgloss.Style) string {
	var row strings.Builder

	if t.border {
		row.WriteString("│ ")
	}

	for i, cell := range cells {
		if i > 0 {
			if t.border {
				row.WriteString(" │ ")
			} else {
				row.WriteString("  ")
			}
		}

		// Pad cell to column width
		padded := cell + strings.Repeat(" ", widths[i]-len(cell))
		row.WriteString(style.Render(padded))
	}

	if t.border {
		row.WriteString(" │")
	}

	return row.String()
}

// renderBorder renders a border line with the given characters.
func (t *Table) renderBorder(widths []int, left, fill, sep, right string) string {
	var border strings.Builder

	border.WriteString(left)
	for i, width := range widths {
		if i > 0 {
			border.WriteString(sep)
		}
		border.WriteString(strings.Repeat(fill, width+2))
	}
	border.WriteString(right)

	return border.String()
}
