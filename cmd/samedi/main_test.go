// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionVariables_HaveDefaultValues(t *testing.T) {
	// Version variables should have default values
	// These can be overridden at build time via ldflags

	assert.NotEmpty(t, Version)
	assert.NotEmpty(t, Commit)
	assert.NotEmpty(t, BuildDate)

	// Default values
	assert.Equal(t, "dev", Version)
	assert.Equal(t, "none", Commit)
	assert.Equal(t, "unknown", BuildDate)
}

func TestVersionVariables_CanBeModified(t *testing.T) {
	// Store original values
	origVersion := Version
	origCommit := Commit
	origBuildDate := BuildDate

	// Modify variables (simulating ldflags build-time injection)
	Version = "1.0.0"
	Commit = "abc123"
	BuildDate = "2025-01-15"

	assert.Equal(t, "1.0.0", Version)
	assert.Equal(t, "abc123", Commit)
	assert.Equal(t, "2025-01-15", BuildDate)

	// Restore original values
	Version = origVersion
	Commit = origCommit
	BuildDate = origBuildDate
}

func TestMain_PackageImports(t *testing.T) {
	// Test that required packages are importable
	// This ensures the main package compiles correctly

	// The main function itself is hard to test directly since it calls os.Exit
	// But we can verify the package structure is valid
	assert.NotEmpty(t, Version, "Version should be accessible")
}
