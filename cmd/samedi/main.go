// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func main() {
	// TODO(#1): Implement CLI with cobra
	// For now, just show version info
	fmt.Printf("samedi version %s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
	fmt.Println("\nA learning operating system for the terminal.")
	fmt.Println("\nRun 'samedi --help' for usage information.")

	os.Exit(0)
}
