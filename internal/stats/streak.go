// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package stats

import (
	"sort"
	"time"

	"github.com/pezware/samedi.dev/internal/session"
)

// CalculateStreak calculates the current and longest learning streaks.
// A streak is consecutive days with at least one learning session.
// Returns (currentStreak, longestStreak).
func CalculateStreak(sessions []session.Session) (int, int) {
	return calculateStreakAsOf(sessions, time.Now())
}

// calculateStreakAsOf calculates streaks as of a specific point in time.
// This is useful for testing and historical analysis.
func calculateStreakAsOf(sessions []session.Session, now time.Time) (int, int) {
	if len(sessions) == 0 {
		return 0, 0
	}

	// Get all active days sorted chronologically
	activeDays := GetActiveDays(sessions)
	if len(activeDays) == 0 {
		return 0, 0
	}

	// Find all streaks
	streaks := findStreaks(activeDays)
	if len(streaks) == 0 {
		return 0, 0
	}

	// Find longest streak
	longestStreak := 0
	for _, streak := range streaks {
		if streak > longestStreak {
			longestStreak = streak
		}
	}

	// Check if current streak is active (last session was today or yesterday)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)

	lastDay := activeDays[len(activeDays)-1]

	currentStreak := 0
	if sameDay(lastDay, today) || sameDay(lastDay, yesterday) {
		// Current streak is active - it's the last streak in the list
		currentStreak = streaks[len(streaks)-1]
	}

	return currentStreak, longestStreak
}

// GetActiveDays returns a sorted list of unique days with learning activity.
// Days are normalized to midnight in the session's timezone.
func GetActiveDays(sessions []session.Session) []time.Time {
	if len(sessions) == 0 {
		return []time.Time{}
	}

	// Use map to track unique days
	dayMap := make(map[string]time.Time)

	for i := range sessions {
		// Normalize to midnight on the start day
		dayStart := time.Date(
			sessions[i].StartTime.Year(),
			sessions[i].StartTime.Month(),
			sessions[i].StartTime.Day(),
			0, 0, 0, 0,
			sessions[i].StartTime.Location(),
		)

		key := dayStart.Format("2006-01-02")
		dayMap[key] = dayStart
	}

	// Convert map to slice
	days := make([]time.Time, 0, len(dayMap))
	for _, day := range dayMap {
		days = append(days, day)
	}

	// Sort chronologically
	sort.Slice(days, func(i, j int) bool {
		return days[i].Before(days[j])
	})

	return days
}

// DetectStreakBreaks identifies days where streaks were broken (gaps in activity).
func DetectStreakBreaks(sessions []session.Session) []time.Time {
	activeDays := GetActiveDays(sessions)
	if len(activeDays) < 2 {
		return []time.Time{}
	}

	breaks := []time.Time{}

	for i := 0; i < len(activeDays)-1; i++ {
		currentDay := activeDays[i]
		nextDay := activeDays[i+1]

		// Calculate days between
		daysBetween := int(nextDay.Sub(currentDay).Hours() / 24)

		// If more than 1 day gap, we have break(s)
		if daysBetween > 1 {
			// Add all gap days as breaks
			for d := 1; d < daysBetween; d++ {
				breakDay := currentDay.AddDate(0, 0, d)
				breaks = append(breaks, breakDay)
			}
		}
	}

	return breaks
}

// Helper functions

// findStreaks identifies all consecutive day streaks and returns their lengths.
func findStreaks(activeDays []time.Time) []int {
	if len(activeDays) == 0 {
		return []int{}
	}

	streaks := []int{}
	currentStreak := 1

	for i := 1; i < len(activeDays); i++ {
		prevDay := activeDays[i-1]
		currentDay := activeDays[i]

		// Check if days are consecutive
		daysBetween := int(currentDay.Sub(prevDay).Hours() / 24)

		if daysBetween == 1 {
			// Consecutive day - continue streak
			currentStreak++
		} else {
			// Gap found - save current streak and start new one
			streaks = append(streaks, currentStreak)
			currentStreak = 1
		}
	}

	// Don't forget to add the last streak
	streaks = append(streaks, currentStreak)

	return streaks
}

// sameDay checks if two times represent the same calendar day.
func sameDay(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() &&
		t1.Month() == t2.Month() &&
		t1.Day() == t2.Day()
}
