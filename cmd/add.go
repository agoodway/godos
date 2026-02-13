package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var addListFlag string

var addCmd = &cobra.Command{
	Use:   "add <text>",
	Short: "Add a new todo",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := strings.Join(args, " ")
		if err := Store.Add(addListFlag, text); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added \"%s\" to %s\n", text, addListFlag)
	},
}

func init() {
	addCmd.Flags().StringVar(&addListFlag, "list", "todo", "target list name")
	rootCmd.AddCommand(addCmd)
}
