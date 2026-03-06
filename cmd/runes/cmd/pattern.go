package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sleeplesslord/runes/internal/store"
	"github.com/spf13/cobra"
)

var patternCmd = &cobra.Command{
	Use:   "pattern",
	Short: "List and search named patterns",
	Long: `Browse named patterns in the rune collection.

Patterns are reusable solution approaches that can have multiple implementations.
Use this to discover what solution patterns exist before solving.

Examples:
  runes pattern              # List all patterns
  runes pattern auth         # Search for auth-related patterns`,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.New(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("initializing store: %w", err)
		}

		runes, err := st.LoadAll()
		if err != nil {
			return fmt.Errorf("loading runes: %w", err)
		}

		// Collect patterns
		type patternInfo struct {
			name   string
			count  int
			runes  []string
			titles []string
		}
		patterns := make(map[string]*patternInfo)

		for _, r := range runes {
			if r.Pattern == "" {
				continue
			}
			p, exists := patterns[r.Pattern]
			if !exists {
				p = &patternInfo{name: r.Pattern}
				patterns[r.Pattern] = p
			}
			p.count++
			p.runes = append(p.runes, r.ID)
			p.titles = append(p.titles, r.Title)
		}

		if len(patterns) == 0 {
			fmt.Println("No patterns found.")
			fmt.Println("\nCreate patterns when adding runes:")
			fmt.Println("  runes add \"Title\" --pattern \"pattern-name\" ...")
			return nil
		}

		// Filter if search term provided
		var filtered []*patternInfo
		if len(args) > 0 {
			query := strings.ToLower(args[0])
			for _, p := range patterns {
				if strings.Contains(strings.ToLower(p.name), query) {
					filtered = append(filtered, p)
				}
			}
		} else {
			for _, p := range patterns {
				filtered = append(filtered, p)
			}
		}

		// Sort by count desc, then name
		sort.Slice(filtered, func(i, j int) bool {
			if filtered[i].count != filtered[j].count {
				return filtered[i].count > filtered[j].count
			}
			return filtered[i].name < filtered[j].name
		})

		// Display
		if len(args) > 0 {
			fmt.Printf("Found %d pattern(s) matching '%s':\n\n", len(filtered), args[0])
		} else {
			fmt.Printf("Found %d pattern(s):\n\n", len(filtered))
		}

		fmt.Println("Pattern               Count  Examples")
		fmt.Println("--------------------  -----  --------")
		for _, p := range filtered {
			examples := p.titles[0]
			if len(examples) > 30 {
				examples = examples[:27] + "..."
			}
			fmt.Printf("%-20s  %5d  %s\n", p.name, p.count, examples)
		}

		fmt.Println("\nExplore a pattern:")
		if len(filtered) > 0 {
			fmt.Printf("  runes search \"%s\"\n", filtered[0].name)
		}
		fmt.Println("  runes list --tag <pattern-tag>")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(patternCmd)
}
