// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PlanService defines the interface for plan operations needed by session service.
// This allows SessionService to verify that plans exist before starting sessions.
type PlanService interface {
	// Get retrieves a plan by ID.
	Get(ctx context.Context, id string) (interface{}, error)
}

// Service provides business logic for session management.
// It orchestrates between the session repository and plan service.
type Service struct {
	repo        Repository
	planService PlanService // Optional - can be nil
}

// NewService creates a new session service.
func NewService(repo Repository, planService PlanService) *Service {
	return &Service{
		repo:        repo,
		planService: planService,
	}
}

// StartRequest contains parameters for starting a new session.
type StartRequest struct {
	PlanID  string
	ChunkID string // Optional
	Notes   string // Optional initial notes
}

// Start creates and starts a new learning session.
// This implements FR-003 (Session Start) from the specifications.
//
// Returns an error if:
// - An active session already exists
// - The plan does not exist (if planService is configured)
// - The request is invalid
func (s *Service) Start(ctx context.Context, req StartRequest) (*Session, error) {
	// Validate request
	if req.PlanID == "" {
		return nil, fmt.Errorf("plan ID cannot be empty")
	}

	// Check for active session
	active, err := s.repo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check for active session: %w", err)
	}

	if active != nil {
		return nil, fmt.Errorf("active session already exists: %s (plan: %s). Stop it with 'samedi stop'", active.ID, active.PlanID)
	}

	// Verify plan exists (if plan service is available)
	if s.planService != nil {
		_, err := s.planService.Get(ctx, req.PlanID)
		if err != nil {
			return nil, fmt.Errorf("plan not found: %s", req.PlanID)
		}
	}

	// Create new session
	now := time.Now()
	session := &Session{
		ID:        uuid.New().String(),
		PlanID:    req.PlanID,
		ChunkID:   req.ChunkID,
		StartTime: now,
		EndTime:   nil, // Active session
		Duration:  0,
		Notes:     req.Notes,
		Artifacts: []string{},
		CreatedAt: now,
	}

	// Validate session
	if err := session.Validate(); err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// Save to repository
	if err := s.repo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// StopRequest contains parameters for stopping an active session.
type StopRequest struct {
	Notes     string
	Artifacts []string
}

// Stop completes the currently active session.
// This implements FR-004 (Session Stop) from the specifications.
//
// Returns an error if:
// - No active session exists
// - The session update fails
func (s *Service) Stop(ctx context.Context, req StopRequest) (*Session, error) {
	// Find active session
	session, err := s.repo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("no active session to stop. Start one with 'samedi start <plan-id>'")
	}

	// Complete the session
	now := time.Now()
	if err := session.Complete(now); err != nil {
		return nil, fmt.Errorf("failed to complete session: %w", err)
	}

	// Add notes if provided
	if req.Notes != "" {
		session.AddNotes(req.Notes)
	}

	// Add artifacts if provided
	for _, artifact := range req.Artifacts {
		session.AddArtifact(artifact)
	}

	// Validate the updated session
	if err := session.Validate(); err != nil {
		return nil, fmt.Errorf("invalid session after update: %w", err)
	}

	// Update in repository
	if err := s.repo.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return session, nil
}

// GetActive retrieves the currently active session, if any.
// Returns nil if no active session exists (not an error).
func (s *Service) GetActive(ctx context.Context) (*Session, error) {
	session, err := s.repo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}

	return session, nil
}

// Status represents the current state of sessions.
type Status struct {
	Active  *Session   // Currently active session, or nil
	Recent  []*Session // Recent sessions (most recent first)
	HasMore bool       // True if there are more sessions beyond Recent
}

// GetStatus retrieves the current session status (active + recent sessions).
// This implements FR-005 (Session Status) from the specifications.
func (s *Service) GetStatus(ctx context.Context) (*Status, error) {
	// Get active session
	active, err := s.repo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}

	// Get recent sessions from all plans
	// For now, we'll get recent sessions from the active session's plan
	// or all sessions if no active session
	var recent []*Session
	if active != nil {
		recent, err = s.repo.List(ctx, active.PlanID, 5)
		if err != nil {
			return nil, fmt.Errorf("failed to get recent sessions: %w", err)
		}
	}

	status := &Status{
		Active:  active,
		Recent:  recent,
		HasMore: len(recent) >= 5, // Simple heuristic
	}

	return status, nil
}

// List retrieves sessions for a specific plan.
func (s *Service) List(ctx context.Context, planID string, limit int) ([]*Session, error) {
	if planID == "" {
		return nil, fmt.Errorf("plan ID cannot be empty")
	}

	sessions, err := s.repo.List(ctx, planID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return sessions, nil
}

// GetByPlan retrieves all sessions for a specific plan.
func (s *Service) GetByPlan(ctx context.Context, planID string) ([]*Session, error) {
	if planID == "" {
		return nil, fmt.Errorf("plan ID cannot be empty")
	}

	sessions, err := s.repo.GetByPlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions for plan: %w", err)
	}

	return sessions, nil
}

// GetSessionCount returns the number of sessions for a plan.
func (s *Service) GetSessionCount(ctx context.Context, planID string) (int, error) {
	sessions, err := s.GetByPlan(ctx, planID)
	if err != nil {
		return 0, err
	}

	return len(sessions), nil
}

// GetTotalDuration returns the total duration in minutes for all sessions of a plan.
func (s *Service) GetTotalDuration(ctx context.Context, planID string) (int, error) {
	sessions, err := s.GetByPlan(ctx, planID)
	if err != nil {
		return 0, err
	}

	totalMinutes := 0
	for _, session := range sessions {
		if !session.IsActive() {
			totalMinutes += session.Duration
		}
	}

	return totalMinutes, nil
}
