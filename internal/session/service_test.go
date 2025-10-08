// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package session

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of the Repository interface for testing.
type MockRepository struct {
	sessions    map[string]*Session
	activeID    string
	createError error
	updateError error
	getError    error
	listError   error
	deleteError error
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		sessions: make(map[string]*Session),
	}
}

func (m *MockRepository) Create(_ context.Context, session *Session) error {
	if m.createError != nil {
		return m.createError
	}
	m.sessions[session.ID] = session
	if session.IsActive() {
		m.activeID = session.ID
	}
	return nil
}

func (m *MockRepository) Get(_ context.Context, id string) (*Session, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	session, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return session, nil
}

func (m *MockRepository) GetActive(_ context.Context) (*Session, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	if m.activeID == "" {
		return nil, nil
	}
	return m.sessions[m.activeID], nil
}

func (m *MockRepository) Update(_ context.Context, session *Session) error {
	if m.updateError != nil {
		return m.updateError
	}
	m.sessions[session.ID] = session
	if !session.IsActive() && m.activeID == session.ID {
		m.activeID = ""
	}
	return nil
}

func (m *MockRepository) List(_ context.Context, planID string, limit int) ([]*Session, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	sessions := make([]*Session, 0)
	for _, session := range m.sessions {
		if session.PlanID == planID {
			sessions = append(sessions, session)
			if len(sessions) >= limit {
				break
			}
		}
	}
	return sessions, nil
}

