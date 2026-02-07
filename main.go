package main

import (
	"fmt"
	"os"

	"github.com/arimxyer/pass-cli/cmd"
	"github.com/arimxyer/pass-cli/cmd/tui"
)

func main() {
	// Default to TUI if no subcommand provided
	shouldUseTUI := true
	vaultPath := ""

	// Parse args to detect subcommands or flags
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		// Check for help or version flags
		if arg == "--help" || arg == "-h" || arg == "--version" || arg == "-v" {
			shouldUseTUI = false
			break
		}

		// Check for subcommands (any non-flag argument)
		if arg != "" && arg[0] != '-' {
			shouldUseTUI = false
			break
		}

		// Extract vault path if provided
		if arg == "--vault" && i+1 < len(os.Args) {
			vaultPath = os.Args[i+1]
			i++ // Skip next arg (vault path value)
		}
	}

	// Route to TUI or CLI
	if shouldUseTUI {
		if err := tui.Run(vaultPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		cmd.Execute()
	}
}
