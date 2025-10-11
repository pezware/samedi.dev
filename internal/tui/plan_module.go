// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/storage"
	"github.com/pezware/samedi.dev/internal/tui/app"
	"github.com/pezware/samedi.dev/internal/tui/components"
)

type planModuleState string

const (
	statePlanList    planModuleState = "list"
	statePlanDetail  planModuleState = "detail"
	statePlanEdit    planModuleState = "edit"
	statePlanCreate  planModuleState = "create"
	statePlanConfirm planModuleState = "confirm"
)

// PlanModule provides CRUD operations for learning plans.
type PlanModule struct {
	service *plan.Service

	state planModuleState

	plans      []*storage.PlanRecord
	listCursor int

	detailPlan  *plan.Plan
	chunkCursor int

	form       *planForm
	confirm    *confirmDialog
	loading    bool
	loadErr    error
	dataLoaded bool
}

type planFormMode string

const (
	formModeEdit   planFormMode = "edit"
	formModeCreate planFormMode = "create"
)

type planForm struct {
	mode          planFormMode
	inputs        []*inputField
	focusIndex    int
	targetPlanID  string
	validationErr error
}

type confirmAction string

const (
	confirmDelete confirmAction = "delete"
)

type confirmDialog struct {
	action  confirmAction
	message string
}

type plansLoadedMsg struct {
	records []*storage.PlanRecord
	err     error
}

type planLoadedMsg struct {
	plan *plan.Plan
	err  error
}

type planSavedMsg struct {
	plan *plan.Plan
	err  error
}

type planDeletedMsg struct {
	planID string
	err    error
}

type chunkStatusUpdatedMsg struct {
	planID  string
	chunkID string
	status  plan.Status
	err     error
}

type planCreatedMsg struct {
	planID string
	title  string
	err    error
}

// NewPlanModule returns a plan management module.
func NewPlanModule(service *plan.Service) *PlanModule {
	return &PlanModule{
		service: service,
		state:   statePlanList,
	}
}

// ID satisfies app.Module.
func (m *PlanModule) ID() string {
	return "plans"
}

// Title satisfies app.Module.
func (m *PlanModule) Title() string {
	return "Plans"
}

// Shortcuts satisfies app.Module.
func (m *PlanModule) Shortcuts() []app.Shortcut {
	switch m.state {
	case statePlanList:
		return []app.Shortcut{
			{Key: "Enter", Description: "view plan"},
			{Key: "n", Description: "new plan"},
			{Key: "d", Description: "delete plan"},
		}
	case statePlanDetail:
		return []app.Shortcut{
			{Key: "space", Description: "toggle chunk status"},
			{Key: "e", Description: "edit metadata"},
			{Key: "d", Description: "delete plan"},
		}
	case statePlanEdit, statePlanCreate:
		return []app.Shortcut{
			{Key: "Tab", Description: "next field"},
			{Key: "Enter", Description: "submit"},
			{Key: "Esc", Description: "cancel"},
		}
	default:
		return []app.Shortcut{}
	}
}

// Init satisfies tea.Model.
func (m *PlanModule) Init() tea.Cmd {
	return nil
}

// Update satisfies tea.Model.
func (m *PlanModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case plansLoadedMsg:
		return m.handlePlansLoaded(msg)
	case planLoadedMsg:
		return m.handlePlanLoaded(msg)
	case planSavedMsg:
		return m.handlePlanSaved(msg)
	case planDeletedMsg:
		return m.handlePlanDeleted(msg)
	case chunkStatusUpdatedMsg:
		return m.handleChunkStatusUpdated(msg)
	case planCreatedMsg:
		return m.handlePlanCreated(msg)
	case app.ModuleActivatedMsg:
		if msg.ID == m.ID() && m.state != statePlanEdit && m.state != statePlanCreate && m.state != statePlanConfirm {
			cmd := m.loadPlans()
			return m, cmd
		}
	}
	return m, nil
}

