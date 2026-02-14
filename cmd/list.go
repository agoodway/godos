package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	listListFlag string
	listAllFlag  bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List todos",
	RunE: func(cmd *cobra.Command, args []string) error {
		if listAllFlag {
			return listAll()
		}
		return listOne(listListFlag)
	},
}

func listOne(name string) error {
	todos, err := todoStore.ListTodos(name)
	if err != nil {
		return err
	}
	if len(todos) == 0 {
		fmt.Printf("No todos in %s\n", name)
		return nil
	}
	for i, t := range todos {
		status := "[ ]"
		if t.Done {
			status = "[x]"
		}
		fmt.Printf("%d. %s %s\n", i+1, status, t.Text)
	}
	return nil
}

func listAll() error {
	names, err := todoStore.Lists()
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Println("No lists found")
		return nil
	}
	for i, name := range names {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("=== %s ===\n", name)
		if err := listOne(name); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	listCmd.Flags().StringVar(&listListFlag, "list", "todo", "list name to display")
	listCmd.Flags().BoolVar(&listAllFlag, "all", false, "show all lists")
	rootCmd.AddCommand(listCmd)
}
