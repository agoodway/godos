package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var doneListFlag string

var doneCmd = &cobra.Command{
	Use:   "done <number>",
	Short: "Mark a todo as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 {
			return fmt.Errorf("invalid todo number %q", args[0])
		}

		text, alreadyDone, err := todoStore.Complete(doneListFlag, n)
		if err != nil {
			return err
		}
		if alreadyDone {
			fmt.Printf("Todo #%d \"%s\" is already done\n", n, text)
			return nil
		}
		fmt.Printf("Completed #%d \"%s\"\n", n, text)
		return nil
	},
}

func init() {
	doneCmd.Flags().StringVar(&doneListFlag, "list", "todo", "list name")
	rootCmd.AddCommand(doneCmd)
}
