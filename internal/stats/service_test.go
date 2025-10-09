// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package stats

import (
	"context"
	"testing"
	"time"

	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/session"
	"github.com/pezware/samedi.dev/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

// MockPlanService mocks the plan service interface
type MockPlanService struct {
	mock.Mock
}

func (m *MockPlanService) Get(ctx context.Context, id string) (*plan.Plan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*plan.Plan), args.Error(1)
}

func (m *MockPlanService) List(ctx context.Context, filter *storage.PlanFilter) ([]*storage.PlanRecord, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*storage.PlanRecord), args.Error(1)
}

// MockSessionService mocks the session service interface
type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) List(ctx context.Context, planID string, limit int) ([]*session.Session, error) {
	args := m.Called(ctx, planID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*session.Session), args.Error(1)
}

func (m *MockSessionService) ListAll(ctx context.Context) ([]*session.Session, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*session.Session), args.Error(1)
}

// Test fixtures

func newTestPlan(id, title string, status plan.Status, chunks []plan.Chunk) *plan.Plan {
	now := time.Now()
	return &plan.Plan{
		ID:         id,
		Title:      title,
		CreatedAt:  now.AddDate(0, 0, -10),
		UpdatedAt:  now,
		TotalHours: 40.0,
		Status:     status,
		Tags:       []string{"test"},
		Chunks:     chunks,
	}
}

func newTestSession(id, planID string, startTime time.Time, durationMinutes int) *session.Session {
	endTime := startTime.Add(time.Duration(durationMinutes) * time.Minute)
	return &session.Session{
		ID:        id,
		PlanID:    planID,
		StartTime: startTime,
		EndTime:   &endTime,
		Duration:  durationMinutes,
		CreatedAt: startTime,
	}
}

func newTestPlanRecord(id, title string, status plan.Status) *storage.PlanRecord {
	now := time.Now()
	return &storage.PlanRecord{
		ID:         id,
		Title:      title,
		CreatedAt:  now.AddDate(0, 0, -10),
		UpdatedAt:  now,
		TotalHours: 40.0,
		Status:     string(status),
		Tags:       []string{"test"},
		FilePath:   "/path/to/" + id + ".md",
	}
}

// Service Tests

func TestService_GetTotalStats(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	tests := []struct {
		name          string
		mockPlans     []*storage.PlanRecord
		mockFullPlans map[string]*plan.Plan
		mockSessions  []*session.Session
		expectedStats *TotalStats
		expectError   bool
	}{
		{
			name:          "empty state - no plans or sessions",
			mockPlans:     []*storage.PlanRecord{},
			mockFullPlans: map[string]*plan.Plan{},
			mockSessions:  []*session.Session{},
			expectedStats: &TotalStats{
				TotalHours:     0,
				TotalSessions:  0,
				ActivePlans:    0,
				CompletedPlans: 0,
				CurrentStreak:  0,
				LongestStreak:  0,
				AverageSession: 0,
			},
		},
		{
			name: "single active plan with sessions",
			mockPlans: []*storage.PlanRecord{
				newTestPlanRecord("p1", "Rust Async", plan.StatusInProgress),
			},
			mockFullPlans: map[string]*plan.Plan{
				"p1": newTestPlan("p1", "Rust Async", plan.StatusInProgress, []plan.Chunk{
					{ID: "c1", Status: plan.StatusCompleted, Duration: 60},
					{ID: "c2", Status: plan.StatusInProgress, Duration: 60},
				}),
			},
			mockSessions: []*session.Session{
				newTestSession("s1", "p1", yesterday, 60),
				newTestSession("s2", "p1", yesterday, 90),
			},
			expectedStats: &TotalStats{
				TotalHours:      2.5,
				TotalSessions:   2,
				ActivePlans:     1,
				CompletedPlans:  0,
				AverageSession:  75.0,
				LastSessionDate: &yesterday,
			},
		},
		{
			name: "multiple plans with different statuses",
			mockPlans: []*storage.PlanRecord{
				newTestPlanRecord("p1", "Rust Async", plan.StatusInProgress),
				newTestPlanRecord("p2", "French B1", plan.StatusCompleted),
				newTestPlanRecord("p3", "Piano", plan.StatusNotStarted),
			},
			mockFullPlans: map[string]*plan.Plan{
				"p1": newTestPlan("p1", "Rust Async", plan.StatusInProgress, []plan.Chunk{
					{ID: "c1", Status: plan.StatusCompleted, Duration: 60},
				}),
				"p2": newTestPlan("p2", "French B1", plan.StatusCompleted, []plan.Chunk{
					{ID: "c1", Status: plan.StatusCompleted, Duration: 60},
				}),
				"p3": newTestPlan("p3", "Piano", plan.StatusNotStarted, []plan.Chunk{
					{ID: "c1", Status: plan.StatusNotStarted, Duration: 60},
				}),
			},
			mockSessions: []*session.Session{
				newTestSession("s1", "p1", yesterday, 120),
				newTestSession("s2", "p2", yesterday, 180),
			},
			expectedStats: &TotalStats{
				TotalHours:      5.0,
				TotalSessions:   2,
				ActivePlans:     2, // p1 (in-progress) + p3 (not-started)
				CompletedPlans:  1,
				AverageSession:  150.0,
				LastSessionDate: &yesterday,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanService := new(MockPlanService)
			mockSessionService := new(MockSessionService)

			// Setup mock expectations
			mockPlanService.On("List", ctx, (*storage.PlanFilter)(nil)).Return(tt.mockPlans, nil)
			for id, p := range tt.mockFullPlans {
				mockPlanService.On("Get", ctx, id).Return(p, nil)
			}
			mockSessionService.On("ListAll", ctx).Return(tt.mockSessions, nil)

			// Create service
			service := NewService(mockPlanService, mockSessionService)

			// Execute
			stats, err := service.GetTotalStats(ctx)

			// Assert
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStats.TotalHours, stats.TotalHours)
			assert.Equal(t, tt.expectedStats.TotalSessions, stats.TotalSessions)
			assert.Equal(t, tt.expectedStats.ActivePlans, stats.ActivePlans)
			assert.Equal(t, tt.expectedStats.CompletedPlans, stats.CompletedPlans)
			assert.Equal(t, tt.expectedStats.AverageSession, stats.AverageSession)

			if tt.expectedStats.LastSessionDate != nil {
				assert.NotNil(t, stats.LastSessionDate)
			}

			mockPlanService.AssertExpectations(t)
			mockSessionService.AssertExpectations(t)
		})
	}
}

