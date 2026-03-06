package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sleeplesslord/runes/internal/rune"
	"github.com/sleeplesslord/runes/internal/store"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit an existing rune",
	Long: `Edit a rune using your default editor.

Opens the rune in YAML format for editing. Save and exit to update.
Supports --title, --problem, --solution, --pattern, --tag, --saga, --learned flags
as an alternative to interactive editing.

Examples:
  runes edit abc123                    # Open in editor
  runes edit abc123 --title "New title" # Update single field`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		st, err := store.New(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("initializing store: %w", err)
		}

		// Load existing rune
		r, err := st.GetByID(id)
		if err != nil {
			return fmt.Errorf("rune not found: %s", id)
		}

		// Check if any flags were provided
		title, _ := cmd.Flags().GetString("title")
		problem, _ := cmd.Flags().GetString("problem")
		solution, _ := cmd.Flags().GetString("solution")
		pattern, _ := cmd.Flags().GetString("pattern")
		tags, _ := cmd.Flags().GetStringArray("tag")
		sagas, _ := cmd.Flags().GetStringArray("saga")
		learned, _ := cmd.Flags().GetString("learned")

		// If any flags provided, update directly
		hasFlags := title != "" || problem != "" || solution != "" ||
			pattern != "" || len(tags) > 0 || len(sagas) > 0 || learned != ""

		if hasFlags {
			if title != "" {
				r.Title = title
			}
			if problem != "" {
				r.Problem = problem
			}
			if solution != "" {
				r.Solution = solution
			}
			if pattern != "" {
				r.Pattern = pattern
			}
			if len(tags) > 0 {
				r.Tags = tags
			}
			if len(sagas) > 0 {
				r.Sagas = sagas
			}
			if learned != "" {
				r.Learned = learned
			}

			if err := st.Update(r); err != nil {
				return fmt.Errorf("updating rune: %w", err)
			}

			fmt.Printf("Updated rune %s\n", r.ID)
			return nil
		}

		// Otherwise, open in editor
		return editInteractive(r, st)
	},
}

func editInteractive(r *rune.Rune, st *store.Store) error {
	// Create temp file with YAML content
	tmpFile, err := os.CreateTemp("", "rune-*.yaml")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write rune as YAML
	content := fmt.Sprintf(`# Edit this rune and save to update
# Lines starting with # are comments and will be ignored

id: %s
created: %s
updated: %s

title: %s

problem: |
%s

solution: |
%s

pattern: %s

tags:
%s

sagas:
%s

learned: |
%s
`, r.ID, r.CreatedAt.Format("2006-01-02"), r.UpdatedAt.Format("2006-01-02"),
		r.Title,
		indent(r.Problem, 2),
		indent(r.Solution, 2),
		r.Pattern,
		formatList(r.Tags, 2),
		formatList(r.Sagas, 2),
		indent(r.Learned, 2))

	if _, err := tmpFile.WriteString(content); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	tmpFile.Close()

	// Open in editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	// Read back the edited content
	_, err = os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("reading edited file: %w", err)
	}

	// Parse and update (simplified - just check for changes)
	// For now, we just support flag-based editing
	// Full YAML parsing would require adding a dependency

	fmt.Println("Note: Interactive editing not fully implemented yet.")
	fmt.Println("Use flags instead: runes edit <id> --title \"new title\"")

	return nil
}

func indent(s string, spaces int) string {
	if s == "" {
		return "  "
	}
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func formatList(items []string, spaces int) string {
	if len(items) == 0 {
		return "  []"
	}
	prefix := strings.Repeat(" ", spaces)
	var lines []string
	for _, item := range items {
		lines = append(lines, prefix+"- "+item)
	}
	return strings.Join(lines, "\n")
}

func init() {
	editCmd.Flags().String("title", "", "Update title")
	editCmd.Flags().String("problem", "", "Update problem description")
	editCmd.Flags().String("solution", "", "Update solution")
	editCmd.Flags().String("pattern", "", "Update pattern name")
	editCmd.Flags().StringArray("tag", nil, "Replace tags (use multiple times)")
	editCmd.Flags().StringArray("saga", nil, "Replace saga links (use multiple times)")
	editCmd.Flags().String("learned", "", "Update learned insight")
	rootCmd.AddCommand(editCmd)
}
