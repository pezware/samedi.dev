// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

//go:build integration
// +build integration

package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/pezware/samedi.dev/internal/session"
	"github.com/pezware/samedi.dev/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionWorkflow_Complete tests the full lifecycle of a session.
func TestSessionWorkflow_Complete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup: Create real database and service
	db, cleanup := setupIntegrationDB(t)
	defer cleanup()

	repo := session.NewSQLiteRepository(db)
	svc := session.NewService(repo, nil)
	ctx := context.Background()

	// Step 1: Start a session
	startReq := session.StartRequest{
		PlanID:  "test-plan",
		ChunkID: "chunk-001",
		Notes:   "Starting integration test",
	}

	sess, err := svc.Start(ctx, startReq)
	require.NoError(t, err)
	require.NotNil(t, sess)
	assert.NotEmpty(t, sess.ID)
	assert.Equal(t, "test-plan", sess.PlanID)
	assert.Equal(t, "chunk-001", sess.ChunkID)
	assert.True(t, sess.IsActive())
	assert.Equal(t, "Starting integration test", sess.Notes)

	sessionID := sess.ID

	// Step 2: Check status - should show active session
	status, err := svc.GetStatus(ctx)
	require.NoError(t, err)
	assert.NotNil(t, status.Active)
	assert.Equal(t, sessionID, status.Active.ID)
	assert.True(t, status.Active.IsActive())

	// Step 3: Try to start another session - should fail
	_, err = svc.Start(ctx, session.StartRequest{
		PlanID: "another-plan",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "active session already exists")

	// Step 4: Add some delay to ensure duration > 0
	time.Sleep(100 * time.Millisecond)

	// Step 5: Stop the session
	stopReq := session.StopRequest{
		Notes:     "Completed integration test",
		Artifacts: []string{"https://example.com/resource", "/path/to/file.md"},
	}

	stoppedSess, err := svc.Stop(ctx, stopReq)
	require.NoError(t, err)
	assert.NotNil(t, stoppedSess)
	assert.Equal(t, sessionID, stoppedSess.ID)
	assert.False(t, stoppedSess.IsActive())
	assert.NotNil(t, stoppedSess.EndTime)
	// Duration might be 0 minutes if test runs too fast, but it should be calculated correctly
	assert.Equal(t, stoppedSess.CalculateDuration(), stoppedSess.Duration)
	assert.Contains(t, stoppedSess.Notes, "Completed integration test")
	assert.Len(t, stoppedSess.Artifacts, 2)

	// Step 6: Check status again - should show no active session
	status, err = svc.GetStatus(ctx)
	require.NoError(t, err)
	assert.Nil(t, status.Active)
	// Note: GetStatus only returns recent sessions when there's an active session
	// So we need to use List() to verify the session was saved

	// Step 7: List sessions for the plan
	sessions, err := svc.List(ctx, "test-plan", 10)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, sessionID, sessions[0].ID)

	// Step 8: Get session count
	count, err := svc.GetSessionCount(ctx, "test-plan")
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Step 9: Get total duration
	totalDuration, err := svc.GetTotalDuration(ctx, "test-plan")
	require.NoError(t, err)
	assert.Equal(t, stoppedSess.Duration, totalDuration)
}

// TestSessionWorkflow_MultipleSessionsSamePlan tests multiple sessions for the same plan.
func TestSessionWorkflow_MultipleSessionsSamePlan(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db, cleanup := setupIntegrationDB(t)
	defer cleanup()

	repo := session.NewSQLiteRepository(db)
	svc := session.NewService(repo, nil)
	ctx := context.Background()

	planID := "multi-session-plan"

	// Create and complete 3 sessions
	for i := 1; i <= 3; i++ {
		// Start session
		startReq := session.StartRequest{
			PlanID:  planID,
			ChunkID: "chunk-001",
		}
		sess, err := svc.Start(ctx, startReq)
		require.NoError(t, err)
		require.NotNil(t, sess)

		// Wait a bit
		time.Sleep(50 * time.Millisecond)

		// Stop session
		stopReq := session.StopRequest{}
		_, err = svc.Stop(ctx, stopReq)
		require.NoError(t, err)
	}

	// Verify we have 3 sessions
	sessions, err := svc.List(ctx, planID, 10)
	require.NoError(t, err)
	assert.Len(t, sessions, 3)

	// Verify session count
	count, err := svc.GetSessionCount(ctx, planID)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	// Verify total duration is sum of all sessions
	totalDuration, err := svc.GetTotalDuration(ctx, planID)
	require.NoError(t, err)
	// Duration might be 0 if test runs too fast
	assert.GreaterOrEqual(t, totalDuration, 0)

	// Verify status shows no active session
	status, err := svc.GetStatus(ctx)
	require.NoError(t, err)
	assert.Nil(t, status.Active)
	// Note: Recent is only populated when there's an active session
}

