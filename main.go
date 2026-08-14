// Command wtx manages tmux workspaces on top of wt-managed git worktrees.
package main

import (
	"fmt"
	"os"

	"github.com/timvw/wtx/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "wtx:", err)
		os.Exit(1)
	}
}
