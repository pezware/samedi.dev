// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package plan

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/pezware/samedi.dev/internal/llm"
	"github.com/pezware/samedi.dev/internal/session"
	"github.com/pezware/samedi.dev/internal/storage"
)

// Service provides business logic for plan management.
// It orchestrates between SQLite metadata storage, filesystem markdown storage,
// and LLM providers for plan generation.
type Service struct {
	sqliteRepo     *SQLiteRepository
	filesystemRepo *FilesystemRepository
	llmProvider    llm.Provider
	fs             *storage.FilesystemStorage
	paths          *storage.Paths
	sessionService *session.Service // Optional - for session integration
}

// NewService creates a new plan service with all required dependencies.
func NewService(
	sqliteRepo *SQLiteRepository,
	filesystemRepo *FilesystemRepository,
	llmProvider llm.Provider,
	fs *storage.FilesystemStorage,
	paths *storage.Paths,
) *Service {
	return &Service{
		sqliteRepo:     sqliteRepo,
		filesystemRepo: filesystemRepo,
		llmProvider:    llmProvider,
		fs:             fs,
		paths:          paths,
	}
}

// SetSessionService sets the session service for session integration.
// This is optional and used for displaying session history in plan views.
func (s *Service) SetSessionService(sessionService *session.Service) {
	s.sessionService = sessionService
}

// CreateRequest contains parameters for creating a new plan.
type CreateRequest struct {
	Topic      string
	TotalHours float64
	Level      string // beginner, intermediate, advanced
	Goals      string // Optional specific goals
}

// Create generates a new learning plan using LLM and saves it to both stores.
// This implements FR-001 (Plan Generation) from the specifications.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Plan, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Generate plan ID (slug) from topic
	planID := slugify(req.Topic)

	// Check if plan already exists
	if s.filesystemRepo.Exists(ctx, planID) {
		return nil, fmt.Errorf("plan already exists: %s", planID)
	}

	// Load and render template
	prompt, err := s.renderTemplate(req, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// Call LLM to generate plan
	llmOutput, err := s.llmProvider.Call(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse LLM output into Plan struct
	plan, err := Parse(llmOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM output: %w", err)
	}

	// Ensure plan ID matches
	plan.ID = planID

	// Validate generated plan
	if err := plan.Validate(); err != nil {
		return nil, fmt.Errorf("generated plan is invalid: %w", err)
	}

	// Save to filesystem
	if err := s.filesystemRepo.Save(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to save plan file: %w", err)
	}

	// Index in SQLite
	record := ToRecord(plan, s.filesystemRepo.Path(planID))
	if err := s.sqliteRepo.Upsert(ctx, record); err != nil {
		// Rollback: delete the file we just created
		// We ignore the delete error since the primary error is more important
		_ = s.filesystemRepo.Delete(ctx, planID) //nolint:errcheck
		return nil, fmt.Errorf("failed to index plan: %w", err)
	}

	return plan, nil
}

// Get retrieves a plan by ID from filesystem.
func (s *Service) Get(ctx context.Context, id string) (*Plan, error) {
	// Load from filesystem (includes full plan with chunks)
	plan, err := s.filesystemRepo.Load(ctx, id)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// Update saves changes to an existing plan in both stores.
// This implements FR-002 (Manual Plan Editing) from the specifications.
func (s *Service) Update(ctx context.Context, plan *Plan) error {
	// Validate plan
	if err := plan.Validate(); err != nil {
		return fmt.Errorf("invalid plan: %w", err)
	}

	// Check if plan exists
	if !s.filesystemRepo.Exists(ctx, plan.ID) {
		return fmt.Errorf("plan not found: %s", plan.ID)
	}

	// Update timestamp
	plan.UpdatedAt = time.Now()

	// Save to filesystem
	if err := s.filesystemRepo.Save(ctx, plan); err != nil {
		return fmt.Errorf("failed to save plan file: %w", err)
	}

	// Update SQLite metadata
	record := ToRecord(plan, s.filesystemRepo.Path(plan.ID))
	if err := s.sqliteRepo.Upsert(ctx, record); err != nil {
		// Note: We don't rollback the file write here since the file was already updated
		// This is an acceptable tradeoff - the SQLite index can be rebuilt
		return fmt.Errorf("failed to update plan index: %w", err)
	}

	return nil
}

// Delete removes a plan from both filesystem and SQLite.
func (s *Service) Delete(ctx context.Context, id string) error {
	// Check if plan exists
	if !s.filesystemRepo.Exists(ctx, id) {
		return fmt.Errorf("plan not found: %s", id)
	}

	// Delete from SQLite first (less critical if it fails)
	if err := s.sqliteRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete from index: %w", err)
	}

	// Delete from filesystem
	if err := s.filesystemRepo.Delete(ctx, id); err != nil {
		// Note: SQLite record is already deleted, but file still exists
		// This is less than ideal but acceptable
		return fmt.Errorf("failed to delete plan file: %w", err)
	}

	return nil
}

// List retrieves plan metadata from SQLite with optional filtering.
func (s *Service) List(ctx context.Context, filter *storage.PlanFilter) ([]*storage.PlanRecord, error) {
	records, err := s.sqliteRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}

	return records, nil
}