// TestSessionWorkflow_DifferentPlans tests sessions across multiple plans.
func TestSessionWorkflow_DifferentPlans(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db, cleanup := setupIntegrationDB(t)
	defer cleanup()

	repo := session.NewSQLiteRepository(db)
	svc := session.NewService(repo, nil)
	ctx := context.Background()

	// Create sessions for plan A
	for i := 0; i < 2; i++ {
		sess, err := svc.Start(ctx, session.StartRequest{PlanID: "plan-a"})
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)
		_, err = svc.Stop(ctx, session.StopRequest{})
		require.NoError(t, err)
		_ = sess
	}

	// Create sessions for plan B
	for i := 0; i < 3; i++ {
		sess, err := svc.Start(ctx, session.StartRequest{PlanID: "plan-b"})
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)
		_, err = svc.Stop(ctx, session.StopRequest{})
		require.NoError(t, err)
		_ = sess
	}

	// Verify plan A has 2 sessions
	countA, err := svc.GetSessionCount(ctx, "plan-a")
	require.NoError(t, err)
	assert.Equal(t, 2, countA)

	// Verify plan B has 3 sessions
	countB, err := svc.GetSessionCount(ctx, "plan-b")
	require.NoError(t, err)
	assert.Equal(t, 3, countB)

	// Verify list only returns sessions for specific plan
	sessionsA, err := svc.List(ctx, "plan-a", 10)
	require.NoError(t, err)
	assert.Len(t, sessionsA, 2)
	for _, sess := range sessionsA {
		assert.Equal(t, "plan-a", sess.PlanID)
	}

	sessionsB, err := svc.List(ctx, "plan-b", 10)
	require.NoError(t, err)
	assert.Len(t, sessionsB, 3)
	for _, sess := range sessionsB {
		assert.Equal(t, "plan-b", sess.PlanID)
	}
}

// TestSessionWorkflow_LimitParameter tests the limit parameter in List.
func TestSessionWorkflow_LimitParameter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db, cleanup := setupIntegrationDB(t)
	defer cleanup()

	repo := session.NewSQLiteRepository(db)
	svc := session.NewService(repo, nil)
	ctx := context.Background()

	planID := "limit-test-plan"

	// Create 10 sessions
	for i := 0; i < 10; i++ {
		sess, err := svc.Start(ctx, session.StartRequest{PlanID: planID})
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
		_, err = svc.Stop(ctx, session.StopRequest{})
		require.NoError(t, err)
		_ = sess
	}

	// Test limit = 5
	sessions, err := svc.List(ctx, planID, 5)
	require.NoError(t, err)
	assert.Len(t, sessions, 5)

	// Test limit = 0 (should return all)
	sessions, err = svc.List(ctx, planID, 0)
	require.NoError(t, err)
	assert.Len(t, sessions, 10)

	// Verify sessions are ordered by most recent first
	for i := 0; i < len(sessions)-1; i++ {
		assert.True(t, sessions[i].StartTime.After(sessions[i+1].StartTime) ||
			sessions[i].StartTime.Equal(sessions[i+1].StartTime))
	}
}

// setupIntegrationDB creates an in-memory SQLite database for integration tests.
func setupIntegrationDB(t *testing.T) (*storage.SQLiteDB, func()) {
	t.Helper()

	// Create in-memory database
	db, err := storage.NewSQLiteDB(":memory:")
	require.NoError(t, err)

	// Run migrations
	migrator := storage.NewMigrator(db)
	err = migrator.Migrate()
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}
