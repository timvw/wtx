// Package wt is a thin wrapper around the wt CLI. wtx never touches git
// directly: it shells out to `wt --format json` so wt keeps ownership of
// worktree strategies, naming, hooks, and forge (PR/MR) integration.
package wt

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Worktree mirrors an entry of `wt --format json list`.
type Worktree struct {
	Path   string `json:"path"`
	Head   string `json:"head"`
	Branch string `json:"branch"`
}

// envelope is the common JSON shape wt wraps every command in.
type envelope struct {
	OK    bool            `json:"ok"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"`
}

func runJSON(args ...string) (json.RawMessage, error) {
	full := append([]string{"--format", "json"}, args...)
	cmd := exec.Command("wt", full...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		// wt prints its {"ok":false,"error":...} envelope to stdout even when
		// it exits non-zero; surface that message rather than a bare exit code.
		var env envelope
		if json.Unmarshal(out, &env) == nil && env.Error != "" {
			return nil, fmt.Errorf("wt %s: %s", strings.Join(args, " "), env.Error)
		}
		return nil, fmt.Errorf("wt %s: %w", strings.Join(args, " "), err)
	}
	var env envelope
	if err := json.Unmarshal(out, &env); err != nil {
		return nil, fmt.Errorf("parsing wt output: %w", err)
	}
	if !env.OK {
		return nil, fmt.Errorf("wt reported failure: %s", env.Error)
	}
	return env.Data, nil
}

// List returns all worktrees known to wt for the current repo.
func List() ([]Worktree, error) {
	data, err := runJSON("list")
	if err != nil {
		return nil, err
	}
	var payload struct {
		Worktrees []Worktree `json:"worktrees"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parsing worktree list: %w", err)
	}
	return payload.Worktrees, nil
}

// CreateResult reports what `wt create` did and where the worktree lives.
type CreateResult struct {
	Status string `json:"status"` // "created" or "exists"
	Branch string `json:"branch"`
	Path   string `json:"path"`
}

// Create asks wt to create (or locate an existing) worktree for branch and
// returns its filesystem path. wt runs its own pre/post-create hooks.
func Create(branch string, base string) (CreateResult, error) {
	args := []string{"create", branch}
	if base != "" {
		args = append(args, base)
	}
	data, err := runJSON(args...)
	if err != nil {
		return CreateResult{}, err
	}
	var res CreateResult
	if err := json.Unmarshal(data, &res); err != nil {
		return CreateResult{}, fmt.Errorf("parsing create result: %w", err)
	}
	if res.Path == "" {
		return CreateResult{}, fmt.Errorf("wt create returned no path for %q", branch)
	}
	return res, nil
}

// Remove asks wt to delete the worktree for branch. It uses text mode because
// removal needs no structured output; wt's own pre/post-remove hooks run.
func Remove(branch string) error {
	cmd := exec.Command("wt", "rm", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wt rm %s: %w", branch, err)
	}
	return nil
}