// Exists checks if a plan exists by checking the filesystem.
func (s *Service) Exists(ctx context.Context, id string) bool {
	return s.filesystemRepo.Exists(ctx, id)
}

// GetMetadata retrieves plan metadata from SQLite without loading the full plan.
// This is faster than Get() when you only need metadata.
func (s *Service) GetMetadata(ctx context.Context, id string) (*storage.PlanRecord, error) {
	record, err := s.sqliteRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// GetRecentSessions retrieves the most recent learning sessions for a plan.
// Returns session data as maps for display purposes.
func (s *Service) GetRecentSessions(ctx context.Context, planID string, limit int) ([]map[string]interface{}, error) {
	// If no session service is configured, return empty
	if s.sessionService == nil {
		return []map[string]interface{}{}, nil
	}

	// Get sessions from session service
	sessions, err := s.sessionService.List(ctx, planID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Convert sessions to map format for display
	result := make([]map[string]interface{}, len(sessions))
	for i, sess := range sessions {
		result[i] = map[string]interface{}{
			"id":         sess.ID,
			"chunk_id":   sess.ChunkID,
			"start_time": sess.StartTime,
			"end_time":   sess.EndTime,
			"duration":   sess.Duration,
			"notes":      sess.Notes,
			"is_active":  sess.IsActive(),
		}
	}

	return result, nil
}

// GetCardCount returns the total number of flashcards for a plan.
// Stage 4 implementation: Will query CardRepository once flashcards are implemented.
// Currently returns 0 as placeholder.
func (s *Service) GetCardCount(_ context.Context, _ string) (int, error) {
	// Stage 4: Will query CardRepository
	return 0, nil
}

// validateCreateRequest checks if the create request has valid parameters.
func (s *Service) validateCreateRequest(req CreateRequest) error {
	if strings.TrimSpace(req.Topic) == "" {
		return fmt.Errorf("topic cannot be empty")
	}

	if req.TotalHours <= 0 {
		return fmt.Errorf("total hours must be positive, got %.1f", req.TotalHours)
	}

	if req.TotalHours > 1000 {
		return fmt.Errorf("total hours too large (max 1000), got %.1f", req.TotalHours)
	}

	// Level is optional, but if provided, should be valid
	if req.Level != "" {
		validLevels := map[string]bool{
			"beginner":     true,
			"intermediate": true,
			"advanced":     true,
		}
		if !validLevels[strings.ToLower(req.Level)] {
			return fmt.Errorf("invalid level: %s (must be beginner, intermediate, or advanced)", req.Level)
		}
	}

	return nil
}

// renderTemplate loads and renders the plan generation template with request parameters.
func (s *Service) renderTemplate(req CreateRequest, slug string) (string, error) {
	// Load template file
	templatePath := s.paths.TemplatePath("plan-generation")
	content, err := s.fs.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	// Parse template
	tmpl, err := template.New("plan-generation").Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Set default level if not provided
	level := req.Level
	if level == "" {
		level = "beginner"
	}

	// Set default goals if not provided
	goals := req.Goals
	if goals == "" {
		goals = fmt.Sprintf("Master %s through structured learning", req.Topic)
	}

	// Prepare template data
	data := map[string]interface{}{
		"Topic":      req.Topic,
		"TotalHours": req.TotalHours,
		"Level":      level,
		"Goals":      goals,
		"Slug":       slug,
		"Now":        time.Now().Format(time.RFC3339),
	}

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// slugify converts a topic string into a filesystem-safe slug.
// Examples:
//   - "Rust Async Programming" -> "rust-async-programming"
//   - "French B1 Mastery" -> "french-b1-mastery"
//   - "Music Theory (Basics)" -> "music-theory-basics"
func slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")

	// Trim hyphens from start and end
	s = strings.Trim(s, "-")

	// Replace multiple consecutive hyphens with single hyphen
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")

	return s
}