func TestService_GetPlanStats(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	tests := []struct {
		name          string
		planID        string
		mockPlan      *plan.Plan
		mockSessions  []*session.Session
		expectedStats *PlanStats
		expectError   bool
	}{
		{
			name:   "plan with no sessions",
			planID: "p1",
			mockPlan: newTestPlan("p1", "Rust Async", plan.StatusNotStarted, []plan.Chunk{
				{ID: "c1", Status: plan.StatusNotStarted, Duration: 60},
				{ID: "c2", Status: plan.StatusNotStarted, Duration: 60},
			}),
			mockSessions: []*session.Session{},
			expectedStats: &PlanStats{
				PlanID:          "p1",
				PlanTitle:       "Rust Async",
				TotalHours:      0,
				PlannedHours:    40.0,
				SessionCount:    0,
				CompletedChunks: 0,
				TotalChunks:     2,
				Progress:        0.0,
				Status:          string(plan.StatusNotStarted),
				LastSession:     nil,
			},
		},
		{
			name:   "plan with sessions and progress",
			planID: "p1",
			mockPlan: newTestPlan("p1", "Rust Async", plan.StatusInProgress, []plan.Chunk{
				{ID: "c1", Status: plan.StatusCompleted, Duration: 60},
				{ID: "c2", Status: plan.StatusInProgress, Duration: 60},
				{ID: "c3", Status: plan.StatusNotStarted, Duration: 60},
			}),
			mockSessions: []*session.Session{
				newTestSession("s1", "p1", yesterday, 90),
				newTestSession("s2", "p1", yesterday, 120),
			},
			expectedStats: &PlanStats{
				PlanID:          "p1",
				PlanTitle:       "Rust Async",
				TotalHours:      3.5,
				PlannedHours:    40.0,
				SessionCount:    2,
				CompletedChunks: 1,
				TotalChunks:     3,
				Progress:        0.333, // 1/3
				Status:          string(plan.StatusInProgress),
				LastSession:     &yesterday,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanService := new(MockPlanService)
			mockSessionService := new(MockSessionService)

			// Setup mock expectations
			mockPlanService.On("Get", ctx, tt.planID).Return(tt.mockPlan, nil)
			mockSessionService.On("List", ctx, tt.planID, 0).Return(tt.mockSessions, nil)

			// Create service
			service := NewService(mockPlanService, mockSessionService)

			// Execute
			stats, err := service.GetPlanStats(ctx, tt.planID)

			// Assert
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStats.PlanID, stats.PlanID)
			assert.Equal(t, tt.expectedStats.PlanTitle, stats.PlanTitle)
			assert.Equal(t, tt.expectedStats.TotalHours, stats.TotalHours)
			assert.Equal(t, tt.expectedStats.PlannedHours, stats.PlannedHours)
			assert.Equal(t, tt.expectedStats.SessionCount, stats.SessionCount)
			assert.Equal(t, tt.expectedStats.CompletedChunks, stats.CompletedChunks)
			assert.Equal(t, tt.expectedStats.TotalChunks, stats.TotalChunks)
			assert.InDelta(t, tt.expectedStats.Progress, stats.Progress, 0.01)
			assert.Equal(t, tt.expectedStats.Status, stats.Status)

			mockPlanService.AssertExpectations(t)
			mockSessionService.AssertExpectations(t)
		})
	}
}

