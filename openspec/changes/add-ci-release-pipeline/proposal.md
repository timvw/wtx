## Why

`wtx` has a buildable CLI skeleton, a `Makefile`, and unit tests, but nothing runs them: every pull request is verified only by whatever the author happened to run locally, and there is no way to hand someone a built `wtx` binary. The repository also cannot adopt branch protection, because `.github/settings.yml` deliberately left `required_status_checks` empty until a workflow exists to require. Adding continuous integration and a tag-triggered release closes both gaps and establishes the release plumbing that later package-manager publishing (Homebrew tap, AUR, WinGet) plugs into.

## What Changes

- Add a **CI workflow** (`.github/workflows/ci.yml`) that runs on pull requests targeting `main`, on pushes to `main`, and on manual dispatch. It performs:
  - formatting check (`gofmt -l .` must report nothing) and `go vet`
  - `golangci-lint`, configured by a new `.golangci.yml`
  - `go build` and `go test` on **ubuntu, macOS and Windows** runners
  - a **cross-compile smoke build** for every platform the release targets, so a release-time cross-compile break is caught on the pull request rather than at tag time
- Add a **release workflow** (`.github/workflows/release.yml`) triggered by pushing a `vMAJOR.MINOR.PATCH` tag, which runs **GoReleaser** to build all target platforms and publish a GitHub Release.
- Add a **`.goreleaser.yml`** producing, for this change, **GitHub release archives only**: `tar.gz` for Linux/macOS, `zip` for Windows, plus a `checksums.txt`. Binaries are built with `CGO_ENABLED=0` and `-trimpath`, with the version stamped into `github.com/timvw/wtx/cmd.version`.
- Add **`LICENSE`** (MIT, Tim Van Wassenhove) and a minimal **`README.md`**, both shipped inside every release archive. The repository currently has neither, and a release without a license is not redistributable.
- Explicitly **out of scope**: publishing to `timvw/homebrew-tap`, AUR, WinGet, Scoop, or Linux packages (deb/rpm); Docker images; code coverage upload; end-to-end tests; and the `branches:` protection block in `.github/settings.yml`. Branch protection is deferred to a follow-up change so that it lands only after CI is green on `main` — a pull request that introduces its own required check cannot satisfy it.

## Capabilities

### New Capabilities

- `ci`: Automated verification of every pull request and every push to the default branch — formatting, vetting, linting, unit tests across the supported operating systems, and a cross-compile check covering the platforms the release targets.
- `release`: Tag-triggered production of downloadable, checksummed release artifacts for all supported platforms, published as a GitHub Release, with the release version stamped into the binary and the license shipped alongside it.

### Modified Capabilities

_None._ The `cli` spec already requires that a binary built with a version stamped in reports that stamped version; the release pipeline satisfies that existing requirement rather than changing it. `repo-settings` is untouched — its branch-protection extension point stays open for the follow-up change.

## Impact

- **New files**: `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `.goreleaser.yml`, `.golangci.yml`, `LICENSE`, `README.md`.
- **Modified files**: `.gitignore` gains GoReleaser's `dist/` output directory. `Makefile` may gain a `lint`/`fmt-check` target so contributors can run locally what CI runs.
- **External state**: pushing a `v*` tag creates a public GitHub Release with attached assets. This is irreversible in practice — a published tag and its assets are consumed by downstream tooling — so the tag format is constrained and the workflow triggers on nothing else.
- **Licensing**: adding `LICENSE` makes the project's terms explicit (MIT) for the first time. This is an intentional, externally visible decision.
- **Credentials**: the release workflow needs only `contents: write` on the default `GITHUB_TOKEN`. No new repository secrets are introduced by this change; cross-repository tokens (for the Homebrew tap and friends) arrive with the follow-up publishing change.
- **Cost**: the CI matrix uses macOS and Windows runners, which bill at a multiple of Linux minutes. The repository is public, so GitHub-hosted minutes are free.
- **Unblocks**: the deferred `.github/settings.yml` branch-protection change, which can now name concrete required check jobs.
