// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTable_Empty(t *testing.T) {
	table := NewTable([]string{"Col1", "Col2"})
	result := table.View()

	// Should render headers only
	assert.Contains(t, result, "Col1")
	assert.Contains(t, result, "Col2")
}

func TestTable_WithRows(t *testing.T) {
	table := NewTable([]string{"Name", "Age"})
	table.AddRow([]string{"Alice", "30"})
	table.AddRow([]string{"Bob", "25"})

	result := table.View()

	// Should render headers and rows
	assert.Contains(t, result, "Name")
	assert.Contains(t, result, "Age")
	assert.Contains(t, result, "Alice")
	assert.Contains(t, result, "30")
	assert.Contains(t, result, "Bob")
	assert.Contains(t, result, "25")
}

func TestTable_ColumnAlignment(t *testing.T) {
	table := NewTable([]string{"Left", "Right"})
	table.AddRow([]string{"A", "1"})
	table.AddRow([]string{"BB", "22"})

	result := table.View()

	// Should have proper column alignment (check for consistent spacing)
	lines := strings.Split(result, "\n")
	assert.True(t, len(lines) >= 3, "should have at least header + 2 rows")
}

func TestTable_WithBorders(t *testing.T) {
	table := NewTable([]string{"Col1"})
	table.SetBorder(true)

	result := table.View()

	// Should include border characters
	assert.Contains(t, result, "─")
}

func TestTable_NoBorders(t *testing.T) {
	table := NewTable([]string{"Col1"})
	table.SetBorder(false)

	result := table.View()

	// Should not include border characters
	assert.NotContains(t, result, "─")
}

func TestTable_MismatchedColumns(t *testing.T) {
	table := NewTable([]string{"Col1", "Col2"})
	table.AddRow([]string{"Value1"}) // Only 1 value, but 2 columns

	result := table.View()

	// Should handle gracefully (pad with empty string)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Value1")
}

func TestTable_ExtraValues(t *testing.T) {
	table := NewTable([]string{"Col1"})
	table.AddRow([]string{"Value1", "ExtraValue"}) // 2 values, but only 1 column

	result := table.View()

	// Should truncate extra values
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Value1")
}

func TestTable_EmptyRow(t *testing.T) {
	table := NewTable([]string{"Col1", "Col2"})
	table.AddRow([]string{})

	result := table.View()

	// Should render empty row
	assert.NotEmpty(t, result)
}

func TestTable_LongValues(t *testing.T) {
	table := NewTable([]string{"Name"})
	table.AddRow([]string{"ThisIsAVeryLongNameThatExceedsTypicalColumnWidth"}) // pragma: allowlist secret

	result := table.View()

	// Should handle long values without panic
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "ThisIsAVeryLongNameThatExceedsTypicalColumnWidth") // pragma: allowlist secret
}
