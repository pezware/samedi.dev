// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockModule is a test implementation of the Module interface
type MockModule struct {
	id        string
	title     string
	shortcuts []Shortcut
	initCalls int
	viewCalls int
}

func NewMockModule(id, title string) *MockModule {
	return &MockModule{
		id:    id,
		title: title,
		shortcuts: []Shortcut{
			{Key: "enter", Description: "select"},
		},
	}
}

func (m *MockModule) Init() tea.Cmd {
	m.initCalls++
	return nil
}

func (m *MockModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *MockModule) View() string {
	m.viewCalls++
	return m.title + " view"
}

func (m *MockModule) ID() string {
	return m.id
}

func (m *MockModule) Title() string {
	return m.title
}

func (m *MockModule) Shortcuts() []Shortcut {
	return m.shortcuts
}

// Tests for App.New

func TestNew_WithValidModules_Success(t *testing.T) {
	modules := []Module{
		NewMockModule("plans", "Plans"),
		NewMockModule("stats", "Stats"),
	}

	app, err := New(modules)

	require.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, 2, len(app.modules))
	assert.Equal(t, 2, len(app.order))
	assert.Equal(t, "plans", app.activeID)
}

func TestNew_EmptyModules_ReturnsError(t *testing.T) {
	app, err := New([]Module{})

	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "at least one module is required")
}

func TestNew_ModuleWithEmptyID_ReturnsError(t *testing.T) {
	badModule := NewMockModule("", "Bad Module")

	app, err := New([]Module{badModule})

	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "empty ID")
}

func TestNew_DuplicateModuleIDs_ReturnsError(t *testing.T) {
	modules := []Module{
		NewMockModule("duplicate", "First"),
		NewMockModule("duplicate", "Second"),
	}

	app, err := New(modules)

	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "duplicate module ID")
}

func TestNew_FirstModuleBecomesActive(t *testing.T) {
	modules := []Module{
		NewMockModule("first", "First"),
		NewMockModule("second", "Second"),
	}

	app, err := New(modules)

	require.NoError(t, err)
	assert.Equal(t, "first", app.activeID)
}

// Tests for App.Init

func TestInit_InitializesActiveModule(t *testing.T) {
	mockModule := NewMockModule("test", "Test")
	modules := []Module{mockModule}

	app, err := New(modules)
	require.NoError(t, err)

	cmd := app.Init()

	assert.NotNil(t, cmd)
	assert.True(t, app.initialized["test"])
	assert.Equal(t, 1, mockModule.initCalls)
}

// Tests for App.Update with KeyMsg

func TestUpdate_QuitKey_ReturnsQuitCmd(t *testing.T) {
	modules := []Module{NewMockModule("test", "Test")}
	app, _ := New(modules)

	// Test 'q' key
	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	assert.NotNil(t, model)
	assert.NotNil(t, cmd)
}

func TestUpdate_CtrlC_ReturnsQuitCmd(t *testing.T) {
	modules := []Module{NewMockModule("test", "Test")}
	app, _ := New(modules)

	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	assert.NotNil(t, model)
	assert.NotNil(t, cmd)
}

func TestUpdate_TabKey_RotatesModuleForward(t *testing.T) {
	modules := []Module{
		NewMockModule("first", "First"),
		NewMockModule("second", "Second"),
	}
	app, _ := New(modules)

	assert.Equal(t, "first", app.activeID)

	app.Update(tea.KeyMsg{Type: tea.KeyTab})

	assert.Equal(t, "second", app.activeID)
}

func TestUpdate_ShiftTabKey_RotatesModuleBackward(t *testing.T) {
	modules := []Module{
		NewMockModule("first", "First"),
		NewMockModule("second", "Second"),
	}
	app, _ := New(modules)
	app.activeID = "second"

	app.Update(tea.KeyMsg{Type: tea.KeyShiftTab})

	assert.Equal(t, "first", app.activeID)
}

func TestUpdate_NumberKeys_JumpsToModule(t *testing.T) {
	modules := []Module{
		NewMockModule("first", "First"),
		NewMockModule("second", "Second"),
		NewMockModule("third", "Third"),
	}
	app, _ := New(modules)

	// Jump to module 3 (index 2)
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})

	assert.Equal(t, "third", app.activeID)
}

func TestUpdate_NumberKey_OutOfRange_NoChange(t *testing.T) {
	modules := []Module{
		NewMockModule("first", "First"),
	}
	app, _ := New(modules)

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})

	assert.Equal(t, "first", app.activeID)
}

// Tests for App.Update with StatusMsg

func TestUpdate_StatusMsg_SetsStatus(t *testing.T) {
	modules := []Module{NewMockModule("test", "Test")}
	app, _ := New(modules)

	statusMsg := StatusMsg{Message: "Test status", IsError: false}
	app.Update(statusMsg)

	assert.NotNil(t, app.status)
	assert.Equal(t, "Test status", app.status.Message)
	assert.False(t, app.status.IsError)
}

func TestUpdate_ErrorStatusMsg_SetsErrorStatus(t *testing.T) {
	modules := []Module{NewMockModule("test", "Test")}
	app, _ := New(modules)

	statusMsg := StatusMsg{Message: "Error occurred", IsError: true}
	app.Update(statusMsg)

	assert.NotNil(t, app.status)
	assert.True(t, app.status.IsError)
}

