// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package stats

import (
	"context"
	"fmt"

	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/session"
	"github.com/pezware/samedi.dev/internal/storage"
)

// PlanService defines the interface for plan operations needed by stats service.
type PlanService interface {
	Get(ctx context.Context, id string) (*plan.Plan, error)
	List(ctx context.Context, filter *storage.PlanFilter) ([]*storage.PlanRecord, error)
}

// SessionService defines the interface for session operations needed by stats service.
type SessionService interface {
	List(ctx context.Context, planID string, limit int) ([]*session.Session, error)
	ListAll(ctx context.Context) ([]*session.Session, error)
}

// Service provides statistics calculation using plan and session data.
// It acts as a facade over the calculator functions, handling data loading
// and conversion.
type Service struct {
	planService    PlanService
	sessionService SessionService
}

// NewService creates a new stats service with required dependencies.
func NewService(planService PlanService, sessionService SessionService) *Service {
	return &Service{
		planService:    planService,
		sessionService: sessionService,
	}
}

// GetTotalStats computes aggregate statistics across all learning activity.
// It loads all plans and sessions, then uses the calculator functions.
func (s *Service) GetTotalStats(ctx context.Context) (*TotalStats, error) {
	// Load all plan records (metadata only)
	planRecords, err := s.planService.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}

	// Load full plans with chunks (needed for progress calculations)
	plans := make([]plan.Plan, 0, len(planRecords))
	for _, record := range planRecords {
		fullPlan, err := s.planService.Get(ctx, record.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load plan %s: %w", record.ID, err)
		}
		plans = append(plans, *fullPlan)
	}

	// Load all sessions
	sessions, err := s.sessionService.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	// Convert pointers to values
	sessionValues := make([]session.Session, len(sessions))
	for i := range sessions {
		sessionValues[i] = *sessions[i]
	}

	// Calculate stats
	stats := CalculateTotalStats(sessionValues, plans)

	return &stats, nil
}

// GetPlanStats computes statistics for a specific learning plan.
// It loads the plan and its sessions, then calculates plan-specific metrics.
func (s *Service) GetPlanStats(ctx context.Context, planID string) (*PlanStats, error) {
	// Load full plan with chunks
	p, err := s.planService.Get(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to load plan: %w", err)
	}

	// Load plan sessions (limit 0 = all sessions)
	sessions, err := s.sessionService.List(ctx, planID, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	// Convert pointers to values
	sessionValues := make([]session.Session, len(sessions))
	for i := range sessions {
		sessionValues[i] = *sessions[i]
	}

	// Calculate stats
	stats := CalculatePlanStats(planID, sessionValues, p)

	return &stats, nil
}

// GetDailyStats groups sessions by day and returns daily statistics.
// It loads all sessions and filters them by the given time range.
func (s *Service) GetDailyStats(ctx context.Context, timeRange TimeRange) ([]DailyStats, error) {
	// Validate time range
	if err := timeRange.Validate(); err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	// Load all sessions
	sessions, err := s.sessionService.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	// Convert pointers to values
	sessionValues := make([]session.Session, len(sessions))
	for i := range sessions {
		sessionValues[i] = *sessions[i]
	}

	// Calculate daily stats
	stats := CalculateDailyStats(sessionValues, timeRange)

	return stats, nil
}

// GetStreakInfo returns current and longest learning streaks.
// A streak is consecutive days with at least one learning session.
func (s *Service) GetStreakInfo(ctx context.Context) (currentStreak, longestStreak int, err error) {
	// Load all sessions
	sessions, err := s.sessionService.ListAll(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to list sessions: %w", err)
	}

	// Convert pointers to values
	sessionValues := make([]session.Session, len(sessions))
	for i := range sessions {
		sessionValues[i] = *sessions[i]
	}

	// Calculate streaks
	current, longest := CalculateStreak(sessionValues)

	return current, longest, nil
}

// GetActiveDays returns all unique days with learning activity.
// Days are normalized to midnight in the session's timezone.
func (s *Service) GetActiveDays(ctx context.Context) ([]DailyStats, error) {
	// Load all sessions
	sessions, err := s.sessionService.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	// Convert pointers to values
	sessionValues := make([]session.Session, len(sessions))
	for i := range sessions {
		sessionValues[i] = *sessions[i]
	}

	// Get active days and create daily stats
	activeDays := GetActiveDays(sessionValues)

	// Convert to daily stats by calculating stats for each day
	result := make([]DailyStats, 0, len(activeDays))
	for i := range activeDays {
		dayStart := activeDays[i]
		dayEnd := dayStart.AddDate(0, 0, 1)
		timeRange := TimeRange{Start: dayStart, End: dayEnd}

		dailyStats := CalculateDailyStats(sessionValues, timeRange)
		if len(dailyStats) > 0 {
			result = append(result, dailyStats[0])
		}
	}

	return result, nil
}

// GetAllPlanStats computes statistics for all plans and returns them as a map.
// This is useful for dashboard views that show all plans at once.
func (s *Service) GetAllPlanStats(ctx context.Context) (map[string]PlanStats, error) {
	// Load all plan records
	planRecords, err := s.planService.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}

	// Load full plans with chunks
	plans := make([]plan.Plan, 0, len(planRecords))
	for _, record := range planRecords {
		fullPlan, err := s.planService.Get(ctx, record.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load plan %s: %w", record.ID, err)
		}
		plans = append(plans, *fullPlan)
	}

	// Load all sessions
	sessions, err := s.sessionService.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	// Convert pointers to values
	sessionValues := make([]session.Session, len(sessions))
	for i := range sessions {
		sessionValues[i] = *sessions[i]
	}

	// Aggregate by plan
	stats := AggregateByPlan(sessionValues, plans)

	return stats, nil
}
