// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTotalStats_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		stats   TotalStats
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid stats",
			stats: TotalStats{
				TotalHours:      10.5,
				TotalSessions:   5,
				ActivePlans:     2,
				CompletedPlans:  1,
				CurrentStreak:   3,
				LongestStreak:   5,
				AverageSession:  126.0,
				LastSessionDate: &now,
			},
			wantErr: false,
		},
		{
			name: "negative total hours",
			stats: TotalStats{
				TotalHours:    -1.0,
				TotalSessions: 5,
			},
			wantErr: true,
			errMsg:  "total hours cannot be negative",
		},
		{
			name: "negative total sessions",
			stats: TotalStats{
				TotalHours:    10.0,
				TotalSessions: -1,
			},
			wantErr: true,
			errMsg:  "total sessions cannot be negative",
		},
		{
			name: "negative active plans",
			stats: TotalStats{
				ActivePlans: -1,
			},
			wantErr: true,
			errMsg:  "active plans cannot be negative",
		},
		{
			name: "negative completed plans",
			stats: TotalStats{
				CompletedPlans: -1,
			},
			wantErr: true,
			errMsg:  "completed plans cannot be negative",
		},
		{
			name: "negative current streak",
			stats: TotalStats{
				CurrentStreak: -1,
			},
			wantErr: true,
			errMsg:  "current streak cannot be negative",
		},
		{
			name: "negative longest streak",
			stats: TotalStats{
				LongestStreak: -1,
			},
			wantErr: true,
			errMsg:  "longest streak cannot be negative",
		},
		{
			name: "current streak exceeds longest streak",
			stats: TotalStats{
				CurrentStreak: 10,
				LongestStreak: 5,
			},
			wantErr: true,
			errMsg:  "current streak (10) cannot exceed longest streak (5)",
		},
		{
			name: "negative average session",
			stats: TotalStats{
				AverageSession: -1.0,
			},
			wantErr: true,
			errMsg:  "average session cannot be negative",
		},
		{
			name: "sessions exist but zero average",
			stats: TotalStats{
				TotalSessions:  5,
				AverageSession: 0,
			},
			wantErr: true,
			errMsg:  "average session should be > 0 when sessions exist",
		},
		{
			name: "no sessions but positive hours",
			stats: TotalStats{
				TotalSessions: 0,
				TotalHours:    5.0,
			},
			wantErr: true,
			errMsg:  "total hours should be 0 when no sessions exist",
		},
		{
			name: "zero values (valid empty state)",
			stats: TotalStats{
				TotalHours:     0,
				TotalSessions:  0,
				ActivePlans:    0,
				CompletedPlans: 0,
				CurrentStreak:  0,
				LongestStreak:  0,
				AverageSession: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.stats.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanStats_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		stats   PlanStats
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid plan stats",
			stats: PlanStats{
				PlanID:          "rust-async",
				PlanTitle:       "Rust Async Programming",
				TotalHours:      10.5,
				PlannedHours:    40.0,
				SessionCount:    12,
				CompletedChunks: 10,
				TotalChunks:     40,
				Progress:        0.25,
				Status:          "in-progress",
				LastSession:     &now,
			},
			wantErr: false,
		},
		{
			name: "empty plan ID",
			stats: PlanStats{
				PlanTitle: "Test",
				Status:    "active",
			},
			wantErr: true,
			errMsg:  "plan ID cannot be empty",
		},
		{
			name: "empty plan title",
			stats: PlanStats{
				PlanID: "test",
				Status: "active",
			},
			wantErr: true,
			errMsg:  "plan title cannot be empty",
		},
		{
			name: "negative total hours",
			stats: PlanStats{
				PlanID:     "test",
				PlanTitle:  "Test",
				TotalHours: -1.0,
				Status:     "active",
			},
			wantErr: true,
			errMsg:  "total hours cannot be negative",
		},
		{
			name: "negative planned hours",
			stats: PlanStats{
				PlanID:       "test",
				PlanTitle:    "Test",
				PlannedHours: -1.0,
				Status:       "active",
			},
			wantErr: true,
			errMsg:  "planned hours cannot be negative",
		},
		{
			name: "negative session count",
			stats: PlanStats{
				PlanID:       "test",
				PlanTitle:    "Test",
				SessionCount: -1,
				Status:       "active",
			},
			wantErr: true,
			errMsg:  "session count cannot be negative",
		},
		{
			name: "negative completed chunks",
			stats: PlanStats{
				PlanID:          "test",
				PlanTitle:       "Test",
				CompletedChunks: -1,
				Status:          "active",
			},
			wantErr: true,
			errMsg:  "completed chunks cannot be negative",
		},
		{
			name: "negative total chunks",
			stats: PlanStats{
				PlanID:      "test",
				PlanTitle:   "Test",
				TotalChunks: -1,
				Status:      "active",
			},
			wantErr: true,
			errMsg:  "total chunks cannot be negative",
		},
		{
			name: "completed exceeds total chunks",
			stats: PlanStats{
				PlanID:          "test",
				PlanTitle:       "Test",
				CompletedChunks: 10,
				TotalChunks:     5,
				Status:          "active",
			},
			wantErr: true,
			errMsg:  "completed chunks (10) cannot exceed total chunks (5)",
		},
		{
			name: "progress below 0",
			stats: PlanStats{
				PlanID:    "test",
				PlanTitle: "Test",
				Progress:  -0.1,
				Status:    "active",
			},
			wantErr: true,
			errMsg:  "progress must be between 0 and 1",
		},
		{
			name: "progress above 1",
			stats: PlanStats{
				PlanID:    "test",
				PlanTitle: "Test",
				Progress:  1.5,
				Status:    "active",
			},
			wantErr: true,
			errMsg:  "progress must be between 0 and 1",
		},
		{
			name: "empty status",
			stats: PlanStats{
				PlanID:    "test",
				PlanTitle: "Test",
			},
			wantErr: true,
			errMsg:  "status cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.stats.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanStats_ProgressPercent(t *testing.T) {
	tests := []struct {
		name     string
		progress float64
		want     int
	}{
		{"0%", 0.0, 0},
		{"25%", 0.25, 25},
		{"50%", 0.50, 50},
		{"75%", 0.75, 75},
		{"100%", 1.0, 100},
		{"33%", 0.33, 33},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := PlanStats{Progress: tt.progress}
			assert.Equal(t, tt.want, ps.ProgressPercent())
		})
	}
}

