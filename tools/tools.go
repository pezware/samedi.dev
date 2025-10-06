// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

//go:build tools
// +build tools

package tools

// This file ensures tool dependencies are tracked in go.mod.
// Import tools here to pin their versions in go.mod.
//
// Install all tools with:
//   make install-tools
//
// Or manually:
//   go install $(go list -f '{{.ImportPath}}' -tags=tools ./tools)

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint" // Linter
	_ "golang.org/x/vuln/cmd/govulncheck"                   // Vulnerability scanner
	_ "gotest.tools/gotestsum"                              // Better test output
)

// Note: Some tools don't support being imported as libraries.
// For those, install via Makefile with pinned versions.
