package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/goodway/godos/internal/todex"
	"github.com/spf13/cobra"
)

var (
	notesFolderFlag   string
	notesQueryFlag    string
	notesPinnedFlag   bool
	notesDeletedFlag  bool
	noteAddFolderFlag string
	noteRmForceFlag   bool
)

var notesCmd = &cobra.Command{
	Use:   "notes",
	Short: "List remote notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		filters := todex.NoteFilters{FolderName: notesFolderFlag, Query: notesQueryFlag}
		if cmd.Flags().Changed("pinned") {
			filters.Pinned = &notesPinnedFlag
		}
		if cmd.Flags().Changed("deleted") {
			filters.Deleted = &notesDeletedFlag
		}
		notes, err := svc.ListNotes(context.Background(), filters)
		if err != nil {
			return err
		}
		if len(notes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No notes found")
			return nil
		}
		for _, note := range notes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s  %s\n", note.ShortID, note.Title)
		}
		return nil
	},
}

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage remote notes",
}

var noteAddCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Create a remote note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		body, err := editText("")
		if err != nil {
			return err
		}
		note, err := svc.CreateNote(context.Background(), args[0], noteAddFolderFlag, body)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created %s \"%s\"\n", note.ShortID, note.Title)
		return nil
	},
}

var noteShowCmd = &cobra.Command{
	Use:   "show <id-prefix>",
	Short: "Show a remote note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		note, err := svc.GetNote(context.Background(), args[0])
		if err != nil {
			return err
		}
		if note.Body == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "(empty note)")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), note.Body)
		}
		return nil
	},
}

var noteEditCmd = &cobra.Command{
	Use:   "edit <id-prefix>",
	Short: "Edit a remote note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		note, err := svc.GetNote(context.Background(), args[0])
		if err != nil {
			return err
		}
		body, err := editText(note.Body)
		if err != nil {
			return err
		}
		note, err = svc.UpdateNoteBody(context.Background(), note.ID, body)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Updated %s \"%s\"\n", note.ShortID, note.Title)
		return nil
	},
}

var noteRmCmd = &cobra.Command{
	Use:   "rm <id-prefix>",
	Short: "Remove a remote note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !noteRmForceFlag && !confirm(cmd, fmt.Sprintf("Delete note %q? [y/N] ", args[0])) {
			fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
			return nil
		}
		svc, err := getAPIService(true)
		if err != nil {
			return err
		}
		note, err := svc.DeleteNote(context.Background(), args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Removed %s \"%s\"\n", note.ShortID, note.Title)
		return nil
	},
}

var noteRestoreCmd = noteActionCmd("restore", "Restore a remote note", func(ctx context.Context, svc *todex.Service, prefix string) (todex.Note, error) {
	return svc.RestoreNote(ctx, prefix)
}, "Restored")

var notePinCmd = noteActionCmd("pin", "Pin a remote note", func(ctx context.Context, svc *todex.Service, prefix string) (todex.Note, error) {
	return svc.PinNote(ctx, prefix)
}, "Pinned")

var noteUnpinCmd = noteActionCmd("unpin", "Unpin a remote note", func(ctx context.Context, svc *todex.Service, prefix string) (todex.Note, error) {
	return svc.UnpinNote(ctx, prefix)
}, "Unpinned")

func noteActionCmd(use, short string, action func(context.Context, *todex.Service, string) (todex.Note, error), verb string) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <id-prefix>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := getAPIService(true)
			if err != nil {
				return err
			}
			note, err := action(context.Background(), svc, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s \"%s\"\n", verb, note.ShortID, note.Title)
			return nil
		},
	}
}

func editText(initial string) (string, error) {
	file, err := os.CreateTemp("", "godos-note-*.md")
	if err != nil {
		return "", err
	}
	name := file.Name()
	defer func() {
		if err := os.Remove(name); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not remove temp file %s: %v\n", name, err)
		}
	}()
	if _, err := file.WriteString(initial); err != nil {
		file.Close()
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	body, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func confirm(cmd *cobra.Command, prompt string) bool {
	fmt.Fprint(cmd.OutOrStdout(), prompt)
	reader := bufio.NewReader(cmd.InOrStdin())
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

func init() {
	notesCmd.Flags().StringVar(&notesFolderFlag, "folder", "", "remote note folder name")
	notesCmd.Flags().StringVar(&notesQueryFlag, "query", "", "search query")
	notesCmd.Flags().BoolVar(&notesPinnedFlag, "pinned", false, "show pinned notes")
	notesCmd.Flags().BoolVar(&notesDeletedFlag, "deleted", false, "show deleted notes")
	noteAddCmd.Flags().StringVar(&noteAddFolderFlag, "folder", "", "remote note folder name")
	noteRmCmd.Flags().BoolVarP(&noteRmForceFlag, "force", "f", false, "skip confirmation prompt")
	noteCmd.AddCommand(noteAddCmd, noteShowCmd, noteEditCmd, noteRmCmd, noteRestoreCmd, notePinCmd, noteUnpinCmd)
	rootCmd.AddCommand(notesCmd, noteCmd)
}
