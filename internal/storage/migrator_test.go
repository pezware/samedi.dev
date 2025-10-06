// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package storage

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrator_Migrate(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	migrator := NewMigrator(db)

	// Run migrations
	err = migrator.Migrate()
	require.NoError(t, err)

	// Verify schema version
	var version int
	err = db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	require.NoError(t, err)
	assert.Equal(t, 1, version)

	// Verify tables created
	tables := []string{"plans", "sessions", "cards", "schema_migrations"}
	for _, table := range tables {
		var exists bool
		err = db.QueryRow(`
			SELECT COUNT(*) > 0
			FROM sqlite_master
			WHERE type='table' AND name=?
		`, table).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "table %s should exist", table)
	}
}

func TestMigrator_Migrate_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	migrator := NewMigrator(db)

	// Run migrations twice
	err = migrator.Migrate()
	require.NoError(t, err)

	err = migrator.Migrate()
	require.NoError(t, err)

	// Version should still be 1
	var version int
	err = db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	require.NoError(t, err)
	assert.Equal(t, 1, version)
}

func TestNewStorage(t *testing.T) {
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

	// Ensure base directory exists before creating storage
	err := paths.EnsureDirectories()
	require.NoError(t, err)

	storage, err := NewStorage(paths.DatabasePath, paths)
	require.NoError(t, err)
	require.NotNil(t, storage)
	defer storage.Close()

	// Verify database is initialized
	var version int
	err = storage.DB.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	require.NoError(t, err)
	assert.Equal(t, 1, version)

	// Verify directories created
	assert.DirExists(t, paths.BaseDir)
	assert.DirExists(t, paths.PlansDir)
	assert.DirExists(t, paths.CardsDir)
}
