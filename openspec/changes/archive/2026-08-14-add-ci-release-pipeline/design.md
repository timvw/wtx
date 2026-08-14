## Context

See `proposal.md` — Why. What shapes the approach here:

- The repository is a single-module Go project (`github.com/timvw/wtx`, `go 1.26.0`) with one binary at the root, a `cmd` package, and unit tests in `cmd/cmd_test.go`. There is no `internal/`, no e2e harness, no coverage tooling, and no lint configuration.
- `cmd/version.go` already resolves a version with a documented precedence: an ldflag-stamped `version` variable first, then embedded build metadata, then `dev`. The `Makefile` already stamps it via `-X github.com/timvw/wtx/cmd.version=$(VERSION)`. The release pipeline has to hit that same symbol; nothing in the Go code needs to change.
- `.github/settings.yml` is applied by the Settings GitHub App on every push to `main`, and its `branches:` block is commented out with a note that required checks land once a workflow exists. That note constrains sequencing, not this change's content.
- The `timvw-ci-bot` GitHub App is installed on the repository. This change does not need it — but the sibling project `timvw/wt` does need it for cross-repository publishing, which is where this pipeline is headed.
- `timvw/wt` is a working reference implementation of exactly this pipeline, several releases deep, with comments recording which failure modes forced which configuration. Reusing its shape is cheaper and safer than rediscovering those failures.

## Goals / Non-Goals

**Goals:**

- One place — `.goreleaser.yml` — defines the target platform matrix, so CI's cross-compile check and the release build cannot drift apart.
- Verification that is reproducible locally: a contributor can run what CI runs without reading the workflow file.
- Stable, individually addressable check names, so the follow-up branch-protection change is a pure addition to `.github/settings.yml`.
- A release configuration whose publishing pipes (Homebrew, AUR, WinGet) are additive later, not a rewrite.

**Non-Goals:**

- Reproducible / bit-identical builds. `wt` pins `mod_timestamp` and archive member mtimes because it runs GoReleaser three times per release and the rebuilt archives must hash identically. `wtx` runs it once, so that machinery is not carried over. It becomes necessary at the same moment WinGet publishing does, and is added then.
- Fork pull requests producing signed or uploadable artifacts. CI on a fork verifies; it does not publish.
- Optimising CI wall-clock. Correctness and legibility first; the suite is seconds long.

## Decisions

### GoReleaser for the release, not a hand-rolled matrix build

**Chosen:** `.goreleaser.yml` + `goreleaser/goreleaser-action`, run by a tag-triggered workflow.

**Alternative — a build matrix plus `gh release upload`:** fewer moving parts and no third-party action, and for "archives only" it is genuinely comparable in size. Rejected on trajectory: the stated destination is a Homebrew tap, AUR, and WinGet. Each of those is a declarative block in GoReleaser and a bespoke script otherwise, and `wt` already demonstrates the GoReleaser versions working. Paying the dependency now avoids migrating a hand-rolled pipeline later.

**Alternative — copy `wt`'s `.goreleaser.yml` wholesale:** rejected. It carries `nfpms`, `brews`, `scoops`, `winget`, `aurs`, the `WT_SKIP_RELEASE_UPLOAD` disable expression, and the reproducibility pins — all of which exist to solve problems `wtx` does not have yet, and each of which would need a secret or an external repository to work at all. Configuration that is present but non-functional is worse than absent.

### Two workflows, not one

`ci.yml` (pull requests, pushes to `main`, manual dispatch) and `release.yml` (version tags only) stay separate files. Their triggers, permissions, and failure semantics differ: CI is `contents: read` and runs constantly; release is `contents: write`, runs rarely, and has irreversible external effects. Merging them into one workflow guarded by `if:` expressions would put write permission on the token for every pull request run, which is the wrong default.

### The tag pattern is `v[0-9]+.[0-9]+.[0-9]+`, matched in the trigger

Not `v*`. A glob would fire the release pipeline on `v-experiment` or a moved tag. The release spec makes "no other event publishes anything" normative; encoding that in the `on.push.tags` filter enforces it at the platform level rather than in a job condition, so a malformed tag produces no run at all rather than a skipped one.

Pre-release tags (`v1.2.3-rc.1`) do not match this pattern and therefore publish nothing. That is a deliberate omission, not an oversight: widening the pattern later is a one-line change, and GoReleaser's `prerelease: auto` already knows how to mark such a release once the pattern admits it.

### CI's cross-compile check runs `goreleaser build --snapshot`, not a hand-written loop

The `ci` spec requires the platform set CI compiles to be the same set the release publishes. A `for GOOS/GOARCH` loop in `ci.yml` would restate that list in a second place, and the two would drift the first time a platform is added. Running GoReleaser in build-only snapshot mode reads the list from `.goreleaser.yml` — the same file the release run uses — so the requirement holds by construction rather than by vigilance.

This also smoke-tests the release configuration itself on every pull request: a `.goreleaser.yml` syntax error or bad ldflag template surfaces on the PR that introduced it, not at tag time when the fix has to race a published tag.

### Go toolchain version comes from `go.mod`, not a literal in the workflow

`actions/setup-go` with `go-version-file: go.mod`. `wt` hard-codes `'1.26'` in five places; bumping the module's Go directive there means remembering to bump the workflows too, and the failure mode when you forget is a confusing compile error rather than an obvious one. Reading the version from `go.mod` makes the module the single source of truth.

### Lint configuration mirrors `wt`'s `.golangci.yml`