// View satisfies tea.Model.
func (m *PlanModule) View() string {
	if m.loading {
		return "Loading plans…"
	}

	if m.loadErr != nil {
		return fmt.Sprintf("Failed to load plans: %v", m.loadErr)
	}

	switch m.state {
	case statePlanList:
		return m.renderPlanList()
	case statePlanDetail:
		return m.renderPlanDetail()
	case statePlanEdit, statePlanCreate:
		return m.renderForm()
	case statePlanConfirm:
		return m.renderConfirm()
	default:
		return "Unknown state."
	}
}

// --------------------------------------------------------------------
// Internal helpers

func (m *PlanModule) loadPlans() tea.Cmd {
	if m.service == nil {
		m.loadErr = fmt.Errorf("plan service unavailable")
		return func() tea.Msg {
			return app.StatusMsg{
				Message: "Plan service unavailable",
				IsError: true,
			}
		}
	}
	m.loading = true
	m.loadErr = nil
	return func() tea.Msg {
		records, err := m.service.List(context.Background(), nil)
		return plansLoadedMsg{records: records, err: err}
	}
}

func (m *PlanModule) handlePlansLoaded(msg plansLoadedMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		m.loadErr = msg.err
		return m, func() tea.Msg {
			return app.StatusMsg{
				Message: fmt.Sprintf("Failed to load plans: %v", msg.err),
				IsError: true,
			}
		}
	}

	m.plans = msg.records
	m.dataLoaded = true
	if m.listCursor >= len(m.plans) {
		m.listCursor = maxInt(0, len(m.plans)-1)
	}

	return m, func() tea.Msg {
		return app.StatusMsg{Message: "Plans refreshed"}
	}
}

func (m *PlanModule) handlePlanLoaded(msg planLoadedMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		return m, func() tea.Msg {
			return app.StatusMsg{
				Message: fmt.Sprintf("Failed to load plan: %v", msg.err),
				IsError: true,
			}
		}
	}

	m.detailPlan = msg.plan
	m.state = statePlanDetail
	m.chunkCursor = 0

	return m, nil
}

func (m *PlanModule) handlePlanSaved(msg planSavedMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		return m, func() tea.Msg {
			return app.StatusMsg{
				Message: fmt.Sprintf("Failed to save plan: %v", msg.err),
				IsError: true,
			}
		}
	}

	m.detailPlan = msg.plan
	m.state = statePlanDetail
	m.form = nil

	return m, tea.Batch(
		func() tea.Msg {
			return app.StatusMsg{Message: "Plan updated"}
		},
		func() tea.Msg {
			return app.BroadcastMsg{Topic: app.TopicPlansChanged, Payload: msg.plan.ID}
		},
		m.loadPlans(),
	)
}

func (m *PlanModule) handlePlanDeleted(msg planDeletedMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		return m, func() tea.Msg {
			return app.StatusMsg{
				Message: fmt.Sprintf("Failed to delete plan: %v", msg.err),
				IsError: true,
			}
		}
	}

	m.state = statePlanList
	m.detailPlan = nil
	m.confirm = nil

	return m, tea.Batch(
		func() tea.Msg {
			return app.StatusMsg{Message: "Plan deleted"}
		},
		func() tea.Msg {
			return app.BroadcastMsg{Topic: app.TopicPlansChanged, Payload: msg.planID}
		},
		m.loadPlans(),
	)
}

func (m *PlanModule) handleChunkStatusUpdated(msg chunkStatusUpdatedMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		return m, func() tea.Msg {
			return app.StatusMsg{
				Message: fmt.Sprintf("Failed to update chunk: %v", msg.err),
				IsError: true,
			}
		}
	}

	cmds := []tea.Cmd{
		func() tea.Msg {
			return app.StatusMsg{Message: "Chunk status updated"}
		},
		func() tea.Msg {
			return app.BroadcastMsg{Topic: app.TopicPlansChanged, Payload: msg.planID}
		},
	}

	if m.detailPlan != nil && m.detailPlan.ID == msg.planID {
		cmds = append(cmds, m.reloadPlan(msg.planID))
	}

	return m, tea.Batch(cmds...)
}

