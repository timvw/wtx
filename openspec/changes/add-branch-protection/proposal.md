## Why

`main` is unprotected: anyone with push access — in practice, the sole maintainer — can push straight to it, and a pull request whose CI is red is still mergeable. The `add-repo-settings-yml` change deliberately left the `branches:` block in `.github/settings.yml` commented out because there were no workflows and therefore no checks to require, and `add-ci-release-pipeline` deliberately stopped short of enforcement for the same reason: a pull request cannot satisfy a required check that the pull request itself introduces. CI now reports five checks and is green on `main`, so the precondition both changes waited on is met. This is that deferred follow-up.

## What Changes

- Replace the commented-out `branches:` block in `.github/settings.yml` with a live one protecting `main`, and rewrite the surrounding comment, which currently asserts that protection is deliberately absent and would otherwise be actively misleading.
- Require all five checks `.github/workflows/ci.yml` reports — `Lint`, `Test (ubuntu-latest)`, `Test (macos-latest)`, `Test (windows-latest)`, `Cross-compile` — to pass before a pull request can merge into `main`. Direct pushes to `main` are refused.
- Set `strict: false`: a branch does not have to be rebased onto the tip of `main` before it merges.
- Set `required_pull_request_reviews: null` and `restrictions: null`, and `enforce_admins: false`. The repository has one maintainer; requiring an approving review would leave him unable to merge his own pull requests, and the admin escape hatch is what makes a stuck required check recoverable without editing protection by hand.
- Explicitly **out of scope**: `topics:`, deferred alongside branch protection by the same comment in `settings.yml` and still deferred — it adds no reviewable behaviour and belongs with a change that has something to say about how the repository is discovered. Also out of scope: migrating to GitHub rulesets, requiring signed commits, requiring linear history, and requiring conversation resolution.

## Capabilities

### New Capabilities

_None._ Branch protection is part of the existing `repo-settings` capability — the same file, applied by the same App — not a new one.

### Modified Capabilities

- `repo-settings`: The requirement "Branch protection remains an open extension point" is **removed**; its premise (no CI exists, so nothing can be required) no longer holds. Three requirements replace it: merges into the default branch require the CI checks to pass, the required check names track the workflow's job names, and the protection profile deliberately requires no review and does not bind administrators.

## Impact

- **Modified file**: `.github/settings.yml` only. No workflow, no source, no build configuration changes.
- **External state**: on merge to `main`, the Settings App calls the branch-protection API and `main` becomes protected. From that moment a red or missing check blocks merging, and `git push` to `main` is rejected for everyone including the maintainer.
- **Sequencing**: this pull request is itself unprotected — protection does not exist until it merges — so it cannot be blocked by the rule it introduces. The *next* pull request is the first one the rule applies to.
- **New failure mode**: renaming a job in `ci.yml` without updating `settings.yml` leaves a required check that nothing ever reports, which blocks every pull request indefinitely. `ci.yml` already carries a comment saying so; this change makes the consequence real rather than hypothetical.
- **Recovery**: `enforce_admins: false` means the maintainer can merge past a stuck check from the GitHub UI without touching the protection rule, so a bad required-check name is an inconvenience rather than a lockout.
- **No code impact**: nothing in the Go module, the Makefile, or the release pipeline is affected.
