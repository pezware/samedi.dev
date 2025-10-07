// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package plan

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// frontmatterDelimiter is the YAML frontmatter delimiter
	frontmatterDelimiter = "---"
)

var (
	// Regular expressions for parsing chunks
	chunkHeaderRegex = regexp.MustCompile(`^##\s+Chunk\s+\d+:\s+(.+?)\s+\{#([^}]+)\}\s*$`)
	durationRegex    = regexp.MustCompile(`^\*\*Duration\*\*:\s*(.+)$`)
	statusRegex      = regexp.MustCompile(`^\*\*Status\*\*:\s*(.+)$`)
	deliverableRegex = regexp.MustCompile(`^\*\*Deliverable\*\*:\s*(.+)$`)
	objectivesRegex  = regexp.MustCompile(`^\*\*Objectives\*\*:\s*$`)
	resourcesRegex   = regexp.MustCompile(`^\*\*Resources\*\*:\s*$`)
)

// ParseFile reads a plan markdown file and returns a Plan struct.
func ParseFile(path string) (*Plan, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return Parse(string(content))
}

// Parse parses markdown content with YAML frontmatter into a Plan struct.
func Parse(content string) (*Plan, error) {
	// Split frontmatter and body
	frontmatter, body, err := splitFrontmatter(content)
	if err != nil {
		return nil, err
	}

	// Parse frontmatter into Plan
	var plan Plan
	if err := yaml.Unmarshal([]byte(frontmatter), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Parse chunks from body
	chunks, err := parseChunks(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chunks: %w", err)
	}
	plan.Chunks = chunks

	return &plan, nil
}

// splitFrontmatter separates YAML frontmatter from markdown body.
// Returns frontmatter, body, and error.
func splitFrontmatter(content string) (string, string, error) {
	lines := strings.Split(content, "\n")

	// Check for frontmatter delimiter at start
	if len(lines) == 0 || lines[0] != frontmatterDelimiter {
		return "", "", fmt.Errorf("missing frontmatter delimiter at start")
	}

	// Find closing delimiter
	endIndex := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == frontmatterDelimiter {
			endIndex = i
			break
		}
	}

	if endIndex == -1 {
		return "", "", fmt.Errorf("missing closing frontmatter delimiter")
	}

	frontmatter := strings.Join(lines[1:endIndex], "\n")
	body := strings.Join(lines[endIndex+1:], "\n")

	return frontmatter, body, nil
}

// parseChunks extracts chunk information from markdown body.
func parseChunks(body string) ([]Chunk, error) {
	var chunks []Chunk
	var currentChunk *Chunk
	var inObjectives, inResources bool

	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check for chunk header
		if matches := chunkHeaderRegex.FindStringSubmatch(trimmed); matches != nil {
			if currentChunk != nil {
				chunks = append(chunks, *currentChunk)
			}
			currentChunk = &Chunk{
				ID:     matches[2],
				Title:  matches[1],
				Status: StatusNotStarted,
			}
			inObjectives, inResources = false, false
			continue
		}

		if currentChunk == nil {
			continue
		}

		// Parse chunk metadata and update section flags
		parsed, newInObjectives, newInResources := parseChunkLine(trimmed, currentChunk, inObjectives, inResources)
		if parsed {
			inObjectives, inResources = newInObjectives, newInResources
		}
	}

	if currentChunk != nil {
		chunks = append(chunks, *currentChunk)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading chunks: %w", err)
	}

	return chunks, nil
}

