## 1. Local verification tooling

- [ ] 1.1 Add `.golangci.yml` at the repository root: golangci-lint v2 schema, `run.timeout: 5m`, `linters.default: none` enabling `govet`, `errcheck`, `staticcheck`, `unused`, `ineffassign`, with `errcheck` excluded for `_test\.go` paths, and the `gofmt` formatter enabled.
- [ ] 1.2 Add `lint` and `fmt-check` targets to the `Makefile` (add both to `.PHONY`): `lint` runs `golangci-lint run`, `fmt-check` fails when `gofmt -l .` prints anything and echoes the offending paths. These are what a contributor runs to reproduce CI locally.
- [ ] 1.3 Verify locally: `make fmt-check`, `make vet`, `make test` all pass on the current tree. If `golangci-lint` is installed, `make lint` passes too; if it is not, note that CI is the first place it runs.

## 2. Project metadata shipped in releases

- [ ] 2.1 Add `LICENSE` at the repository root: MIT, copyright Tim Van Wassenhove, current year.
- [ ] 2.2 Add `README.md`: what `wtx` is (tmux workspaces on top of `wt`-managed git worktrees), how to install from a GitHub release (download the archive for your platform, extract, put `wtx` on `PATH`), how to build from source (`make build`), and the license. Keep it minimal — do not document subcommands that do not exist yet.
- [ ] 2.3 Add `dist/` to `.gitignore` (GoReleaser's output directory).

## 3. GoReleaser configuration

- [ ] 3.1 Add `.goreleaser.yml` with `version: 2`, `project_name: wtx`, and a `before.hooks` entry running `go mod tidy`.
- [ ] 3.2 Configure the build: `main: .`, `binary: wtx`, `flags: [-trimpath]`, `env: [CGO_ENABLED=0]`, and `ldflags: -s -w -X github.com/timvw/wtx/cmd.version={{.Version}}` — the symbol must match the unexported `version` var in `cmd/version.go`.
- [ ] 3.3 Declare the target matrix: `goos: [linux, darwin, windows]` × `goarch: [amd64, arm64]`, with `windows/arm64` under `ignore:` and a comment saying why it is excluded.
- [ ] 3.4 Configure archives: `tar.gz` by default with a `zip` `format_overrides` for Windows, `name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"`, and the default `LICENSE*` / `README*` file globs so both ship in every archive.
- [ ] 3.5 Configure `checksum.name_template: 'checksums.txt'`.
- [ ] 3.6 Configure `changelog` with `sort: asc` and filters excluding `^docs:`, `^test:`, `^chore:`, `Merge pull request`, and `Merge branch`.
- [ ] 3.7 Configure `release`: `github.owner: timvw`, `github.name: wtx`, `draft: false`, `prerelease: auto`, `name_template: "{{.Tag}}"`, and `replace_existing_artifacts: true` (re-run safety — see design.md).
- [ ] 3.8 Verify locally with `goreleaser release --snapshot --clean` (or `goreleaser check` if GoReleaser is not installed): confirm five archives plus `checksums.txt` land in `dist/`, that each archive contains `wtx`, `LICENSE`, and `README.md`, and that the extracted Linux or macOS binary prints the snapshot version from `wtx version`.

## 4. CI workflow

- [ ] 4.1 Add `.github/workflows/ci.yml` named `CI`, triggered on `pull_request` targeting `main`, `push` to `main`, and `workflow_dispatch`; set `permissions: contents: read` and a concurrency group keyed on the ref with `cancel-in-progress: true`.
- [ ] 4.2 Add a leading comment stating that job names are a public interface consumed by `required_status_checks` in `.github/settings.yml`, and that renaming one silently breaks branch protection.
- [ ] 4.3 Add the `Lint` job (ubuntu-latest): checkout, `actions/setup-go` with `go-version-file: go.mod` and `cache: true`, `go mod verify`, a `gofmt -l .` check that fails on non-empty output and prints the paths, `go vet ./...`, then `golangci/golangci-lint-action`.
- [ ] 4.4 Add the `Test` job as a matrix over `ubuntu-latest`, `macos-latest`, and `windows-latest` with `fail-fast: false`: checkout, setup-go from `go.mod`, `go build ./...`, and `go test ./...` — with `-race` on ubuntu and macOS but not on Windows (no cgo toolchain; see design.md).
- [ ] 4.5 Add the `Cross-compile` job (ubuntu-latest): checkout, setup-go from `go.mod`, then `goreleaser/goreleaser-action` with `args: build --snapshot --clean`, so the platform list comes from `.goreleaser.yml` rather than being restated here.
- [ ] 4.6 Record the exact check names the workflow reports (`Lint`, `Test (ubuntu-latest)`, `Test (macos-latest)`, `Test (windows-latest)`, `Cross-compile`) in a comment in `ci.yml`, so the follow-up branch-protection change can copy them verbatim.

## 5. Release workflow

- [ ] 5.1 Add `.github/workflows/release.yml` named `Release`, triggered only on `push.tags: ['v[0-9]+.[0-9]+.[0-9]+']`, with `permissions: contents: write` and no concurrency cancellation.
- [ ] 5.2 Add the `goreleaser` job (ubuntu-latest): checkout with `fetch-depth: 0` (GoReleaser needs full tag history for version and changelog), setup-go with `go-version-file: go.mod`, then `goreleaser/goreleaser-action` with `args: release --clean` and `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}`.
- [ ] 5.3 Add a comment recording that publishing to the Homebrew tap, AUR, and WinGet is a deliberate follow-up: each needs a cross-repository token minted from the `timvw-ci-bot` App, and `timvw/wt`'s release workflow is the reference for that shape.

## 6. Verification

- [ ] 6.1 Confirm the workflow files parse: run `actionlint` if available, otherwise re-read each file against the GitHub Actions schema (trigger keys, `permissions`, matrix syntax, action versions).
- [ ] 6.2 Confirm the platform sets agree: the pairs built by `goreleaser build --snapshot` are exactly the five in `.goreleaser.yml`, and CI compiles that same set because it reads the same file — no second list exists anywhere.
- [ ] 6.3 Deliberately break formatting in a scratch file, confirm `make fmt-check` fails and names the file, then revert. Do the same for a `go vet` finding if cheap.
- [ ] 6.4 Open the pull request and confirm all five checks run and pass on it. Do not push a version tag as part of this change — cutting the first release is a separate, maintainer-timed decision.
- [ ] 6.5 After the pull request merges, confirm CI is green on `main`. This is the precondition the deferred branch-protection change waits on.
