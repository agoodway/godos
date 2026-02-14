package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var addListFlag string

var addCmd = &cobra.Command{
	Use:   "add <text>",
	Short: "Add a new todo",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")
		if err := todoStore.Add(addListFlag, text); err != nil {
			return err
		}
		fmt.Printf("Added \"%s\" to %s\n", text, addListFlag)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addListFlag, "list", "todo", "target list name")
	rootCmd.AddCommand(addCmd)
}
