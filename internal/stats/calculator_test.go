// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package stats

import (
	"testing"
	"time"

	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/session"
	"github.com/stretchr/testify/assert"
)

func TestCalculateTotalStats(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	twoDaysAgo := now.AddDate(0, 0, -2)

	tests := []struct {
		name     string
		sessions []session.Session
		plans    []plan.Plan
		want     TotalStats
	}{
		{
			name:     "empty sessions and plans",
			sessions: []session.Session{},
			plans:    []plan.Plan{},
			want: TotalStats{
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
			name:     "plans exist but no sessions (bug fix test)",
			sessions: []session.Session{},
			plans: []plan.Plan{
				{ID: "p1", Status: plan.StatusNotStarted},
				{ID: "p2", Status: plan.StatusInProgress},
				{ID: "p3", Status: plan.StatusCompleted},
			},
			want: TotalStats{
				TotalHours:     0,
				TotalSessions:  0,
				ActivePlans:    2, // p1 (not-started) + p2 (in-progress)
				CompletedPlans: 1, // p3
				CurrentStreak:  0,
				LongestStreak:  0,
				AverageSession: 0,
			},
		},
		{
			name: "single session",
			sessions: []session.Session{
				{
					ID:        "s1",
					PlanID:    "p1",
					StartTime: yesterday,
					EndTime:   ptrTime(yesterday.Add(60 * time.Minute)),
					Duration:  60,
				},
			},
			plans: []plan.Plan{
				{ID: "p1", Status: plan.StatusInProgress},
			},
			want: TotalStats{
				TotalHours:      1.0,
				TotalSessions:   1,
				ActivePlans:     1,
				CompletedPlans:  0,
				AverageSession:  60.0,
				LastSessionDate: &yesterday,
			},
		},
		{
			name: "multiple sessions same plan",
			sessions: []session.Session{
				{
					ID:        "s1",
					PlanID:    "p1",
					StartTime: twoDaysAgo,
					EndTime:   ptrTime(twoDaysAgo.Add(90 * time.Minute)),
					Duration:  90,
				},
				{
					ID:        "s2",
					PlanID:    "p1",
					StartTime: yesterday,
					EndTime:   ptrTime(yesterday.Add(30 * time.Minute)),
					Duration:  30,
				},
			},
			plans: []plan.Plan{
				{ID: "p1", Status: plan.StatusInProgress},
			},
			want: TotalStats{
				TotalHours:      2.0,
				TotalSessions:   2,
				ActivePlans:     1,
				CompletedPlans:  0,
				AverageSession:  60.0, // (90 + 30) / 2
				LastSessionDate: &yesterday,
			},
		},
		{
			name: "multiple plans with different statuses",
			sessions: []session.Session{
				{
					ID:        "s1",
					PlanID:    "p1",
					StartTime: yesterday,
					EndTime:   ptrTime(yesterday.Add(60 * time.Minute)),
					Duration:  60,
				},
				{
					ID:        "s2",
					PlanID:    "p2",
					StartTime: yesterday,
					EndTime:   ptrTime(yesterday.Add(120 * time.Minute)),
					Duration:  120,
				},
			},
			plans: []plan.Plan{
				{ID: "p1", Status: plan.StatusInProgress},
				{ID: "p2", Status: plan.StatusCompleted},
				{ID: "p3", Status: plan.StatusNotStarted},
			},
			want: TotalStats{
				TotalHours:      3.0,
				TotalSessions:   2,
				ActivePlans:     2, // p1 (in-progress) + p3 (not-started)
				CompletedPlans:  1,
				AverageSession:  90.0, // (60 + 120) / 2
				LastSessionDate: &yesterday,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTotalStats(tt.sessions, tt.plans)

			assert.Equal(t, tt.want.TotalHours, got.TotalHours)
			assert.Equal(t, tt.want.TotalSessions, got.TotalSessions)
			assert.Equal(t, tt.want.ActivePlans, got.ActivePlans)
			assert.Equal(t, tt.want.CompletedPlans, got.CompletedPlans)
			assert.Equal(t, tt.want.AverageSession, got.AverageSession)

			if tt.want.LastSessionDate != nil {
				assert.NotNil(t, got.LastSessionDate)
				assert.WithinDuration(t, *tt.want.LastSessionDate, *got.LastSessionDate, time.Second)
			} else {
				assert.Nil(t, got.LastSessionDate)
			}
		})
	}
}

func TestCalculatePlanStats(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	tests := []struct {
		name     string
		planID   string
		sessions []session.Session
		p        plan.Plan
		want     PlanStats
	}{
		{
			name:     "no sessions for plan",
			planID:   "p1",
			sessions: []session.Session{},
			p: plan.Plan{
				ID:         "p1",
				Title:      "Test Plan",
				TotalHours: 10.0,
				Status:     plan.StatusNotStarted,
				Chunks: []plan.Chunk{
					{ID: "c1", Status: plan.StatusNotStarted},
					{ID: "c2", Status: plan.StatusNotStarted},
				},
			},
			want: PlanStats{
				PlanID:          "p1",
				PlanTitle:       "Test Plan",
				TotalHours:      0,
				PlannedHours:    10.0,
				SessionCount:    0,
				CompletedChunks: 0,
				TotalChunks:     2,
				Progress:        0.0,
				Status:          string(plan.StatusNotStarted),
			},
		},
		{
			name:   "plan with sessions and progress",
			planID: "p1",
			sessions: []session.Session{
				{
					ID:        "s1",
					PlanID:    "p1",
					StartTime: yesterday,
					EndTime:   ptrTime(yesterday.Add(60 * time.Minute)),
					Duration:  60,
				},
				{
					ID:        "s2",
					PlanID:    "p1",
					StartTime: yesterday,
					EndTime:   ptrTime(yesterday.Add(90 * time.Minute)),
					Duration:  90,
				},
			},
			p: plan.Plan{
				ID:         "p1",
				Title:      "Rust Async",
				TotalHours: 40.0,
				Status:     plan.StatusInProgress,
				Chunks: []plan.Chunk{
					{ID: "c1", Status: plan.StatusCompleted},
					{ID: "c2", Status: plan.StatusCompleted},
					{ID: "c3", Status: plan.StatusInProgress},
					{ID: "c4", Status: plan.StatusNotStarted},
				},
			},
			want: PlanStats{
				PlanID:          "p1",
				PlanTitle:       "Rust Async",
				TotalHours:      2.5,
				PlannedHours:    40.0,
				SessionCount:    2,
				CompletedChunks: 2,
				TotalChunks:     4,
				Progress:        0.5, // 2 / 4
				Status:          string(plan.StatusInProgress),
				LastSession:     &yesterday,
			},
		},
		{
			name:   "completed plan",
			planID: "p1",
			sessions: []session.Session{
				{
					ID:        "s1",
					PlanID:    "p1",
					StartTime: yesterday,
					EndTime:   ptrTime(yesterday.Add(180 * time.Minute)),
					Duration:  180,
				},
			},
			p: plan.Plan{
				ID:         "p1",
				Title:      "French B1",
				TotalHours: 50.0,
				Status:     plan.StatusCompleted,
				Chunks: []plan.Chunk{
					{ID: "c1", Status: plan.StatusCompleted},
					{ID: "c2", Status: plan.StatusCompleted},
				},
			},
			want: PlanStats{
				PlanID:          "p1",
				PlanTitle:       "French B1",
				TotalHours:      3.0,
				PlannedHours:    50.0,
				SessionCount:    1,
				CompletedChunks: 2,
				TotalChunks:     2,
				Progress:        1.0, // 2 / 2
				Status:          string(plan.StatusCompleted),
				LastSession:     &yesterday,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculatePlanStats(tt.planID, tt.sessions, &tt.p)

			assert.Equal(t, tt.want.PlanID, got.PlanID)
			assert.Equal(t, tt.want.PlanTitle, got.PlanTitle)
			assert.Equal(t, tt.want.TotalHours, got.TotalHours)
			assert.Equal(t, tt.want.PlannedHours, got.PlannedHours)
			assert.Equal(t, tt.want.SessionCount, got.SessionCount)
			assert.Equal(t, tt.want.CompletedChunks, got.CompletedChunks)
			assert.Equal(t, tt.want.TotalChunks, got.TotalChunks)
			assert.Equal(t, tt.want.Progress, got.Progress)
			assert.Equal(t, tt.want.Status, got.Status)

			if tt.want.LastSession != nil {
				assert.NotNil(t, got.LastSession)
				assert.WithinDuration(t, *tt.want.LastSession, *got.LastSession, time.Second)
			} else {
				assert.Nil(t, got.LastSession)
			}
		})
	}
}

func TestCalculateDailyStats(t *testing.T) {
	// Create specific dates for testing
	day1 := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC)
	day3 := time.Date(2024, 10, 3, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		sessions  []session.Session
		timeRange TimeRange
		want      []DailyStats
	}{
		{
			name:      "no sessions",
			sessions:  []session.Session{},
			timeRange: TimeRange{Start: day1, End: day3},
			want:      []DailyStats{},
		},
		{
			name: "single day single session",
			sessions: []session.Session{
				{
					ID:        "s1",
					PlanID:    "p1",
					StartTime: day1.Add(10 * time.Hour),
					EndTime:   ptrTime(day1.Add(11 * time.Hour)),
					Duration:  60,
				},
			},
			timeRange: TimeRange{Start: day1, End: day1.Add(24 * time.Hour)},
			want: []DailyStats{
				{
					Date:         day1,
					Duration:     60,
					SessionCount: 1,
					Plans:        []string{"p1"},
				},
			},
		},
		{
			name: "multiple sessions same day",
			sessions: []session.Session{
				{
					ID:        "s1",
					PlanID:    "p1",
					StartTime: day1.Add(10 * time.Hour),
					EndTime:   ptrTime(day1.Add(11 * time.Hour)),
					Duration:  60,
				},
				{
					ID:        "s2",
					PlanID:    "p2",
					StartTime: day1.Add(14 * time.Hour),
					EndTime:   ptrTime(day1.Add(15 * time.Hour)),
					Duration:  60,
				},
			},
			timeRange: TimeRange{Start: day1, End: day1.Add(24 * time.Hour)},
			want: []DailyStats{
				{
					Date:         day1,
					Duration:     120,
					SessionCount: 2,
					Plans:        []string{"p1", "p2"},
				},
			},
		},
		{
			name: "multiple days with sessions",
			sessions: []session.Session{
				{
					ID:        "s1",
					PlanID:    "p1",
					StartTime: day1.Add(10 * time.Hour),
					EndTime:   ptrTime(day1.Add(11 * time.Hour)),
					Duration:  60,
				},
				{
					ID:        "s2",
					PlanID:    "p1",
					StartTime: day2.Add(10 * time.Hour),
					EndTime:   ptrTime(day2.Add(12 * time.Hour)),
					Duration:  120,
				},
				{
					ID:        "s3",
					PlanID:    "p2",
					StartTime: day3.Add(10 * time.Hour),
					EndTime:   ptrTime(day3.Add(10*time.Hour + 30*time.Minute)),
					Duration:  30,
				},
			},
			timeRange: TimeRange{Start: day1, End: day3.Add(24 * time.Hour)},
			want: []DailyStats{
				{
					Date:         day1,
					Duration:     60,
					SessionCount: 1,
					Plans:        []string{"p1"},
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
					Plans:        []string{"p2"},
				},
			},
		},
		{
			name: "sessions outside time range excluded",
			sessions: []session.Session{
				{
					ID:        "s1",
					PlanID:    "p1",
					StartTime: day1.Add(-24 * time.Hour), // Before range
					EndTime:   ptrTime(day1.Add(-23 * time.Hour)),
					Duration:  60,
				},
				{
					ID:        "s2",
					PlanID:    "p1",
					StartTime: day2.Add(10 * time.Hour), // In range
					EndTime:   ptrTime(day2.Add(11 * time.Hour)),
					Duration:  60,
				},
				{
					ID:        "s3",
					PlanID:    "p1",
					StartTime: day3.Add(25 * time.Hour), // After range
					EndTime:   ptrTime(day3.Add(26 * time.Hour)),
					Duration:  60,
				},
			},
			timeRange: TimeRange{Start: day1, End: day3},
			want: []DailyStats{
				{
					Date:         day2,
					Duration:     60,
					SessionCount: 1,
					Plans:        []string{"p1"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateDailyStats(tt.sessions, tt.timeRange)

			assert.Len(t, got, len(tt.want))
			for i, wantDay := range tt.want {
				assert.Equal(t, wantDay.Date.Year(), got[i].Date.Year())
				assert.Equal(t, wantDay.Date.Month(), got[i].Date.Month())
				assert.Equal(t, wantDay.Date.Day(), got[i].Date.Day())
				assert.Equal(t, wantDay.Duration, got[i].Duration)
				assert.Equal(t, wantDay.SessionCount, got[i].SessionCount)
				assert.ElementsMatch(t, wantDay.Plans, got[i].Plans)
			}
		})
	}
}

func TestAggregateByPlan(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)

	sessions := []session.Session{
		{
			ID:        "s1",
			PlanID:    "p1",
			StartTime: yesterday,
			EndTime:   ptrTime(yesterday.Add(60 * time.Minute)),
			Duration:  60,
		},
		{
			ID:        "s2",
			PlanID:    "p1",
			StartTime: yesterday,
			EndTime:   ptrTime(yesterday.Add(90 * time.Minute)),
			Duration:  90,
		},
		{
			ID:        "s3",
			PlanID:    "p2",
			StartTime: yesterday,
			EndTime:   ptrTime(yesterday.Add(120 * time.Minute)),
			Duration:  120,
		},
	}

	plans := []plan.Plan{
		{
			ID:         "p1",
			Title:      "Plan 1",
			TotalHours: 10.0,
			Status:     plan.StatusInProgress,
			Chunks: []plan.Chunk{
				{ID: "c1", Status: plan.StatusCompleted},
				{ID: "c2", Status: plan.StatusNotStarted},
			},
		},
		{
			ID:         "p2",
			Title:      "Plan 2",
			TotalHours: 20.0,
			Status:     plan.StatusCompleted,
			Chunks: []plan.Chunk{
				{ID: "c1", Status: plan.StatusCompleted},
			},
		},
	}

	result := AggregateByPlan(sessions, plans)

	assert.Len(t, result, 2)

	// Check p1 stats
	assert.Contains(t, result, "p1")
	p1Stats := result["p1"]
	assert.Equal(t, "p1", p1Stats.PlanID)
	assert.Equal(t, "Plan 1", p1Stats.PlanTitle)
	assert.Equal(t, 2.5, p1Stats.TotalHours) // 150 minutes / 60
	assert.Equal(t, 2, p1Stats.SessionCount)
	assert.Equal(t, 1, p1Stats.CompletedChunks)
	assert.Equal(t, 2, p1Stats.TotalChunks)
	assert.Equal(t, 0.5, p1Stats.Progress)

	// Check p2 stats
	assert.Contains(t, result, "p2")
	p2Stats := result["p2"]
	assert.Equal(t, "p2", p2Stats.PlanID)
	assert.Equal(t, "Plan 2", p2Stats.PlanTitle)
	assert.Equal(t, 2.0, p2Stats.TotalHours) // 120 minutes / 60
	assert.Equal(t, 1, p2Stats.SessionCount)
	assert.Equal(t, 1, p2Stats.CompletedChunks)
	assert.Equal(t, 1, p2Stats.TotalChunks)
	assert.Equal(t, 1.0, p2Stats.Progress)
}

// Helper function to create time pointers
func ptrTime(t time.Time) *time.Time {
	return &t
}
