package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hbn/runes/internal/rune"
	"github.com/hbn/runes/internal/store"
	"github.com/spf13/cobra"
)

var (
	listTag   string
	listLimit int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all runes",
	Long: `Display all runes with optional filtering.

Examples:
  runes list                    # All runes
  runes list --tag auth         # Filter by tag
  runes list --limit 5          # Limit results`,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.New(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("initializing store: %w", err)
		}

		runes, err := st.LoadAll()
		if err != nil {
			return fmt.Errorf("loading runes: %w", err)
		}

		// Filter by tag
		if listTag != "" {
			var filtered []*rune.Rune
			for _, r := range runes {
				if r.HasTag(listTag) {
					filtered = append(filtered, r)
				}
			}
			runes = filtered
		}

		if len(runes) == 0 {
			fmt.Println("No runes found.")
			return nil
		}

		// Sort by created date (newest first)
		sort.Slice(runes, func(i, j int) bool {
			return runes[i].CreatedAt.After(runes[j].CreatedAt)
		})

		// Apply limit
		if listLimit > 0 && listLimit < len(runes) {
			runes = runes[:listLimit]
		}

		fmt.Printf("Found %d rune(s):\n\n", len(runes))
		fmt.Printf("%-6s %-20s %s\n", "ID", "Title", "Tags")
		fmt.Println(strings.Repeat("-", 50))

		for _, r := range runes {
			title := r.Title
			if len(title) > 20 {
				title = title[:17] + "..."
			}

			tags := ""
			if len(r.Tags) > 0 {
				tags = fmt.Sprintf("[%s]", strings.Join(r.Tags, ", "))
			}

			fmt.Printf("%-6s %-20s %s\n", r.ID, title, tags)
		}

		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listTag, "tag", "", "Filter by tag")
	listCmd.Flags().IntVar(&listLimit, "limit", 0, "Limit results (0 = no limit)")
	rootCmd.AddCommand(listCmd)
}
