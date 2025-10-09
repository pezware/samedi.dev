// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package stats

import (
	"time"

	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/session"
)

// CalculateTotalStats computes aggregate statistics across all sessions and plans.
func CalculateTotalStats(sessions []session.Session, plans []plan.Plan) TotalStats {
	stats := TotalStats{}

	if len(sessions) == 0 {
		return stats
	}

	// Calculate total hours and sessions
	totalMinutes := 0
	var lastSession *time.Time

	for i := range sessions {
		totalMinutes += sessions[i].Duration

		// Track most recent session
		if lastSession == nil || sessions[i].StartTime.After(*lastSession) {
			lastSession = &sessions[i].StartTime
		}
	}

	stats.TotalHours = float64(totalMinutes) / 60.0
	stats.TotalSessions = len(sessions)
	stats.AverageSession = float64(totalMinutes) / float64(len(sessions))
	stats.LastSessionDate = lastSession

	// Count active and completed plans
	// Active = not-started or in-progress (exclude archived)
	// Completed = completed status
	for i := range plans {
		switch plans[i].Status {
		case plan.StatusNotStarted, plan.StatusInProgress:
			stats.ActivePlans++
		case plan.StatusCompleted:
			stats.CompletedPlans++
			// StatusArchived and others are not counted
		}
	}

	// Calculate streak (will be implemented in streak.go)
	stats.CurrentStreak, stats.LongestStreak = CalculateStreak(sessions)

	return stats
}

// CalculatePlanStats computes statistics for a specific plan.
func CalculatePlanStats(planID string, sessions []session.Session, p *plan.Plan) PlanStats {
	stats := PlanStats{
		PlanID:       planID,
		PlanTitle:    p.Title,
		PlannedHours: p.TotalHours,
		TotalChunks:  len(p.Chunks),
		Status:       string(p.Status),
	}

	// Filter sessions for this plan
	planSessions := filterSessionsByPlan(sessions, planID)

	if len(planSessions) == 0 {
		stats.Progress = p.Progress()
		stats.CompletedChunks = countCompletedChunks(p.Chunks)
		return stats
	}

	// Calculate total hours and session count
	totalMinutes := 0
	var lastSession *time.Time

	for i := range planSessions {
		totalMinutes += planSessions[i].Duration

		// Track most recent session
		if lastSession == nil || planSessions[i].StartTime.After(*lastSession) {
			lastSession = &planSessions[i].StartTime
		}
	}

	stats.TotalHours = float64(totalMinutes) / 60.0
	stats.SessionCount = len(planSessions)
	stats.LastSession = lastSession

	// Calculate progress from chunks
	stats.Progress = p.Progress()
	stats.CompletedChunks = countCompletedChunks(p.Chunks)

	return stats
}

// CalculateDailyStats groups sessions by day and returns daily statistics.
func CalculateDailyStats(sessions []session.Session, timeRange TimeRange) []DailyStats {
	if len(sessions) == 0 {
		return []DailyStats{}
	}

	// Group sessions by day
	dailyMap := make(map[string]*DailyStats)

	for i := range sessions {
		sess := sessions[i]

		// Skip sessions outside time range
		if !timeRange.Contains(sess.StartTime) {
			continue
		}

		// Get day key (date at midnight)
		dayKey := getDayKey(sess.StartTime)

		// Initialize daily stats if not exists
		if dailyMap[dayKey] == nil {
			dayStart := time.Date(
				sess.StartTime.Year(),
				sess.StartTime.Month(),
				sess.StartTime.Day(),
				0, 0, 0, 0,
				sess.StartTime.Location(),
			)
			dailyMap[dayKey] = &DailyStats{
				Date:  dayStart,
				Plans: []string{},
			}
		}

		// Accumulate stats
		dailyMap[dayKey].Duration += sess.Duration
		dailyMap[dayKey].SessionCount++

		// Add plan ID if not already present
		if !contains(dailyMap[dayKey].Plans, sess.PlanID) {
			dailyMap[dayKey].Plans = append(dailyMap[dayKey].Plans, sess.PlanID)
		}
	}

	// Convert map to sorted slice
	result := make([]DailyStats, 0, len(dailyMap))
	for _, stats := range dailyMap {
		result = append(result, *stats)
	}

	// Sort by date
	sortDailyStats(result)

	return result
}

// AggregateByPlan creates a map of plan IDs to their statistics.
func AggregateByPlan(sessions []session.Session, plans []plan.Plan) map[string]PlanStats {
	result := make(map[string]PlanStats)

	// Calculate stats for each plan
	for i := range plans {
		stats := CalculatePlanStats(plans[i].ID, sessions, &plans[i])
		result[plans[i].ID] = stats
	}

	return result
}

// Helper functions

// filterSessionsByPlan returns sessions that belong to the specified plan.
func filterSessionsByPlan(sessions []session.Session, planID string) []session.Session {
	result := make([]session.Session, 0)
	for i := range sessions {
		if sessions[i].PlanID == planID {
			result = append(result, sessions[i])
		}
	}
	return result
}

// countCompletedChunks counts the number of completed chunks in a plan.
func countCompletedChunks(chunks []plan.Chunk) int {
	count := 0
	for _, chunk := range chunks {
		if chunk.Status == plan.StatusCompleted {
			count++
		}
	}
	return count
}

// getDayKey returns a unique key for a day (YYYY-MM-DD).
func getDayKey(t time.Time) string {
	return t.Format("2006-01-02")
}

// contains checks if a string slice contains a value.
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// sortDailyStats sorts daily stats by date (ascending).
func sortDailyStats(stats []DailyStats) {
	// Simple bubble sort for small slices
	n := len(stats)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if stats[j].Date.After(stats[j+1].Date) {
				stats[j], stats[j+1] = stats[j+1], stats[j]
			}
		}
	}
}