golangci-lint v2 schema, `default: none` with `govet`, `errcheck`, `staticcheck`, `unused`, `ineffassign` enabled, `errcheck` relaxed inside `_test.go`, and the `gofmt` formatter enabled. An explicit allowlist rather than golangci-lint's defaults: the enabled set then survives a linter upgrade unchanged, instead of a new default check turning an unrelated dependency bump into a red pull request.

`go vet` runs as its own step even though `govet` is enabled in golangci-lint. It is redundant by design — the `ci` spec requires both, `go vet` is what a contributor reaches for without installing anything, and it costs a second.

### `gofmt` is checked, not applied

The formatting step fails on non-empty `gofmt -l .` output and prints the offending paths. It never writes. A CI job that reformats and commits would need write permission on the token and would rewrite a contributor's branch under them; `make fmt` already exists for fixing it locally.

### Job names are treated as an interface

`Lint`, `Test (ubuntu-latest)` / `(macos-latest)` / `(windows-latest)`, and `Cross-compile` are the names the follow-up change will list in `required_status_checks.contexts`. Renaming a job silently drops it from branch protection — GitHub matches required checks by name, and a rule requiring a name nothing reports blocks every pull request forever. The workflow file gets a comment saying so.

Because the test job is a matrix, its per-leg check names include the matrix value; branch protection must therefore list each leg, not the job. That is a consequence worth writing down before the follow-up change trips over it.

### `-race` on Linux only

The race detector needs a cgo toolchain; enabling it on the Windows runner means installing one for no benefit, since the point of the Windows leg is proving the suite passes on a Windows host, not detecting races a second time. macOS runs it — its cgo toolchain is present by default. Same reasoning `wt` records.

### Release runs on the default `GITHUB_TOKEN` with `contents: write`

No App token, no repository secret. Everything this change publishes lives in this repository. The App token becomes necessary at the first cross-repository publish (the Homebrew tap), and `wt`'s release workflow shows the shape: mint a scoped installation token with `actions/create-github-app-token`, listing each destination repository. Introducing it now would add an unused secret dependency and a fork-PR edge case for nothing.

### `fetch-depth: 0` on the release checkout

GoReleaser derives the version and the changelog from tag history; a shallow clone gives it neither. CI keeps the default shallow checkout.

### Re-run safety via `replace_existing_artifacts`

The release spec requires re-running a tag to converge rather than fail on an already-uploaded asset. GoReleaser's `release.replace_existing_artifacts: true` deletes and re-uploads instead of erroring on the 422. Cheaper and more predictable than deleting the release by hand after a transient failure.

### Concurrency: cancel superseded CI runs, never cancel a release

`ci.yml` gets a concurrency group keyed on the ref with `cancel-in-progress: true`, so pushing twice to a pull request does not leave a stale run reporting. `release.yml` gets no cancellation: a half-cancelled publish is exactly the partial state the spec forbids.

### `LICENSE` and `README.md` are part of this change, not a prerequisite for it

GoReleaser's default archive `files` globs pick up `LICENSE*` and `README*`, and the release spec requires both inside every archive. MIT matches `wt` and the rest of the author's projects. The README stays minimal — what `wtx` is, how to install a release, how to build from source — because the CLI itself is still a skeleton and documenting commands that do not exist yet ages badly.

## Risks / Trade-offs

- **A required check name is renamed later and branch protection silently stops enforcing it, or blocks every PR.** → Job names are documented as an interface in a comment in `ci.yml`; the follow-up protection change lists them explicitly, and any rename must update both.
- **The first tag push is the first time the release pipeline runs end to end, and it runs against a public repository.** → The CI cross-compile step exercises the same `.goreleaser.yml` on every pull request, so configuration errors surface before the tag. Before the real `v0.1.0`, run `goreleaser release --snapshot --clean` locally to see the artifacts without publishing.
- **`goreleaser/goreleaser-action` and the other actions are pinned by major version tag, which is a mutable ref.** → Accepted: it is the same posture as `wt` and the reason releases keep working across minor action updates. Immutable SHA pinning is the stricter alternative and needs Renovate (which `wt` has and `wtx` does not) to stay current; without automation, SHA pins rot into an unmaintained pipeline.
- **The Windows and macOS legs bill at 2× and 10× Linux minutes.** → The repository is public, so GitHub-hosted minutes are free. If it ever goes private this is the first thing to trim.
- **Adding a fifth check to a repository with no branch protection changes nothing on its own — a red pull request is still mergeable.** → Expected. Enforcement is the follow-up change's job, and it is deliberately sequenced after CI proves green on `main`.
- **`goreleaser build --snapshot` in CI is slower than a bare `go build` loop** (it builds five binaries with full ldflags). → Seconds, on a project this size, in exchange for eliminating a duplicated platform list.

## Migration Plan

1. Land CI first (workflow, `.golangci.yml`, `Makefile` targets), and confirm it is green on `main`. Nothing downstream is safe to configure until a run has actually succeeded.
2. Land `LICENSE`, `README.md`, `.goreleaser.yml`, and `release.yml`. Verify locally with `goreleaser release --snapshot --clean`, which builds and archives into `dist/` without touching GitHub.
3. Push the first version tag when the maintainer chooses to cut a release. Nothing about this change forces a release to happen.

**Rollback:** delete the workflow files. There is no state to unwind — CI writes nothing, and the release pipeline is inert until a version tag is pushed. If a bad release is published, delete the GitHub Release and its tag; because no package manager consumes `wtx` yet, nothing downstream has pinned it.

## Open Questions

- Which version the first tag should be (`v0.1.0` vs `v0.0.1`) — a maintainer call at release time. It changes no spec, no configuration, and no task here.
