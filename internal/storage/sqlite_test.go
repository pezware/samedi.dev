// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package storage

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQLiteDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db)

	defer db.Close()

	// Test connection
	err = db.DB().Ping()
	assert.NoError(t, err)
}

func TestSQLiteDB_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)

	err = db.Close()
	assert.NoError(t, err)

	// Closing again should return an error (database is already closed)
	// In practice, sql.DB.Close() returns nil even when already closed,
	// so we just verify it doesn't panic
	err = db.Close()
	// sql.DB allows multiple Close() calls without error
	assert.NoError(t, err)
}

func TestSQLiteDB_Exec(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create test table
	_, err = db.Exec(`CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)

	// Insert data
	result, err := db.Exec(`INSERT INTO test (name) VALUES (?)`, "Alice")
	require.NoError(t, err)

	rowsAffected, err := result.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)
}

func TestSQLiteDB_QueryRow(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create and populate test table
	_, err = db.Exec(`CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO test (name) VALUES (?)`, "Bob")
	require.NoError(t, err)

	// Query single row
	var name string
	err = db.QueryRow(`SELECT name FROM test WHERE id = 1`).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "Bob", name)
}

func TestSQLiteDB_Query(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create and populate test table
	_, err = db.Exec(`CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO test (name) VALUES (?), (?)`, "Alice", "Bob")
	require.NoError(t, err)

	// Query multiple rows
	rows, err := db.Query(`SELECT name FROM test ORDER BY id`)
	require.NoError(t, err)
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		require.NoError(t, err)
		names = append(names, name)
	}

	assert.Equal(t, []string{"Alice", "Bob"}, names)
}

func TestSQLiteDB_Transaction(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create test table
	_, err = db.Exec(`CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)

	// Begin transaction
	tx, err := db.Begin()
	require.NoError(t, err)

	// Insert in transaction
	_, err = tx.Exec(`INSERT INTO test (name) VALUES (?)`, "Charlie")
	require.NoError(t, err)

	// Rollback
	err = tx.Rollback()
	require.NoError(t, err)

	// Verify data not committed
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM test`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
