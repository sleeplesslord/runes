package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "runes",
	Short: "Runes - Knowledge management for solutions",
	Long: `Runes captures and discovers solutions to problems.

Each rune documents a solved problem with structure:
- What was the problem?
- What was the solution?
- What pattern emerged?
- What did we learn?

Use runes to avoid solving the same problem twice.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
