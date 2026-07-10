package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/timvw/wtx/internal/tmux"
	"github.com/timvw/wtx/internal/wt"
)

var removeCmd = &cobra.Command{
	Use:     "remove <branch>",
	Aliases: []string{"rm"},
	Short:   "Kill the tmux workspace and remove the worktree via wt",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := args[0]

		// The repo name resolves from the current directory's git common dir,
		// which is identical for the main checkout and every linked worktree,
		// so the session name matches the one used at creation even when the
		// branch's worktree is already gone.
		session := tmux.SessionName(repoName("."), branch)
		if err := tmux.Kill(session); err != nil {
			return err
		}
		fmt.Printf("✓ killed tmux session %q\n", session)

		return wt.Remove(branch)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
