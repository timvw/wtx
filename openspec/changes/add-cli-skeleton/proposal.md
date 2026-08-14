## Why

`timvw/wtx` currently contains no source code at all — only repository metadata (`.github/settings.yml`) and OpenSpec planning artifacts. Every feature planned for wtx (tmux workspaces on top of `wt`-managed git worktrees, with per-pane code-assistant layouts) has to hang off an executable that does not exist yet. This change lays that foundation: a buildable, testable Go binary named `wtx` with the two commands every CLI needs from day one — `help` and `version` — so subsequent changes add real subcommands instead of also inventing the program's entry point, argument parsing, and exit-code conventions.

Doing it now also unblocks the CI workflow deferred in `.github/settings.yml`: a follow-up change can add build/vet/test workflows and then the branch protection that depends on them, because there will finally be something to build.

## What Changes

- Add a Go module for the `wtx` executable, with `main` as a thin entry point that delegates to a command layer and translates command failure into a non-zero exit code.
- Adopt **Cobra** (`github.com/spf13/cobra`) as the command framework. It supplies the `help` subcommand, `-h`/`--help` flags, and usage text for the root command and every future subcommand.
- Add a **root command** (`wtx`) that, invoked with no arguments, prints its help text and exits successfully. It carries a short and long description of what wtx is.
- Add a **`version` subcommand** (`wtx version`) that prints the program version, plus a `--version` flag on the root command that prints the same string.
- **Version resolution**: default to `runtime/debug.ReadBuildInfo()` so `go install`/`go build` report a real module version or VCS revision, with a `-ldflags -X` override so release builds can stamp an exact tag, falling back to a `dev` placeholder when neither is available.
- Add unit tests covering version resolution and the observable output of `wtx version`, `wtx help`, and unknown-command handling.
- Add a `Makefile` with `build`, `test`, `vet`, and `fmt` targets, where `build` stamps the version via `-ldflags`.
- Add a `.gitignore` for the built binary.

**Out of scope**, deliberately: any real wtx functionality (worktree or tmux operations), GitHub Actions workflows, release automation / GoReleaser, shell completions, and man-page generation. Those are follow-up changes.

## Capabilities

### New Capabilities

- `cli`: The `wtx` executable's command-line surface — how the program is invoked, how it reports its own version, how it presents help and usage, and what exit codes it returns. This is the contract every future subcommand plugs into.

### Modified Capabilities

_None — `repo-settings` is the only existing capability and none of its requirements change. This change adds no workflow and no branch protection._

## Impact

- **New files**: `go.mod`, `go.sum`, `main.go`, `cmd/root.go`, `cmd/version.go`, tests alongside them, `Makefile`, `.gitignore`.
- **New dependencies**: `github.com/spf13/cobra` and its transitive dependencies (`spf13/pflag`, `inconshreveable/mousetrap`). This is the first dependency the repository takes on.
- **Toolchain**: introduces a Go toolchain requirement for building and testing the repository, where previously none existed.
- **User-visible surface**: establishes `wtx`, `wtx help`, `wtx version`, `wtx --version`, and the exit-code convention. Once released, these are compatibility commitments.
- **Unblocks**: the CI workflow and branch-protection follow-up signposted in `.github/settings.yml`.
- **Risk**: low. No existing behavior can regress because there is no existing code. The main durable decision is Cobra as the command framework — cheap to adopt now, progressively more expensive to replace once many subcommands exist.