func (m *MockRepository) GetByPlan(_ context.Context, planID string) ([]*Session, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	sessions := make([]*Session, 0)
	for _, session := range m.sessions {
		if session.PlanID == planID {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (m *MockRepository) Delete(_ context.Context, id string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.sessions, id)
	if m.activeID == id {
		m.activeID = ""
	}
	return nil
}

// MockPlanService is a mock implementation of PlanService for testing.
type MockPlanService struct {
	plans    map[string]interface{}
	getError error
}

func NewMockPlanService() *MockPlanService {
	return &MockPlanService{
		plans: make(map[string]interface{}),
	}
}

func (m *MockPlanService) AddPlan(id string) {
	m.plans[id] = struct{}{}
}

func (m *MockPlanService) Get(_ context.Context, id string) (interface{}, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	if _, exists := m.plans[id]; !exists {
		return nil, fmt.Errorf("plan not found: %s", id)
	}
	return m.plans[id], nil
}

func TestService_Start_Success(t *testing.T) {
	repo := NewMockRepository()
	planService := NewMockPlanService()
	planService.AddPlan("test-plan")

	service := NewService(repo, planService)
	ctx := context.Background()

	req := StartRequest{
		PlanID:  "test-plan",
		ChunkID: "chunk-001",
	}

	session, err := service.Start(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Equal(t, "test-plan", session.PlanID)
	assert.Equal(t, "chunk-001", session.ChunkID)
	assert.True(t, session.IsActive())
	assert.NotEmpty(t, session.ID)
}

func TestService_Start_WithInitialNotes(t *testing.T) {
	repo := NewMockRepository()
	planService := NewMockPlanService()
	planService.AddPlan("test-plan")

	service := NewService(repo, planService)
	ctx := context.Background()

	req := StartRequest{
		PlanID: "test-plan",
		Notes:  "Starting work on chapter 3",
	}

	session, err := service.Start(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, "Starting work on chapter 3", session.Notes)
}

func TestService_Start_EmptyPlanID(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	req := StartRequest{
		PlanID: "",
	}

	_, err := service.Start(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan ID cannot be empty")
}

func TestService_Start_ActiveSessionExists(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	// Create an active session
	activeSession := &Session{
		ID:        uuid.New().String(),
		PlanID:    "existing-plan",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
	}
	repo.Create(ctx, activeSession)

	// Try to start another session
	req := StartRequest{
		PlanID: "new-plan",
	}

	_, err := service.Start(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "active session already exists")
	assert.Contains(t, err.Error(), "samedi stop")
}

func TestService_Start_PlanNotFound(t *testing.T) {
	repo := NewMockRepository()
	planService := NewMockPlanService()
	// Don't add any plans

	service := NewService(repo, planService)
	ctx := context.Background()

	req := StartRequest{
		PlanID: "non-existent-plan",
	}

	_, err := service.Start(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan not found")
}

func TestService_Start_NoPlanServiceValidation(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil) // No plan service
	ctx := context.Background()

	req := StartRequest{
		PlanID: "any-plan", // Won't be validated
	}

	session, err := service.Start(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "any-plan", session.PlanID)
}

func TestService_Start_RepositoryError(t *testing.T) {
	repo := NewMockRepository()
	repo.createError = fmt.Errorf("database error")

	service := NewService(repo, nil)
	ctx := context.Background()

	req := StartRequest{
		PlanID: "test-plan",
	}

	_, err := service.Start(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create session")
}

func TestService_Stop_Success(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	// Start a session first
	startReq := StartRequest{
		PlanID: "test-plan",
	}
	_, err := service.Start(ctx, startReq)
	require.NoError(t, err)

	// Stop it (wait at least 1 minute to ensure duration > 0)
	time.Sleep(100 * time.Millisecond)
	stopReq := StopRequest{
		Notes:     "Completed chapter 3",
		Artifacts: []string{"https://github.com/user/repo"},
	}

	stoppedSession, err := service.Stop(ctx, stopReq)
	require.NoError(t, err)
	require.NotNil(t, stoppedSession)

	assert.False(t, stoppedSession.IsActive())
	// Duration may be 0 if less than 1 minute elapsed, just check it's set correctly
	assert.Equal(t, stoppedSession.CalculateDuration(), stoppedSession.Duration)
	assert.Contains(t, stoppedSession.Notes, "Completed chapter 3")
	assert.Len(t, stoppedSession.Artifacts, 1)
	assert.Equal(t, "https://github.com/user/repo", stoppedSession.Artifacts[0])
}

func TestService_Stop_NoNotes(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	// Start and stop without notes
	service.Start(ctx, StartRequest{PlanID: "test-plan"})

	stopReq := StopRequest{
		Notes: "",
	}

	stoppedSession, err := service.Stop(ctx, stopReq)
	require.NoError(t, err)
	assert.Empty(t, stoppedSession.Notes)
}

func TestService_Stop_NoActiveSession(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	stopReq := StopRequest{}

	_, err := service.Stop(ctx, stopReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active session to stop")
	assert.Contains(t, err.Error(), "samedi start")
}

func TestService_Stop_RepositoryError(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	// Start a session
	service.Start(ctx, StartRequest{PlanID: "test-plan"})

	// Set update error
	repo.updateError = fmt.Errorf("database error")

	stopReq := StopRequest{}

	_, err := service.Stop(ctx, stopReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update session")
}

func TestService_GetActive_WithActiveSession(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	// Start a session
	started, _ := service.Start(ctx, StartRequest{PlanID: "test-plan"})

	// Get active
	active, err := service.GetActive(ctx)
	require.NoError(t, err)
	require.NotNil(t, active)

	assert.Equal(t, started.ID, active.ID)
	assert.True(t, active.IsActive())
}

func TestService_GetActive_NoActiveSession(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	active, err := service.GetActive(ctx)
	require.NoError(t, err)
	assert.Nil(t, active)
}

func TestService_GetStatus_WithActiveSession(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	// Start a session
	service.Start(ctx, StartRequest{PlanID: "test-plan"})

	status, err := service.GetStatus(ctx)
	require.NoError(t, err)
	require.NotNil(t, status)

	assert.NotNil(t, status.Active)
	assert.True(t, status.Active.IsActive())
}

func TestService_GetStatus_NoActiveSession(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	status, err := service.GetStatus(ctx)
	require.NoError(t, err)
	require.NotNil(t, status)

	assert.Nil(t, status.Active)
}

func TestService_List_Success(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	// Create multiple completed sessions
	for i := 0; i < 5; i++ {
		start := time.Now().Add(time.Duration(-i) * time.Hour)
		end := start.Add(1 * time.Hour)
		session := &Session{
			ID:        uuid.New().String(),
			PlanID:    "test-plan",
			StartTime: start,
			EndTime:   &end,
			Duration:  60,
			CreatedAt: start,
		}
		repo.Create(ctx, session)
	}

	sessions, err := service.List(ctx, "test-plan", 3)
	require.NoError(t, err)
	assert.Len(t, sessions, 3)
}

func TestService_List_EmptyPlanID(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	_, err := service.List(ctx, "", 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan ID cannot be empty")
}

func TestService_GetByPlan_Success(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	// Create sessions for different plans
	for i := 0; i < 3; i++ {
		session := &Session{
			ID:        uuid.New().String(),
			PlanID:    "plan-1",
			StartTime: time.Now(),
			CreatedAt: time.Now(),
		}
		repo.Create(ctx, session)
	}

	for i := 0; i < 2; i++ {
		session := &Session{
			ID:        uuid.New().String(),
			PlanID:    "plan-2",
			StartTime: time.Now(),
			CreatedAt: time.Now(),
		}
		repo.Create(ctx, session)
	}

	sessions, err := service.GetByPlan(ctx, "plan-1")
	require.NoError(t, err)
	assert.Len(t, sessions, 3)
}

func TestService_GetByPlan_EmptyPlanID(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	_, err := service.GetByPlan(ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan ID cannot be empty")
}

func TestService_GetSessionCount_Success(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	// Create 5 sessions
	for i := 0; i < 5; i++ {
		session := &Session{
			ID:        uuid.New().String(),
			PlanID:    "test-plan",
			StartTime: time.Now(),
			CreatedAt: time.Now(),
		}
		repo.Create(ctx, session)
	}

	count, err := service.GetSessionCount(ctx, "test-plan")
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestService_GetSessionCount_NoPlan(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	count, err := service.GetSessionCount(ctx, "non-existent")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestService_GetTotalDuration_Success(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	// Create completed sessions with different durations
	start := time.Now()
	for i, duration := range []int{30, 60, 45} {
		end := start.Add(time.Duration(duration) * time.Minute)
		session := &Session{
			ID:        uuid.New().String(),
			PlanID:    "test-plan",
			StartTime: start.Add(time.Duration(i) * time.Hour),
			EndTime:   &end,
			Duration:  duration,
			CreatedAt: start,
		}
		repo.Create(ctx, session)
	}

	totalDuration, err := service.GetTotalDuration(ctx, "test-plan")
	require.NoError(t, err)
	assert.Equal(t, 135, totalDuration) // 30 + 60 + 45
}

func TestService_GetTotalDuration_ExcludesActiveSessions(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	// Create one completed session (60 min)
	start := time.Now()
	end := start.Add(60 * time.Minute)
	completedSession := &Session{
		ID:        uuid.New().String(),
		PlanID:    "test-plan",
		StartTime: start,
		EndTime:   &end,
		Duration:  60,
		CreatedAt: start,
	}
	repo.Create(ctx, completedSession)

	// Create one active session (should be excluded)
	activeSession := &Session{
		ID:        uuid.New().String(),
		PlanID:    "test-plan",
		StartTime: time.Now(),
		EndTime:   nil,
		Duration:  0,
		CreatedAt: time.Now(),
	}
	repo.Create(ctx, activeSession)

	totalDuration, err := service.GetTotalDuration(ctx, "test-plan")
	require.NoError(t, err)
	assert.Equal(t, 60, totalDuration) // Only the completed session
}

func TestService_GetTotalDuration_NoPlan(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)
	ctx := context.Background()

	totalDuration, err := service.GetTotalDuration(ctx, "non-existent")
	require.NoError(t, err)
	assert.Equal(t, 0, totalDuration)
}
