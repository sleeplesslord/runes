package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/hbn/runes/internal/store"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize local runes storage",
	Long:  `Creates a .runes directory in the current project for local rune storage.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.New(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("initializing store: %w", err)
		}

		if err := st.InitLocal(); err != nil {
			return fmt.Errorf("initializing local runes: %w", err)
		}

		fmt.Printf("Initialized local runes storage in %s\n", filepath.Dir(st.LocalPath()))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
