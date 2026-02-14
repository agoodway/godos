package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/goodway/godos/internal/store"
	"github.com/spf13/cobra"
)

func defaultStoreDir() string {
	if dir := os.Getenv("GODOS_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return home + "/.godos"
}

var listsCmd = &cobra.Command{
	Use:   "lists",
	Short: "Manage todo lists",
	Long:  `List, create, rename, and delete todo lists.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		s := store.New(defaultStoreDir())
		summaries, err := s.ListAll()
		if err != nil {
			return err
		}
		if len(summaries) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No lists found.")
			return nil
		}
		for _, ls := range summaries {
			fmt.Fprintf(cmd.OutOrStdout(), "%s  (%d/%d done)\n", ls.Name, ls.Completed, ls.Total)
		}
		return nil
	},
}

var listsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new empty list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := store.ValidateName(name); err != nil {
			return err
		}
		s := store.New(defaultStoreDir())
		if err := s.CreateList(name); err != nil {
			if errors.Is(err, store.ErrListExists) {
				return fmt.Errorf("list %q already exists", name)
			}
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created list %q.\n", name)
		return nil
	},
}

var listsRenameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename an existing list",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldName, newName := args[0], args[1]
		if err := store.ValidateName(newName); err != nil {
			return fmt.Errorf("invalid new name: %w", err)
		}
		s := store.New(defaultStoreDir())
		if err := s.RenameList(oldName, newName); err != nil {
			if errors.Is(err, store.ErrListNotFound) {
				return fmt.Errorf("list %q does not exist", oldName)
			}
			if errors.Is(err, store.ErrListExists) {
				return fmt.Errorf("list %q already exists", newName)
			}
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Renamed list %q to %q.\n", oldName, newName)
		return nil
	},
}

var listsDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a list and all its todos",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		s := store.New(defaultStoreDir())

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			total, _, err := s.CountTodos(name)
			if err != nil {
				if errors.Is(err, store.ErrListNotFound) {
					return fmt.Errorf("list %q does not exist", name)
				}
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Delete list %q with %d todos? [y/N] ", name, total)
			reader := bufio.NewReader(cmd.InOrStdin())
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
				return nil
			}
		}

		if err := s.DeleteList(name); err != nil {
			if errors.Is(err, store.ErrListNotFound) {
				return fmt.Errorf("list %q does not exist", name)
			}
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted list %q.\n", name)
		return nil
	},
}

func init() {
	listsDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	listsCmd.AddCommand(listsCreateCmd)
	listsCmd.AddCommand(listsRenameCmd)
	listsCmd.AddCommand(listsDeleteCmd)
	rootCmd.AddCommand(listsCmd)
}
