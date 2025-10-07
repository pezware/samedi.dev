// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package llm

import (
	"context"
	"fmt"
	"strings"
)

// MockProvider is a mock LLM provider for testing.
// It returns predefined responses based on prompt patterns.
type MockProvider struct {
	// Responses maps prompt substrings to canned responses
	Responses map[string]string

	// DefaultResponse is returned if no pattern matches
	DefaultResponse string

	// CallCount tracks how many times Call was invoked
	CallCount int

	// LastPrompt stores the most recent prompt
	LastPrompt string

	// ShouldError if true, Call returns an error
	ShouldError bool

	// ErrorMessage is the error message to return when ShouldError is true
	ErrorMessage string
}

// NewMockProvider creates a new mock provider with default responses.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		Responses:       make(map[string]string),
		DefaultResponse: defaultPlanMarkdown,
	}
}

// Call implements the Provider interface.
// Returns a canned response based on prompt content.
func (m *MockProvider) Call(_ context.Context, prompt string) (string, error) {
	m.CallCount++
	m.LastPrompt = prompt

	if m.ShouldError {
		return "", &ProviderError{
			Provider:  "mock",
			Err:       fmt.Errorf("%s", m.ErrorMessage),
			Retryable: false,
		}
	}

	// Check for pattern matches
	for pattern, response := range m.Responses {
		if strings.Contains(strings.ToLower(prompt), strings.ToLower(pattern)) {
			return response, nil
		}
	}

	return m.DefaultResponse, nil
}

// Reset clears the call history.
func (m *MockProvider) Reset() {
	m.CallCount = 0
	m.LastPrompt = ""
	m.ShouldError = false
}

// defaultPlanMarkdown is a valid plan response for testing.
const defaultPlanMarkdown = `---
id: test-plan
title: Test Learning Plan
created: 2024-01-15T10:00:00Z
updated: 2024-01-15T10:00:00Z
total_hours: 3
status: not-started
tags:
  - test
---

# Test Learning Plan

## Chunk 1: Introduction {#chunk-001}

**Duration**: 1 hour
**Status**: not-started
**Objectives**:

- Learn the basics
- Understand key concepts

**Resources**:

- Example tutorial
- Documentation

**Deliverable**: Complete exercises

---

## Chunk 2: Advanced Topics {#chunk-002}

**Duration**: 1.5 hours
**Status**: not-started
**Objectives**:

- Deep dive into advanced features
- Build a project

**Resources**:

- Advanced guide

**Deliverable**: Working project

---

## Chunk 3: Review {#chunk-003}

**Duration**: 30 minutes
**Status**: not-started
`
