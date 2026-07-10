# Agent Guidelines

## Project

`wtx` is a small Go CLI that opens tmux workspaces on top of
[`wt`](https://github.com/timvw/wt)-managed git worktrees. `wt` owns the
worktrees; `wtx` adds multi-pane tmux sessions with per-pane code-assistant
commands driven by named layout presets.

## Build / test

```bash
go build ./...        # build all packages
go build -o wtx .     # build the binary (gitignored)
go test ./...         # run the unit suite
go vet ./... && gofmt -l .   # must be clean before commit (gofmt -l prints nothing)
```

## Layout

| Package          | Responsibility                                                        |
| ---------------- | --------------------------------------------------------------------- |
| `internal/wt`    | Thin wrapper over the `wt` CLI (`wt --format json`). No direct git.    |
| `internal/config`| TOML config: agents, layout presets, `agent:<name>` + `{{.Var}}` expansion. |
| `internal/tmux`  | `layout.go` = pure sequence builders (tested); `tmux.go` = I/O layer.  |
| `internal/cmd`   | Cobra commands: `add`, `open`, `list`, `remove`.                       |

## Conventions

- **`wtx` never shells out to `git` for worktree state** — always go through
  `wt --format json` so `wt` keeps ownership of strategies, naming, and hooks.
  (The one exception is `repoName`, which reads git's common-dir for naming.)
- **Keep tmux logic pure where possible.** Sequence construction lives in
  `layout.go` with no I/O so it stays unit-testable; `tmux.go` only executes.
- **Address tmux panes by captured `#{pane_id}`, never by index** — splits
  renumber indices and would misroute commands.
- Match the surrounding style; run `gofmt` and `go vet` before committing.

## Scope notes

Deliberately out of scope for now (both would layer on the same
`wt --format json` data): a TUI dashboard and multi-machine sync.
