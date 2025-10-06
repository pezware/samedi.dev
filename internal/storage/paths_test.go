// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPaths(t *testing.T) {
	paths, err := DefaultPaths()
	require.NoError(t, err)
	require.NotNil(t, paths)

	// Check paths are set
	assert.NotEmpty(t, paths.BaseDir)
	assert.NotEmpty(t, paths.PlansDir)
	assert.NotEmpty(t, paths.CardsDir)
	assert.NotEmpty(t, paths.TemplatesDir)
	assert.NotEmpty(t, paths.DatabasePath)
	assert.NotEmpty(t, paths.ConfigPath)

	// Check paths contain expected components
	assert.Contains(t, paths.BaseDir, ".samedi")
	assert.Contains(t, paths.PlansDir, "plans")
	assert.Contains(t, paths.CardsDir, "cards")
	assert.Contains(t, paths.DatabasePath, "sessions.db")
}

func TestPaths_EnsureDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	paths := &Paths{
		BaseDir:      filepath.Join(tmpDir, ".samedi"),
		PlansDir:     filepath.Join(tmpDir, ".samedi", "plans"),
		CardsDir:     filepath.Join(tmpDir, ".samedi", "cards"),
		TemplatesDir: filepath.Join(tmpDir, ".samedi", "templates"),
		BackupDir:    filepath.Join(tmpDir, "samedi-backups"),
		DatabasePath: filepath.Join(tmpDir, ".samedi", "sessions.db"),
		ConfigPath:   filepath.Join(tmpDir, ".samedi", "config.toml"),
	}

	// Ensure directories
	err := paths.EnsureDirectories()
	require.NoError(t, err)

	// Verify directories exist
	assert.DirExists(t, paths.BaseDir)
	assert.DirExists(t, paths.PlansDir)
	assert.DirExists(t, paths.CardsDir)
	assert.DirExists(t, paths.TemplatesDir)
	assert.DirExists(t, paths.BackupDir)
}

func TestPaths_Exists(t *testing.T) {
	tmpDir := t.TempDir()

	paths := &Paths{
		BaseDir:      filepath.Join(tmpDir, ".samedi"),
		PlansDir:     filepath.Join(tmpDir, ".samedi", "plans"),
		CardsDir:     filepath.Join(tmpDir, ".samedi", "cards"),
		TemplatesDir: filepath.Join(tmpDir, ".samedi", "templates"),
		BackupDir:    filepath.Join(tmpDir, "samedi-backups"),
	}

	// Directories don't exist yet
	assert.False(t, paths.Exists())

	// Create directories
	err := paths.EnsureDirectories()
	require.NoError(t, err)

	// Now they exist
	assert.True(t, paths.Exists())
}

func TestPaths_Clean(t *testing.T) {
	tmpDir := t.TempDir()

	paths := &Paths{
		BaseDir:      filepath.Join(tmpDir, ".samedi"),
		PlansDir:     filepath.Join(tmpDir, ".samedi", "plans"),
		CardsDir:     filepath.Join(tmpDir, ".samedi", "cards"),
		TemplatesDir: filepath.Join(tmpDir, ".samedi", "templates"),
		BackupDir:    filepath.Join(tmpDir, "samedi-backups"),
	}

	// Create directories
	err := paths.EnsureDirectories()
	require.NoError(t, err)
	assert.DirExists(t, paths.BaseDir)

	// Clean
	err = paths.Clean()
	require.NoError(t, err)

	// Verify base directory removed
	_, err = os.Stat(paths.BaseDir)
	assert.True(t, os.IsNotExist(err))
}

func TestPaths_PlanPath(t *testing.T) {
	paths := &Paths{
		PlansDir: "/home/user/.samedi/plans",
	}

	path := paths.PlanPath("rust-async")
	assert.Equal(t, "/home/user/.samedi/plans/rust-async.md", path)
}

func TestPaths_CardsPath(t *testing.T) {
	paths := &Paths{
		CardsDir: "/home/user/.samedi/cards",
	}

	path := paths.CardsPath("french-b1")
	assert.Equal(t, "/home/user/.samedi/cards/french-b1.cards.md", path)
}

func TestPaths_TemplatePath(t *testing.T) {
	paths := &Paths{
		TemplatesDir: "/home/user/.samedi/templates",
	}

	path := paths.TemplatePath("plan-generation")
	assert.Equal(t, "/home/user/.samedi/templates/plan-generation.md", path)
}
