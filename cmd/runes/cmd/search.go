package cmd

import (
	"fmt"
	"strings"

	"github.com/hbn/runes/internal/rune"
	"github.com/hbn/runes/internal/store"
	"github.com/spf13/cobra"
)

var searchLimit int
var searchSaga string

var searchCmd = &cobra.Command{
	Use:   "search [<query>...]",
	Short: "Search runes",
	Long: `Find runes matching query in title, problem, solution, tags, or pattern.

Multiple queries can be provided to search for different terms at once.
Each query produces separate results.

Use --saga to filter by linked saga ID (can be used alone or with queries).

Examples:
  runes search "auth timeout"
  runes search "database" --limit 5
  runes search "auth" "database" "timeout"
  runes search --saga abc123
  runes search "timeout" --saga abc123`,
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

		// Filter by saga if specified
		if searchSaga != "" {
			var filtered []*rune.Rune
			for _, r := range runes {
				if r.HasSaga(searchSaga) {
					filtered = append(filtered, r)
				}
			}
			runes = filtered
		}

		// If no queries provided, show all (optionally saga-filtered) runes
		if len(args) == 0 {
			if searchSaga == "" {
				return fmt.Errorf("search requires at least one query or --saga flag")
			}
			// Display saga-filtered results
			if len(runes) == 0 {
				fmt.Println("No runes found.")
				return nil
			}
			fmt.Printf("Saga: %s\n", searchSaga)
			fmt.Printf("Found %d rune(s):\n\n", len(runes))
			for _, r := range runes {
				displayRune(r)
			}
			return nil
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

			fmt.Printf("Query: %q", query)
			if searchSaga != "" {
				fmt.Printf(" (saga: %s)", searchSaga)
			}
			fmt.Println()

			if len(results) == 0 {
				fmt.Println("No runes found.")
				continue
			}

			fmt.Printf("Found %d rune(s):\n\n", len(results))

			for _, r := range results {
				displayRune(r)
			}
		}

		return nil
	},
}

func displayRune(r *rune.Rune) {
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
	if len(r.Sagas) > 0 {
		fmt.Printf("         Sagas: [%s]\n", strings.Join(r.Sagas, ", "))
	}
	fmt.Println()
}

func init() {
	searchCmd.Flags().IntVar(&searchLimit, "limit", 10, "Maximum results")
	searchCmd.Flags().StringVar(&searchSaga, "saga", "", "Filter by saga ID")
	rootCmd.AddCommand(searchCmd)
}
