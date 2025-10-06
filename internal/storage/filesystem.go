// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package storage

import (
	"fmt"
	"os"
)

// FilesystemStorage handles file operations for plans and cards.
type FilesystemStorage struct {
	paths *Paths
}

// NewFilesystemStorage creates a new filesystem storage instance.
func NewFilesystemStorage(paths *Paths) *FilesystemStorage {
	return &FilesystemStorage{
		paths: paths,
	}
}

// Initialize ensures all directories exist.
func (fs *FilesystemStorage) Initialize() error {
	return fs.paths.EnsureDirectories()
}

// ReadFile reads a file from the filesystem.
func (fs *FilesystemStorage) ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return data, nil
}

// WriteFile writes data to a file with secure permissions.
func (fs *FilesystemStorage) WriteFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

// DeleteFile removes a file from the filesystem.
func (fs *FilesystemStorage) DeleteFile(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}
	return nil
}

// FileExists checks if a file exists.
func (fs *FilesystemStorage) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Paths returns the filesystem paths.
func (fs *FilesystemStorage) Paths() *Paths {
	return fs.paths
}
