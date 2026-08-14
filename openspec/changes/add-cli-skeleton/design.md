## Context

The repository has no Go code, no module, and no dependencies — see `proposal.md` — so every structural decision here is greenfield and unconstrained by existing layout. The constraints that do apply come from what the repository already commits to:

- `.github/settings.yml` declares squash-only merges, so the change lands as one commit and should be self-consistent.
- No CI workflow exists yet, so correctness has to be verifiable locally with a single command (`make test`) that a later workflow can call verbatim.
- The behavior contract is `specs/cli/spec.md`; this document only covers how it is realised.

A local Go 1.26.5 toolchain is available. Cobra, build-info-with-`ldflags`-override versioning, and a source-plus-build-tooling scope were chosen by the user up front; the rationale is recorded under Decisions.

## Goals / Non-Goals

**Goals:**

- A package layout where adding a real subcommand means adding one file and one `AddCommand` call, with no edits to `main.go`.
- Command behavior testable in-process — no compiling a binary and shelling out to it in unit tests.
- Version resolution correct in all three build modes (`go run`, `go install`, release build) without per-release source edits.

**Non-Goals:**

- Configuration file loading, environment-variable binding, or a `--config` flag. No command needs configuration yet; adding Viper now would be dependency weight with no user.
- Structured logging, `--verbose`/`--quiet`, or a global output-format flag. These are cross-cutting decisions better made once there is real output to shape.
- Shell completions and man pages. Cobra can generate both later; wiring them now would ship generated artifacts for two commands.
- Cross-compilation matrices, checksums, or signing — release concerns, deferred with the release workflow itself.

## Decisions

### Cobra as the command framework

`github.com/spf13/cobra` supplies the `help` subcommand, `-h`/`--help` on every command, per-command usage text, and "unknown command" diagnostics — most of `specs/cli/spec.md` is satisfied by its defaults rather than by hand-written code. It is the de facto standard for Go CLIs (kubectl, gh, hugo), so the layout is familiar to contributors and to code assistants working in this repo.

*Alternatives considered.* **Stdlib `flag` + manual dispatch**: zero dependencies, but help text, usage errors, and subcommand routing all become hand-written code that must be re-tested for every new subcommand — and wtx is planned to grow several. **`urfave/cli` v3**: lighter and more declarative, but a smaller ecosystem, and its v2→v3 API break makes it the less settled choice for a long-lived tool.

*Cost accepted.* Cobra pulls in `spf13/pflag` and `inconshreveable/mousetrap` — the repository's first dependencies, and therefore the first thing needing update tooling later.

### `cmd/` package holding commands, `main.go` as a thin shim

`main.go` at the repository root does exactly one thing: call `cmd.Execute()` and translate an error into `os.Exit(1)`. All command construction lives in package `cmd`.

This matters for testability: `package main` cannot be imported by a test in another package, and `os.Exit` is untestable in-process. Keeping commands in an importable package lets tests build a command, redirect its output with `SetOut`/`SetErr` into a buffer, execute it, and assert on the bytes — which is how the spec's stream-and-exit-status scenarios get covered without spawning processes.

*Alternative considered.* Everything in `package main`: fewer files, but forces every behavioral test to build and exec a binary, making the test suite slow and platform-sensitive from day one.

Package `cmd` at the repository root (rather than `internal/cli/`) follows the dominant Cobra convention. It is importable by outside code, which is a minor downside; `internal/` can be adopted later if the command layer should be sealed off, and doing so is a mechanical move while there are only two commands.

### Each command is built by a constructor, not a package-level variable

Commands are produced by functions (`newRootCmd()`, `newVersionCmd()`) returning a fresh `*cobra.Command`, rather than declared as package-level `var`s.

Cobra commands carry mutable state — parsed flag values, output writers, and the `args` slice — and a package-level command shares that state across tests in the same binary. A test that sets `--version` would leak into the next test's execution. Fresh construction per call makes each test independent and each execution deterministic, which is what the spec's "invoking the same binary twice produces byte-identical output" scenario needs.

*Alternative considered.* Package-level `var rootCmd = &cobra.Command{...}` with `init()` wiring — the layout `cobra-cli` generates. Rejected for the state-leak reason above; the constructor form costs one extra line per command.

### Version resolution: stamped value, then build info, then `dev`

A single exported function resolves the version, in this precedence:

1. A package-level `version` string, empty by default, set at build time via `-ldflags "-X <module>/cmd.version=<tag>"`. Non-empty wins.
2. Otherwise `runtime/debug.ReadBuildInfo()`: prefer `bi.Main.Version` when it is a real version (not the `(devel)` placeholder the toolchain writes for non-module builds), else fall back to the `vcs.revision` build setting, shortened, suffixed with `+dirty` when `vcs.modified` is true.
3. Otherwise the literal `dev`.

