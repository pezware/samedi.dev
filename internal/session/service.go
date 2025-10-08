// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PlanChunk represents a chunk for the session service's needs.
type PlanChunk struct {
	ID       string
	Duration int // in minutes
	Status   string
}

// PlanService defines the interface for plan operations needed by session service.
// This allows SessionService to verify that plans exist and update chunk statuses.
type PlanService interface {
	// Get retrieves a plan by ID.
	Get(ctx context.Context, id string) (interface{}, error)

	// GetChunk retrieves a specific chunk from a plan.
	GetChunk(ctx context.Context, planID, chunkID string) (*PlanChunk, error)

	// UpdateChunkStatus updates a chunk's status (for smart inference).
	UpdateChunkStatus(ctx context.Context, planID, chunkID, newStatus string) error
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

	// Smart inference: Mark chunk as in-progress if it's not-started
	if s.planService != nil && req.ChunkID != "" {
		chunk, err := s.planService.GetChunk(ctx, req.PlanID, req.ChunkID)
		if err == nil && chunk.Status == "not-started" {
			// Best-effort update: silently ignore errors as this is not critical to session creation
			//nolint:errcheck // intentionally ignoring error for best-effort status update
			s.planService.UpdateChunkStatus(ctx, req.PlanID, req.ChunkID, "in-progress")
		}
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

	// Smart inference: Auto-complete chunk if total time >= chunk duration
	// Best-effort update: silently ignore errors as session was successfully stopped
	if s.planService != nil && session.ChunkID != "" {
		//nolint:errcheck // intentionally ignoring error for best-effort status update
		s.checkAndCompleteChunk(ctx, session.PlanID, session.ChunkID)
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

	// Get recent sessions
	// If there's an active session, show recent sessions from that plan
	// If no active session, show recent sessions across all plans
	var recent []*Session
	var planID string
	if active != nil {
		planID = active.PlanID
	} else {
		planID = "" // Empty planID means "all plans"
	}

	recent, err = s.repo.List(ctx, planID, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent sessions: %w", err)
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

// GetChunkSessions retrieves all sessions for a specific chunk within a plan.
func (s *Service) GetChunkSessions(ctx context.Context, planID, chunkID string) ([]*Session, error) {
	if planID == "" {
		return nil, fmt.Errorf("plan ID cannot be empty")
	}
	if chunkID == "" {
		return nil, fmt.Errorf("chunk ID cannot be empty")
	}

	// Get all sessions for the plan
	sessions, err := s.repo.GetByPlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Filter to only sessions for this chunk
	var chunkSessions []*Session
	for _, session := range sessions {
		if session.ChunkID == chunkID {
			chunkSessions = append(chunkSessions, session)
		}
	}

	return chunkSessions, nil
}

// ChunkStats contains statistics about sessions for a specific chunk.
type ChunkStats struct {
	SessionCount  int // Total number of sessions for this chunk
	TotalDuration int // Total time spent in minutes (completed sessions only)
}

// GetChunkStats returns statistics about sessions for a specific chunk.
func (s *Service) GetChunkStats(ctx context.Context, planID, chunkID string) (*ChunkStats, error) {
	sessions, err := s.GetChunkSessions(ctx, planID, chunkID)
	if err != nil {
		return nil, err
	}

	stats := &ChunkStats{
		SessionCount: len(sessions),
	}

	// Calculate total duration from completed sessions
	for _, session := range sessions {
		if !session.IsActive() {
			stats.TotalDuration += session.Duration
		}
	}

	return stats, nil
}

// checkAndCompleteChunk checks if a chunk should be auto-completed based on session time.
// If total session time for the chunk >= chunk duration, marks it as completed.
func (s *Service) checkAndCompleteChunk(ctx context.Context, planID, chunkID string) error {
	// Get the chunk to find its expected duration
	chunk, err := s.planService.GetChunk(ctx, planID, chunkID)
	if err != nil {
		return fmt.Errorf("failed to get chunk: %w", err)
	}

	// Skip if already completed or skipped
	if chunk.Status == "completed" || chunk.Status == "skipped" {
		return nil
	}

	// Get all sessions for this plan
	sessions, err := s.repo.GetByPlan(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	// Calculate total time spent on this specific chunk
	totalMinutes := 0
	for _, session := range sessions {
		if session.ChunkID == chunkID && !session.IsActive() {
			totalMinutes += session.Duration
		}
	}

	// If total time >= chunk duration, mark as completed
	if totalMinutes >= chunk.Duration {
		if err := s.planService.UpdateChunkStatus(ctx, planID, chunkID, "completed"); err != nil {
			return fmt.Errorf("failed to mark chunk as completed: %w", err)
		}
	}

	return nil
}