func (m *PlanModule) handlePlanCreated(msg planCreatedMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		return m, func() tea.Msg {
			return app.StatusMsg{
				Message: fmt.Sprintf("Failed to create plan: %v", msg.err),
				IsError: true,
			}
		}
	}

	m.form = nil
	m.state = statePlanList

	return m, tea.Batch(
		func() tea.Msg {
			return app.StatusMsg{Message: fmt.Sprintf("Plan created: %s", msg.title)}
		},
		func() tea.Msg {
			return app.BroadcastMsg{Topic: app.TopicPlansChanged, Payload: msg.planID}
		},
		m.loadPlans(),
	)
}

// --------------------------------------------------------------------
// Input handling

func (m *PlanModule) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case statePlanList:
		return m.handleListKeys(msg)
	case statePlanDetail:
		return m.handleDetailKeys(msg)
	case statePlanEdit, statePlanCreate:
		return m.handleFormKeys(msg)
	case statePlanConfirm:
		return m.handleConfirmKeys(msg)
	default:
		return m, nil
	}
}

func (m *PlanModule) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if len(m.plans) == 0 {
			return m, nil
		}
		m.listCursor--
		if m.listCursor < 0 {
			m.listCursor = len(m.plans) - 1
		}
	case tea.KeyDown:
		if len(m.plans) == 0 {
			return m, nil
		}
		m.listCursor++
		if m.listCursor >= len(m.plans) {
			m.listCursor = 0
		}
	case tea.KeyEnter:
		return m.openSelectedPlan()
	case tea.KeyRunes:
		if len(msg.Runes) == 0 {
			return m, nil
		}
		switch msg.Runes[0] {
		case 'n', 'N':
			return m.showCreateForm()
		case 'd', 'D':
			return m.startDeleteSelectedPlan()
		}
	}
	return m, nil
}

func (m *PlanModule) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.detailPlan == nil {
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		m.state = statePlanList
		m.detailPlan = nil
		return m, nil
	case tea.KeyUp:
		if len(m.detailPlan.Chunks) == 0 {
			return m, nil
		}
		m.chunkCursor--
		if m.chunkCursor < 0 {
			m.chunkCursor = len(m.detailPlan.Chunks) - 1
		}
	case tea.KeyDown:
		if len(m.detailPlan.Chunks) == 0 {
			return m, nil
		}
		m.chunkCursor++
		if m.chunkCursor >= len(m.detailPlan.Chunks) {
			m.chunkCursor = 0
		}
	case tea.KeyRunes:
		if len(msg.Runes) == 0 {
			return m, nil
		}
		switch msg.Runes[0] {
		case 'e', 'E':
			return m.showEditForm()
		case 'd', 'D':
			m.confirm = &confirmDialog{
				action:  confirmDelete,
				message: "Delete this plan? This will remove the markdown file.",
			}
			m.state = statePlanConfirm
			return m, nil
		case ' ':
			return m.toggleSelectedChunk()
		}
	}
	return m, nil
}

func (m *PlanModule) handleFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.form == nil {
		m.state = statePlanList
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		m.form = nil
		m.state = statePlanList
		return m, nil
	case tea.KeyTab, tea.KeyShiftTab:
		direction := 1
		if msg.Type == tea.KeyShiftTab {
			direction = -1
		}
		m.form.focusIndex = (m.form.focusIndex + direction + len(m.form.inputs)) % len(m.form.inputs)
		m.updateFormFocus()
		return m, nil
	case tea.KeyEnter:
		return m.submitForm()
	}

	if m.form.focusIndex < len(m.form.inputs) {
		input := m.form.inputs[m.form.focusIndex]
		if input.Update(msg) {
			m.form.validationErr = nil
		}
		return m, nil
	}

	return m, nil
}

