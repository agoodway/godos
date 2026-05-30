package cmd

import (
	"context"
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
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		task, err := svc.AddTask(context.Background(), addListFlag, text)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Added %s \"%s\" to %s\n", task.ShortID, task.Title, addListFlag)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addListFlag, "list", "todo", "target list name")
	rootCmd.AddCommand(addCmd)
}
