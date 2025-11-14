// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package testutil

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/session"
	"github.com/pezware/samedi.dev/internal/storage"
	"github.com/stretchr/testify/require"
)

// NewTestPlan creates a test plan with default values.
func NewTestPlan(t *testing.T) *plan.Plan {
	t.Helper()
	return &plan.Plan{
		ID:    "test-plan-id",
		Title: "Test Plan",
		Chunks: []plan.Chunk{
			{
				ID:          "chunk-001",
				Title:       "Test Chunk 1",
				Duration:    60,
				Status:      plan.StatusNotStarted,
				Objectives:  []string{"Learn basics", "Practice exercises"},
				Resources:   []string{"https://example.com/docs"},
				Deliverable: "Working example",
			},
			{
				ID:       "chunk-002",
				Title:    "Test Chunk 2",
				Duration: 90,
				Status:   plan.StatusInProgress,
			},
			{
				ID:       "chunk-003",
				Title:    "Test Chunk 3",
				Duration: 45,
				Status:   plan.StatusCompleted,
			},
		},
		TotalHours: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Status:     plan.StatusInProgress,
	}
}

// NewTestSession creates a test session with default values.
func NewTestSession(t *testing.T) *session.Session {
	t.Helper()
	now := time.Now()
	return &session.Session{
		ID:        "test-session-id",
		PlanID:    "test-plan-id",
		ChunkID:   "chunk-001",
		StartTime: now,
		Notes:     "Test session notes",
		Artifacts: []string{"https://example.com/artifact"},
	}
}

// NewTestDB creates an in-memory SQLite database for testing.
// Note: This does not run migrations. Use NewTestSQLiteDB for a fully initialized database.
func NewTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "Failed to create test database")

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

// NewTestSQLiteDB creates a SQLite database with migrations for testing.
func NewTestSQLiteDB(t *testing.T) (*storage.SQLiteDB, func()) {
	t.Helper()

	// Use a temp file for SQLite
	tmpFile := filepath.Join(t.TempDir(), "test.db")

	sqliteDB, err := storage.NewSQLiteDB(tmpFile)
	require.NoError(t, err, "Failed to create test SQLite database")

	// Run migrations
	migrator := storage.NewMigrator(sqliteDB)
	err = migrator.Migrate()
	require.NoError(t, err, "Failed to run migrations")

	cleanup := func() {
		sqliteDB.Close()
		os.Remove(tmpFile)
	}

	return sqliteDB, cleanup
}

// NewTestPaths creates temporary paths for testing.
func NewTestPaths(t *testing.T) (*storage.Paths, func()) {
	t.Helper()

	tmpDir := t.TempDir()

	paths := &storage.Paths{
		BaseDir:      tmpDir,
		PlansDir:     filepath.Join(tmpDir, "plans"),
		CardsDir:     filepath.Join(tmpDir, "cards"),
		TemplatesDir: filepath.Join(tmpDir, "templates"),
		BackupDir:    filepath.Join(tmpDir, "backups"),
		DatabasePath: filepath.Join(tmpDir, "samedi.db"),
		ConfigPath:   filepath.Join(tmpDir, "config.toml"),
	}

	// Create directories
	err := paths.EnsureDirectories()
	require.NoError(t, err, "Failed to create test directories")

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return paths, cleanup
}

// LoadFixture loads a test fixture file from testdata directory.
func LoadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err, "Failed to load fixture: %s", name)
	return data
}

// LoadFixtureString loads a test fixture file as a string.
func LoadFixtureString(t *testing.T, name string) string {
	t.Helper()
	return string(LoadFixture(t, name))
}
