package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var rmListFlag string

var rmCmd = &cobra.Command{
	Use:   "rm <number>",
	Short: "Remove a todo",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 {
			fmt.Fprintf(os.Stderr, "Error: invalid todo number %q\n", args[0])
			os.Exit(1)
		}

		text, err := Store.Remove(rmListFlag, n)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Removed #%d \"%s\"\n", n, text)
	},
}

func init() {
	rmCmd.Flags().StringVar(&rmListFlag, "list", "todo", "list name")
	rootCmd.AddCommand(rmCmd)
}