func (m *PlanModule) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirm == nil {
		m.state = statePlanList
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		m.confirm = nil
		m.state = statePlanDetail
		return m, nil
	case tea.KeyEnter:
		if m.confirm.action == confirmDelete {
			cmd := m.deleteCurrentPlan()
			return m, cmd
		}
	}

	return m, nil
}

// --------------------------------------------------------------------
// Actions

func (m *PlanModule) openSelectedPlan() (tea.Model, tea.Cmd) {
	if len(m.plans) == 0 || m.listCursor >= len(m.plans) {
		return m, nil
	}
	record := m.plans[m.listCursor]
	m.loading = true
	return m, func() tea.Msg {
		planData, err := m.service.Get(context.Background(), record.ID)
		return planLoadedMsg{plan: planData, err: err}
	}
}

func (m *PlanModule) showCreateForm() (tea.Model, tea.Cmd) {
	inputs := []*inputField{
		newInputField("Topic (e.g. Rust async)"),
		newInputField("Total hours (e.g. 40)"),
		newInputField("Level (beginner/intermediate/advanced)"),
	}
	inputs[0].Focus()
	m.form = &planForm{
		mode:       formModeCreate,
		inputs:     inputs,
		focusIndex: 0,
	}
	m.state = statePlanCreate
	m.updateFormFocus()
	return m, nil
}

func (m *PlanModule) startDeleteSelectedPlan() (tea.Model, tea.Cmd) {
	if len(m.plans) == 0 || m.listCursor >= len(m.plans) {
		return m, nil
	}
	record := m.plans[m.listCursor]

	m.confirm = &confirmDialog{
		action:  confirmDelete,
		message: fmt.Sprintf("Delete plan %q? This will remove the markdown file.", record.Title),
	}
	m.state = statePlanConfirm
	return m, nil
}

func (m *PlanModule) deleteCurrentPlan() tea.Cmd {
	var planID string
	switch {
	case m.detailPlan != nil:
		planID = m.detailPlan.ID
	case len(m.plans) > 0 && m.listCursor < len(m.plans):
		planID = m.plans[m.listCursor].ID
	default:
		return func() tea.Msg {
			return app.StatusMsg{
				Message: "No plan selected",
				IsError: true,
			}
		}
	}

	m.loading = true
	return func() tea.Msg {
		err := m.service.Delete(context.Background(), planID)
		return planDeletedMsg{planID: planID, err: err}
	}
}

func (m *PlanModule) showEditForm() (tea.Model, tea.Cmd) {
	if m.detailPlan == nil {
		return m, nil
	}

	inputs := []*inputField{
		newInputField("Title"),
		newInputField("Total hours"),
		newInputField("Tags (comma separated)"),
	}
	inputs[0].SetValue(m.detailPlan.Title)
	inputs[1].SetValue(fmt.Sprintf("%.1f", m.detailPlan.TotalHours))
	inputs[2].SetValue(strings.Join(m.detailPlan.Tags, ", "))
	inputs[0].Focus()

	m.form = &planForm{
		mode:         formModeEdit,
		inputs:       inputs,
		focusIndex:   0,
		targetPlanID: m.detailPlan.ID,
	}
	m.state = statePlanEdit
	m.updateFormFocus()
	return m, nil
}

func (m *PlanModule) submitForm() (tea.Model, tea.Cmd) {
	if m.form == nil {
		return m, nil
	}

	switch m.form.mode {
	case formModeCreate:
		cmd := m.createPlanFromForm()
		return m, cmd
	case formModeEdit:
		cmd := m.updatePlanFromForm()
		return m, cmd
	default:
		return m, nil
	}
}

