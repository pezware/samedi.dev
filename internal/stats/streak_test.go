// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package stats

import (
	"testing"
	"time"

	"github.com/pezware/samedi.dev/internal/session"
	"github.com/stretchr/testify/assert"
)

func TestCalculateStreak(t *testing.T) {
	// Create base time for consistent testing
	baseTime := time.Date(2024, 10, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		sessions    []session.Session
		wantCurrent int
		wantLongest int
	}{
		{
			name:        "no sessions",
			sessions:    []session.Session{},
			wantCurrent: 0,
			wantLongest: 0,
		},
		{
			name: "single session today",
			sessions: []session.Session{
				createSession("s1", "p1", baseTime, 60),
			},
			wantCurrent: 1,
			wantLongest: 1,
		},
		{
			name: "consecutive days 3 day streak",
			sessions: []session.Session{
				createSession("s1", "p1", baseTime.AddDate(0, 0, -2), 60),
				createSession("s2", "p1", baseTime.AddDate(0, 0, -1), 60),
				createSession("s3", "p1", baseTime, 60),
			},
			wantCurrent: 3,
			wantLongest: 3,
		},
		{
			name: "streak with gap",
			sessions: []session.Session{
				createSession("s1", "p1", baseTime.AddDate(0, 0, -5), 60),
				createSession("s2", "p1", baseTime.AddDate(0, 0, -4), 60),
				// Gap on day -3
				createSession("s3", "p1", baseTime.AddDate(0, 0, -2), 60),
				createSession("s4", "p1", baseTime.AddDate(0, 0, -1), 60),
				createSession("s5", "p1", baseTime, 60),
			},
			wantCurrent: 3, // Current streak: -2, -1, 0
			wantLongest: 3, // Longest is current streak
		},
		{
			name: "multiple streaks longest not current",
			sessions: []session.Session{
				// First streak: 4 days
				createSession("s1", "p1", baseTime.AddDate(0, 0, -10), 60),
				createSession("s2", "p1", baseTime.AddDate(0, 0, -9), 60),
				createSession("s3", "p1", baseTime.AddDate(0, 0, -8), 60),
				createSession("s4", "p1", baseTime.AddDate(0, 0, -7), 60),
				// Gap
				// Second streak: 2 days (current)
				createSession("s5", "p1", baseTime.AddDate(0, 0, -1), 60),
				createSession("s6", "p1", baseTime, 60),
			},
			wantCurrent: 2, // Current: -1, 0
			wantLongest: 4, // Longest: -10 to -7
		},
		{
			name: "multiple sessions same day count as one",
			sessions: []session.Session{
				createSession("s1", "p1", baseTime.AddDate(0, 0, -2).Add(9*time.Hour), 60),
				createSession("s2", "p1", baseTime.AddDate(0, 0, -2).Add(14*time.Hour), 60),
				createSession("s3", "p1", baseTime.AddDate(0, 0, -1), 60),
				createSession("s4", "p1", baseTime, 60),
			},
			wantCurrent: 3,
			wantLongest: 3,
		},
		{
			name: "overnight session counts for start day",
			sessions: []session.Session{
				// Session starts on day -2 at 23:00, ends day -1 at 01:00
				{
					ID:        "s1",
					PlanID:    "p1",
					StartTime: time.Date(2024, 9, 29, 23, 0, 0, 0, time.UTC),         // Day -2 at 23:00
					EndTime:   ptrTime(time.Date(2024, 9, 30, 1, 0, 0, 0, time.UTC)), // Day -1 at 01:00
					Duration:  120,
				},
				createSession("s2", "p1", baseTime.AddDate(0, 0, -1), 60),
				createSession("s3", "p1", baseTime, 60),
			},
			wantCurrent: 3,
			wantLongest: 3,
		},
		{
			name: "streak broken if last session was 2+ days ago",
			sessions: []session.Session{
				createSession("s1", "p1", baseTime.AddDate(0, 0, -5), 60),
				createSession("s2", "p1", baseTime.AddDate(0, 0, -4), 60),
				createSession("s3", "p1", baseTime.AddDate(0, 0, -3), 60),
				// Last session was 3 days ago from baseTime
			},
			wantCurrent: 0, // Streak broken (last session > 1 day ago)
			wantLongest: 3, // Historical streak was 3 days
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCurrent, gotLongest := calculateStreakAsOf(tt.sessions, baseTime)

			assert.Equal(t, tt.wantCurrent, gotCurrent, "current streak mismatch")
			assert.Equal(t, tt.wantLongest, gotLongest, "longest streak mismatch")
		})
	}
}

