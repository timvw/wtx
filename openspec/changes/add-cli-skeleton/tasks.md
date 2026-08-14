## 1. Module and dependencies

- [x] 1.1 Initialise the Go module as `github.com/timvw/wtx` with `go mod init`, then set the language version in `go.mod` to `go 1.26.0` (trim the patch version `go mod init` writes, so the floor is the 1.26 release line rather than the local 1.26.5)
- [x] 1.2 Add `github.com/spf13/cobra` with `go get`, then run `go mod tidy` and confirm `go.mod` and `go.sum` are both written and pin `spf13/pflag` and `inconshreveable/mousetrap` transitively
- [x] 1.3 Add `.gitignore` ignoring the built `wtx` binary at the repository root

## 2. Version resolution

- [x] 2.1 Create `cmd/version.go` with a package-level `var version string` (empty by default, the `-ldflags -X` stamping target) and an exported `Version() string` resolver
- [x] 2.2 Implement the resolver precedence: stamped `version` if non-empty; else `debug.ReadBuildInfo()` using `bi.Main.Version` when it is not empty and not `(devel)`; else the `vcs.revision` build setting shortened to 12 characters, suffixed with `+dirty` when `vcs.modified` is `"true"`; else the literal `dev`
- [x] 2.3 Confirm the resolver is total — it returns a non-empty string and does not panic when `ReadBuildInfo` reports `ok == false` or returns no usable settings

## 3. Command layer

- [x] 3.1 Create `cmd/root.go` with `newRootCmd() *cobra.Command`: `Use: "wtx"`, a one-line `Short`, a `Long` describing wtx as tmux workspaces on top of `wt`-managed git worktrees, `SilenceUsage` and `SilenceErrors` set so error output is emitted once by the caller rather than duplicated with a usage dump
- [x] 3.2 Set the root command's `Version` from `Version()` and set `SetVersionTemplate` so `--version` prints the bare version string plus a newline, byte-identical to what `wtx version` prints
- [x] 3.3 Give the root command a `RunE` that prints help via `cmd.Help()` and returns `nil`, so bare `wtx` exits `0` regardless of Cobra's default for an unset `Run`
- [x] 3.4 Add `newVersionCmd() *cobra.Command` in `cmd/version.go` — `Use: "version"`, a `Short` line, `Args: cobra.NoArgs` — whose `RunE` writes `Version()` and a newline to `cmd.OutOrStdout()`
- [x] 3.5 Register the version command on the root command inside `newRootCmd()`, and add an exported `Execute() error` that constructs a fresh root command and returns `rootCmd.Execute()`
- [x] 3.6 Create `main.go` in `package main` that calls `cmd.Execute()` and, on a non-nil error, writes the error to `os.Stderr` and calls `os.Exit(1)` — no other logic

## 4. Tests

- [x] 4.1 Add a test helper in `package cmd` that builds a fresh root command, redirects output with `SetOut`/`SetErr` into separate buffers, sets args with `SetArgs`, executes, and returns stdout, stderr, and the error
- [x] 4.2 Test `Version()`: returns the stamped value when `version` is non-empty, and returns a non-empty string (never panicking) when it is empty
- [x] 4.3 Test `wtx version`: writes exactly one non-empty line to stdout with no leading or trailing whitespace beyond the newline, writes nothing to stderr, and returns a nil error
- [x] 4.4 Test that `wtx --version` and `wtx version` produce byte-identical stdout, and that repeating the same invocation twice produces byte-identical stdout
- [x] 4.5 Test bare invocation (empty args): help text is printed and the returned error is nil
- [x] 4.6 Test that root help output names `wtx` and lists both `help` and `version` among the available commands
- [x] 4.7 Test `wtx help version` and `wtx version --help`: usage for the version subcommand is printed with a nil error
- [x] 4.8 Test usage errors: `wtx no-such-command`, `wtx help no-such-command`, and `wtx --no-such-flag` each return a non-nil error whose message names the offending input, and leave stdout free of error text

## 5. Build tooling

- [x] 5.1 Add a `Makefile` with a `build` target that derives a version from `git describe --tags --always --dirty` (tolerating a tagless repository) and passes it as `-ldflags "-X github.com/timvw/wtx/cmd.version=$(VERSION)"`
- [x] 5.2 Add `test` (`go test ./...`), `vet` (`go vet ./...`), and `fmt` (`gofmt -l -w .`) targets, plus a `.PHONY` declaration and a default target
- [x] 5.3 Verify the ldflags path end to end: `make build` then `./wtx version` prints the `git describe` value, confirming the `-X` symbol path matches the package

## 6. Verification against the spec

- [x] 6.1 Run `make fmt`, `make vet`, and `make test` and confirm all pass with no reported issues
- [x] 6.2 Verify an unstamped `go build` reports a build-info-derived version (not `dev`) in the git checkout, confirming the build-metadata fallback works. (Corrected during apply: this task originally named `go run`, which is precisely the mode the toolchain does *not* stamp VCS info into — `go run . version` correctly reports `dev`, exercising the third precedence step instead.)
- [x] 6.3 Walk every scenario in `specs/cli/spec.md` and confirm each is covered by a test or by a manual check recorded in the apply notes, including exit status `0` for bare invocation and non-zero for an unknown command as observed from a real shell
- [x] 6.4 Run `openspec validate add-cli-skeleton --strict` and resolve anything it reports
