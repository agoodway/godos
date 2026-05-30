package cmd

import (
	"errors"
	"fmt"
	"sort"

	"github.com/goodway/godos/config"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Manage godos configuration",
	Long:  `Read and write persistent configuration values for godos.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
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
		fmt.Fprintf(cmd.OutOrStdout(), "Set %s = %s\n", key, displayConfigValue(key, value))
		return nil
	},
}

var configureGetShowSecret bool

var configureGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if key == config.APITokenKey && !configureGetShowSecret {
			return fmt.Errorf("refusing to print api_token; rerun with --show-secret to reveal it")
		}
		value, err := config.Get(key)
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
		if errors.Is(err, config.ErrNotFound) {
			fmt.Fprintln(cmd.OutOrStdout(), "No configuration is set.")
			return nil
		}
		if err != nil {
			return err
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
			value := m[k]
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", k, displayConfigValue(k, value))
		}
		return nil
	},
}

var configureDeleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Remove a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if err := config.Delete(key); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted %s\n", key)
		return nil
	},
}

func init() {
	configureGetCmd.Flags().BoolVar(&configureGetShowSecret, "show-secret", false, "print secret values without redaction")
	configureCmd.AddCommand(configureSetCmd)
	configureCmd.AddCommand(configureGetCmd)
	configureCmd.AddCommand(configureListCmd)
	configureCmd.AddCommand(configureDeleteCmd)
	rootCmd.AddCommand(configureCmd)
}

func displayConfigValue(key, value string) string {
	if key == config.APITokenKey && value != "" {
		return "****"
	}
	return value
}
