// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package app

import tea "github.com/charmbracelet/bubbletea"

// Module defines the interface that all TUI modules must implement.
// Modules are standard Bubble Tea models with additional metadata used
// by the shared application shell for navigation and presentation.
type Module interface {
	tea.Model

	// ID returns a stable identifier for the module (e.g. "plan").
	ID() string

	// Title is displayed in the navigation bar (e.g. "Plans").
	Title() string

	// Shortcuts returns module-specific shortcut hints rendered in the footer.
	Shortcuts() []Shortcut
}

// Shortcut describes a keyboard shortcut exposed by a module or the shell.
type Shortcut struct {
	Key         string
	Description string
}

// ModuleActivatedMsg is sent to a module when it becomes the active module
// in the shared shell. Modules can use FirstActivation to lazy-load data.
type ModuleActivatedMsg struct {
	ID              string
	FirstActivation bool
}

// StatusMsg is emitted by modules to display notifications in the shell footer.
type StatusMsg struct {
	Message string
	IsError bool
}

// BroadcastMsg notifies all modules about cross-cutting events (e.g. plan updates).
// Modules can choose to react by inspecting the Topic.
type BroadcastMsg struct {
	Topic   string
	Payload interface{}
}

// Predefined broadcast topics.
const (
	TopicPlansChanged = "plan:changed"
)
