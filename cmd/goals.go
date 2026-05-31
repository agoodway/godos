package cmd

import (
	"context"
	"fmt"

	"github.com/goodway/godos/internal/todex"
	"github.com/spf13/cobra"
)

var (
	goalAddDescriptionFlag  string
	goalAddReasonFlag       string
	goalEditTitleFlag       string
	goalEditDescriptionFlag string
	goalEditReasonFlag      string
	goalRmForceFlag         bool
)

var goalsCmd = &cobra.Command{
	Use:   "goals",
	Short: "List remote goals",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		goals, err := svc.ListGoals(context.Background())
		if err != nil {
			return err
		}
		if len(goals) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No goals found")
			return nil
		}
		for _, goal := range goals {
			fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %d%%\n", goal.ShortID, goal.Title, goal.Progress)
		}
		return nil
	},
}

var goalCmd = &cobra.Command{
	Use:   "goal",
	Short: "Manage remote goals",
}

var goalAddCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Create a remote goal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		changes := todex.GoalChanges{Title: stringValue(args[0]), Description: optionalString(goalAddDescriptionFlag), Reason: optionalString(goalAddReasonFlag)}
		goal, err := svc.CreateGoal(context.Background(), changes)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created %s \"%s\"\n", goal.ShortID, goal.Title)
		return nil
	},
}

var goalShowCmd = &cobra.Command{
	Use:   "show <id-prefix>",
	Short: "Show a remote goal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		goal, err := svc.GetGoal(context.Background(), args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Title: %s\nDescription: %s\nReason: %s\nProgress: %d%%\n", goal.Title, goal.Description, goal.Reason, goal.Progress)
		return nil
	},
}

var goalEditCmd = &cobra.Command{
	Use:   "edit <id-prefix>",
	Short: "Edit a remote goal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		changes := todex.GoalChanges{}
		if cmd.Flags().Changed("title") {
			changes.Title = stringValue(goalEditTitleFlag)
		}
		if cmd.Flags().Changed("description") {
			changes.Description = stringValue(goalEditDescriptionFlag)
		}
		if cmd.Flags().Changed("reason") {
			changes.Reason = stringValue(goalEditReasonFlag)
		}
		if changes.Title == nil && changes.Description == nil && changes.Reason == nil {
			return fmt.Errorf("no changes specified; provide --title, --description, or --reason")
		}
		goal, err := svc.UpdateGoal(context.Background(), args[0], changes)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Updated %s \"%s\"\n", goal.ShortID, goal.Title)
		return nil
	},
}

var goalRmCmd = &cobra.Command{
	Use:   "rm <id-prefix>",
	Short: "Remove a remote goal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !goalRmForceFlag && !confirm(cmd, fmt.Sprintf("Delete goal %q? [y/N] ", args[0])) {
			fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
			return nil
		}
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		goal, err := svc.DeleteGoal(context.Background(), args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Removed %s \"%s\"\n", goal.ShortID, goal.Title)
		return nil
	},
}

var goalLinkCmd = goalTaskCmd("link", "Link a task to a goal", func(ctx context.Context, svc *todex.Service, goalPrefix, taskPrefix string) (todex.Goal, error) {
	return svc.LinkGoalTask(ctx, goalPrefix, taskPrefix)
}, "Linked")

var goalUnlinkCmd = goalTaskCmd("unlink", "Unlink a task from a goal", func(ctx context.Context, svc *todex.Service, goalPrefix, taskPrefix string) (todex.Goal, error) {
	return svc.UnlinkGoalTask(ctx, goalPrefix, taskPrefix)
}, "Unlinked")

func goalTaskCmd(use, short string, action func(context.Context, *todex.Service, string, string) (todex.Goal, error), verb string) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <goal-id-prefix> <task-id-prefix>",
		Short: short,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := getAPIService(true)
			if err != nil {
				return err
			}
			goal, err := action(context.Background(), svc, args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s task for %s \"%s\"\n", verb, goal.ShortID, goal.Title)
			return nil
		},
	}
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return stringValue(value)
}

func stringValue(value string) *string {
	return &value
}

func init() {
	goalAddCmd.Flags().StringVar(&goalAddDescriptionFlag, "description", "", "goal description")
	goalAddCmd.Flags().StringVar(&goalAddReasonFlag, "reason", "", "goal reason")
	goalEditCmd.Flags().StringVar(&goalEditTitleFlag, "title", "", "goal title")
	goalEditCmd.Flags().StringVar(&goalEditDescriptionFlag, "description", "", "goal description")
	goalEditCmd.Flags().StringVar(&goalEditReasonFlag, "reason", "", "goal reason")
	goalRmCmd.Flags().BoolVarP(&goalRmForceFlag, "force", "f", false, "skip confirmation prompt")
	goalCmd.AddCommand(goalAddCmd, goalShowCmd, goalEditCmd, goalRmCmd, goalLinkCmd, goalUnlinkCmd)
	rootCmd.AddCommand(goalsCmd, goalCmd)
}
