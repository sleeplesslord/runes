package cmd

import (
	"fmt"
	"strings"

	"github.com/hbn/runes/internal/store"
	"github.com/spf13/cobra"
)

var searchLimit int

var searchCmd = &cobra.Command{
	Use:   "search <query>...",
	Short: "Search runes",
	Long: `Find runes matching query in title, problem, solution, tags, or pattern.

Multiple queries can be provided to search for different terms at once.
Each query produces separate results.

Examples:
  runes search "auth timeout"
  runes search "database" --limit 5
  runes search "auth" "database" "timeout"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.New(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("initializing store: %w", err)
		}

		// Load all runes once for all queries
		runes, err := st.LoadAll()
		if err != nil {
			return fmt.Errorf("loading runes: %w", err)
		}

		// Search each query
		for i, query := range args {
			// Add separator between queries (but not before first)
			if i > 0 {
				fmt.Println(strings.Repeat("-", 40))
			}

			results, err := store.SearchRunes(runes, query, searchLimit)
			if err != nil {
				return fmt.Errorf("searching for %q: %w", query, err)
			}

			fmt.Printf("Query: %q\n", query)

			if len(results) == 0 {
				fmt.Println("No runes found.")
				continue
			}

			fmt.Printf("Found %d rune(s):\n\n", len(results))

			for _, r := range results {
				fmt.Printf("  %-6s %s\n", r.ID, r.Title)
				if r.Problem != "" {
					problem := r.Problem
					if len(problem) > 50 {
						problem = problem[:47] + "..."
					}
					fmt.Printf("         Problem: %s\n", problem)
				}
				if len(r.Tags) > 0 {
					fmt.Printf("         Tags: [%s]\n", strings.Join(r.Tags, ", "))
				}
				fmt.Println()
			}
		}

		return nil
	},
}

func init() {
	searchCmd.Flags().IntVar(&searchLimit, "limit", 10, "Maximum results")
	rootCmd.AddCommand(searchCmd)
}
