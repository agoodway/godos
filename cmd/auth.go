package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/goodway/godos/config"
	"github.com/goodway/godos/internal/todex"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login <email>",
	Short: "Authenticate with Todex",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := getAPIService(false)
		if err != nil {
			return err
		}
		password, err := readPassword(cmd, "Password: ")
		if err != nil {
			return err
		}
		token, err := svc.Login(context.Background(), args[0], password)
		if err != nil {
			return err
		}
		if err := config.Set(config.APITokenKey, token); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Logged in.")
		return nil
	},
}

var registerCmd = &cobra.Command{
	Use:   "register <email>",
	Short: "Register a Todex account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := getAPIService(false)
		if err != nil {
			return err
		}
		password, err := readPassword(cmd, "Password: ")
		if err != nil {
			return err
		}
		token, err := svc.Register(context.Background(), args[0], password)
		if err != nil {
			return err
		}
		if err := config.Set(config.APITokenKey, token); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Registered and logged in.")
		return nil
	},
}

func readPassword(cmd *cobra.Command, prompt string) (string, error) {
	input := cmd.InOrStdin()
	if file, ok := input.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		fmt.Fprint(cmd.OutOrStdout(), prompt)
		password, err := term.ReadPassword(int(file.Fd()))
		fmt.Fprintln(cmd.OutOrStdout())
		if err != nil {
			return "", err
		}
		return strings.TrimRight(string(password), "\r\n"), nil
	}
	line, err := bufio.NewReader(input).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear the Todex session",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := config.APIToken()
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "No active session.")
			return nil
		}

		baseURL, baseErr := config.APIBaseURL()
		if baseErr == nil {
			if svc, err := todex.New(todex.ServiceConfig{BaseURL: baseURL, Token: token}); err == nil {
				_ = svc.Logout(context.Background())
			}
		}
		if err := config.Delete(config.APITokenKey); err != nil && !errors.Is(err, config.ErrNotFound) {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Logged out.")
		return nil
	},
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Inspect Todex authentication state",
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the authenticated Todex user",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		user, err := svc.CurrentUser(context.Background())
		if err != nil {
			return fmt.Errorf("authentication required or expired; run godos login: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Authenticated as %s\n", user.Email)
		return nil
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(logoutCmd)
}
