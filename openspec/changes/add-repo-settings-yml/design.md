## Context

See `proposal.md` — Why. Design-relevant state of `timvw/wtx` at the time of writing:

- The repository holds no application source yet — only `openspec/` and `.claude/`. There is no `.github/` directory and no workflows.
- Live GitHub state: visibility `PUBLIC`; description "Tmux workspaces on top of wt-managed git worktrees, with per-pane code-assistant layouts"; default branch `main` with **no** protection and **no** rulesets; all three merge strategies allowed; `delete_branch_on_merge` false; issues and projects on, wiki off; no topics; no license.
- Both the `repository-settings` **Settings** App and `timvw-ci-bot` are already installed on the repository. Neither needs to be installed by this change.

The Settings App reconciles on push to the default branch (and periodically). It only touches keys the file declares.

## Goals / Non-Goals

**Goals:**

- One tracked file, reviewable in a PR, that fully determines the repository settings it declares.
- A file shaped so that adding branch protection and required status checks later is an additive edit, not a restructure.
- Record the `timvw-ci-bot` credential contract where a contributor will find it.

**Non-Goals:**

- Managing organisation- or user-level settings — the file is repository-scoped only.
- Managing collaborators, teams, or labels. The repo is a single-owner project; declaring `labels:` in the Settings App is destructive (it deletes labels not listed), so it is deliberately left out.
- Adding any CI workflow. Required checks cannot be declared before the checks exist.

## Decisions

**Use the Settings App (`.github/settings.yml`) rather than Terraform or a `gh api` script.**
The App is already installed and reconciles continuously with zero local state — no provider credentials, no state file, no runner. Terraform's `github_repository` resource is more expressive (it can manage rulesets), but it needs a token, a state backend, and a workflow to run it — disproportionate for one repository. A `gh api` script in a workflow would need the very credentials this change is only documenting. Revisit if repository count grows or if rulesets become necessary.

**Declare branch protection under `branches:`, not `rulesets:`.**
The Settings App's stable, documented surface is the `branches:` key mapping to the classic branch-protection API. Rulesets are the newer GitHub mechanism and are better supported by Terraform than by the App. Since this change declares no protection at all, the decision only fixes the shape of the follow-up: the extension point is a commented `branches:` block. If the follow-up needs ruleset-only features (e.g. merge queues), that is the trigger to reconsider the tool, not to fight the App.

**Squash-only merging, with automatic head-branch deletion.**
The repository has no history worth preserving and a single primary author; squash-only keeps `main` linear and makes "one PR = one commit" true, which in turn makes required-check history easy to read once CI exists. This *changes* current behaviour (all three strategies are allowed today) — it is an assumption made in the absence of a stated preference, and it is a one-line edit if wrong.

**Declare feature toggles explicitly, even where the value matches today's state.**
The point of the file is that reading it tells you the configuration. A toggle left undeclared is a toggle someone must check the UI for. Cost is a few lines; benefit is that the file is complete for its declared surface.

**Set the description to the literal string `wtx`, as requested.**
This discards an informative description on a public repository, where the description is the primary signal in search results and on the profile page. Recorded here because it is user-visible and was explicitly confirmed; the value is a one-line edit in the file if it is later reconsidered.

**Document `timvw-ci-bot` in a comment rather than wiring it up.**
The App is installed but nothing consumes it yet. A comment in `settings.yml` naming the App as the credential source for automated PR/check workflows costs nothing and is where someone adding the first workflow will already be looking. The actual token exchange (App ID + private key → installation token, e.g. via `actions/create-github-app-token`) belongs to the change that adds the first workflow.

**Do not declare `topics:` in this change.**
It was floated in scoping but adds no reviewable behaviour and invites bikeshedding. It can be added in the same follow-up as branch protection.

## Risks / Trade-offs

- **The Settings App silently overwrites manual UI changes** → That is the intent, but it surprises people. Mitigated by a header comment in the file stating that the repository is App-managed and that UI edits will be reverted.
- **A malformed `settings.yml` fails quietly** — the App reports errors in its own logs (repo → Settings → GitHub Apps → Settings → advanced/deliveries), not as a check on the PR → Mitigated by verifying the applied state with `gh repo view` after merge rather than assuming success. Long-term mitigation is a schema-lint workflow, which belongs with the CI change.
- **Squash-only may be wrong for this project** → Single-line revert; called out above as an assumption.
- **Losing the descriptive repo description hurts discoverability** → Explicitly confirmed by the user; noted above so the trade-off is visible if revisited.
- **Chicken-and-egg on branch protection**: once `main` requires status checks, the PR that *changes* `settings.yml` must itself pass them → Deliberately deferred to the follow-up change, which can sequence protection after CI is green.

## Migration Plan

1. Add `.github/settings.yml` on a feature branch; open a PR.
2. Merge to `main`. The Settings App reconciles automatically — no manual trigger.
3. Verify with `gh repo view timvw/wtx --json description,visibility,deleteBranchOnMerge,mergeCommitAllowed,squashMergeAllowed,rebaseMergeAllowed,hasIssuesEnabled,hasProjectsEnabled,hasWikiEnabled`.
4. If the App did not apply, check its delivery log before editing anything by hand.

**Rollback:** revert the commit and push to `main`; the App reconciles back. Settings the file never declared were never touched, so rollback is complete for the declared surface. The one value that does not restore itself is the original description — recorded verbatim in `proposal.md` — Impact, so it can be restored by hand.

## Open Questions

- Which status checks become required, and whether `main` also requires an approving review, is deferred to the follow-up change — it depends on what the first CI workflow actually produces, and it changes neither this file's shape nor the spec.
