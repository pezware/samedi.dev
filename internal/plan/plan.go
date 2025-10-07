// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package plan

import (
	"fmt"
	"time"
)

// Plan represents a learning curriculum broken into time-boxed chunks.
// Plans are stored as markdown files with YAML frontmatter and indexed in SQLite.
type Plan struct {
	ID         string    `json:"id" yaml:"id"`
	Title      string    `json:"title" yaml:"title"`
	CreatedAt  time.Time `json:"created_at" yaml:"created"`
	UpdatedAt  time.Time `json:"updated_at" yaml:"updated"`
	TotalHours float64   `json:"total_hours" yaml:"total_hours"`
	Status     Status    `json:"status" yaml:"status"`
	Tags       []string  `json:"tags,omitempty" yaml:"tags,omitempty"`
	Chunks     []Chunk   `json:"chunks" yaml:"-"`
}

// Chunk represents a single learning session within a plan.
// Each chunk is time-boxed and has specific objectives.
type Chunk struct {
	ID          string   `json:"id" yaml:"id"`
	Title       string   `json:"title" yaml:"title"`
	Duration    int      `json:"duration" yaml:"duration"` // Duration in minutes
	Status      Status   `json:"status" yaml:"status"`
	Objectives  []string `json:"objectives,omitempty" yaml:"objectives,omitempty"`
	Resources   []string `json:"resources,omitempty" yaml:"resources,omitempty"`
	Deliverable string   `json:"deliverable,omitempty" yaml:"deliverable,omitempty"`
}

// Status represents the current state of a plan or chunk.
type Status string

const (
	// StatusNotStarted indicates the plan/chunk hasn't been started yet.
	StatusNotStarted Status = "not-started"
	// StatusInProgress indicates active work on the plan/chunk.
	StatusInProgress Status = "in-progress"
	// StatusCompleted indicates the plan/chunk is finished.
	StatusCompleted Status = "completed"
	// StatusSkipped indicates the chunk was intentionally skipped.
	StatusSkipped Status = "skipped"
	// StatusArchived indicates the plan is archived (completed or abandoned).
	StatusArchived Status = "archived"
)

// Valid statuses for validation.
var validStatuses = map[Status]bool{
	StatusNotStarted: true,
	StatusInProgress: true,
	StatusCompleted:  true,
	StatusSkipped:    true,
	StatusArchived:   true,
}

// IsValid checks if the status is a valid value.
func (s Status) IsValid() bool {
	return validStatuses[s]
}

// Validate checks if the plan has all required fields and valid values.
func (p *Plan) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}
	if p.Title == "" {
		return fmt.Errorf("plan title cannot be empty")
	}
	if p.TotalHours <= 0 {
		return fmt.Errorf("total hours must be positive, got %.1f", p.TotalHours)
	}
	if !p.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", p.Status)
	}
	if p.CreatedAt.IsZero() {
		return fmt.Errorf("created_at cannot be zero")
	}
	if p.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at cannot be zero")
	}
	if p.UpdatedAt.Before(p.CreatedAt) {
		return fmt.Errorf("updated_at cannot be before created_at")
	}

	// Validate chunks
	chunkIDs := make(map[string]bool)
	for i, chunk := range p.Chunks {
		if err := chunk.Validate(); err != nil {
			return fmt.Errorf("chunk %d (%s): %w", i, chunk.ID, err)
		}
		// Check for duplicate chunk IDs
		if chunkIDs[chunk.ID] {
			return fmt.Errorf("duplicate chunk ID: %s", chunk.ID)
		}
		chunkIDs[chunk.ID] = true
	}

	return nil
}

// Validate checks if the chunk has all required fields and valid values.
func (c *Chunk) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("chunk ID cannot be empty")
	}
	if c.Title == "" {
		return fmt.Errorf("chunk title cannot be empty")
	}
	if c.Duration <= 0 {
		return fmt.Errorf("duration must be positive, got %d", c.Duration)
	}
	if !c.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", c.Status)
	}

	return nil
}

// Progress calculates the completion percentage of the plan.
// Returns a value between 0.0 and 1.0.
func (p *Plan) Progress() float64 {
	if len(p.Chunks) == 0 {
		return 0.0
	}

	completed := 0
	for _, chunk := range p.Chunks {
		if chunk.Status == StatusCompleted {
			completed++
		}
	}

	return float64(completed) / float64(len(p.Chunks))
}

// ProgressPercent returns the completion percentage as an integer (0-100).
func (p *Plan) ProgressPercent() int {
	return int(p.Progress() * 100)
}

// TotalMinutes calculates total plan duration in minutes.
func (p *Plan) TotalMinutes() int {
	total := 0
	for _, chunk := range p.Chunks {
		total += chunk.Duration
	}
	return total
}

// CompletedHours calculates hours spent on completed chunks.
func (p *Plan) CompletedHours() float64 {
	minutes := 0
	for _, chunk := range p.Chunks {
		if chunk.Status == StatusCompleted {
			minutes += chunk.Duration
		}
	}
	return float64(minutes) / 60.0
}

// RemainingHours calculates hours remaining for incomplete chunks.
func (p *Plan) RemainingHours() float64 {
	minutes := 0
	for _, chunk := range p.Chunks {
		if chunk.Status != StatusCompleted && chunk.Status != StatusSkipped {
			minutes += chunk.Duration
		}
	}
	return float64(minutes) / 60.0
}

// NextChunk returns the next chunk to work on (first not-started or in-progress chunk).
// Returns nil if no chunks are available.
func (p *Plan) NextChunk() *Chunk {
	for i := range p.Chunks {
		if p.Chunks[i].Status == StatusInProgress {
			return &p.Chunks[i]
		}
	}
	for i := range p.Chunks {
		if p.Chunks[i].Status == StatusNotStarted {
			return &p.Chunks[i]
		}
	}
	return nil
}
