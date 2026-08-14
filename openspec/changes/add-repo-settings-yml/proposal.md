## Why

Repository configuration for `timvw/wtx` currently lives only in the GitHub web UI: visibility, description, merge behaviour and branch protection are invisible to the repo, unreviewable, and undocumented. The `repository-settings` **Settings** GitHub App is installed on the repo and continuously applies `.github/settings.yml` from the default branch, so we can move that configuration into source control where it is diffable and reviewable. Doing this now — before CI workflows and branch protection exist — establishes the file that a later change will extend with required status checks.

## What Changes

- Add `.github/settings.yml`, read by the `repository-settings` Settings GitHub App, as the single source of truth for repository configuration.
- Declare repository identity: name `wtx`, visibility **public**, description **`wtx`** (replaces the current longer description — an intentional, externally visible change).
- Declare merge and branch hygiene: allowed merge strategies, `delete_branch_on_merge`, and feature toggles (issues / projects / wiki).
- Document — in the file itself and in `design.md` — that the `timvw-ci-bot` GitHub App is already installed and is the intended credential source for automated PR/check workflows. No workflow is added by this change.
- Explicitly **out of scope**: branch protection rules, required status checks, and `topics:`. There is no CI yet, so there are no checks to require; a follow-up change adds the `branches:` block once workflows exist. That follow-up must land protection **after** the first CI workflow is green, otherwise the pull request that introduces the protection cannot satisfy it itself.

## Capabilities

### New Capabilities
- `repo-settings`: Declarative, source-controlled configuration of the GitHub repository (identity, visibility, merge behaviour, feature toggles) applied automatically by the Settings GitHub App, plus the documented extension point for branch protection and required checks.

### Modified Capabilities

_None — no existing capability's requirements change._

## Impact

- **New file**: `.github/settings.yml`.
- **External state**: on merge to `main`, the Settings App rewrites live GitHub repository settings. The repo description changes from "Tmux workspaces on top of wt-managed git worktrees, with per-pane code-assistant layouts" to "wtx". Visibility stays `public` (already public — the file codifies it rather than changing it).
- **Prerequisite already satisfied**: the Settings App and `timvw-ci-bot` App are both installed on `timvw/wtx`.
- **No code impact**: the repository currently contains no application source; this change touches only repository metadata.
- **Risk**: the Settings App applies the file on push to the default branch and will overwrite manual UI edits. Any setting not declared in the file is left untouched by the App, so partial declaration is safe.
