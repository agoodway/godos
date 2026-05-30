package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	listNameFlag string
	listAllFlag  bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List todos",
	RunE: func(cmd *cobra.Command, args []string) error {
		if listAllFlag {
			return listAll(cmd)
		}
		return listOne(cmd, listNameFlag)
	},
}

func listOne(cmd *cobra.Command, name string) error {
	svc, err := getAPIService(true)
	if err != nil {
		return err
	}
	todos, err := svc.ListTasks(context.Background(), name)
	if err != nil {
		return err
	}
	if len(todos) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No todos in %s\n", name)
		return nil
	}
	for _, t := range todos {
		status := "[ ]"
		if t.Done {
			status = "[x]"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s. %s %s\n", t.ShortID, status, t.Title)
	}
	return nil
}

func listAll(cmd *cobra.Command) error {
	svc, err := getAPIService(true)
	if err != nil {
		return err
	}
	lists, err := svc.ListAllTasks(context.Background())
	if err != nil {
		return err
	}
	if len(lists) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lists found")
		return nil
	}
	for i, list := range lists {
		if i > 0 {
			fmt.Fprintln(cmd.OutOrStdout())
		}
		fmt.Fprintf(cmd.OutOrStdout(), "=== %s ===\n", list.Name)
		if len(list.Tasks) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No todos in %s\n", list.Name)
			continue
		}
		for _, t := range list.Tasks {
			status := "[ ]"
			if t.Done {
				status = "[x]"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s. %s %s\n", t.ShortID, status, t.Title)
		}
	}
	return nil
}

func init() {
	listCmd.Flags().StringVar(&listNameFlag, "list", "todo", "list name to display")
	listCmd.Flags().BoolVar(&listAllFlag, "all", false, "show all lists")
	rootCmd.AddCommand(listCmd)
}
