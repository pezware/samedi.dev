// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package stats

import (
	"fmt"
	"time"
)

// TotalStats represents aggregate statistics across all learning activity.
type TotalStats struct {
	TotalHours      float64    `json:"total_hours"`                 // Total learning time in hours
	TotalSessions   int        `json:"total_sessions"`              // Number of sessions
	ActivePlans     int        `json:"active_plans"`                // Plans in progress
	CompletedPlans  int        `json:"completed_plans"`             // Plans completed
	CurrentStreak   int        `json:"current_streak"`              // Consecutive days with activity
	LongestStreak   int        `json:"longest_streak"`              // Longest streak achieved
	AverageSession  float64    `json:"average_session"`             // Average session duration in minutes
	LastSessionDate *time.Time `json:"last_session_date,omitempty"` // Most recent session
}

// Validate checks if total stats have valid values.
func (ts *TotalStats) Validate() error {
	if ts.TotalHours < 0 {
		return fmt.Errorf("total hours cannot be negative: %.2f", ts.TotalHours)
	}
	if ts.TotalSessions < 0 {
		return fmt.Errorf("total sessions cannot be negative: %d", ts.TotalSessions)
	}
	if ts.ActivePlans < 0 {
		return fmt.Errorf("active plans cannot be negative: %d", ts.ActivePlans)
	}
	if ts.CompletedPlans < 0 {
		return fmt.Errorf("completed plans cannot be negative: %d", ts.CompletedPlans)
	}
	if ts.CurrentStreak < 0 {
		return fmt.Errorf("current streak cannot be negative: %d", ts.CurrentStreak)
	}
	if ts.LongestStreak < 0 {
		return fmt.Errorf("longest streak cannot be negative: %d", ts.LongestStreak)
	}
	if ts.CurrentStreak > ts.LongestStreak && ts.LongestStreak > 0 {
		return fmt.Errorf("current streak (%d) cannot exceed longest streak (%d)", ts.CurrentStreak, ts.LongestStreak)
	}
	if ts.AverageSession < 0 {
		return fmt.Errorf("average session cannot be negative: %.2f", ts.AverageSession)
	}
	// If there are sessions, average should be > 0
	if ts.TotalSessions > 0 && ts.AverageSession == 0 {
		return fmt.Errorf("average session should be > 0 when sessions exist")
	}
	// If no sessions, hours should be 0
	if ts.TotalSessions == 0 && ts.TotalHours > 0 {
		return fmt.Errorf("total hours should be 0 when no sessions exist")
	}

	return nil
}

// PlanStats represents statistics for a specific learning plan.
type PlanStats struct {
	PlanID          string     `json:"plan_id"`
	PlanTitle       string     `json:"plan_title"`
	TotalHours      float64    `json:"total_hours"`            // Time spent on this plan
	PlannedHours    float64    `json:"planned_hours"`          // Total planned hours
	SessionCount    int        `json:"session_count"`          // Number of sessions
	CompletedChunks int        `json:"completed_chunks"`       // Chunks completed
	TotalChunks     int        `json:"total_chunks"`           // Total chunks in plan
	Progress        float64    `json:"progress"`               // Completion percentage (0.0-1.0)
	Status          string     `json:"status"`                 // Plan status
	LastSession     *time.Time `json:"last_session,omitempty"` // Most recent session
}

