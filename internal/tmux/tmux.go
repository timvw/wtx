// Package tmux builds and drives tmux workspaces. The layout technique is
// borrowed from kwt: construct panes capturing stable #{pane_id}s, then target
// panes by ID (never index) when sending commands, so splits can't misroute.
//
// The pure sequence-building logic lives in layout.go and is unit-tested; this
// file is the thin I/O layer that runs those sequences against tmux.
package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// HasSession reports whether a tmux session with the given name already exists.
func HasSession(name string) bool {
	return exec.Command("tmux", "has-session", "-t", name).Run() == nil
}

// Launch creates a detached tmux session named `session` rooted at `dir`, with
// one pane per entry of `panes` arranged via `arrange` (ignored for a single
// pane). Non-empty pane entries are run as commands; "" leaves a plain shell.
// It is a no-op if the session already exists.
func Launch(session, dir, arrange string, panes []string) error {
	if HasSession(session) {
		return nil
	}
	if len(panes) == 0 {
		panes = []string{""}
	}

	// Phase 1: construction. Capture each new pane's stable ID; only
	// new-session and split-window print output, so non-empty lines are IDs.
	var paneIDs []string
	created := false
	for i, args := range BuildConstructionSequence(session, dir, arrange, len(panes)) {
		out, err := runCapture(args...)
		if err != nil {
			// The first command is new-session; if it lost a race with a
			// concurrent creator the session now exists, so treat it as done.
			if i == 0 && HasSession(session) {
				return nil
			}
			return abort(created, session, args[0], err)
		}
		if args[0] == "new-session" {
			created = true
		}
		if out != "" {
			paneIDs = append(paneIDs, out)
		}
	}
	if len(paneIDs) != len(panes) {
		return abort(created, session, "construction",
			fmt.Errorf("expected %d panes but tmux created %d", len(panes), len(paneIDs)))
	}

	// Phase 2: dispatch commands by captured pane ID.
	for _, args := range BuildPaneCommandSequence(paneIDs, panes) {
		if err := run(args...); err != nil {
			return abort(created, session, args[0], err)
		}
	}
	return nil
}

// abort tears down a session this call created before returning the error, so a
// failed launch never leaves a half-built session behind.
func abort(created bool, session, stage string, cause error) error {
	if created {
		_ = Kill(session)
	}
	return fmt.Errorf("tmux %s: %w", stage, cause)
}

// Attach connects the terminal to the named session. Inside tmux it switches
// the client; otherwise it replaces the process with `tmux attach` so the
// session takes over the terminal.
func Attach(session string) error {
	if os.Getenv("TMUX") != "" {
		return run("switch-client", "-t", session)
	}
	bin, err := exec.LookPath("tmux")
	if err != nil {
		return err
	}
	return syscall.Exec(bin, []string{"tmux", "attach-session", "-t", session}, os.Environ())
}

// Kill terminates a session if it exists.
func Kill(session string) error {
	if !HasSession(session) {
		return nil
	}
	return run("kill-session", "-t", session)
}

func run(args ...string) error {
	cmd := exec.Command("tmux", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCapture(args ...string) (string, error) {
	out, err := exec.Command("tmux", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
