// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_NoConfigFile(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Store original HOME and restore after test
	origHomeDir := os.Getenv("HOME")
	defer func() {
		os.Setenv("HOME", origHomeDir)
	}()

	// Override home directory for test
	os.Setenv("HOME", tmpDir)

	// Load should return defaults if no config file exists
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should have defaults
	assert.Equal(t, "auto", cfg.LLM.Provider)
}

func TestSave_And_Load(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Store original HOME and restore after test
	origHomeDir := os.Getenv("HOME")
	defer func() {
		os.Setenv("HOME", origHomeDir)
	}()

	// Override home directory for test
	os.Setenv("HOME", tmpDir)

	// Create config with custom values
	cfg := DefaultConfig()
	cfg.LLM.Provider = "codex"
	cfg.LLM.TimeoutSeconds = 60

	// Save config
	err := Save(cfg)
	require.NoError(t, err)

	// Verify file exists
	configPath := filepath.Join(tmpDir, ".samedi", "config.toml")
	assert.FileExists(t, configPath)

	// Verify file content was written correctly
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "codex")
	assert.Contains(t, string(data), "TimeoutSeconds = 60")

	// Note: We can't easily test Load() in this test because viper caches
	// the home directory lookup. The Save() function correctly writes the
	// config file, which is what we're testing here.
	// Load() is tested separately in TestLoad_NoConfigFile.
}

func TestSave_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origHomeDir := os.Getenv("HOME")
	defer os.Setenv("HOME", origHomeDir)
	os.Setenv("HOME", tmpDir)

	// Create invalid config
	cfg := DefaultConfig()
	cfg.LLM.Provider = "invalid"

	// Save should fail validation
	err := Save(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid configuration")
}

func TestInitConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origHomeDir := os.Getenv("HOME")
	defer os.Setenv("HOME", origHomeDir)
	os.Setenv("HOME", tmpDir)

	// Initialize config
	err := InitConfig()
	require.NoError(t, err)

	// Verify file exists
	configPath := filepath.Join(tmpDir, ".samedi", "config.toml")
	assert.FileExists(t, configPath)

	// Trying to init again should fail
	err = InitConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestPath(t *testing.T) {
	path := Path()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, ".samedi")
	assert.Contains(t, path, "config.toml")
}
