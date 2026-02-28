package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hbn/runes/internal/rune"
	"github.com/hbn/runes/internal/store"
	"github.com/spf13/cobra"
)

var quickCmd = &cobra.Command{
	Use:   "quick [title]",
	Short: "Quick interactive capture of a solution",
	Long: `Interactive mode for quickly capturing a rune.

Prompts for each field step by step. Lower friction than using flags.
Useful when you want to capture something quickly without remembering flags.

Examples:
  runes quick                    # Prompt for title
  runes quick "Auth timeout"     # Start with title, prompt for rest`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.New(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("initializing store: %w", err)
		}

		reader := bufio.NewReader(os.Stdin)

		// Get title
		var title string
		if len(args) > 0 {
			title = args[0]
			fmt.Printf("Title: %s\n", title)
		} else {
			fmt.Print("Title (brief summary): ")
			title, _ = reader.ReadString('\n')
			title = strings.TrimSpace(title)
		}

		if title == "" {
			return fmt.Errorf("title required")
		}

		// Create rune
		r := rune.New(title)

		// Prompt for problem
		fmt.Println()
		fmt.Println("What was the problem? (empty line to finish)")
		fmt.Print("> ")
		problem, _ := reader.ReadString('\n')
		r.Problem = strings.TrimSpace(problem)

		// Prompt for solution
		fmt.Println()
		fmt.Println("What was the solution? (empty line to finish)")
		fmt.Print("> ")
		solution, _ := reader.ReadString('\n')
		r.Solution = strings.TrimSpace(solution)

		// Validate required fields
		if r.Problem == "" || r.Solution == "" {
			fmt.Println()
			fmt.Println("⚠️  Skipping capture (problem and solution required)")
			fmt.Println("   Use 'runes add' with flags for full control")
			return nil
		}

		// Optional: pattern
		fmt.Println()
		fmt.Print("Pattern name (optional): ")
		pattern, _ := reader.ReadString('\n')
		r.Pattern = strings.TrimSpace(pattern)

		// Optional: tags
		fmt.Print("Tags (comma-separated, optional): ")
		tagsInput, _ := reader.ReadString('\n')
		tagsInput = strings.TrimSpace(tagsInput)
		if tagsInput != "" {
			for _, t := range strings.Split(tagsInput, ",") {
				r.AddTag(strings.TrimSpace(t))
			}
		}

		// Optional: learned insight
		fmt.Println()
		fmt.Println("Key insight for next time? (optional)")
		fmt.Print("> ")
		learned, _ := reader.ReadString('\n')
		r.Learned = strings.TrimSpace(learned)

		// Confirm
		fmt.Println()
		fmt.Println("--- Preview ---")
		fmt.Printf("Title:    %s\n", r.Title)
		fmt.Printf("Problem:  %s\n", r.Problem)
		fmt.Printf("Solution: %s\n", r.Solution)
		if r.Pattern != "" {
			fmt.Printf("Pattern:  %s\n", r.Pattern)
		}
		if len(r.Tags) > 0 {
			fmt.Printf("Tags:     %s\n", strings.Join(r.Tags, ", "))
		}
		if r.Learned != "" {
			fmt.Printf("Learned:  %s\n", r.Learned)
		}

		fmt.Println()
		fmt.Print("Save this rune? [Y/n]: ")
		confirm, _ := reader.ReadString('\n')
		confirm = strings.ToLower(strings.TrimSpace(confirm))

		if confirm == "n" || confirm == "no" {
			fmt.Println("Aborted.")
			return nil
		}

		// Save
		if err := st.Save(r); err != nil {
			return fmt.Errorf("saving rune: %w", err)
		}

		fmt.Println()
		fmt.Printf("Created rune %s: %s\n", r.ID, r.Title)
		fmt.Printf("Tags: [%s]\n", strings.Join(r.Tags, ", "))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(quickCmd)
}
