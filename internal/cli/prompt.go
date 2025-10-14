package cli

import (
	"os"

	"golang.org/x/term"
)

// isInteractive returns true when prompts should be shown.
func isInteractive(noPrompt bool) bool {
	if noPrompt {
		return false
	}
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}
