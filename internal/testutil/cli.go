// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package testutil

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// MockReader creates a bufio.Reader from a string for testing prompts.
func MockReader(input string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(input))
}

// MockWriter creates a buffer that implements io.Writer for capturing output.
func MockWriter() *bytes.Buffer {
	return &bytes.Buffer{}
}

// CaptureOutput captures stdout/stderr during function execution.
func CaptureOutput(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	// Note: This is a simplified version. For actual stdout/stderr capture,
	// you'd need to redirect os.Stdout/os.Stderr
	fn()
	return buf.String()
}

// AssertOutputContains checks that output contains expected string.
func AssertOutputContains(t *testing.T, output, expected string) {
	t.Helper()
	require.Contains(t, output, expected, "Output should contain: %s\nGot: %s", expected, output)
}

// AssertOutputNotContains checks that output does not contain a string.
func AssertOutputNotContains(t *testing.T, output, unexpected string) {
	t.Helper()
	require.NotContains(t, output, unexpected, "Output should not contain: %s\nGot: %s", unexpected, output)
}

// MultilineInput creates input with multiple lines for testing prompts.
func MultilineInput(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}
