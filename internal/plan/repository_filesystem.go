// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package plan

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pezware/samedi.dev/internal/storage"
)

// FilesystemRepository handles markdown file operations for plans.
type FilesystemRepository struct {
	fs    *storage.FilesystemStorage
	paths *storage.Paths
}

// NewFilesystemRepository creates a new filesystem-backed plan repository.
func NewFilesystemRepository(fs *storage.FilesystemStorage, paths *storage.Paths) *FilesystemRepository {
	return &FilesystemRepository{
		fs:    fs,
		paths: paths,
	}
}

// Save writes a plan to a markdown file with YAML frontmatter.
func (r *FilesystemRepository) Save(_ context.Context, plan *Plan) error {
	// Validate plan before saving
	if err := plan.Validate(); err != nil {
		return fmt.Errorf("invalid plan: %w", err)
	}

	// Convert plan to markdown
	content, err := Format(plan)
	if err != nil {
		return fmt.Errorf("failed to format plan: %w", err)
	}

	// Get file path
	filePath := r.paths.PlanPath(plan.ID)

	// Write to filesystem
	if err := r.fs.WriteFile(filePath, []byte(content)); err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}

	return nil
}

// Load reads and parses a plan from a markdown file.
func (r *FilesystemRepository) Load(_ context.Context, id string) (*Plan, error) {
	filePath := r.paths.PlanPath(id)

	// Check if file exists
	if !r.fs.FileExists(filePath) {
		return nil, fmt.Errorf("plan not found: %s", id)
	}

	// Read file
	content, err := r.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan file: %w", err)
	}

	// Parse markdown
	plan, err := Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse plan: %w", err)
	}

	return plan, nil
}

// Delete removes a plan's markdown file.
func (r *FilesystemRepository) Delete(_ context.Context, id string) error {
	filePath := r.paths.PlanPath(id)

	// Check if file exists
	if !r.fs.FileExists(filePath) {
		return fmt.Errorf("plan not found: %s", id)
	}

	// Delete file
	if err := r.fs.DeleteFile(filePath); err != nil {
		return fmt.Errorf("failed to delete plan file: %w", err)
	}

	return nil
}

// Exists checks if a plan file exists.
func (r *FilesystemRepository) Exists(_ context.Context, id string) bool {
	filePath := r.paths.PlanPath(id)
	return r.fs.FileExists(filePath)
}

// List returns all plan IDs by scanning the plans directory.
func (r *FilesystemRepository) List(_ context.Context) ([]string, error) {
	// Read directory entries
	entries, err := os.ReadDir(r.paths.PlansDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read plans directory: %w", err)
	}

	planIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Only process .md files
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		// Extract plan ID from filename (remove .md extension)
		planID := strings.TrimSuffix(name, ".md")
		planIDs = append(planIDs, planID)
	}

	return planIDs, nil
}

// Path returns the full file path for a plan ID.
func (r *FilesystemRepository) Path(id string) string {
	return r.paths.PlanPath(id)
}

// LoadAll loads all plans from the filesystem.
// This is useful for bulk operations but can be slow for many plans.
func (r *FilesystemRepository) LoadAll(ctx context.Context) ([]*Plan, error) {
	planIDs, err := r.List(ctx)
	if err != nil {
		return nil, err
	}

	plans := make([]*Plan, 0, len(planIDs))
	for _, id := range planIDs {
		plan, err := r.Load(ctx, id)
		if err != nil {
			// Log error but continue with other plans
			continue
		}
		plans = append(plans, plan)
	}

	return plans, nil
}
