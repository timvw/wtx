package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/timvw/wtx/internal/config"
	"github.com/timvw/wtx/internal/tmux"
	"github.com/timvw/wtx/internal/wt"
)

var (
	addLayout   string
	addAgent    string
	addNoLaunch bool
)

var addCmd = &cobra.Command{
	Use:   "add <branch> [base-branch]",
	Short: "Create a worktree via wt and open its tmux workspace",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := args[0]
		base := ""
		if len(args) > 1 {
			base = args[1]
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := cfg.OverrideDefaultAgent(effectiveAgent(addAgent)); err != nil {
			return err
		}
		layout, err := cfg.Resolve(addLayout)
		if err != nil {
			return err
		}

		res, err := wt.Create(branch, base)
		if err != nil {
			return err
		}
		fmt.Printf("✓ worktree %s: %s\n", res.Status, res.Path)

		repo := repoName(res.Path)
		session := tmux.SessionName(repo, res.Branch)
		panes, err := config.ExpandPanes(layout.Panes, config.TemplateVars{
			Branch:  res.Branch,
			Path:    res.Path,
			Repo:    repo,
			Session: session,
		})
		if err != nil {
			return err
		}
		if err := tmux.Launch(session, res.Path, layout.Arrange, panes); err != nil {
			return err
		}
		fmt.Printf("✓ tmux session %q (layout: %s)\n", session, layout.Name)

		if addNoLaunch {
			return nil
		}
		return tmux.Attach(session)
	},
}

func init() {
	addCmd.Flags().StringVarP(&addLayout, "layout", "l", "", "layout preset to use (default: config default_layout)")
	addCmd.Flags().StringVarP(&addAgent, "agent", "a", "", "agent for agent:default panes; overrides $WTX_AGENT and config default_agent")
	addCmd.Flags().BoolVar(&addNoLaunch, "no-launch", false, "create the session but do not attach")
	rootCmd.AddCommand(addCmd)
}
