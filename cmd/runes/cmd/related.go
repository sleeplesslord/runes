package cmd

import (
	"fmt"
	"sort"

	"github.com/sleeplesslord/runes/internal/rune"
	"github.com/sleeplesslord/runes/internal/store"
	"github.com/spf13/cobra"
)

var relatedCmd = &cobra.Command{
	Use:   "related <id>",
	Short: "Show runes related by content similarity",
	Long: `Find runes with similar content to the given rune.

Uses content-based similarity to discover related knowledge.
Helpful for finding alternative solutions or related patterns.

Examples:
  runes related xr5h           # Show runes similar to xr5h
  runes related xr5h --limit 10  # Show top 10 matches`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		limit, _ := cmd.Flags().GetInt("limit")

		st, err := store.New(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("initializing store: %w", err)
		}

		// Load the source rune
		source, err := st.GetByID(id)
		if err != nil {
			return fmt.Errorf("rune not found: %s", id)
		}

		// Load all runes
		runes, err := st.LoadAll()
		if err != nil {
			return fmt.Errorf("loading runes: %w", err)
		}

		// Calculate similarity scores
		type scoredRune struct {
			rune  *rune.Rune
			score float64
		}
		var scored []scoredRune

		for _, r := range runes {
			if r.ID == source.ID {
				continue // Skip self
			}
			score := source.SimilarityScore(r)
			if score > 0.0 {
				scored = append(scored, scoredRune{r, score})
			}
		}

		if len(scored) == 0 {
			fmt.Println("No related runes found.")
			fmt.Println("\nAdd more runes to build connections.")
			return nil
		}

		// Sort by score descending
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].score > scored[j].score
		})

		// Apply limit
		if limit > 0 && limit < len(scored) {
			scored = scored[:limit]
		}

		// Display
		fmt.Printf("Found %d related rune(s) for %s:\n\n", len(scored), source.ID)
		fmt.Println("ID     Similarity  Title")
		fmt.Println("---    ----------  -----")

		for _, s := range scored {
			simPct := int(s.score * 100)
			fmt.Printf("%-6s %3d%%        %s\n", s.rune.ID, simPct, s.rune.Title)
		}

		fmt.Println("\nExplore a related rune:")
		fmt.Printf("  runes show %s\n", scored[0].rune.ID)

		return nil
	},
}

func init() {
	relatedCmd.Flags().Int("limit", 5, "Maximum number of results")
	rootCmd.AddCommand(relatedCmd)
}
