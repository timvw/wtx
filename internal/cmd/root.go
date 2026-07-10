package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timvw/wtx/internal/wt"
)

var rootCmd = &cobra.Command{
	Use:   "wtx",
	Short: "Launch tmux workspaces on top of wt-managed git worktrees",
	Long: "wtx layers named tmux layouts and per-pane agent commands on top of\n" +
		"wt. wt owns the worktrees (strategies, hooks, PR/MR checkout); wtx opens\n" +
		"a workspace for the branch you want to work on.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "wtx: "+err.Error())
		os.Exit(1)
	}
}

// repoName derives a repository name from any worktree path via git's common
// dir, so linked worktrees and the main checkout resolve to the same name.
// The path is made absolute first so a relative input like "." still resolves
// to the repository directory rather than "." itself.
func repoName(worktreePath string) string {
	if abs, err := filepath.Abs(worktreePath); err == nil {
		worktreePath = abs
	}
	out, err := exec.Command("git", "-C", worktreePath, "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return filepath.Base(worktreePath)
	}
	commonDir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(worktreePath, commonDir)
	}
	return filepath.Base(filepath.Dir(commonDir))
}

// branchLabel is a display label that stays meaningful for detached worktrees,
// whose branch field is empty.
func branchLabel(w wt.Worktree) string {
	if w.Branch == "" {
		return "(detached)"
	}
	return w.Branch
}

// sessionKey is the branch component of the tmux session name. Detached
// worktrees have no branch, so they fall back to a per-commit key; otherwise
// every detached worktree in a repo would collide onto the same session.
func sessionKey(w wt.Worktree) string {
	if w.Branch != "" {
		return w.Branch
	}
	head := w.Head
	if len(head) > 12 {
		head = head[:12]
	}
	if head == "" {
		head = "unknown"
	}
	return "detached-" + head
}

// pickBranch lets the user choose a worktree interactively, preferring fzf when
// available and falling back to a numbered prompt.
func pickBranch(worktrees []wt.Worktree) (wt.Worktree, error) {
	if len(worktrees) == 0 {
		return wt.Worktree{}, fmt.Errorf("no worktrees found")
	}
	if _, err := exec.LookPath("fzf"); err == nil {
		return pickWithFzf(worktrees)
	}
	return pickWithPrompt(worktrees)
}

func pickWithFzf(worktrees []wt.Worktree) (wt.Worktree, error) {
	// Prefix each row with a stable index and select on that, so the choice is
	// unambiguous even for detached or same-named branches.
	var b strings.Builder
	for i, w := range worktrees {
		fmt.Fprintf(&b, "%d\t%s\t%s\n", i, branchLabel(w), w.Path)
	}
	cmd := exec.Command("fzf", "--with-nth=2,3", "--delimiter=\t", "--prompt=worktree> ")
	cmd.Stdin = strings.NewReader(b.String())
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return wt.Worktree{}, fmt.Errorf("selection cancelled")
	}
	idx, err := strconv.Atoi(strings.SplitN(strings.TrimSpace(string(out)), "\t", 2)[0])
	if err != nil || idx < 0 || idx >= len(worktrees) {
		return wt.Worktree{}, fmt.Errorf("invalid selection")
	}
	return worktrees[idx], nil
}

func pickWithPrompt(worktrees []wt.Worktree) (wt.Worktree, error) {
	for i, w := range worktrees {
		fmt.Printf("  %2d) %-30s %s\n", i+1, branchLabel(w), w.Path)
	}
	fmt.Print("select> ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return wt.Worktree{}, fmt.Errorf("selection cancelled")
	}
	n, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || n < 1 || n > len(worktrees) {
		return wt.Worktree{}, fmt.Errorf("invalid selection")
	}
	return worktrees[n-1], nil
}

// effectiveAgent applies the agent-selection precedence: an explicit --agent
// flag, then $WTX_AGENT, then "" (meaning: keep the config's default_agent).
func effectiveAgent(flag string) string {
	if flag != "" {
		return flag
	}
	return os.Getenv("WTX_AGENT")
}

// findByBranch returns the worktree whose branch matches, else an error.
func findByBranch(worktrees []wt.Worktree, branch string) (wt.Worktree, error) {
	for _, w := range worktrees {
		if w.Branch == branch {
			return w, nil
		}
	}
	return wt.Worktree{}, fmt.Errorf("no worktree for branch %q", branch)
}
