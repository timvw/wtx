## 1. Author the configuration file

- [x] 1.1 Create `.github/` and add `.github/settings.yml` with a header comment stating that this repository is managed by the `repository-settings` Settings GitHub App and that manual edits in the GitHub UI will be reverted
- [x] 1.2 Declare repository identity under `repository:` — `name: wtx`, `private: false`, `description: wtx`
- [x] 1.3 Declare merge behaviour — `allow_squash_merge: true`, `allow_merge_commit: false`, `allow_rebase_merge: false`, `delete_branch_on_merge: true`
- [x] 1.4 Declare feature toggles explicitly — `has_issues: true`, `has_projects: false`, `has_wiki: false`, `has_downloads: false`
- [x] 1.5 Add a comment naming `timvw-ci-bot` as the installed App that automated PR/check workflows use to mint short-lived installation tokens (no PAT)
- [x] 1.6 Add a commented-out `branches:` block marking where branch protection and required status checks will be declared by the follow-up change — leave it inert

## 2. Validate before merging

- [x] 2.1 Confirm the file parses as valid YAML and has a top-level `repository:` mapping
- [x] 2.2 Re-read the file against `specs/repo-settings/spec.md` and confirm every ADDED requirement is covered, and that no active `branches:` rules are present
- [ ] 2.3 Record the current live description verbatim in the PR body so it can be restored by hand if the rename is reverted
- [ ] 2.4 Open a PR with the change; do not push directly to `main`

## 3. Apply and verify

- [ ] 3.1 Merge the PR to `main` and let the Settings App reconcile (no manual trigger needed)
- [ ] 3.2 Verify applied state: `gh repo view timvw/wtx --json description,visibility,deleteBranchOnMerge,mergeCommitAllowed,squashMergeAllowed,rebaseMergeAllowed,hasIssuesEnabled,hasProjectsEnabled,hasWikiEnabled`
- [ ] 3.3 Confirm `description` is exactly `wtx` and `visibility` is `PUBLIC`
- [ ] 3.4 Confirm `main` still has no branch protection: `gh api /repos/timvw/wtx/branches/main/protection` returns 404
- [ ] 3.5 If any declared value did not apply, inspect the Settings App delivery log (repo Settings → GitHub Apps → Settings) before changing anything through the UI

## 4. Hand off to the follow-up

- [x] 4.1 Note in the change record that branch protection, required status checks, and `topics:` are deferred, and that the follow-up must sequence protection after the first CI workflow is green so the settings PR itself can still merge