func (m *PlanModule) createPlanFromForm() tea.Cmd {
	topic := strings.TrimSpace(m.form.inputs[0].Value())
	hoursStr := strings.TrimSpace(m.form.inputs[1].Value())
	level := strings.TrimSpace(strings.ToLower(m.form.inputs[2].Value()))

	if topic == "" {
		m.form.validationErr = fmt.Errorf("topic is required")
		return nil
	}

	hours, err := strconv.ParseFloat(hoursStr, 64)
	if err != nil || hours <= 0 {
		m.form.validationErr = fmt.Errorf("hours must be positive number")
		return nil
	}

	m.loading = true
	req := plan.CreateRequest{
		Topic:      topic,
		TotalHours: hours,
		Level:      level,
	}

	return func() tea.Msg {
		createdPlan, err := m.service.Create(context.Background(), req)
		if err != nil {
			return planCreatedMsg{err: err}
		}

		return planCreatedMsg{
			planID: createdPlan.ID,
			title:  createdPlan.Title,
		}
	}
}

func (m *PlanModule) updatePlanFromForm() tea.Cmd {
	if m.detailPlan == nil {
		return func() tea.Msg {
			return planSavedMsg{err: fmt.Errorf("plan not loaded")}
		}
	}

	title := strings.TrimSpace(m.form.inputs[0].Value())
	hoursStr := strings.TrimSpace(m.form.inputs[1].Value())
	tagsStr := strings.TrimSpace(m.form.inputs[2].Value())

	if title == "" {
		m.form.validationErr = fmt.Errorf("title is required")
		return nil
	}

	hours, err := strconv.ParseFloat(hoursStr, 64)
	if err != nil || hours <= 0 {
		m.form.validationErr = fmt.Errorf("total hours must be positive")
		return nil
	}

	var tags []string
	if tagsStr != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			t := strings.TrimSpace(tag)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	planCopy := *m.detailPlan
	planCopy.Title = title
	planCopy.TotalHours = hours
	planCopy.Tags = tags

	m.loading = true
	return func() tea.Msg {
		err := m.service.Update(context.Background(), &planCopy)
		return planSavedMsg{plan: &planCopy, err: err}
	}
}

func (m *PlanModule) toggleSelectedChunk() (tea.Model, tea.Cmd) {
	if m.detailPlan == nil || len(m.detailPlan.Chunks) == 0 {
		return m, nil
	}

	chunk := m.detailPlan.Chunks[m.chunkCursor]
	nextStatus := nextChunkStatus(chunk.Status)

	m.loading = true
	return m, func() tea.Msg {
		err := m.service.UpdateChunkStatus(context.Background(), m.detailPlan.ID, chunk.ID, nextStatus)
		return chunkStatusUpdatedMsg{
			planID:  m.detailPlan.ID,
			chunkID: chunk.ID,
			status:  nextStatus,
			err:     err,
		}
	}
}

func (m *PlanModule) reloadPlan(planID string) tea.Cmd {
	return func() tea.Msg {
		planData, err := m.service.Get(context.Background(), planID)
		return planLoadedMsg{plan: planData, err: err}
	}
}

// --------------------------------------------------------------------
// Rendering helpers

func (m *PlanModule) renderPlanList() string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render("Plans")
	b.WriteString(title)
	b.WriteString("\n\n")

	if len(m.plans) == 0 {
		b.WriteString("No plans found.\n")
		b.WriteString("Press 'n' to create a new plan.")
		return b.String()
	}

	table := components.NewTable([]string{"ID", "Title", "Status", "Hours"})

	for i, record := range m.plans {
		row := []string{
			record.ID,
			record.Title,
			record.Status,
			fmt.Sprintf("%.1f", record.TotalHours),
		}

		if i == m.listCursor {
			table.AddHighlightedRow(row)
		} else {
			table.AddRow(row)
		}
	}

	b.WriteString(table.View())
	return b.String()
}

