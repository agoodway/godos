package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var rmListFlag string

var rmCmd = &cobra.Command{
	Use:   "rm <number>",
	Short: "Remove a todo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 {
			return fmt.Errorf("invalid todo number %q", args[0])
		}

		text, err := getStore().Remove(rmListFlag, n)
		if err != nil {
			return err
		}
		fmt.Printf("Removed #%d \"%s\"\n", n, text)
		return nil
	},
}

func init() {
	rmCmd.Flags().StringVar(&rmListFlag, "list", "todo", "list name")
	rootCmd.AddCommand(rmCmd)
}
