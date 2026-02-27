package cmd

import (
	"fmt"
	"strings"

	"github.com/hbn/runes/internal/rune"
	"github.com/hbn/runes/internal/store"
	"github.com/spf13/cobra"
)

var (
	addProblem  string
	addSolution string
	addPattern  string
	addTags     []string
	addSagas    []string
	addLearned  string
	addGlobal   bool
)

var addCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a new rune",
	Long: `Create a new rune documenting a solved problem.

Required: title
Optional: problem, solution, pattern, tags, sagas, learned

Examples:
  runes add "Fixed auth timeout"
  runes add "Database connection pooling" --problem "Too many connections" --solution "Use connection pool"
  runes add "OAuth retry logic" --tags auth,oauth --sagas abc123
  runes add "Global pattern" --global  # Force global storage`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]

		st, err := store.New(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("initializing store: %w", err)
		}

		r := rune.New(title)
		r.Problem = addProblem
		r.Solution = addSolution
		r.Pattern = addPattern
		r.Learned = addLearned

		for _, tag := range addTags {
			r.AddTag(tag)
		}
		for _, sagaID := range addSagas {
			r.LinkSaga(sagaID)
		}

		// Determine scope
		var scope []store.Scope
		if addGlobal {
			scope = []store.Scope{store.ScopeGlobal}
		}

		if err := st.Save(r, scope...); err != nil {
			return fmt.Errorf("saving rune: %w", err)
		}

		fmt.Printf("Created rune %s: %s\n", r.ID, r.Title)
		if len(r.Tags) > 0 {
			fmt.Printf("Tags: [%s]\n", strings.Join(r.Tags, ", "))
		}

		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addProblem, "problem", "", "Problem description")
	addCmd.Flags().StringVar(&addSolution, "solution", "", "Solution description")
	addCmd.Flags().StringVar(&addPattern, "pattern", "", "Reusable pattern name")
	addCmd.Flags().StringArrayVar(&addTags, "tag", nil, "Add tag (can use multiple)")
	addCmd.Flags().StringArrayVar(&addSagas, "saga", nil, "Link to saga ID")
	addCmd.Flags().StringVar(&addLearned, "learned", "", "What we learned")
	addCmd.Flags().BoolVar(&addGlobal, "global", false, "Force global storage (default: local if in project)")
	rootCmd.AddCommand(addCmd)
}