func TestDailyStats_Validate(t *testing.T) {
	date := time.Date(2024, 10, 8, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		stats   DailyStats
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid daily stats",
			stats: DailyStats{
				Date:         date,
				Duration:     120,
				SessionCount: 2,
				Plans:        []string{"plan-1", "plan-2"},
			},
			wantErr: false,
		},
		{
			name: "zero date",
			stats: DailyStats{
				Duration:     60,
				SessionCount: 1,
			},
			wantErr: true,
			errMsg:  "date cannot be zero",
		},
		{
			name: "negative duration",
			stats: DailyStats{
				Date:         date,
				Duration:     -1,
				SessionCount: 1,
			},
			wantErr: true,
			errMsg:  "duration cannot be negative",
		},
		{
			name: "negative session count",
			stats: DailyStats{
				Date:         date,
				Duration:     60,
				SessionCount: -1,
			},
			wantErr: true,
			errMsg:  "session count cannot be negative",
		},
		{
			name: "sessions exist but zero duration",
			stats: DailyStats{
				Date:         date,
				Duration:     0,
				SessionCount: 1,
			},
			wantErr: true,
			errMsg:  "duration should be > 0 when sessions exist",
		},
		{
			name: "duration exists but zero sessions",
			stats: DailyStats{
				Date:         date,
				Duration:     60,
				SessionCount: 0,
			},
			wantErr: true,
			errMsg:  "session count should be > 0 when duration exists",
		},
		{
			name: "zero values (valid empty day)",
			stats: DailyStats{
				Date:         date,
				Duration:     0,
				SessionCount: 0,
				Plans:        []string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.stats.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDailyStats_Hours(t *testing.T) {
	tests := []struct {
		name     string
		duration int
		want     float64
	}{
		{"60 minutes", 60, 1.0},
		{"90 minutes", 90, 1.5},
		{"120 minutes", 120, 2.0},
		{"30 minutes", 30, 0.5},
		{"0 minutes", 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := DailyStats{Duration: tt.duration}
			assert.Equal(t, tt.want, ds.Hours())
		})
	}
}

func TestTimeRange_Validate(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	tests := []struct {
		name    string
		tr      TimeRange
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid range",
			tr:      TimeRange{Start: yesterday, End: now},
			wantErr: false,
		},
		{
			name:    "zero start time",
			tr:      TimeRange{Start: time.Time{}, End: now},
			wantErr: true,
			errMsg:  "start time cannot be zero",
		},
		{
			name:    "zero end time",
			tr:      TimeRange{Start: now, End: time.Time{}},
			wantErr: true,
			errMsg:  "end time cannot be zero",
		},
		{
			name:    "end before start",
			tr:      TimeRange{Start: tomorrow, End: yesterday},
			wantErr: true,
			errMsg:  "end time cannot be before start time",
		},
		{
			name:    "same start and end (valid)",
			tr:      TimeRange{Start: now, End: now},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tr.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTimeRange_Contains(t *testing.T) {
	start := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 10, 31, 23, 59, 59, 0, time.UTC)
	tr := TimeRange{Start: start, End: end}

	tests := []struct {
		name string
		t    time.Time
		want bool
	}{
		{
			name: "time before range",
			t:    time.Date(2024, 9, 30, 12, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "time at start",
			t:    start,
			want: true,
		},
		{
			name: "time in middle",
			t:    time.Date(2024, 10, 15, 12, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "time at end",
			t:    end,
			want: true,
		},
		{
			name: "time after range",
			t:    time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tr.Contains(tt.t))
		})
	}
}

func TestTimeRange_Helpers(t *testing.T) {
	t.Run("NewTimeRangeToday", func(t *testing.T) {
		tr := NewTimeRangeToday()
		assert.NoError(t, tr.Validate())

		// Should start at midnight today
		now := time.Now()
		expectedStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		assert.Equal(t, expectedStart.Year(), tr.Start.Year())
		assert.Equal(t, expectedStart.Month(), tr.Start.Month())
		assert.Equal(t, expectedStart.Day(), tr.Start.Day())
		assert.Equal(t, 0, tr.Start.Hour())

		// End should be now
		assert.WithinDuration(t, now, tr.End, time.Second)
	})

	t.Run("NewTimeRangeThisWeek", func(t *testing.T) {
		tr := NewTimeRangeThisWeek()
		assert.NoError(t, tr.Validate())

		// Start should be Monday at midnight
		assert.Equal(t, time.Monday, tr.Start.Weekday())
		assert.Equal(t, 0, tr.Start.Hour())
		assert.Equal(t, 0, tr.Start.Minute())

		// End should be now
		assert.WithinDuration(t, time.Now(), tr.End, time.Second)
	})

	t.Run("NewTimeRangeThisMonth", func(t *testing.T) {
		tr := NewTimeRangeThisMonth()
		assert.NoError(t, tr.Validate())

		// Start should be 1st of current month at midnight
		now := time.Now()
		assert.Equal(t, now.Year(), tr.Start.Year())
		assert.Equal(t, now.Month(), tr.Start.Month())
		assert.Equal(t, 1, tr.Start.Day())
		assert.Equal(t, 0, tr.Start.Hour())

		// End should be now
		assert.WithinDuration(t, now, tr.End, time.Second)
	})

	t.Run("NewTimeRangeSince", func(t *testing.T) {
		since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		tr := NewTimeRangeSince(since)
		assert.NoError(t, tr.Validate())

		assert.Equal(t, since, tr.Start)
		assert.WithinDuration(t, time.Now(), tr.End, time.Second)
	})

	t.Run("NewTimeRangeAll", func(t *testing.T) {
		tr := NewTimeRangeAll()
		assert.NoError(t, tr.Validate())

		// Start should be epoch
		assert.Equal(t, int64(0), tr.Start.Unix())

		// End should be now
		assert.WithinDuration(t, time.Now(), tr.End, time.Second)
	})
}
