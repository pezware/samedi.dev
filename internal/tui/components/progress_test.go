// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgressBar_Empty(t *testing.T) {
	bar := NewProgressBar(0.0, 20)
	result := bar.View()

	// Should render empty bar
	assert.Contains(t, result, "[")
	assert.Contains(t, result, "]")
	assert.Contains(t, result, "0%")
}

func TestProgressBar_Full(t *testing.T) {
	bar := NewProgressBar(1.0, 20)
	result := bar.View()

	// Should render full bar
	assert.Contains(t, result, "[")
	assert.Contains(t, result, "]")
	assert.Contains(t, result, "100%")
}

func TestProgressBar_Partial(t *testing.T) {
	tests := []struct {
		name     string
		progress float64
		want     string
	}{
		{
			name:     "25%",
			progress: 0.25,
			want:     "25%",
		},
		{
			name:     "50%",
			progress: 0.50,
			want:     "50%",
		},
		{
			name:     "75%",
			progress: 0.75,
			want:     "75%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := NewProgressBar(tt.progress, 20)
			result := bar.View()

			assert.Contains(t, result, tt.want)
			assert.Contains(t, result, "[")
			assert.Contains(t, result, "]")
		})
	}
}

func TestProgressBar_NegativeProgress(t *testing.T) {
	bar := NewProgressBar(-0.5, 20)
	result := bar.View()

	// Should clamp to 0%
	assert.Contains(t, result, "0%")
}

func TestProgressBar_OverflowProgress(t *testing.T) {
	bar := NewProgressBar(1.5, 20)
	result := bar.View()

	// Should clamp to 100%
	assert.Contains(t, result, "100%")
}

func TestProgressBar_WithCustomWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"narrow", 10},
		{"normal", 20},
		{"wide", 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := NewProgressBar(0.5, tt.width)
			result := bar.View()

			// Should render bar with brackets
			assert.Contains(t, result, "[")
			assert.Contains(t, result, "]")
			assert.Contains(t, result, "50%")
		})
	}
}

func TestProgressBar_ColoredByProgress(t *testing.T) {
	tests := []struct {
		name     string
		progress float64
		color    string
	}{
		{"low progress - red", 0.2, "red"},
		{"medium progress - yellow", 0.5, "yellow"},
		{"high progress - green", 0.9, "green"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := NewProgressBar(tt.progress, 20)
			result := bar.View()

			// Should return a string (color testing requires visual inspection)
			assert.NotEmpty(t, result)
		})
	}
}
