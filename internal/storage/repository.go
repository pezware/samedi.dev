// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package storage

// Repository interfaces define storage operations.
// Implementations will be in domain-specific packages (plan, session, flashcard).

// Storage combines database and filesystem storage.
type Storage struct {
	DB         *SQLiteDB
	Filesystem *FilesystemStorage
	Paths      *Paths
}

// NewStorage creates a new storage instance.
func NewStorage(dbPath string, paths *Paths) (*Storage, error) {
	// Initialize database
	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		return nil, err
	}

	// Run migrations
	migrator := NewMigrator(db)
	if err := migrator.Migrate(); err != nil {
		db.Close()
		return nil, err
	}

	// Initialize filesystem
	fs := NewFilesystemStorage(paths)
	if err := fs.Initialize(); err != nil {
		db.Close()
		return nil, err
	}

	return &Storage{
		DB:         db,
		Filesystem: fs,
		Paths:      paths,
	}, nil
}

// Close closes all storage connections.
func (s *Storage) Close() error {
	return s.DB.Close()
}
