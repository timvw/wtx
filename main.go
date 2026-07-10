// Command wtx launches tmux workspaces on top of wt-managed git worktrees.
//
// wtx treats wt as the worktree engine (creation, strategies, hooks, PR/MR
// checkout) and adds the one thing wt leaves out: multi-pane tmux sessions
// with per-pane agent commands, driven by named layout presets.
package main

import "github.com/timvw/wtx/internal/cmd"

func main() {
	cmd.Execute()
}