// Validate checks if plan stats have valid values.
func (ps *PlanStats) Validate() error {
	if ps.PlanID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}
	if ps.PlanTitle == "" {
		return fmt.Errorf("plan title cannot be empty")
	}
	if ps.TotalHours < 0 {
		return fmt.Errorf("total hours cannot be negative: %.2f", ps.TotalHours)
	}
	if ps.PlannedHours < 0 {
		return fmt.Errorf("planned hours cannot be negative: %.2f", ps.PlannedHours)
	}
	if ps.SessionCount < 0 {
		return fmt.Errorf("session count cannot be negative: %d", ps.SessionCount)
	}
	if ps.CompletedChunks < 0 {
		return fmt.Errorf("completed chunks cannot be negative: %d", ps.CompletedChunks)
	}
	if ps.TotalChunks < 0 {
		return fmt.Errorf("total chunks cannot be negative: %d", ps.TotalChunks)
	}
	if ps.CompletedChunks > ps.TotalChunks {
		return fmt.Errorf("completed chunks (%d) cannot exceed total chunks (%d)", ps.CompletedChunks, ps.TotalChunks)
	}
	if ps.Progress < 0 || ps.Progress > 1.0 {
		return fmt.Errorf("progress must be between 0 and 1, got %.2f", ps.Progress)
	}
	if ps.Status == "" {
		return fmt.Errorf("status cannot be empty")
	}

	return nil
}

// ProgressPercent returns the completion percentage as an integer (0-100).
func (ps *PlanStats) ProgressPercent() int {
	return int(ps.Progress * 100)
}

// DailyStats represents statistics for a single day.
type DailyStats struct {
	Date         time.Time `json:"date"`          // The day (time set to midnight)
	Duration     int       `json:"duration"`      // Total minutes for the day
	SessionCount int       `json:"session_count"` // Number of sessions
	Plans        []string  `json:"plans"`         // Plan IDs worked on
}

// Validate checks if daily stats have valid values.
func (ds *DailyStats) Validate() error {
	if ds.Date.IsZero() {
		return fmt.Errorf("date cannot be zero")
	}
	if ds.Duration < 0 {
		return fmt.Errorf("duration cannot be negative: %d", ds.Duration)
	}
	if ds.SessionCount < 0 {
		return fmt.Errorf("session count cannot be negative: %d", ds.SessionCount)
	}
	if ds.SessionCount > 0 && ds.Duration == 0 {
		return fmt.Errorf("duration should be > 0 when sessions exist")
	}
	if ds.SessionCount == 0 && ds.Duration > 0 {
		return fmt.Errorf("session count should be > 0 when duration exists")
	}

	return nil
}

// Hours returns the duration in hours.
func (ds *DailyStats) Hours() float64 {
	return float64(ds.Duration) / 60.0
}

// TimeRange represents a time range for filtering statistics.
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Validate checks if the time range is valid.
func (tr *TimeRange) Validate() error {
	if tr.Start.IsZero() {
		return fmt.Errorf("start time cannot be zero")
	}
	if tr.End.IsZero() {
		return fmt.Errorf("end time cannot be zero")
	}
	if tr.End.Before(tr.Start) {
		return fmt.Errorf("end time cannot be before start time")
	}

	return nil
}

// Contains checks if a given time falls within the range.
func (tr *TimeRange) Contains(t time.Time) bool {
	return !t.Before(tr.Start) && !t.After(tr.End)
}

// NewTimeRangeToday creates a time range for today (midnight to now).
func NewTimeRangeToday() TimeRange {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return TimeRange{Start: start, End: now}
}

// NewTimeRangeThisWeek creates a time range for the current week (Monday to now).
func NewTimeRangeThisWeek() TimeRange {
	now := time.Now()
	// Find Monday of current week
	weekday := int(now.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	daysToMonday := weekday - 1
	monday := now.AddDate(0, 0, -daysToMonday)
	start := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
	return TimeRange{Start: start, End: now}
}

// NewTimeRangeThisMonth creates a time range for the current month (1st to now).
func NewTimeRangeThisMonth() TimeRange {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return TimeRange{Start: start, End: now}
}

// NewTimeRangeSince creates a time range from a given date to now.
func NewTimeRangeSince(since time.Time) TimeRange {
	return TimeRange{Start: since, End: time.Now()}
}

// NewTimeRangeAll creates a time range covering all time (from epoch to now).
func NewTimeRangeAll() TimeRange {
	return TimeRange{Start: time.Unix(0, 0), End: time.Now()}
}