// Tests for App.Update with WindowSizeMsg

func TestUpdate_WindowSizeMsg_UpdatesDimensions(t *testing.T) {
	modules := []Module{NewMockModule("test", "Test")}
	app, _ := New(modules)

	app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	assert.Equal(t, 120, app.width)
	assert.Equal(t, 40, app.height)
}

// Tests for App.View

func TestView_RendersNavigation(t *testing.T) {
	modules := []Module{
		NewMockModule("plans", "Plans"),
		NewMockModule("stats", "Stats"),
	}
	app, _ := New(modules)

	view := app.View()

	assert.Contains(t, view, "Plans")
	assert.Contains(t, view, "Stats")
	assert.Contains(t, view, "1·")
	assert.Contains(t, view, "2·")
}

func TestView_RendersActiveModuleView(t *testing.T) {
	mockModule := NewMockModule("test", "Test")
	modules := []Module{mockModule}
	app, _ := New(modules)

	view := app.View()

	assert.Contains(t, view, "Test view")
	assert.Equal(t, 1, mockModule.viewCalls)
}

func TestView_RendersFooterWithShortcuts(t *testing.T) {
	modules := []Module{NewMockModule("test", "Test")}
	app, _ := New(modules)

	view := app.View()

	// Should contain global shortcuts
	assert.Contains(t, view, "Tab")
	assert.Contains(t, view, "quit")

	// Should contain module shortcuts
	assert.Contains(t, view, "enter")
}

func TestView_WithStatus_RendersStatus(t *testing.T) {
	modules := []Module{NewMockModule("test", "Test")}
	app, _ := New(modules)
	app.status = &StatusMsg{Message: "Loading...", IsError: false}

	view := app.View()

	assert.Contains(t, view, "Loading...")
}

// Tests for module rotation

func TestRotateModule_Forward_WrapsAround(t *testing.T) {
	modules := []Module{
		NewMockModule("first", "First"),
		NewMockModule("second", "Second"),
		NewMockModule("third", "Third"),
	}
	app, _ := New(modules)
	app.activeID = "third"

	app.rotateModule(1)

	assert.Equal(t, "first", app.activeID)
}

func TestRotateModule_Backward_WrapsAround(t *testing.T) {
	modules := []Module{
		NewMockModule("first", "First"),
		NewMockModule("second", "Second"),
	}
	app, _ := New(modules)
	app.activeID = "first"

	app.rotateModule(-1)

	assert.Equal(t, "second", app.activeID)
}

func TestRotateModule_EmptyOrder_NoChange(t *testing.T) {
	modules := []Module{NewMockModule("test", "Test")}
	app, _ := New(modules)
	app.order = []string{}

	app.rotateModule(1)

	// Should not crash
	assert.NotNil(t, app)
}

// Tests for module index

func TestModuleIndexFromKey_ValidKeys(t *testing.T) {
	modules := []Module{
		NewMockModule("first", "First"),
		NewMockModule("second", "Second"),
	}
	app, _ := New(modules)

	tests := []struct {
		key      rune
		expected int
		valid    bool
	}{
		{'1', 0, true},
		{'2', 1, true},
		{'9', -1, false}, // Out of range
		{'0', -1, false}, // Invalid key
		{'a', -1, false}, // Invalid key
	}

	for _, tt := range tests {
		idx, ok := app.moduleIndexFromKey(tt.key)
		assert.Equal(t, tt.valid, ok, "key %c should be valid=%v", tt.key, tt.valid)
		if tt.valid {
			assert.Equal(t, tt.expected, idx, "key %c should map to index %d", tt.key, tt.expected)
		}
	}
}

// Tests for setActiveIndex

func TestSetActiveIndex_ValidIndex_SetsActive(t *testing.T) {
	modules := []Module{
		NewMockModule("first", "First"),
		NewMockModule("second", "Second"),
	}
	app, _ := New(modules)

	changed := app.setActiveIndex(1)

	assert.True(t, changed)
	assert.Equal(t, "second", app.activeID)
}

func TestSetActiveIndex_SameIndex_ReturnsFalse(t *testing.T) {
	modules := []Module{
		NewMockModule("first", "First"),
	}
	app, _ := New(modules)

	changed := app.setActiveIndex(0)

	assert.False(t, changed)
	assert.Equal(t, "first", app.activeID)
}

func TestSetActiveIndex_InvalidIndex_ReturnsFalse(t *testing.T) {
	modules := []Module{
		NewMockModule("first", "First"),
	}
	app, _ := New(modules)

	changed := app.setActiveIndex(5)

	assert.False(t, changed)
	assert.Equal(t, "first", app.activeID)
}

func TestSetActiveIndex_NegativeIndex_ReturnsFalse(t *testing.T) {
	modules := []Module{
		NewMockModule("first", "First"),
	}
	app, _ := New(modules)

	changed := app.setActiveIndex(-1)

	assert.False(t, changed)
	assert.Equal(t, "first", app.activeID)
}

// Tests for activeModule

func TestActiveModule_ReturnsCurrentModule(t *testing.T) {
	module1 := NewMockModule("first", "First")
	module2 := NewMockModule("second", "Second")
	modules := []Module{module1, module2}
	app, _ := New(modules)

	active := app.activeModule()

	assert.Equal(t, module1, active)

	app.activeID = "second"
	active = app.activeModule()

	assert.Equal(t, module2, active)
}
