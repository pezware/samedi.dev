// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"testing"

	"github.com/pezware/samedi.dev/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectStopInputs_ValidInputs_Success(t *testing.T) {
	tests := []struct {
		name              string
		opts              stopOptions
		expectedNote      string
		expectedArtifacts []string
	}{
		{
			name: "with note flag",
			opts: stopOptions{
				noPrompt:    true,
				notes:       stringPtr("Completed chapter 3"),
				noteFlagSet: true,
			},
			expectedNote:      "Completed chapter 3",
			expectedArtifacts: []string{},
		},
		{
			name: "with artifacts flag",
			opts: stopOptions{
				noPrompt:        true,
				artifacts:       &[]string{"https://example.com", "file.md"},
				artifactFlagSet: true,
			},
			expectedNote:      "",
			expectedArtifacts: []string{"https://example.com", "file.md"},
		},
		{
			name: "no flags provided",
			opts: stopOptions{
				noPrompt: true,
			},
			expectedNote:      "",
			expectedArtifacts: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note, artifacts, err := collectStopInputs(tt.opts)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedNote, note)
			assert.Len(t, artifacts, len(tt.expectedArtifacts))
		})
	}
}

func TestPromptForStopNote_ValidInput_ReturnsNote(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedNote string
		expectError  bool
	}{
		{
			name:         "simple note",
			input:        "Completed all exercises\n",
			expectedNote: "Completed all exercises",
			expectError:  false,
		},
		{
			name:         "empty note",
			input:        "\n",
			expectedNote: "",
			expectError:  false,
		},
		{
			name:         "EOF without newline",
			input:        "",
			expectedNote: "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := testutil.MockReader(tt.input)
			writer := testutil.MockWriter()

			note, err := promptForStopNote(reader, writer)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNote, note)
			}

			output := writer.String()
			assert.Contains(t, output, "Session notes (optional):")
		})
	}
}

func TestPromptForArtifacts_MultipleArtifacts(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedArtifacts []string
		expectError       bool
	}{
		{
			name:              "single artifact",
			input:             "https://github.com/user/repo\n\n",
			expectedArtifacts: []string{"https://github.com/user/repo"},
			expectError:       false,
		},
		{
			name:              "multiple artifacts",
			input:             "file1.md\nfile2.go\nhttps://example.com\n\n",
			expectedArtifacts: []string{"file1.md", "file2.go", "https://example.com"},
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := testutil.MockReader(tt.input)
			writer := testutil.MockWriter()

			artifacts, err := promptForArtifacts(reader, writer)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedArtifacts, artifacts)
			}
		})
	}
}

func TestStopCmd_FlagsDefaultValues(t *testing.T) {
	cmd := stopCmd()

	// Check default values
	note := cmd.Flags().Lookup("note")
	require.NotNil(t, note)
	assert.Equal(t, "", note.DefValue)

	artifact := cmd.Flags().Lookup("artifact")
	require.NotNil(t, artifact)
	assert.Equal(t, "[]", artifact.DefValue)

	auto := cmd.Flags().Lookup("auto")
	require.NotNil(t, auto)
	assert.Equal(t, "false", auto.DefValue)
}
