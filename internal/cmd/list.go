package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/timvw/wtx/internal/tmux"
	"github.com/timvw/wtx/internal/wt"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List worktrees and whether each has a live tmux workspace",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		worktrees, err := wt.List()
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "BRANCH\tWORKSPACE\tPATH")
		for _, t := range worktrees {
			session := tmux.SessionName(repoName(t.Path), sessionKey(t))
			state := "-"
			if tmux.HasSession(session) {
				state = "live"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", branchLabel(t), state, t.Path)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