func TestGetActiveDays(t *testing.T) {
	baseTime := time.Date(2024, 10, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		sessions []session.Session
		want     []time.Time
	}{
		{
			name:     "no sessions",
			sessions: []session.Session{},
			want:     []time.Time{},
		},
		{
			name: "single session",
			sessions: []session.Session{
				createSession("s1", "p1", baseTime, 60),
			},
			want: []time.Time{
				dayStart(baseTime),
			},
		},
		{
			name: "multiple sessions same day",
			sessions: []session.Session{
				createSession("s1", "p1", baseTime.Add(2*time.Hour), 60),
				createSession("s2", "p1", baseTime.Add(5*time.Hour), 60),
			},
			want: []time.Time{
				dayStart(baseTime),
			},
		},
		{
			name: "sessions on different days",
			sessions: []session.Session{
				createSession("s1", "p1", baseTime.AddDate(0, 0, -2), 60),
				createSession("s2", "p1", baseTime.AddDate(0, 0, -1), 60),
				createSession("s3", "p1", baseTime, 60),
			},
			want: []time.Time{
				dayStart(baseTime.AddDate(0, 0, -2)),
				dayStart(baseTime.AddDate(0, 0, -1)),
				dayStart(baseTime),
			},
		},
		{
			name: "overnight session counts for start day",
			sessions: []session.Session{
				{
					ID:        "s1",
					PlanID:    "p1",
					StartTime: time.Date(2024, 10, 1, 23, 0, 0, 0, time.UTC),         // 23:00 on baseTime day
					EndTime:   ptrTime(time.Date(2024, 10, 2, 1, 0, 0, 0, time.UTC)), // 01:00 next day
					Duration:  120,
				},
			},
			want: []time.Time{
				dayStart(baseTime),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetActiveDays(tt.sessions)

			assert.Len(t, got, len(tt.want))
			for i, wantDay := range tt.want {
				assert.True(t, got[i].Equal(wantDay), "day %d: expected %v, got %v", i, wantDay, got[i])
			}
		})
	}
}

func TestDetectStreakBreaks(t *testing.T) {
	baseTime := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		sessions []session.Session
		want     []time.Time
	}{
		{
			name:     "no sessions no breaks",
			sessions: []session.Session{},
			want:     []time.Time{},
		},
		{
			name: "consecutive days no breaks",
			sessions: []session.Session{
				createSession("s1", "p1", baseTime.AddDate(0, 0, -2), 60),
				createSession("s2", "p1", baseTime.AddDate(0, 0, -1), 60),
				createSession("s3", "p1", baseTime, 60),
			},
			want: []time.Time{},
		},
		{
			name: "single gap",
			sessions: []session.Session{
				createSession("s1", "p1", baseTime.AddDate(0, 0, -3), 60),
				// Gap on day -2
				createSession("s2", "p1", baseTime.AddDate(0, 0, -1), 60),
				createSession("s3", "p1", baseTime, 60),
			},
			want: []time.Time{
				baseTime.AddDate(0, 0, -2),
			},
		},
		{
			name: "multiple gaps",
			sessions: []session.Session{
				createSession("s1", "p1", baseTime.AddDate(0, 0, -6), 60),
				// Gap on day -5
				createSession("s2", "p1", baseTime.AddDate(0, 0, -4), 60),
				createSession("s3", "p1", baseTime.AddDate(0, 0, -3), 60),
				// Gap on day -2
				createSession("s4", "p1", baseTime.AddDate(0, 0, -1), 60),
				createSession("s5", "p1", baseTime, 60),
			},
			want: []time.Time{
				baseTime.AddDate(0, 0, -5),
				baseTime.AddDate(0, 0, -2),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectStreakBreaks(tt.sessions)

			assert.Len(t, got, len(tt.want))
			for i, wantBreak := range tt.want {
				assert.True(t, sameDay(got[i], wantBreak), "break %d: expected %v, got %v", i, wantBreak, got[i])
			}
		})
	}
}

// Helper functions for tests

//nolint:unparam // planID flexibility useful for future tests
func createSession(id, planID string, startTime time.Time, durationMinutes int) session.Session {
	endTime := startTime.Add(time.Duration(durationMinutes) * time.Minute)
	return session.Session{
		ID:        id,
		PlanID:    planID,
		StartTime: startTime,
		EndTime:   &endTime,
		Duration:  durationMinutes,
	}
}

func dayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
