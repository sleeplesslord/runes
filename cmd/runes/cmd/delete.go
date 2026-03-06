package cmd

import (
	"fmt"
	"strings"

	"github.com/sleeplesslord/runes/internal/store"
	"github.com/spf13/cobra"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a rune",
	Long: `Permanently delete a rune from the store.

By default, asks for confirmation. Use --force to skip confirmation.

Examples:
  runes delete abc123              # Confirm before delete
  runes delete abc123 --force      # Delete without confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		st, err := store.New(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("initializing store: %w", err)
		}

		// Load rune to show what will be deleted
		r, err := st.GetByID(id)
		if err != nil {
			return fmt.Errorf("rune not found: %s", id)
		}

		// Show what will be deleted
		fmt.Printf("Will delete rune:\n")
		fmt.Printf("  ID:    %s\n", r.ID)
		fmt.Printf("  Title: %s\n", r.Title)
		if r.Pattern != "" {
			fmt.Printf("  Pattern: %s\n", r.Pattern)
		}

		// Confirm unless --force
		if !deleteForce {
			fmt.Printf("\nAre you sure? [y/N]: ")
			var response string
			if _, err := fmt.Scanln(&response); err != nil {
				return fmt.Errorf("aborted")
			}
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		// Delete
		if err := st.Delete(id); err != nil {
			return fmt.Errorf("deleting rune: %w", err)
		}

		fmt.Printf("Deleted rune %s\n", id)
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Delete without confirmation")
	rootCmd.AddCommand(deleteCmd)
}
