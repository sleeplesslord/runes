package cmd

import (
	"fmt"
	"sort"

	"github.com/sleeplesslord/runes/internal/store"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all tags with rune counts",
	Long: `Show overview of all tags in the rune collection.

Use this to discover what areas are covered before searching.
If you see a relevant tag, search within it:
  runes list --tag <tag>

Examples:
  runes tags              # Show all tags sorted by count
  runes tags --alpha      # Sort alphabetically`,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.New(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("initializing store: %w", err)
		}

		runes, err := st.LoadAll()
		if err != nil {
			return fmt.Errorf("loading runes: %w", err)
		}

		if len(runes) == 0 {
			fmt.Println("No runes found.")
			fmt.Println("\nCreate your first rune:")
			fmt.Println("  runes add \"Title\" --problem \"...\" --solution \"...\"")
			return nil
		}

		// Count tags
		tagCounts := make(map[string]int)
		for _, r := range runes {
			for _, tag := range r.Tags {
				tagCounts[tag]++
			}
		}

		if len(tagCounts) == 0 {
			fmt.Println("No tags found in runes.")
			fmt.Println("\nAdd tags when creating runes:")
			fmt.Println("  runes add \"Title\" --tag auth --tag api ...")
			return nil
		}

		// Convert to slice for sorting
		type tagCount struct {
			Tag   string
			Count int
		}
		var tags []tagCount
		for tag, count := range tagCounts {
			tags = append(tags, tagCount{Tag: tag, Count: count})
		}

		// Sort
		alpha, _ := cmd.Flags().GetBool("alpha")
		if alpha {
			sort.Slice(tags, func(i, j int) bool {
				return tags[i].Tag < tags[j].Tag
			})
		} else {
			// Sort by count desc, then alpha
			sort.Slice(tags, func(i, j int) bool {
				if tags[i].Count != tags[j].Count {
					return tags[i].Count > tags[j].Count
				}
				return tags[i].Tag < tags[j].Tag
			})
		}

		// Display
		fmt.Printf("Found %d tag(s) across %d rune(s):\n\n", len(tags), len(runes))
		fmt.Println("Tag                  Count")
		fmt.Println("-------------------- -----")
		for _, tc := range tags {
			fmt.Printf("%-20s %5d\n", tc.Tag, tc.Count)
		}

		fmt.Println("\nExplore a tag:")
		fmt.Printf("  runes list --tag %s\n", tags[0].Tag)
		fmt.Println("\nOr search across all runes:")
		fmt.Println("  runes search \"<keywords>\"")

		return nil
	},
}

func init() {
	tagsCmd.Flags().Bool("alpha", false, "Sort alphabetically instead of by count")
	rootCmd.AddCommand(tagsCmd)
}
