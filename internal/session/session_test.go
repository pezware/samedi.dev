// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession_Validate_ValidActiveSession(t *testing.T) {
	now := time.Now()
	session := &Session{
		ID:        "test-session-id",
		PlanID:    "test-plan",
		ChunkID:   "chunk-001",
		StartTime: now,
		EndTime:   nil, // Active session
		Duration:  0,
		CreatedAt: now,
	}

	err := session.Validate()
	assert.NoError(t, err)
}

func TestSession_Validate_ValidCompletedSession(t *testing.T) {
	start := time.Now()
	end := start.Add(1 * time.Hour)

	session := &Session{
		ID:        "test-session-id",
		PlanID:    "test-plan",
		StartTime: start,
		EndTime:   &end,
		Duration:  60,
		CreatedAt: start,
	}

	err := session.Validate()
	assert.NoError(t, err)
}

func TestSession_Validate_EmptyID(t *testing.T) {
	session := &Session{
		PlanID:    "test-plan",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
	}

	err := session.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session ID cannot be empty")
}

func TestSession_Validate_EmptyPlanID(t *testing.T) {
	session := &Session{
		ID:        "test-session-id",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
	}

	err := session.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan ID cannot be empty")
}

func TestSession_Validate_ZeroStartTime(t *testing.T) {
	session := &Session{
		ID:        "test-session-id",
		PlanID:    "test-plan",
		StartTime: time.Time{},
		CreatedAt: time.Now(),
	}

	err := session.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start time cannot be zero")
}

func TestSession_Validate_ZeroCreatedAt(t *testing.T) {
	session := &Session{
		ID:        "test-session-id",
		PlanID:    "test-plan",
		StartTime: time.Now(),
		CreatedAt: time.Time{},
	}

	err := session.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "created_at cannot be zero")
}

func TestSession_Validate_EndTimeBeforeStartTime(t *testing.T) {
	start := time.Now()
	end := start.Add(-1 * time.Hour)

	session := &Session{
		ID:        "test-session-id",
		PlanID:    "test-plan",
		StartTime: start,
		EndTime:   &end,
		Duration:  60,
		CreatedAt: start,
	}

	err := session.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "end time cannot be before start time")
}

func TestSession_Validate_EndTimeEqualsStartTime(t *testing.T) {
	start := time.Now()
	end := start

	session := &Session{
		ID:        "test-session-id",
		PlanID:    "test-plan",
		StartTime: start,
		EndTime:   &end,
		Duration:  0,
		CreatedAt: start,
	}

	err := session.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "end time cannot equal start time")
}

func TestSession_Validate_DurationMismatch(t *testing.T) {
	start := time.Now()
	end := start.Add(1 * time.Hour)

	session := &Session{
		ID:        "test-session-id",
		PlanID:    "test-plan",
		StartTime: start,
		EndTime:   &end,
		Duration:  30, // Should be 60
		CreatedAt: start,
	}

	err := session.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duration mismatch")
}

func TestSession_Validate_ActiveSessionWithDuration(t *testing.T) {
	session := &Session{
		ID:        "test-session-id",
		PlanID:    "test-plan",
		StartTime: time.Now(),
		EndTime:   nil,
		Duration:  30, // Active sessions should have 0 duration
		CreatedAt: time.Now(),
	}

	err := session.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "active session should have zero duration")
}

func TestSession_Validate_NegativeCardsCreated(t *testing.T) {
	session := &Session{
		ID:           "test-session-id",
		PlanID:       "test-plan",
		StartTime:    time.Now(),
		CreatedAt:    time.Now(),
		CardsCreated: -1,
	}

	err := session.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cards created cannot be negative")
}

func TestSession_IsActive_ActiveSession(t *testing.T) {
	session := &Session{
		ID:        "test-session-id",
		PlanID:    "test-plan",
		StartTime: time.Now(),
		EndTime:   nil,
		CreatedAt: time.Now(),
	}

	assert.True(t, session.IsActive())
}

func TestSession_IsActive_CompletedSession(t *testing.T) {
	start := time.Now()
	end := start.Add(1 * time.Hour)

	session := &Session{
		ID:        "test-session-id",
		PlanID:    "test-plan",
		StartTime: start,
		EndTime:   &end,
		Duration:  60,
		CreatedAt: start,
	}

	assert.False(t, session.IsActive())
}

func TestSession_CalculateDuration_ActiveSession(t *testing.T) {
	session := &Session{
		StartTime: time.Now(),
		EndTime:   nil,
	}

	duration := session.CalculateDuration()
	assert.Equal(t, 0, duration)
}

func TestSession_CalculateDuration_OneHourSession(t *testing.T) {
	start := time.Now()
	end := start.Add(1 * time.Hour)

	session := &Session{
		StartTime: start,
		EndTime:   &end,
	}

	duration := session.CalculateDuration()
	assert.Equal(t, 60, duration)
}

func TestSession_CalculateDuration_ThirtyMinuteSession(t *testing.T) {
	start := time.Now()
	end := start.Add(30 * time.Minute)

	session := &Session{
		StartTime: start,
		EndTime:   &end,
	}

	duration := session.CalculateDuration()
	assert.Equal(t, 30, duration)
}

