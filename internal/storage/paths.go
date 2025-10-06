// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths holds all filesystem paths for samedi data.
type Paths struct {
	BaseDir      string
	PlansDir     string
	CardsDir     string
	TemplatesDir string
	BackupDir    string
	DatabasePath string
	ConfigPath   string
}

// DefaultPaths returns the default filesystem paths.
func DefaultPaths() (*Paths, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".samedi")

	return &Paths{
		BaseDir:      baseDir,
		PlansDir:     filepath.Join(baseDir, "plans"),
		CardsDir:     filepath.Join(baseDir, "cards"),
		TemplatesDir: filepath.Join(baseDir, "templates"),
		BackupDir:    filepath.Join(homeDir, "samedi-backups"),
		DatabasePath: filepath.Join(baseDir, "sessions.db"),
		ConfigPath:   filepath.Join(baseDir, "config.toml"),
	}, nil
}

// EnsureDirectories creates all required directories if they don't exist.
func (p *Paths) EnsureDirectories() error {
	dirs := []string{
		p.BaseDir,
		p.PlansDir,
		p.CardsDir,
		p.TemplatesDir,
		p.BackupDir,
	}

	for _, dir := range dirs {
		// Skip empty directory paths
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// Exists checks if all required directories exist.
func (p *Paths) Exists() bool {
	dirs := []string{
		p.BaseDir,
		p.PlansDir,
		p.CardsDir,
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return false
		}
	}

	return true
}

// Clean removes all samedi data directories (dangerous!).
// Used primarily for testing.
func (p *Paths) Clean() error {
	if err := os.RemoveAll(p.BaseDir); err != nil {
		return fmt.Errorf("failed to remove base directory: %w", err)
	}
	return nil
}

// PlanPath returns the full path for a plan markdown file.
func (p *Paths) PlanPath(planID string) string {
	return filepath.Join(p.PlansDir, fmt.Sprintf("%s.md", planID))
}

// CardsPath returns the full path for a cards markdown file.
func (p *Paths) CardsPath(planID string) string {
	return filepath.Join(p.CardsDir, fmt.Sprintf("%s.cards.md", planID))
}

// TemplatePath returns the full path for a template file.
func (p *Paths) TemplatePath(templateName string) string {
	return filepath.Join(p.TemplatesDir, fmt.Sprintf("%s.md", templateName))
}
