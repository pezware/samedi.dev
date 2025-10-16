// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pezware/samedi.dev/internal/plan"
	"github.com/pezware/samedi.dev/internal/session"
	"github.com/spf13/cobra"
)

// startCmd creates the `samedi start` command for starting a learning session.
func startCmd() *cobra.Command {
	var (
		notes      string
		noPrompt   bool
		showChunks bool
	)

	cmd := &cobra.Command{
		Use:   "start <plan-id> [chunk-id]",
		Short: "Start a learning session",
		Long: `Start a new learning session for a specific plan and optional chunk.

The session timer begins immediately and tracks your learning time.
Only one session can be active at a time.

Examples:
  samedi start french-b1
  samedi start french-b1 chunk-003
  samedi start rust-async chunk-015 --note "Working on tokio tutorial"`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			noteFlagSet := cmd.Flags().Changed("note")
			if err := executeStart(cmd, args, startOptions{
				note:         &notes,
				noteFlagSet:  noteFlagSet,
				noPrompt:     noPrompt,
				showAllChunk: showChunks,
			}); err != nil {
				exitWithError("%v", err)
			}
		},
	}

	// Flags
	cmd.Flags().StringVar(&notes, "note", "", "initial notes for the session")
	cmd.Flags().BoolVar(&noPrompt, "no-prompt", false, "skip interactive prompts")
	cmd.Flags().BoolVar(&showChunks, "show-chunks", false, "display chunk details before prompting")

	return cmd
}

type startOptions struct {
	note         *string
	noteFlagSet  bool
	noPrompt     bool
	showAllChunk bool
}

func executeStart(cmd *cobra.Command, args []string, opts startOptions) error {
	planID, chunkID, note, err := gatherStartInputs(cmd, args, opts)
	if err != nil {
		return err
	}

	// Initialize session service
	svc, err := getSessionService(cmd)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Prepare start request
	req := session.StartRequest{
		PlanID:  planID,
		ChunkID: chunkID,
		Notes:   note,
	}

	// Start session
	sess, err := svc.Start(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	// Display session started message
	fmt.Printf("→ Session started: %s", sess.PlanID)
	if sess.ChunkID != "" {
		fmt.Printf(" (%s)", sess.ChunkID)
	}
	fmt.Println()

	// Show session details
	fmt.Printf("  Started at: %s\n", sess.StartTime.Format("15:04"))
	if sess.Notes != "" {
		fmt.Printf("  Notes: %s\n", sess.Notes)
	}

	// Display chunk details if chunk was specified
	if sess.ChunkID != "" {
		info, err := getChunkDisplayInfo(cmd, sess.PlanID, sess.ChunkID)
		if err == nil {
			displayChunkDetails(info)
		}
		// Silently ignore errors - chunk display is optional
	}

	fmt.Println("\nTimer running. Stop with: samedi stop")
	return nil
}

func gatherStartInputs(cmd *cobra.Command, args []string, opts startOptions) (string, string, string, error) {
	if len(args) == 0 {
		return "", "", "", fmt.Errorf("plan ID is required")
	}

	planID := args[0]
	chunkID := ""
	if len(args) > 1 {
		chunkID = args[1]
	}

	note := ""
	if opts.note != nil {
		note = *opts.note
	}

	if !isInteractive(opts.noPrompt) {
		return planID, chunkID, note, nil
	}

	reader := bufio.NewReader(os.Stdin)
	writer := os.Stdout

	if chunkID == "" {
		selected, err := promptForChunkSelection(cmd, planID, reader, writer, chunkPromptOptions{
			showAll: opts.showAllChunk,
		})
		if err != nil {
			return "", "", "", fmt.Errorf("failed to choose chunk: %w", err)
		}
		chunkID = selected
	}

	if !opts.noteFlagSet && note == "" {
		entered, err := promptForInitialNote(reader, writer)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to read note: %w", err)
		}
		note = entered
		if opts.note != nil {
			*opts.note = note
		}
	}

	return planID, chunkID, note, nil
}

type chunkPromptOptions struct {
	showAll bool
}

