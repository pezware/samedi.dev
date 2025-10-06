// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// SQLiteDB wraps a SQLite database connection.
type SQLiteDB struct {
	db *sql.DB
}

// NewSQLiteDB creates a new SQLite database connection.
func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	// Open database with WAL mode for better concurrency
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000", dbPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite doesn't handle concurrent writes well
	db.SetMaxIdleConns(1)

	return &SQLiteDB{db: db}, nil
}

// Close closes the database connection.
func (s *SQLiteDB) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// DB returns the underlying sql.DB for direct access.
func (s *SQLiteDB) DB() *sql.DB {
	return s.db
}

// Begin starts a new transaction.
func (s *SQLiteDB) Begin() (*sql.Tx, error) {
	return s.db.Begin()
}

// Exec executes a query without returning rows.
func (s *SQLiteDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

// Query executes a query that returns rows.
func (s *SQLiteDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.Query(query, args...)
}

// QueryRow executes a query that returns a single row.
func (s *SQLiteDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.db.QueryRow(query, args...)
}
