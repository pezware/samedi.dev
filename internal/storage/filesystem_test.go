// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package storage

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesystemStorage_Initialize(t *testing.T) {
	tmpDir := t.TempDir()

	paths := &Paths{
		BaseDir:      filepath.Join(tmpDir, ".samedi"),
		PlansDir:     filepath.Join(tmpDir, ".samedi", "plans"),
		CardsDir:     filepath.Join(tmpDir, ".samedi", "cards"),
		TemplatesDir: filepath.Join(tmpDir, ".samedi", "templates"),
		BackupDir:    filepath.Join(tmpDir, "samedi-backups"),
	}

	fs := NewFilesystemStorage(paths)

	err := fs.Initialize()
	require.NoError(t, err)

	// Verify directories created
	assert.DirExists(t, paths.BaseDir)
	assert.DirExists(t, paths.PlansDir)
	assert.DirExists(t, paths.CardsDir)
}

func TestFilesystemStorage_WriteAndReadFile(t *testing.T) {
	tmpDir := t.TempDir()

	paths := &Paths{
		BaseDir:      filepath.Join(tmpDir, ".samedi"),
		PlansDir:     filepath.Join(tmpDir, ".samedi", "plans"),
		CardsDir:     filepath.Join(tmpDir, ".samedi", "cards"),
		TemplatesDir: filepath.Join(tmpDir, ".samedi", "templates"),
		BackupDir:    filepath.Join(tmpDir, "samedi-backups"),
	}

	fs := NewFilesystemStorage(paths)
	err := fs.Initialize()
	require.NoError(t, err)

	// Write file
	testPath := filepath.Join(paths.PlansDir, "test.md")
	testData := []byte("# Test Plan\n\nThis is a test.")

	err = fs.WriteFile(testPath, testData)
	require.NoError(t, err)

	// Read file
	data, err := fs.ReadFile(testPath)
	require.NoError(t, err)
	assert.Equal(t, testData, data)
}

func TestFilesystemStorage_DeleteFile(t *testing.T) {
	tmpDir := t.TempDir()

	paths := &Paths{
		BaseDir:      filepath.Join(tmpDir, ".samedi"),
		PlansDir:     filepath.Join(tmpDir, ".samedi", "plans"),
		CardsDir:     filepath.Join(tmpDir, ".samedi", "cards"),
		TemplatesDir: filepath.Join(tmpDir, ".samedi", "templates"),
		BackupDir:    filepath.Join(tmpDir, "samedi-backups"),
	}

	fs := NewFilesystemStorage(paths)
	err := fs.Initialize()
	require.NoError(t, err)

	// Write file
	testPath := filepath.Join(paths.PlansDir, "test.md")
	err = fs.WriteFile(testPath, []byte("test"))
	require.NoError(t, err)
	assert.True(t, fs.FileExists(testPath))

	// Delete file
	err = fs.DeleteFile(testPath)
	require.NoError(t, err)
	assert.False(t, fs.FileExists(testPath))
}

func TestFilesystemStorage_FileExists(t *testing.T) {
	tmpDir := t.TempDir()

	paths := &Paths{
		BaseDir:      filepath.Join(tmpDir, ".samedi"),
		PlansDir:     filepath.Join(tmpDir, ".samedi", "plans"),
		CardsDir:     filepath.Join(tmpDir, ".samedi", "cards"),
		TemplatesDir: filepath.Join(tmpDir, ".samedi", "templates"),
		BackupDir:    filepath.Join(tmpDir, "samedi-backups"),
	}

	fs := NewFilesystemStorage(paths)
	err := fs.Initialize()
	require.NoError(t, err)

	testPath := filepath.Join(paths.PlansDir, "test.md")

	// File doesn't exist initially
	assert.False(t, fs.FileExists(testPath))

	// Write file
	err = fs.WriteFile(testPath, []byte("test"))
	require.NoError(t, err)

	// Now it exists
	assert.True(t, fs.FileExists(testPath))
}