func TestSession_CalculateDuration_OvernightSession(t *testing.T) {
	start := time.Date(2024, 1, 15, 23, 30, 0, 0, time.UTC)
	end := time.Date(2024, 1, 16, 1, 30, 0, 0, time.UTC)

	session := &Session{
		StartTime: start,
		EndTime:   &end,
	}

	duration := session.CalculateDuration()
	assert.Equal(t, 120, duration) // 2 hours
}

func TestSession_CalculateDuration_MultiDaySession(t *testing.T) {
	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC)

	session := &Session{
		StartTime: start,
		EndTime:   &end,
	}

	duration := session.CalculateDuration()
	assert.Equal(t, 1440, duration) // 24 hours
}

func TestSession_ElapsedMinutes_CompletedSession(t *testing.T) {
	start := time.Now()
	end := start.Add(45 * time.Minute)

	session := &Session{
		StartTime: start,
		EndTime:   &end,
		Duration:  45,
	}

	elapsed := session.ElapsedMinutes()
	assert.Equal(t, 45, elapsed)
}

func TestSession_ElapsedMinutes_ActiveSession(t *testing.T) {
	// Start session 10 minutes ago
	start := time.Now().Add(-10 * time.Minute)

	session := &Session{
		StartTime: start,
		EndTime:   nil,
		Duration:  0,
	}

	elapsed := session.ElapsedMinutes()
	// Should be approximately 10 minutes (allow 1 minute tolerance)
	assert.InDelta(t, 10, elapsed, 1)
}

func TestSession_ElapsedTime_ShortSession(t *testing.T) {
	start := time.Now()
	end := start.Add(30 * time.Minute)

	session := &Session{
		StartTime: start,
		EndTime:   &end,
		Duration:  30,
	}

	elapsed := session.ElapsedTime()
	assert.Equal(t, "30m", elapsed)
}

func TestSession_ElapsedTime_OneHourSession(t *testing.T) {
	start := time.Now()
	end := start.Add(1*time.Hour + 15*time.Minute)

	session := &Session{
		StartTime: start,
		EndTime:   &end,
		Duration:  75,
	}

	elapsed := session.ElapsedTime()
	assert.Equal(t, "1h 15m", elapsed)
}

func TestSession_ElapsedTime_MultiHourSession(t *testing.T) {
	start := time.Now()
	end := start.Add(3*time.Hour + 45*time.Minute)

	session := &Session{
		StartTime: start,
		EndTime:   &end,
		Duration:  225,
	}

	elapsed := session.ElapsedTime()
	assert.Equal(t, "3h 45m", elapsed)
}

func TestSession_Complete_Success(t *testing.T) {
	start := time.Now()
	end := start.Add(1 * time.Hour)

	session := &Session{
		ID:        "test-session-id",
		PlanID:    "test-plan",
		StartTime: start,
		EndTime:   nil,
		Duration:  0,
		CreatedAt: start,
	}

	err := session.Complete(end)
	require.NoError(t, err)

	assert.False(t, session.IsActive())
	assert.Equal(t, end, *session.EndTime)
	assert.Equal(t, 60, session.Duration)
}

func TestSession_Complete_AlreadyCompleted(t *testing.T) {
	start := time.Now()
	end := start.Add(1 * time.Hour)

	session := &Session{
		StartTime: start,
		EndTime:   &end,
		Duration:  60,
	}

	err := session.Complete(time.Now())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session is already complete")
}

func TestSession_Complete_EndTimeBeforeStart(t *testing.T) {
	start := time.Now()
	end := start.Add(-1 * time.Hour)

	session := &Session{
		StartTime: start,
		EndTime:   nil,
	}

	err := session.Complete(end)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "end time cannot be before start time")
}

func TestSession_AddNotes_EmptyNotes(t *testing.T) {
	session := &Session{}
	session.AddNotes("First note")

	assert.Equal(t, "First note", session.Notes)
}

func TestSession_AddNotes_AppendToExisting(t *testing.T) {
	session := &Session{
		Notes: "First note",
	}
	session.AddNotes("Second note")

	assert.Equal(t, "First note\nSecond note", session.Notes)
}

func TestSession_AddArtifact_SingleArtifact(t *testing.T) {
	session := &Session{}
	session.AddArtifact("https://github.com/user/repo")

	assert.Len(t, session.Artifacts, 1)
	assert.Equal(t, "https://github.com/user/repo", session.Artifacts[0])
}

func TestSession_AddArtifact_MultipleArtifacts(t *testing.T) {
	session := &Session{}
	session.AddArtifact("https://github.com/user/repo")
	session.AddArtifact("/path/to/file.md")

	assert.Len(t, session.Artifacts, 2)
	assert.Equal(t, "https://github.com/user/repo", session.Artifacts[0])
	assert.Equal(t, "/path/to/file.md", session.Artifacts[1])
}

func TestSession_AddArtifact_EmptyString(t *testing.T) {
	session := &Session{}
	session.AddArtifact("")

	assert.Len(t, session.Artifacts, 0)
}
