package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "godos",
	Short: "godos - a CLI tool",
	Long:  `godos is a CLI application built with Cobra.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to godos! Use --help to see available commands.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
