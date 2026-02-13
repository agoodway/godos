package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	listListFlag string
	listAllFlag  bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List todos",
	Run: func(cmd *cobra.Command, args []string) {
		if listAllFlag {
			listAll()
			return
		}
		listOne(listListFlag)
	},
}

func listOne(name string) {
	todos, err := Store.ListTodos(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	if len(todos) == 0 {
		fmt.Printf("No todos in %s\n", name)
		return
	}
	for i, t := range todos {
		status := "[ ]"
		if t.Done {
			status = "[x]"
		}
		fmt.Printf("%d. %s %s\n", i+1, status, t.Text)
	}
}

func listAll() {
	names, err := Store.Lists()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	if len(names) == 0 {
		fmt.Println("No lists found")
		return
	}
	for i, name := range names {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("=== %s ===\n", name)
		listOne(name)
	}
}

func init() {
	listCmd.Flags().StringVar(&listListFlag, "list", "todo", "list name to display")
	listCmd.Flags().BoolVar(&listAllFlag, "all", false, "show all lists")
	rootCmd.AddCommand(listCmd)
}
