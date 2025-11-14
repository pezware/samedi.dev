// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package tui

import (
	"testing"

	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/tui/app"
	"github.com/stretchr/testify/assert"
)

func TestNewPlanModule_CreatesModule(t *testing.T) {
	// Create with nil service for basic structure test
	module := NewPlanModule(nil)

	assert.NotNil(t, module)
	assert.Equal(t, "plans", module.ID())
	assert.Equal(t, "Plans", module.Title())
}

func TestPlanModule_ID(t *testing.T) {
	module := NewPlanModule(nil)

	assert.Equal(t, "plans", module.ID())
}

func TestPlanModule_Title(t *testing.T) {
	module := NewPlanModule(nil)

	assert.Equal(t, "Plans", module.Title())
}

func TestPlanModule_Shortcuts_HasShortcuts(t *testing.T) {
	module := NewPlanModule(nil)

	shortcuts := module.Shortcuts()

	assert.NotEmpty(t, shortcuts)

	// Default state is list, which should have Enter, n, d
	hasEnter := false
	hasN := false
	hasD := false

	for _, sc := range shortcuts {
		if sc.Key == "Enter" {
			hasEnter = true
		}
		if sc.Key == "n" {
			hasN = true
		}
		if sc.Key == "d" {
			hasD = true
		}
	}

	assert.True(t, hasEnter, "List state should have 'Enter' shortcut")
	assert.True(t, hasN, "List state should have 'n' shortcut")
	assert.True(t, hasD, "List state should have 'd' shortcut")
}

func TestPlanModule_Init_ReturnsCmd(t *testing.T) {
	module := NewPlanModule(nil)

	// Init should not panic (cmd may be nil)
	assert.NotPanics(t, func() {
		cmd := module.Init()
		_ = cmd // cmd may be nil, which is fine
	})
}

func TestPlanModule_View_DoesNotPanic(t *testing.T) {
	module := NewPlanModule(nil)

	// View should not panic even with nil service
	assert.NotPanics(t, func() {
		view := module.View()
		assert.NotEmpty(t, view)
	})
}

func TestPlanModule_State_DefaultsToList(t *testing.T) {
	module := NewPlanModule(nil)

	assert.Equal(t, statePlanList, module.state)
}

func TestPlanModuleState_Constants(t *testing.T) {
	// Verify state constants are defined
	assert.Equal(t, planModuleState("list"), statePlanList)
	assert.Equal(t, planModuleState("detail"), statePlanDetail)
	assert.Equal(t, planModuleState("edit"), statePlanEdit)
	assert.Equal(t, planModuleState("create"), statePlanCreate)
	assert.Equal(t, planModuleState("confirm"), statePlanConfirm)
}

func TestPlanFormMode_Constants(t *testing.T) {
	// Verify form mode constants
	assert.Equal(t, planFormMode("edit"), formModeEdit)
	assert.Equal(t, planFormMode("create"), formModeCreate)
}

func TestConfirmAction_Constants(t *testing.T) {
	// Verify confirm action constants
	assert.Equal(t, confirmAction("delete"), confirmDelete)
}

func TestPlanModule_Update_HandlesModuleActivatedMsg(t *testing.T) {
	module := NewPlanModule(nil)

	msg := app.ModuleActivatedMsg{
		ID:              "plans",
		FirstActivation: true,
	}

	// Update should not panic
	assert.NotPanics(t, func() {
		model, cmd := module.Update(msg)
		assert.NotNil(t, model)
		// cmd may be nil or not, either is valid
		_ = cmd
	})
}

func TestPlansLoadedMsg_Structure(t *testing.T) {
	msg := plansLoadedMsg{
		records: nil,
		err:     nil,
	}

	assert.NotNil(t, &msg)
}

func TestPlanLoadedMsg_Structure(t *testing.T) {
	testPlan := &plan.Plan{
		ID:    "test-plan",
		Title: "Test Plan",
	}

	msg := planLoadedMsg{
		plan: testPlan,
		err:  nil,
	}

	assert.Equal(t, testPlan, msg.plan)
	assert.Nil(t, msg.err)
}

func TestPlanSavedMsg_Structure(t *testing.T) {
	msg := planSavedMsg{
		plan: nil,
		err:  nil,
	}

	assert.NotNil(t, &msg)
}

func TestPlanDeletedMsg_Structure(t *testing.T) {
	msg := planDeletedMsg{
		planID: "test-plan",
		err:    nil,
	}

	assert.Equal(t, "test-plan", msg.planID)
	assert.Nil(t, msg.err)
}

func TestChunkStatusUpdatedMsg_Structure(t *testing.T) {
	msg := chunkStatusUpdatedMsg{
		planID:  "test-plan",
		chunkID: "chunk-001",
		status:  plan.StatusCompleted,
		err:     nil,
	}

	assert.Equal(t, "test-plan", msg.planID)
	assert.Equal(t, "chunk-001", msg.chunkID)
	assert.Equal(t, plan.StatusCompleted, msg.status)
	assert.Nil(t, msg.err)
}

func TestPlanModule_InitialState(t *testing.T) {
	module := NewPlanModule(nil)

	// Verify initial state
	assert.Equal(t, statePlanList, module.state)
	assert.Equal(t, 0, module.listCursor)
	assert.Equal(t, 0, module.chunkCursor)
	assert.False(t, module.loading)
	assert.False(t, module.dataLoaded)
	assert.Nil(t, module.loadErr)
}

func TestConfirmDialog_Structure(t *testing.T) {
	dialog := &confirmDialog{
		action:  confirmDelete,
		message: "Are you sure?",
	}

	assert.Equal(t, confirmDelete, dialog.action)
	assert.Equal(t, "Are you sure?", dialog.message)
}

func TestPlanForm_Structure(t *testing.T) {
	form := &planForm{
		mode:          formModeCreate,
		inputs:        nil,
		focusIndex:    0,
		targetPlanID:  "test-plan",
		validationErr: nil,
	}

	assert.Equal(t, formModeCreate, form.mode)
	assert.Equal(t, 0, form.focusIndex)
	assert.Equal(t, "test-plan", form.targetPlanID)
	assert.Nil(t, form.validationErr)
}
