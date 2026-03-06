package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sleeplesslord/runes/internal/store"
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
		"### What Are Runes?\n\n" +
		"Runes are atomic, reusable solutions to problems you've solved. Think of them as a personal Stack Overflow " +
		"that lives in your workspace — a searchable database of fixes, configurations, workarounds, and insights.\n\n" +
		"### The Core Idea\n\n" +
		"Every time you solve a problem, you have a choice: let that knowledge evaporate, or encode it as a rune. " +
		"Runes capture not just *what* you did, but *why* and *when* it matters. Over time, this creates compound " +
		"returns — each new problem builds on prior solutions instead of starting from zero.\n\n" +
		"### When to Capture a Rune\n\n" +
		"**Capture after solving:**\n" +
		"- You spent >10 minutes figuring something out\n" +
		"- You fixed an error you might hit again\n" +
		"- You configured something fiddly (tools, APIs, environments)\n" +
		"- You found a non-obvious solution to a common problem\n" +
		"- You want to remember \"how did I do that last time?\"\n\n" +
		"**Don't capture:**\n" +
		"- Routine tasks (already in saga)\n" +
		"- Unfinished investigations\n" +
		"- Hypothetical solutions\n\n" +
		"### The Workflow\n\n" +
		"```bash\n" +
		"# 1. Before solving, check if you already know\n" +
		"runes search \"docker compose networking\"\n\n" +
		"# 2. Solve the problem...\n\n" +
		"# 3. After solving, capture it\n" +
		"runes add \"Fixed Docker Compose DNS\" --problem \"...\" --solution \"...\"\n\n" +
		"# 4. Later, when the same problem appears\n" +
		"runes search \"docker dns\"\n" +
		"# → Returns your solution instantly\n" +
		"```\n\n" +
		"### Quick Reference\n\n" +
		"```bash\n" +
		"runes search \"...\"    # Check if solution exists (FIRST)\n" +
		"runes add \"...\"       # Capture solution after solving (SECOND)\n" +
		"runes list              # Browse all runes\n" +
		"runes show <id>         # View full rune details\n" +
		"runes tags              # List all tags\n" +
		"runes pattern           # Browse named patterns\n" +
		"```\n\n" +
		"### Creating Discoverable Runes\n\n" +
		"1. **Use common keywords** in titles (\"auth timeout\" not \"incident-2024\")\n" +
		"2. **Include synonyms** in problem descriptions\n" +
		"3. **Tag generously** — more tags = more ways to find\n" +
		"4. **Write \"learned\"** — capture the insight, not just the fix\n\n" +
		"### Integration with Saga\n\n" +
		"Runes complement sagas:\n" +
		"- **Sagas** track *work in progress* (tasks, projects, things to do)\n" +
		"- **Runes** capture *knowledge gained* (solutions, fixes, things learned)\n" +
		"When you finish a saga, you often have new runes to add. When you start a saga, search runes to avoid repeating past work.\n\n" +
		"See `skills/SKILL.md` for full agent documentation.\n"

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
