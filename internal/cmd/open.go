package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/timvw/wtx/internal/config"
	"github.com/timvw/wtx/internal/tmux"
	"github.com/timvw/wtx/internal/wt"
)

var (
	openLayout string
	openAgent  string
)

var openCmd = &cobra.Command{
	Use:   "open [branch]",
	Short: "Attach to a worktree's workspace, launching the layout if needed",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		worktrees, err := wt.List()
		if err != nil {
			return err
		}

		var target wt.Worktree
		if len(args) == 1 {
			target, err = findByBranch(worktrees, args[0])
		} else {
			target, err = pickBranch(worktrees)
		}
		if err != nil {
			return err
		}

		repo := repoName(target.Path)
		session := tmux.SessionName(repo, sessionKey(target))
		if !tmux.HasSession(session) {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if err := cfg.OverrideDefaultAgent(effectiveAgent(openAgent)); err != nil {
				return err
			}
			layout, err := cfg.Resolve(openLayout)
			if err != nil {
				return err
			}
			panes, err := config.ExpandPanes(layout.Panes, config.TemplateVars{
				Branch:  target.Branch,
				Path:    target.Path,
				Repo:    repo,
				Session: session,
			})
			if err != nil {
				return err
			}
			if err := tmux.Launch(session, target.Path, layout.Arrange, panes); err != nil {
				return err
			}
			fmt.Printf("✓ tmux session %q (layout: %s)\n", session, layout.Name)
		}
		return tmux.Attach(session)
	},
}

func init() {
	openCmd.Flags().StringVarP(&openLayout, "layout", "l", "", "layout preset when launching (default: config default_layout)")
	openCmd.Flags().StringVarP(&openAgent, "agent", "a", "", "agent for agent:default panes; overrides $WTX_AGENT and config default_agent")
	rootCmd.AddCommand(openCmd)
}
