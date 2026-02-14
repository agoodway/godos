package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/goodway/godos/config"
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
		todoStore = store.New(resolveStoreDir())
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

// resolveStoreDir returns the storage directory using the cascade:
// GODOS_DIR env > explicit --dir > config default_dir > ~/.godos.
func resolveStoreDir() string {
	if dir := os.Getenv("GODOS_DIR"); dir != "" {
		return dir
	}
	if dirFlag != "" {
		return dirFlag
	}
	if dir, err := config.Get("default_dir"); err == nil && dir != "" {
		return dir
	}
	return defaultDir()
}

var rootCmd = &cobra.Command{
	Use:           "godos",
	Short:         "A CLI tool for managing todo lists",
	Long:          `godos is a command-line todo list manager. Create, organize, and track tasks across multiple lists.`,
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
	rootCmd.PersistentFlags().StringVar(&dirFlag, "dir", "", "storage directory for todo lists")
}
