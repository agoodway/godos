package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var doneListFlag string

var doneCmd = &cobra.Command{
	Use:   "done <number>",
	Short: "Mark a todo as complete",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 {
			fmt.Fprintf(os.Stderr, "Error: invalid todo number %q\n", args[0])
			os.Exit(1)
		}

		text, alreadyDone, err := Store.Complete(doneListFlag, n)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		if alreadyDone {
			fmt.Printf("Todo #%d \"%s\" is already done\n", n, text)
			return
		}
		fmt.Printf("Completed #%d \"%s\"\n", n, text)
	},
}

func init() {
	doneCmd.Flags().StringVar(&doneListFlag, "list", "todo", "list name")
	rootCmd.AddCommand(doneCmd)
}
