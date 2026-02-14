package cmd

import (
	"fmt"
	"sort"

	"github.com/goodway/godos/config"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Manage godos configuration",
	Long:  `Read and write persistent configuration values for godos.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var configureSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		if err := config.Set(key, value); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Set %s = %s\n", key, value)
		return nil
	},
}

var configureGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value, err := config.Get(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), value)
		return nil
	},
}

var configureListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := config.Load()
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "No configuration is set.")
			return nil
		}
		if len(m) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No configuration is set.")
			return nil
		}

		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", k, m[k])
		}
		return nil
	},
}

func init() {
	configureCmd.AddCommand(configureSetCmd)
	configureCmd.AddCommand(configureGetCmd)
	configureCmd.AddCommand(configureListCmd)
	rootCmd.AddCommand(configureCmd)
}