func promptForChunkSelection(cmd *cobra.Command, planID string, reader *bufio.Reader, writer io.Writer, opts chunkPromptOptions) (string, error) {
	ctx := context.Background()

	planSvc, err := getPlanService(cmd, "")
	if err != nil {
		return "", fmt.Errorf("failed to load plan: %w", err)
	}

	p, err := planSvc.Get(ctx, planID)
	if err != nil {
		return "", fmt.Errorf("failed to read plan: %w", err)
	}

	if opts.showAll {
		fmt.Fprintln(writer, "\nPlan chunks:")
		printAllChunkDetails(writer, p)
	}

	if len(p.Chunks) == 0 {
		fmt.Fprintln(writer, "\nThis plan has no chunks yet. Session will not be linked to a chunk.")
		return "", nil
	}

	if id, handled, err := promptWithSuggestedChunk(p, reader, writer); handled {
		return id, err
	}

	return promptManualChunkSelection(p, reader, writer)
}

func promptForInitialNote(reader *bufio.Reader, writer io.Writer) (string, error) {
	fmt.Fprint(writer, "Initial note (optional): ")
	line, err := reader.ReadString('\n')
	if errors.Is(err, io.EOF) && line == "" {
		return "", nil
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func nextActiveChunk(p *plan.Plan) *plan.Chunk {
	for i := range p.Chunks {
		chunk := &p.Chunks[i]
		switch chunk.Status {
		case plan.StatusCompleted, plan.StatusSkipped:
			continue
		default:
			return chunk
		}
	}
	return nil
}

func joinSample(values []string, limit int) string {
	if len(values) == 0 {
		return "-"
	}
	if len(values) > limit {
		values = values[:limit]
	}
	return strings.Join(values, ", ")
}

func printAllChunkDetails(writer io.Writer, p *plan.Plan) {
	for _, chunk := range p.Chunks {
		fmt.Fprintf(writer, "\n%s — %s (%d min) [%s]\n", chunk.ID, chunk.Title, chunk.Duration, chunk.Status)
		if len(chunk.Objectives) > 0 {
			fmt.Fprintln(writer, "  Objectives:")
			for _, obj := range chunk.Objectives {
				fmt.Fprintf(writer, "    • %s\n", obj)
			}
		}
		if len(chunk.Resources) > 0 {
			fmt.Fprintln(writer, "  Resources:")
			for _, res := range chunk.Resources {
				fmt.Fprintf(writer, "    • %s\n", res)
			}
		}
		if chunk.Deliverable != "" {
			fmt.Fprintf(writer, "  Deliverable: %s\n", chunk.Deliverable)
		}
	}
}

func promptWithSuggestedChunk(p *plan.Plan, reader *bufio.Reader, writer io.Writer) (string, bool, error) {
	next := nextActiveChunk(p)
	if next == nil {
		fmt.Fprintln(writer, "\nAll chunks are marked completed. You can still choose one manually.")
		fmt.Fprintln(writer, "(Type '?' to list all chunks)")
		return "", false, nil
	}

	fmt.Fprintf(writer, "\nNext chunk suggestion: %s — %s (%d min)\n", next.ID, next.Title, next.Duration)
	if len(next.Objectives) > 0 {
		fmt.Fprintln(writer, "Objectives:")
		for i, obj := range next.Objectives {
			if i >= 3 {
				break
			}
			fmt.Fprintf(writer, "  • %s\n", obj)
		}
	}

	fmt.Fprintln(writer, "(Type '?' to list all chunks)")
	fmt.Fprint(writer, "Start this chunk? [Y/n]: ")

	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", true, err
	}

	response := strings.TrimSpace(strings.ToLower(line))
	if response == "" || response == "y" || response == "yes" {
		return next.ID, true, nil
	}

	return "", false, nil
}

func promptManualChunkSelection(p *plan.Plan, reader *bufio.Reader, writer io.Writer) (string, error) {
	chunkIndex := make(map[string]struct{}, len(p.Chunks))
	allIDs := make([]string, 0, len(p.Chunks))
	for _, chunk := range p.Chunks {
		chunkIndex[chunk.ID] = struct{}{}
		allIDs = append(allIDs, chunk.ID)
	}

	for {
		fmt.Fprint(writer, "Enter chunk ID (leave blank for no chunk): ")
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}

		id := strings.TrimSpace(line)
		switch {
		case id == "":
			return "", nil
		case id == "?" || strings.EqualFold(id, "details"):
			printAllChunkDetails(writer, p)
		default:
			if _, ok := chunkIndex[id]; ok {
				return id, nil
			}
			fmt.Fprintf(writer, "Chunk %q not found. Available IDs: %s\n", id, joinSample(allIDs, 5))
			if errors.Is(err, io.EOF) {
				return "", nil
			}
		}
	}
}
