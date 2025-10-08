// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"fmt"

	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/session"
	"github.com/spf13/cobra"
)

// ChunkDisplayInfo contains all information needed to display chunk details.
type ChunkDisplayInfo struct {
	Chunk          *plan.Chunk
	SessionStats   *session.ChunkStats
	RecentSessions []*session.Session
	PlanID         string
}

// getChunkDisplayInfo fetches chunk details and session statistics.
func getChunkDisplayInfo(cmd *cobra.Command, planID, chunkID string) (*ChunkDisplayInfo, error) {
	// Get plan service to fetch chunk
	planSvc, err := getPlanService(cmd, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get plan service: %w", err)
	}

	// Get the chunk
	chunk, err := planSvc.GetChunk(context.Background(), planID, chunkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chunk: %w", err)
	}

	// Get session service for statistics
	sessionSvc, err := getSessionService(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get session service: %w", err)
	}

	// Get session statistics
	stats, err := sessionSvc.GetChunkStats(context.Background(), planID, chunkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chunk stats: %w", err)
	}

	// Get recent sessions for this chunk
	sessions, err := sessionSvc.GetChunkSessions(context.Background(), planID, chunkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chunk sessions: %w", err)
	}

	// Keep only the 3 most recent sessions
	recentSessions := sessions
	if len(recentSessions) > 3 {
		recentSessions = recentSessions[:3]
	}

	return &ChunkDisplayInfo{
		Chunk:          chunk,
		SessionStats:   stats,
		RecentSessions: recentSessions,
		PlanID:         planID,
	}, nil
}

// displayChunkDetails displays comprehensive chunk information.
func displayChunkDetails(info *ChunkDisplayInfo) {
	chunk := info.Chunk
	stats := info.SessionStats

	// Header: Title and ID
	if chunk.Title != "" {
		fmt.Printf("\n%s\n", chunk.Title)
		fmt.Printf("ID: %s\n", chunk.ID)
	} else {
		fmt.Printf("\nChunk: %s\n", chunk.ID)
	}

	// Status and Duration
	statusIcon := getStatusIcon(chunk.Status)
	fmt.Printf("Status: %s %s\n", statusIcon, chunk.Status)
	fmt.Printf("Duration: %d min", chunk.Duration)

	// Progress information
	if stats.TotalDuration > 0 {
		progress := 0
		if chunk.Duration > 0 {
			progress = (stats.TotalDuration * 100) / chunk.Duration
			if progress > 100 {
				progress = 100
			}
		}
		fmt.Printf(" (%d/%d min, %d%%)", stats.TotalDuration, chunk.Duration, progress)
	}
	fmt.Println()

	// Session count
	if stats.SessionCount > 0 {
		fmt.Printf("Sessions: %d\n", stats.SessionCount)
	}

	// Objectives
	if len(chunk.Objectives) > 0 {
		fmt.Println("\nObjectives:")
		for _, obj := range chunk.Objectives {
			fmt.Printf("  • %s\n", obj)
		}
	}

	// Resources
	if len(chunk.Resources) > 0 {
		fmt.Println("\nResources:")
		for _, res := range chunk.Resources {
			fmt.Printf("  • %s\n", res)
		}
	}

	// Deliverable
	if chunk.Deliverable != "" {
		fmt.Printf("\nDeliverable: %s\n", chunk.Deliverable)
	}

	// Recent sessions
	if len(info.RecentSessions) > 0 {
		fmt.Println("\nRecent sessions:")
		for _, sess := range info.RecentSessions {
			if sess.IsActive() {
				fmt.Printf("  → Active (started %s)\n", sess.StartTime.Format("Jan 2 15:04"))
			} else {
				fmt.Printf("  ✓ %s - %d min", sess.EndTime.Format("Jan 2 15:04"), sess.Duration)
				if sess.Notes != "" {
					fmt.Printf(" - %s", sess.Notes)
				}
				fmt.Println()
			}
		}
	}
}

// getStatusIcon returns an icon/symbol for the chunk status.
func getStatusIcon(status plan.Status) string {
	switch status {
	case plan.StatusNotStarted:
		return "○"
	case plan.StatusInProgress:
		return "◐"
	case plan.StatusCompleted:
		return "●"
	case plan.StatusSkipped:
		return "⊘"
	default:
		return "?"
	}
}
