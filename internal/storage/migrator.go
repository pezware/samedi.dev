// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package storage

import (
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration represents a database migration.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// Migrator handles database schema migrations.
type Migrator struct {
	db *SQLiteDB
}

// NewMigrator creates a new migrator instance.
func NewMigrator(db *SQLiteDB) *Migrator {
	return &Migrator{db: db}
}

// Migrate runs all pending migrations.
func (m *Migrator) Migrate() error {
	// Get current schema version
	currentVersion, err := m.getCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Load all migrations
	migrations, err := m.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			if err := m.applyMigration(migration); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
			}
		}
	}

	return nil
}

// getCurrentVersion returns the current schema version.
func (m *Migrator) getCurrentVersion() (int, error) {
	// Check if schema_migrations table exists
	var tableExists bool
	err := m.db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM sqlite_master
		WHERE type='table' AND name='schema_migrations'
	`).Scan(&tableExists)
	if err != nil {
		return 0, fmt.Errorf("failed to check schema_migrations table: %w", err)
	}

	if !tableExists {
		return 0, nil
	}

	// Get latest version
	var version int
	err = m.db.QueryRow(`
		SELECT COALESCE(MAX(version), 0)
		FROM schema_migrations
	`).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get current version: %w", err)
	}

	return version, nil
}

// loadMigrations loads all migration files from embedded filesystem.
func (m *Migrator) loadMigrations() ([]Migration, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Parse version from filename (e.g., "001_initial_schema.sql")
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) != 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		// Read migration SQL
		sql, err := migrationsFS.ReadFile(fmt.Sprintf("migrations/%s", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    strings.TrimSuffix(parts[1], ".sql"),
			SQL:     string(sql),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// applyMigration applies a single migration.
func (m *Migrator) applyMigration(migration Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		// Ignore error on rollback - it's expected if commit succeeded
		//nolint:errcheck // Transaction cleanup
		tx.Rollback()
	}()

	// Execute migration SQL
	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration (if not already recorded by the migration itself)
	// The initial migration includes its own INSERT, so we check first
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", migration.Version).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if migration is recorded: %w", err)
	}

	if !exists {
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.Version); err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
