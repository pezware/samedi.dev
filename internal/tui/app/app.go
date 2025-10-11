// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// App is the shared Bubble Tea shell that coordinates registered modules.
type App struct {
	modules     map[string]Module
	order       []string
	activeID    string
	initialized map[string]bool
	activated   map[string]bool

	width  int
	height int

	status *StatusMsg
}

var (
	navStyle         = lipgloss.NewStyle().Bold(true)
	activeNavStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	statusStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStatusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("209"))
	borderStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// New constructs a new shell with the provided modules. The first module becomes active.
func New(modules []Module) (*App, error) {
	if len(modules) == 0 {
		return nil, fmt.Errorf("at least one module is required")
	}

	mods := make(map[string]Module, len(modules))
	order := make([]string, len(modules))
	for i, module := range modules {
		id := module.ID()
		if id == "" {
			return nil, fmt.Errorf("module at index %d has empty ID", i)
		}
		if _, exists := mods[id]; exists {
			return nil, fmt.Errorf("duplicate module ID: %s", id)
		}
		mods[id] = module
		order[i] = id
	}

	return &App{
		modules:     mods,
		order:       order,
		activeID:    order[0],
		initialized: map[string]bool{},
		activated:   map[string]bool{},
	}, nil
}

// Init initializes the currently active module.
func (a *App) Init() tea.Cmd {
	mod := a.activeModule()
	if mod == nil {
		return nil
	}

	a.initialized[a.activeID] = true
	initialCmd := mod.Init()

	activateCmd := func() tea.Msg {
		return ModuleActivatedMsg{
			ID:              a.activeID,
			FirstActivation: true,
		}
	}

	return tea.Batch(initialCmd, activateCmd)
}

// Update processes messages, handling global navigation and delegating to the active module.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if cmd, handled := a.handleKeyMsg(m); handled {
			return a, cmd
		}
	case ModuleActivatedMsg:
		if m.ID != a.activeID {
			return a, nil
		}
	case StatusMsg:
		a.status = &m
		return a, nil
	case BroadcastMsg:
		cmd := a.handleBroadcast(m)
		// Trivial wrapping to satisfy gocritic evalOrder rule
		return a, tea.Batch(cmd)
	case tea.WindowSizeMsg:
		a.width = m.Width
		a.height = m.Height
		return a, nil
	}

	mod := a.activeModule()
	if mod == nil {
		return a, nil
	}

	updated, cmd := mod.Update(msg)
	if updatedModule, ok := updated.(Module); ok {
		a.modules[a.activeID] = updatedModule
	}
	return a, cmd
}

func (a *App) handleKeyMsg(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch {
	case msg.Type == tea.KeyCtrlC || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'q'):
		return tea.Quit, true
	case msg.Type == tea.KeyTab:
		a.rotateModule(1)
		return a.activateCurrentModule(false), true
	case msg.Type == tea.KeyShiftTab:
		a.rotateModule(-1)
		return a.activateCurrentModule(false), true
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1:
		if idx, ok := a.moduleIndexFromKey(msg.Runes[0]); ok {
			if a.setActiveIndex(idx) {
				return a.activateCurrentModule(false), true
			}
			return nil, true
		}
	}
	return nil, false
}

func (a *App) handleBroadcast(msg BroadcastMsg) tea.Cmd {
	var cmds []tea.Cmd
	for id, module := range a.modules {
		if id == a.activeID {
			continue
		}
		updated, cmd := module.Update(msg)
		if updatedModule, ok := updated.(Module); ok {
			a.modules[id] = updatedModule
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// View renders the navigation bar, active module view, and footer.
func (a *App) View() string {
	var b strings.Builder

	b.WriteString(a.renderNavigation())
	b.WriteString("\n")

	if mod := a.activeModule(); mod != nil {
		b.WriteString(mod.View())
	} else {
		b.WriteString("No module available.")
	}

	if footer := a.renderFooter(); footer != "" {
		b.WriteString("\n")
		b.WriteString(footer)
	}

	return b.String()
}

func (a *App) activeModule() Module {
	return a.modules[a.activeID]
}

func (a *App) renderNavigation() string {
	var items []string
	for idx, id := range a.order {
		mod := a.modules[id]
		if mod == nil {
			continue
		}
		label := fmt.Sprintf("%d·%s", idx+1, mod.Title())
		if id == a.activeID {
			items = append(items, activeNavStyle.Render(label))
		} else {
			items = append(items, navStyle.Render(label))
		}
	}

	return borderStyle.Render(strings.Join(items, " │ "))
}

func (a *App) renderFooter() string {
	module := a.activeModule()
	additional := 0
	if module != nil {
		additional = len(module.Shortcuts())
	}
	parts := make([]string, 0, len(globalShortcuts)+additional)

	global := globalShortcuts

	for _, sc := range global {
		parts = append(parts, fmt.Sprintf("%s %s", navStyle.Render(sc.Key), sc.Description))
	}

	if module != nil {
		for _, sc := range module.Shortcuts() {
			parts = append(parts, fmt.Sprintf("%s %s", navStyle.Render(sc.Key), sc.Description))
		}
	}

	status := ""
	if a.status != nil {
		style := statusStyle
		if a.status.IsError {
			style = errorStatusStyle
		}
		status = style.Render(a.status.Message)
	}

	footer := strings.Join(parts, "  ")
	if status != "" {
		footer = fmt.Sprintf("%s\n%s", footer, status)
	}

	return footer
}

func (a *App) moduleIndexFromKey(r rune) (int, bool) {
	if r >= '1' && r <= '9' {
		idx := int(r - '1')
		if idx >= 0 && idx < len(a.order) {
			return idx, true
		}
	}
	return -1, false
}

func (a *App) setActiveIndex(idx int) bool {
	if idx < 0 || idx >= len(a.order) {
		return false
	}
	id := a.order[idx]
	if id == a.activeID {
		return false
	}
	a.activeID = id
	return true
}

func (a *App) rotateModule(direction int) {
	if len(a.order) == 0 {
		return
	}
	currentIdx := 0
	for i, id := range a.order {
		if id == a.activeID {
			currentIdx = i
			break
		}
	}
	next := (currentIdx + direction + len(a.order)) % len(a.order)
	a.activeID = a.order[next]
}

func (a *App) activateCurrentModule(initial bool) tea.Cmd {
	mod := a.activeModule()
	if mod == nil {
		return nil
	}

	var cmds []tea.Cmd

	if !a.initialized[a.activeID] {
		if initCmd := mod.Init(); initCmd != nil {
			cmds = append(cmds, initCmd)
		}
		a.initialized[a.activeID] = true
	}

	activationMsg := ModuleActivatedMsg{
		ID:              a.activeID,
		FirstActivation: initial || !a.activated[a.activeID],
	}

	updated, cmd := mod.Update(activationMsg)
	if updatedModule, ok := updated.(Module); ok {
		a.modules[a.activeID] = updatedModule
	}

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	a.activated[a.activeID] = true

	return tea.Batch(cmds...)
}

var globalShortcuts = []Shortcut{
	{Key: "Tab/Shift+Tab", Description: "switch module"},
	{Key: "1…9", Description: "jump to module"},
	{Key: "q", Description: "quit"},
}