func (m *PlanModule) renderPlanDetail() string {
	if m.detailPlan == nil {
		return "Plan not loaded."
	}

	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).Render(m.detailPlan.Title)
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Status: %s | Total Hours: %.1f\n", m.detailPlan.Status, m.detailPlan.TotalHours))
	if len(m.detailPlan.Tags) > 0 {
		b.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(m.detailPlan.Tags, ", ")))
	}

	b.WriteString("\nChunks:\n")

	table := components.NewTable([]string{"ID", "Title", "Status", "Duration"})
	for i, chunk := range m.detailPlan.Chunks {
		row := []string{
			chunk.ID,
			chunk.Title,
			string(chunk.Status),
			fmt.Sprintf("%d min", chunk.Duration),
		}
		if i == m.chunkCursor {
			table.AddHighlightedRow(row)
		} else {
			table.AddRow(row)
		}
	}

	b.WriteString(table.View())
	b.WriteString("\n[Esc] Back  [space] Toggle status  [e] Edit  [d] Delete")

	return b.String()
}

func (m *PlanModule) renderForm() string {
	if m.form == nil {
		return ""
	}

	var b strings.Builder
	header := "New Plan"
	if m.form.mode == formModeEdit {
		header = "Edit Plan"
	}

	b.WriteString(lipgloss.NewStyle().Bold(true).Render(header))
	b.WriteString("\n\n")

	labels := []string{
		"Topic",
		"Total Hours",
		"Level",
	}
	if m.form.mode == formModeEdit {
		labels = []string{
			"Title",
			"Total Hours",
			"Tags",
		}
	}

	for i, input := range m.form.inputs {
		b.WriteString(labels[i])
		b.WriteString("\n")
		b.WriteString(input.View())
		b.WriteString("\n\n")
	}

	if m.form.validationErr != nil {
		errorMsg := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(m.form.validationErr.Error())
		b.WriteString(errorMsg)
		b.WriteString("\n")
	}

	b.WriteString("[Enter] Submit  [Esc] Cancel")
	return b.String()
}

func (m *PlanModule) renderConfirm() string {
	if m.confirm == nil {
		return ""
	}

	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	var b strings.Builder
	b.WriteString(style.Render(m.confirm.message))
	b.WriteString("\n\n[Enter] Confirm  [Esc] Cancel")
	return b.String()
}

// --------------------------------------------------------------------
// Utility helpers

func (m *PlanModule) updateFormFocus() {
	if m.form == nil {
		return
	}

	for i := range m.form.inputs {
		if i == m.form.focusIndex {
			m.form.inputs[i].Focus()
		} else {
			m.form.inputs[i].Blur()
		}
	}
}

func nextChunkStatus(current plan.Status) plan.Status {
	switch current {
	case plan.StatusNotStarted:
		return plan.StatusInProgress
	case plan.StatusInProgress:
		return plan.StatusCompleted
	case plan.StatusCompleted:
		return plan.StatusSkipped
	case plan.StatusSkipped:
		return plan.StatusNotStarted
	default:
		return plan.StatusNotStarted
	}
}

type inputField struct {
	placeholder string
	value       []rune
	focused     bool
}

func newInputField(placeholder string) *inputField {
	return &inputField{
		placeholder: placeholder,
		value:       []rune{},
	}
}

func (f *inputField) Focus() {
	f.focused = true
}

func (f *inputField) Blur() {
	f.focused = false
}

func (f *inputField) SetValue(v string) {
	f.value = []rune(v)
}

func (f *inputField) Value() string {
	return string(f.value)
}

func (f *inputField) Update(msg tea.KeyMsg) bool {
	if !f.focused {
		return false
	}

	switch msg.Type {
	case tea.KeyRunes:
		f.value = append(f.value, msg.Runes...)
		return true
	case tea.KeyBackspace:
		if len(f.value) > 0 {
			f.value = f.value[:len(f.value)-1]
		}
		return true
	}

	return false
}

func (f *inputField) View() string {
	content := string(f.value)
	if content == "" {
		content = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(f.placeholder)
	}
	cursor := ""
	if f.focused {
		cursor = " ▎"
	}
	style := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	if f.focused {
		style = style.BorderForeground(lipgloss.Color("212"))
	} else {
		style = style.BorderForeground(lipgloss.Color("240"))
	}
	return style.Render(content + cursor)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
