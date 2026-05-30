package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <id-prefix>",
	Short: "Remove a todo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		task, err := svc.DeleteTask(context.Background(), args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Removed %s \"%s\"\n", task.ShortID, task.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