The point of step 2 is that `go install github.com/timvw/wtx@v0.1.0` reports `v0.1.0` with no build flags at all, and a `go build` in a git checkout reports the commit it came from. Without it, every build that forgets the `-ldflags` incantation — which is every build a contributor makes by hand — reports a uselessly identical string.

The function must be total: `ReadBuildInfo` returns `ok == false` in some build modes, and the spec forbids an empty version or a panic, hence the `dev` floor.

*Alternative considered.* `-ldflags` only. One line of code instead of ~20, but `go install` and local `go build` both report `dev`, which is exactly when a version string is most useful for bug reports.

*Trade-off accepted.* The output is not a fixed format — a tag, a revision hash, or `dev` depending on build mode. The spec deliberately requires only "a single non-empty stable line", not a parseable schema, so this stays within contract.

### `--version` and `wtx version` share one resolver

The root command sets `Version` (which makes Cobra render `--version` automatically) from the same resolver the `version` subcommand prints. The spec requires the two to be identical, and a single source makes that true by construction rather than by discipline.

Cobra's default `--version` template prefixes the program name (`wtx version <v>`). The subcommand prints the bare string on its own line. To keep "identical string" unambiguous, the root command's version template is set to print the bare version and a newline, matching the subcommand byte for byte. A test asserts the equality directly.

### The `help` command is replaced, not inherited

*Discovered during implementation.* Cobra's built-in help command (v1.10.2, `command.go:1293`) handles an unknown topic with a plain `Run` that prints `Unknown help topic` and returns nothing — so `wtx help no-such-command` exits `0`. The spec requires a non-zero status there, because a help lookup that silently succeeds for a command that does not exist is a trap for scripts.

`SetHelpCommand` overrides `helpCommand`, and `InitDefaultHelpCmd` skips its default when that field is already set, so supplying a replacement is enough. The replacement (`cmd/help.go`) uses `RunE` and returns an error naming the unknown topic; for a topic that does exist it does what upstream does — initialise the target's help and version flags, then call `Help()`.

*Alternative considered.* Relax the spec to match Cobra's default. Rejected: the exit status is the contract a CI step or shell script actually branches on, and matching a framework default is not a reason to ship the weaker behavior.

### Bare `wtx` prints help and exits 0

The root command's `RunE` prints help and returns nil, rather than being left unset. An unset `Run`/`RunE` also makes Cobra print help, but it returns an error in some Cobra versions' `Execute()` paths; setting it explicitly makes the exit-`0` guarantee independent of Cobra's default.

### Makefile as the build entry point

A `Makefile` with `build`, `test`, `vet`, `fmt` targets. `build` computes a version from `git describe` and passes it through `-ldflags`, so the stamping path is exercised by ordinary local use and not only by an as-yet-unwritten release workflow. `make` is present everywhere Go development happens and needs no extra tooling; a later CI workflow calls the same targets, so local and CI cannot drift.

*Alternative considered.* Raw `go build`/`go test` invocations documented in a README. Simpler, but the `-ldflags` string is long enough that nobody would type it, and CI would end up with its own copy.

### Go language version in `go.mod`

`go.mod` declares `go 1.26.0`, matching the current release line (locally installed: 1.26.5). Confirmed by the user: track the latest Go rather than keeping an older floor for compatibility.

The declared version is the minor release, not the local patch version that `go mod init` writes by default. A patch-level floor (`go 1.26.5`) would make the toolchain refuse to build on 1.26.0–1.26.4 and trigger a toolchain download, for no benefit — nothing here depends on a patch-release fix.

## Risks / Trade-offs

- **First dependency in the repository** → Cobra's tree is small (`pflag`, `mousetrap`) and stable, and `go.sum` pins it. Dependabot or an equivalent belongs with the deferred CI change, not here.
- **Command-layer shape is a durable commitment** → Cobra's `Command` structure is what every future subcommand is written against; switching frameworks later means rewriting them all. Mitigated by the framework being the industry default and by the change being small enough to redo now if the first real subcommand exposes a problem.
- **Version string format is not fixed across build modes** → the spec requires stability and non-emptiness, not a schema, so scripts must not parse it into components. If a machine-readable form is needed later, add `wtx version --json` as an additive change rather than reshaping the plain output.
- **`vcs.revision` is absent when building from a source tarball or with `-buildvcs=false`** → falls through to `dev`, which is correct behavior rather than a failure. Release builds stamp explicitly and are unaffected.
- **No CI yet, so nothing enforces `make test` on a pull request** → accepted for this change; it is precisely the gap the follow-up workflow change closes, and this change is what makes that workflow possible.

## Migration Plan

Not applicable: nothing exists to migrate from, there are no users of the binary yet, and no external state changes. Rollback is reverting the single squash commit.

## Open Questions

- Whether the command layer should eventually move to `internal/cli/` to seal it from external importers. Deferrable — a mechanical rename while the package holds two commands, and it changes no behavior the spec describes.
