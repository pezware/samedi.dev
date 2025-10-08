// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package session

import (
	"fmt"
	"time"
)

// Session represents a tracked learning session linked to a plan and optional chunk.
// Sessions record when learning started, when it ended, and associated notes/artifacts.
type Session struct {
	ID           string     `json:"id"`                      // UUID
	PlanID       string     `json:"plan_id"`                 // References plan
	ChunkID      string     `json:"chunk_id,omitempty"`      // Optional chunk reference
	StartTime    time.Time  `json:"start_time"`              // When session started
	EndTime      *time.Time `json:"end_time,omitempty"`      // When session ended (nil if active)
	Duration     int        `json:"duration_minutes"`        // Calculated duration in minutes
	Notes        string     `json:"notes,omitempty"`         // User notes after session
	Artifacts    []string   `json:"artifacts,omitempty"`     // URLs or file paths
	CardsCreated int        `json:"cards_created,omitempty"` // Number of flashcards generated
	CreatedAt    time.Time  `json:"created_at"`              // Record creation timestamp
}

// Validate checks if the session has all required fields and valid values.
func (s *Session) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	if s.PlanID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}
	if s.StartTime.IsZero() {
		return fmt.Errorf("start time cannot be zero")
	}
	if s.CreatedAt.IsZero() {
		return fmt.Errorf("created_at cannot be zero")
	}

	// If session is completed, validate end time
	if s.EndTime != nil {
		if s.EndTime.Before(s.StartTime) {
			return fmt.Errorf("end time cannot be before start time")
		}
		if s.EndTime.Equal(s.StartTime) {
			return fmt.Errorf("end time cannot equal start time")
		}

		// Validate duration matches calculated value
		expectedDuration := s.CalculateDuration()
		if s.Duration != expectedDuration {
			return fmt.Errorf("duration mismatch: stored %d minutes, calculated %d minutes", s.Duration, expectedDuration)
		}
	} else if s.Duration != 0 {
		// Active sessions should have zero duration
		return fmt.Errorf("active session should have zero duration, got %d", s.Duration)
	}

	if s.CardsCreated < 0 {
		return fmt.Errorf("cards created cannot be negative, got %d", s.CardsCreated)
	}

	return nil
}

// IsActive returns true if the session is currently in progress (no end time).
func (s *Session) IsActive() bool {
	return s.EndTime == nil
}

// CalculateDuration calculates the duration between start and end time in minutes.
// Returns 0 if the session is still active.
func (s *Session) CalculateDuration() int {
	if s.EndTime == nil {
		return 0
	}

	duration := s.EndTime.Sub(s.StartTime)
	return int(duration.Minutes())
}

// ElapsedMinutes returns the current elapsed time for active sessions,
// or the final duration for completed sessions.
func (s *Session) ElapsedMinutes() int {
	if s.EndTime == nil {
		// Active session - calculate from start to now
		duration := time.Since(s.StartTime)
		return int(duration.Minutes())
	}
	// Completed session - return stored duration
	return s.Duration
}

// ElapsedTime returns a human-readable elapsed time string.
// For active sessions, calculates current elapsed time.
// For completed sessions, returns the final duration.
func (s *Session) ElapsedTime() string {
	minutes := s.ElapsedMinutes()
	hours := minutes / 60
	mins := minutes % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %02dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

// Complete marks the session as complete with the given end time.
// It automatically calculates and sets the duration.
func (s *Session) Complete(endTime time.Time) error {
	if !s.IsActive() {
		return fmt.Errorf("session is already complete")
	}
	if endTime.Before(s.StartTime) {
		return fmt.Errorf("end time cannot be before start time")
	}

	s.EndTime = &endTime
	s.Duration = s.CalculateDuration()
	return nil
}

// AddNotes appends notes to the session.
func (s *Session) AddNotes(notes string) {
	if s.Notes == "" {
		s.Notes = notes
	} else {
		s.Notes += "\n" + notes
	}
}

// AddArtifact adds a URL or file path to the artifacts list.
func (s *Session) AddArtifact(artifact string) {
	if artifact != "" {
		s.Artifacts = append(s.Artifacts, artifact)
	}
}
