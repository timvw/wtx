# wtx - Worktree Workspaces

`wtx` opens **tmux workspaces on top of [`wt`](https://github.com/timvw/wt)-managed
git worktrees**. `wt` owns the worktrees — strategies, naming, hooks, and
GitHub PR / GitLab MR checkout. `wtx` adds the one thing `wt` leaves out:
multi-pane tmux sessions with per-pane agent commands, driven by named layout
presets.

It's the layer that [`kwt`](https://github.com/kenn-io/kwt) /
[`gwq`](https://github.com/d-kuro/gwq) build in-tree — but as a small companion
that treats `wt` as a black box (`wt --format json`), so you keep everything
`wt` does better (PR/MR checkout, strategies, CI status).

## How it works

```
wtx add feature/x
  ├─ wt --format json create feature/x   # wt: strategy, naming, hooks
  ├─ reads .data.path from the response
  └─ builds a tmux session at that path from a layout preset
```

Layouts use the same technique as kwt: construct panes capturing stable
`#{pane_id}`s, then dispatch each pane's command **by ID** (never by index), so
splits can't misroute commands.

## Install

```bash
go install github.com/timvw/wtx@latest   # requires wt + tmux on PATH
```

## Commands

| Command                    | Purpose                                                       |
| -------------------------- | ------------------------------------------------------------- |
| `wtx add <branch> [base]`  | Create a worktree via `wt` and open its tmux workspace        |
| `wtx open [branch]`        | Attach to a worktree's workspace (launch layout if not live)  |
| `wtx list` / `wtx ls`      | List worktrees and which have a live tmux workspace           |
| `wtx remove <branch>`      | Kill the tmux workspace, then `wt rm` the worktree            |

Flags: `--layout/-l <name>` picks a preset; `--no-launch` (on `add`) creates
the session without attaching.

## Configuration

Lives at `~/.config/wtx/config.toml` (or `$WTX_HOME/config.toml`). A starter
file is written on first run.

```toml
# The code assistant launched by layouts that use agent:default. Change this
# one line to switch assistants everywhere without editing any layout.
default_agent = "claude"

# Layout used when none is given ("none" = single blank shell).
default_layout = "assistant"

# Named commands, referenced by layouts as agent:<name>. Configure flags once.
[agents]
claude = "claude"
codex  = "codex"

# Each pane is a literal command, "" (plain shell), or agent:<name>.
# agent:default resolves to the agent named by default_agent above.
[[layouts]]
name    = "assistant"
arrange = "even-horizontal"   # even-vertical | tiled | main-vertical | main-horizontal
panes   = ["agent:default", ""]   # assistant on the left, a shell on the right

[[layouts]]
name    = "duo"
arrange = "even-horizontal"
panes   = ["agent:claude", "agent:codex"]
```

Switching assistant is a one-liner — `default_agent = "codex"` (or `aider`,
`gemini`, …, once you add it to `[agents]`) — and every `agent:default` pane
follows.

### Choosing the assistant per shell or per command

There is no cross-tool standard env var for "the code assistant" (unlike
`$EDITOR` for editors), so `wtx` uses its own, `$WTX_AGENT`, with this
precedence for what `agent:default` resolves to:

```
--agent flag   >   $WTX_AGENT   >   config default_agent
```

```bash
export WTX_AGENT=codex          # session default, no config edit
WTX_AGENT=aider wtx add feat/x  # one-off for this command
wtx add feat/x --agent claude   # explicit, wins over everything
```

All three resolve through `[agents]`, so the value names an entry there (its
flags stay configured in one place). An unknown name is rejected. You can even
omit `default_agent` from config entirely and drive it purely from
`$WTX_AGENT` / `--agent`; `wtx` only errors if nothing resolves at launch.

## Code assistant integration

Launching a code assistant *is* the point of `wtx` — an `agent:<name>` pane runs
that command with its working directory set to the worktree. Run `wtx add` once
per branch to get the parallel-agents-per-worktree workflow, each assistant
isolated by directory.

Agent (and literal) pane commands may use template variables so the assistant
launches with **context**, not just the right directory:

| Variable       | Value                          |
| -------------- | ------------------------------ |
| `{{.Branch}}`  | branch checked out             |
| `{{.Path}}`    | absolute worktree path         |
| `{{.Repo}}`    | repository name                |
| `{{.Session}}` | tmux session name              |

```toml
[agents]
claude = "claude \"start working on {{.Branch}}\""   # opening prompt
codex  = "codex \"$(cat {{.Path}}/TASK.md)\""         # per-worktree task file
```

## Requirements

- [`wt`](https://github.com/timvw/wt) on `PATH`
- `tmux`
- Go 1.25+ to build from source
- optional: `fzf` for nicer interactive selection in `wtx open`

## Development

```bash
go build -o wtx .     # build the binary
go test ./...         # unit tests (config resolution, template expansion, layout sequences)
go vet ./... && gofmt -l .
```

Package layout and conventions are documented in [AGENTS.md](AGENTS.md). The
tmux layout logic is split into pure sequence builders (`internal/tmux/layout.go`,
unit-tested) and a thin execution layer (`internal/tmux/tmux.go`).

## Status

v0 — `add` / `open` / `list` / `remove`. Deliberately out of scope for now:
a TUI dashboard and multi-machine sync (the other things kwt adds); both would
layer on the same `wt --format json` data.