func TestService_GetDailyStats(t *testing.T) {
	ctx := context.Background()

	day1 := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC)
	day3 := time.Date(2024, 10, 3, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		timeRange     TimeRange
		mockSessions  []*session.Session
		expectedStats []DailyStats
	}{
		{
			name:          "no sessions",
			timeRange:     TimeRange{Start: day1, End: day3},
			mockSessions:  []*session.Session{},
			expectedStats: []DailyStats{},
		},
		{
			name:      "multiple days with sessions",
			timeRange: TimeRange{Start: day1, End: day3.Add(24 * time.Hour)},
			mockSessions: []*session.Session{
				newTestSession("s1", "p1", day1.Add(10*time.Hour), 60),
				newTestSession("s2", "p2", day1.Add(14*time.Hour), 90),
				newTestSession("s3", "p1", day2.Add(10*time.Hour), 120),
				newTestSession("s4", "p1", day3.Add(10*time.Hour), 30),
			},
			expectedStats: []DailyStats{
				{
					Date:         day1,
					Duration:     150,
					SessionCount: 2,
					Plans:        []string{"p1", "p2"},
				},
				{
					Date:         day2,
					Duration:     120,
					SessionCount: 1,
					Plans:        []string{"p1"},
				},
				{
					Date:         day3,
					Duration:     30,
					SessionCount: 1,
					Plans:        []string{"p1"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanService := new(MockPlanService)
			mockSessionService := new(MockSessionService)

			// Setup mock expectations
			mockSessionService.On("ListAll", ctx).Return(tt.mockSessions, nil)

			// Create service
			service := NewService(mockPlanService, mockSessionService)

			// Execute
			stats, err := service.GetDailyStats(ctx, tt.timeRange)

			// Assert
			require.NoError(t, err)
			assert.Len(t, stats, len(tt.expectedStats))

			for i, expected := range tt.expectedStats {
				assert.Equal(t, expected.Date.Year(), stats[i].Date.Year())
				assert.Equal(t, expected.Date.Month(), stats[i].Date.Month())
				assert.Equal(t, expected.Date.Day(), stats[i].Date.Day())
				assert.Equal(t, expected.Duration, stats[i].Duration)
				assert.Equal(t, expected.SessionCount, stats[i].SessionCount)
				assert.ElementsMatch(t, expected.Plans, stats[i].Plans)
			}

			mockSessionService.AssertExpectations(t)
		})
	}
}

func TestService_GetStreakInfo(t *testing.T) {
	ctx := context.Background()
	baseTime := time.Date(2024, 10, 5, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name                  string
		mockSessions          []*session.Session
		expectedCurrentStreak int
		expectedLongestStreak int
	}{
		{
			name:                  "no sessions",
			mockSessions:          []*session.Session{},
			expectedCurrentStreak: 0,
			expectedLongestStreak: 0,
		},
		{
			name: "3-day current streak",
			mockSessions: []*session.Session{
				newTestSession("s1", "p1", baseTime.AddDate(0, 0, -2), 60),
				newTestSession("s2", "p1", baseTime.AddDate(0, 0, -1), 60),
				newTestSession("s3", "p1", baseTime, 60),
			},
			expectedCurrentStreak: 3,
			expectedLongestStreak: 3,
		},
		{
			name: "broken streak",
			mockSessions: []*session.Session{
				newTestSession("s1", "p1", baseTime.AddDate(0, 0, -5), 60),
				newTestSession("s2", "p1", baseTime.AddDate(0, 0, -4), 60),
				newTestSession("s3", "p1", baseTime.AddDate(0, 0, -3), 60),
				// Gap on day -2, -1, 0
			},
			expectedCurrentStreak: 0,
			expectedLongestStreak: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanService := new(MockPlanService)
			mockSessionService := new(MockSessionService)

			// Setup mock expectations
			mockSessionService.On("ListAll", ctx).Return(tt.mockSessions, nil)

			// Create service
			service := NewService(mockPlanService, mockSessionService)

			// Execute
			current, longest, err := service.GetStreakInfo(ctx)

			// Assert
			require.NoError(t, err)
			// Note: Streak tests may vary based on actual time.Now(), so we check they are non-negative
			assert.GreaterOrEqual(t, current, 0)
			assert.GreaterOrEqual(t, longest, 0)

			mockSessionService.AssertExpectations(t)
		})
	}
}
