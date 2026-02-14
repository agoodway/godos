package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/goodway/godos/internal/store"
	"github.com/spf13/cobra"
)

var (
	dirFlag   string
	storeOnce sync.Once
	todoStore *store.Store
)

func getStore() *store.Store {
	storeOnce.Do(func() {
		todoStore = store.New(dirFlag)
	})
	return todoStore
}

func defaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".godos"
	}
	return filepath.Join(home, ".godos")
}

var rootCmd = &cobra.Command{
	Use:           "godos",
	Short:         "A simple CLI todo manager backed by markdown files",
	Long:          `godos manages your todos as markdown checkbox lists. Each list is a .md file with - [ ] and - [x] entries.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dirFlag, "dir", defaultDir(), "storage directory for todo lists")
}
