// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"os"

	"github.com/pezware/samedi.dev/internal/cli"
)

// Version information (set via ldflags during build)
// These are passed to the CLI package for display
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func main() {
	// Set version information in CLI package
	cli.Version = Version
	cli.Commit = Commit
	cli.BuildDate = BuildDate

	// Execute CLI
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