// parseChunkLine processes a single line of chunk content.
// Returns whether the line was processed and updated section flags.
func parseChunkLine(trimmed string, chunk *Chunk, inObjectives, inResources bool) (bool, bool, bool) {
	// Parse duration
	if matches := durationRegex.FindStringSubmatch(trimmed); matches != nil {
		duration, err := parseDuration(matches[1])
		if err == nil {
			chunk.Duration = duration
		}
		return true, false, false
	}

	// Parse status
	if matches := statusRegex.FindStringSubmatch(trimmed); matches != nil {
		chunk.Status = Status(strings.TrimSpace(matches[1]))
		return true, false, false
	}

	// Parse deliverable
	if matches := deliverableRegex.FindStringSubmatch(trimmed); matches != nil {
		chunk.Deliverable = strings.TrimSpace(matches[1])
		return true, false, false
	}

	// Check for objectives section
	if objectivesRegex.MatchString(trimmed) {
		return true, true, false
	}

	// Check for resources section
	if resourcesRegex.MatchString(trimmed) {
		return true, false, true
	}

	// Parse list items
	if isListItem(trimmed) {
		item := extractListItem(trimmed)
		if item != "" {
			if inObjectives {
				chunk.Objectives = append(chunk.Objectives, item)
			} else if inResources {
				chunk.Resources = append(chunk.Resources, item)
			}
		}
		return true, inObjectives, inResources
	}

	// End of list sections on non-list, non-empty line
	if trimmed != "" && !isListItem(trimmed) {
		return true, false, false
	}

	return false, inObjectives, inResources
}

// isListItem checks if a line is a markdown list item.
func isListItem(line string) bool {
	return (strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*")) && line != frontmatterDelimiter
}

// extractListItem extracts the content from a markdown list item.
func extractListItem(line string) string {
	item := strings.TrimPrefix(line, "-")
	item = strings.TrimPrefix(item, "*")
	return strings.TrimSpace(item)
}

// parseDuration converts duration strings like "1 hour", "90 minutes" to minutes.
func parseDuration(s string) (int, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	// Try to parse as number + unit
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid format: expected '<number> <unit>'")
	}

	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %w", err)
	}

	unit := parts[1]
	// Handle plural forms
	unit = strings.TrimSuffix(unit, "s")

	switch unit {
	case "hour", "hr", "h":
		return int(value * 60), nil
	case "minute", "min", "m":
		return int(value), nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}

// Format serializes a Plan to markdown with YAML frontmatter.
func Format(plan *Plan) (string, error) {
	var buf bytes.Buffer

	// Write frontmatter
	buf.WriteString("---\n")
	frontmatter, err := yaml.Marshal(plan)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}
	buf.Write(frontmatter)
	buf.WriteString("---\n\n")

	// Write title
	buf.WriteString(fmt.Sprintf("# %s\n\n", plan.Title))

	// Write chunks
	for i, chunk := range plan.Chunks {
		if i > 0 {
			buf.WriteString("---\n\n")
		}

		// Chunk header
		buf.WriteString(fmt.Sprintf("## Chunk %d: %s {#%s}\n", i+1, chunk.Title, chunk.ID))

		// Duration
		hours := float64(chunk.Duration) / 60.0
		if hours == float64(int(hours)) {
			buf.WriteString(fmt.Sprintf("**Duration**: %d hour", int(hours)))
		} else {
			buf.WriteString(fmt.Sprintf("**Duration**: %.1f hours", hours))
		}
		if hours != 1.0 && int(hours) != 1 {
			// Add plural 's' if not exactly 1 hour
			buf.WriteString("s")
		}
		buf.WriteString("\n")

		// Status
		buf.WriteString(fmt.Sprintf("**Status**: %s\n", chunk.Status))

		// Objectives
		if len(chunk.Objectives) > 0 {
			buf.WriteString("**Objectives**:\n")
			for _, obj := range chunk.Objectives {
				buf.WriteString(fmt.Sprintf("- %s\n", obj))
			}
			buf.WriteString("\n")
		}

		// Resources
		if len(chunk.Resources) > 0 {
			buf.WriteString("**Resources**:\n")
			for _, res := range chunk.Resources {
				buf.WriteString(fmt.Sprintf("- %s\n", res))
			}
			buf.WriteString("\n")
		}

		// Deliverable
		if chunk.Deliverable != "" {
			buf.WriteString(fmt.Sprintf("**Deliverable**: %s\n\n", chunk.Deliverable))
		}
	}

	return buf.String(), nil
}
