package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var undoneCmd = &cobra.Command{
	Use:   "undone <id-prefix>",
	Short: "Mark a todo as incomplete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		task, err := svc.ReopenTask(context.Background(), args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Reopened %s \"%s\"\n", task.ShortID, task.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(undoneCmd)
}
