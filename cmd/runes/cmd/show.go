package cmd

import (
	"fmt"
	"strings"

	"github.com/hbn/runes/internal/store"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show rune details",
	Long: `Display full rune with all fields.

Examples:
  runes show abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		st, err := store.New(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("initializing store: %w", err)
		}

		r, err := st.GetByID(id)
		if err != nil {
			return fmt.Errorf("rune not found: %s", id)
		}

		fmt.Println(strings.Repeat("═", 50))
		fmt.Printf("RUNE: %s\n", r.ID)
		fmt.Println(strings.Repeat("═", 50))
		fmt.Println()

		fmt.Printf("Title:    %s\n", r.Title)
		fmt.Printf("Created:  %s\n", r.CreatedAt.Format("Jan 02, 2006"))
		fmt.Println()

		if r.Problem != "" {
			fmt.Println("┌─ Problem")
			fmt.Println(r.Problem)
			fmt.Println()
		}

		if r.Solution != "" {
			fmt.Println("┌─ Solution")
			fmt.Println(r.Solution)
			fmt.Println()
		}

		if r.Pattern != "" {
			fmt.Printf("Pattern:  %s\n", r.Pattern)
			fmt.Println()
		}

		if r.Learned != "" {
			fmt.Println("┌─ Learned")
			fmt.Println(r.Learned)
			fmt.Println()
		}

		if len(r.Tags) > 0 {
			fmt.Printf("Tags:     [%s]\n", strings.Join(r.Tags, ", "))
		}

		if len(r.Sagas) > 0 {
			fmt.Printf("Sagas:    [%s]\n", strings.Join(r.Sagas, ", "))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
