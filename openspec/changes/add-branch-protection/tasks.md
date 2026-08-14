## 1. Confirm the precondition

- [x] 1.1 Confirm CI is green on `main`: `gh run list --branch main --limit 5`. Protection must not be declared until a workflow has actually succeeded on the default branch, otherwise the required checks name something that has never passed.
- [x] 1.2 Read the check names off a real run rather than trusting the comment in `ci.yml`: `gh run view <run-id> --json jobs --jq '.jobs[].name'`. Confirm the five reported names are exactly `Lint`, `Test (ubuntu-latest)`, `Test (macos-latest)`, `Test (windows-latest)`, `Cross-compile`, and that they match the comment at the top of `ci.yml` character for character.
- [x] 1.3 Confirm `main` is currently unprotected: `gh api /repos/timvw/wtx/branches/main/protection` returns 404. This is the before-state the post-merge verification is compared against.

## 2. Verify the schema before writing it

- [x] 2.1 Check the shape the Settings App accepts against its `branches` plugin documentation, not from memory: the App forwards the `protection` mapping to `PUT /repos/{owner}/{repo}/branches/{branch}/protection`, so that endpoint's schema governs.
- [x] 2.2 Confirm the four top-level members `required_status_checks`, `enforce_admins`, `required_pull_request_reviews`, and `restrictions` are all required and all nullable, and that omitting any one of them causes the App to apply none of the settings. Every one of the four must appear in the file, with `null` where the feature is disabled.
- [x] 2.3 Confirm `enforce_admins: false` is accepted (the field is boolean-or-null; `false` and `null` both disable enforcement) and that `restrictions` must be `null` on a user-owned repository, where push restrictions are an organisation-only feature.

## 3. Declare the protection

- [x] 3.1 Rewrite the comment block above `branches:` in `.github/settings.yml`. It currently states that protection is deliberately not declared because no CI workflows exist; that is now false and must not survive the change. Replace it with a comment recording the profile actually chosen and why.
- [x] 3.2 Replace the commented-out `branches:` block with an active one: `- name: main` with a `protection` mapping. Do not leave the inert block behind alongside the live one.
- [x] 3.3 Declare `required_status_checks` with `strict: false` and `contexts:` listing all five check names. The test job is a matrix, so each operating-system leg is listed under its own name — there is no combined `Test` context to require.
- [x] 3.4 Declare `required_pull_request_reviews: null`, `enforce_admins: false`, and `restrictions: null`, each with a short comment saying why. The `null`s are load-bearing, not placeholders: a later edit that "tidies" them away disables the whole rule.
- [x] 3.5 Decide the `topics:` deferral. It was deferred in the same comment block and is not in this change's scope; keep that note alive in the rewritten comment rather than dropping it along with the branch-protection text it was attached to.
- [x] 3.6 Add a comment cross-referencing `.github/workflows/ci.yml` — the list of contexts is one half of an interface whose other half is the job names in that file.

## 4. Verify before committing

- [x] 4.1 Confirm the file still parses: `python3 -c "import yaml; yaml.safe_load(open('.github/settings.yml'))"`.
- [x] 4.2 Diff the five declared contexts against the names captured in task 1.2 — exact string comparison, not by eye. A trailing space or a differing dash makes a required check that can never report.
- [x] 4.3 Confirm the `repository:` block is untouched: this change adds a top-level key and must not perturb identity, merge behaviour, or feature toggles.
- [x] 4.4 Run `openspec validate add-branch-protection --strict`.
- [x] 4.5 Open a pull request. Do not push to `main`, and do not apply protection through the GitHub API or UI — the file is the only mechanism.

## 5. Apply and verify after merge

- [ ] 5.1 Merge the pull request and let the Settings App reconcile on the push to `main`.
- [ ] 5.2 Confirm the rule was applied: `gh api /repos/timvw/wtx/branches/main/protection` now returns 200 rather than 404.
- [ ] 5.3 Confirm the applied values: `required_status_checks.contexts` holds the five names, `required_status_checks.strict` is `false`, `enforce_admins.enabled` is `false`, and `required_pull_request_reviews` is absent from the response.
- [ ] 5.4 If nothing was applied, read the Settings App delivery log (repo Settings → GitHub Apps → Settings) before changing anything by hand. The App reports malformed configuration to its own log, not as a check on the pull request — a silent no-op looks identical to success from the repository's side.
- [ ] 5.5 Confirm on the next pull request that all five checks appear under "Required" and that the merge button is blocked until they report.
- [ ] 5.6 Confirm a direct push to `main` is now refused.
