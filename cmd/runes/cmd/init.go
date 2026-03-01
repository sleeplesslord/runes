package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hbn/runes/internal/store"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize local runes storage",
	Long:  `Creates a .runes directory in the current project for local rune storage.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.New(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("initializing store: %w", err)
		}

		if err := st.InitLocal(); err != nil {
			return fmt.Errorf("initializing local runes: %w", err)
		}

		fmt.Printf("Initialized local runes storage in %s\n", filepath.Dir(st.LocalPath()))
		
		// Add runes section to AGENTS.md if it exists
		if err := addRunesToAgents(); err != nil {
			// Non-fatal: just print the help text
			fmt.Println()
			fmt.Println("Runes initialized. Basic usage:")
			fmt.Println("  runes search \"...\"   # Find existing solutions")
			fmt.Println("  runes add \"...\"      # Capture new solution")
			fmt.Println("  runes list            # Browse all runes")
		}
		
		return nil
	},
}

func addRunesToAgents() error {
	agentsPath := "AGENTS.md"
	
	// Check if file exists
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		return err
	}
	
	// Read existing content
	content, err := os.ReadFile(agentsPath)
	if err != nil {
		return err
	}
	
	// Check if runes section already exists
	if strings.Contains(string(content), "## Runes") {
		return nil // Already exists, skip
	}
	
	// Append runes section
	runesSection := "\n## Runes - Knowledge Capture\n\n" +
		"This project uses [Runes](https://github.com/sleeplesslord/runes) for knowledge management.\n\n" +
		"### Quick Start\n\n" +
		"```bash\n" +
		"runes search \"...\"    # Check if solution exists before solving\n" +
		"runes add \"...\"       # Capture solution after solving\n" +
		"runes list            # Browse all captured knowledge\n" +
		"```\n\n" +
		"### Basic Commands\n\n" +
		"| Command | Description |\n" +
		"|---------|-------------|\n" +
		"| `runes search \"query\"` | Find prior solutions |\n" +
		"| `runes add \"title\"` | Create new knowledge entry |\n" +
		"| `runes list` | Show all runes |\n" +
		"| `runes show <id>` | Read full rune details |\n" +
		"| `runes tags` | List all tags |\n" +
		"| `runes pattern` | Browse named patterns |\n\n" +
		"### Creating Discoverable Runes\n\n" +
		"1. **Use common keywords** in titles (\"auth timeout\" not \"incident-2024\")\n" +
		"2. **Include synonyms** in problem descriptions\n" +
		"3. **Tag generously** — more tags = more ways to find\n" +
		"4. **Write \"learned\"** — capture the insight, not just the fix\n\n" +
		"See `skills/runes-agent/SKILL.md` for full agent documentation.\n"
	
	// Append to file
	f, err := os.OpenFile(agentsPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	
	if _, err := f.WriteString(runesSection); err != nil {
		return err
	}
	
	fmt.Println("\n✓ Added Runes section to AGENTS.md")
	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
