## Context

See `proposal.md` — Why. What constrains the shape of this change:

- `.github/settings.yml` is applied by the `repository-settings` Settings App on every push to `main`. Its `branches:` plugin passes the `protection` mapping through, essentially unmodified, to the REST call `PUT /repos/{owner}/{repo}/branches/{branch}/protection`. There is no intermediate schema of the App's own: whatever that endpoint accepts is what the file may contain, and whatever it rejects fails inside the App rather than on the pull request.
- That endpoint has four **required** top-level members — `required_status_checks`, `enforce_admins`, `required_pull_request_reviews`, `restrictions` — each of them nullable. The App's own documentation repeats the warning: omit one and *none* of the settings are applied. Disabling a feature therefore means writing `null`, not leaving the key out.
- `.github/workflows/ci.yml` reports exactly five checks, recorded verbatim in a comment at the top of that file, which also states that job names are a public interface and must be renamed in both places or not at all. This change is the second place.
- CI is green on `main` (run `31802481738`), and a live query of that run's job names returns the same five strings the comment records. The precondition the two earlier changes waited on is therefore satisfied by observation, not by assumption.
- The repository has one active contributor. That single fact drives most of the profile below.

## Goals / Non-Goals

**Goals:**

- Make a red pull request unmergeable, which is the entire point of having built CI.
- Keep the maintainer able to merge his own work without a second person in the loop.
- Keep the mechanism singular: `settings.yml` is the only way protection is configured, so protection stays diffable and a UI edit is reverted rather than persisting invisibly.
- Leave a recovery path that does not require editing the protection rule, because the most likely failure here is self-inflicted and would otherwise be self-locking.

**Non-Goals:**

- Modelling a review workflow for contributors who do not exist yet. When a second maintainer appears, `required_pull_request_reviews` stops being `null`; that is a three-line edit and a different change.
- Migrating to GitHub rulesets. See below.
- Adding `topics:`, which the same comment block deferred. It is unrelated to enforcement and adding it here would make the diff about two things.

## Decisions

### Protection lives in `settings.yml`, not in the API or the UI

The alternative — `gh api ... /branches/main/protection` once, by hand — takes one command and leaves no record. It was rejected on the same grounds the `repo-settings` capability exists at all: a setting that is not in the file is a setting nobody can review, nobody can diff, and the Settings App will happily leave in whatever state a stray UI click put it. Worse, once the file declares `branches:` at all, the App reconciles it; a hand-applied rule that disagrees with the file is overwritten on the next push, so the two mechanisms cannot coexist. One mechanism, and it is the file.

### Classic branch protection, not rulesets

GitHub's rulesets are the newer mechanism and are strictly more expressive — bypass actors, layered rules, org-level inheritance. The Settings App's supported, documented surface is `branches:`, which maps to the classic protection endpoint; it has no ruleset plugin. Choosing rulesets means choosing a different tool to apply them (Terraform, or a workflow calling the API), which means reintroducing credentials and state for a repository that needs neither.

`add-repo-settings-yml` already recorded this decision and named the trigger for revisiting it: needing a ruleset-only feature, such as a merge queue. Nothing in this profile needs one. The decision is restated here only because this is the change that first exercises it.

### `contexts:`, not `checks:`

The REST endpoint accepts both. `checks:` is the newer form and pins each required check to an owning app id, which prevents a different app from satisfying a check by reporting the same name; `contexts:` is a plain list of names and carries a closing-down notice in GitHub's documentation. `contexts:` is nevertheless what the Settings App documents, what its examples use, and what is still accepted today. Using the form the App documents keeps this file readable against the App's own docs, and the spoofing risk `checks:` mitigates requires an attacker who can already install an app on the repository. If the App's docs move to `checks:`, so should this file.

### `strict: false` — do not require the branch to be up to date

`strict: true` makes GitHub refuse to merge a pull request whose base has moved since its checks ran, forcing a rebase and a full re-run of all five checks. Its value is real: it prevents the semantic conflict where two branches each pass in isolation and break once combined.

Rejected anyway, on frequency. With one contributor, near-simultaneous merges are rare, and the cost is paid on every merge that follows any other merge: five checks re-run, including a Windows leg and a full cross-compile, before a button becomes clickable again. That converts branch protection from a safety net into a queue. The failure mode `strict: true` guards against — a broken `main` — is caught within minutes by the push-triggered CI run that `ci.yml` already declares, and is fixed forward. If concurrent work becomes normal, flipping this to `true` is a one-word edit; the honest sequencing is to flip it when the pain appears, not before.

### `required_pull_request_reviews: null` — no approving review

Tim is the sole maintainer. GitHub does not let an author approve their own pull request, so `required_approving_review_count: 1` would make every pull request permanently unmergeable except by an admin override. That is not a stricter policy; it is a policy that converts the admin escape hatch from a rarely-used recovery mechanism into the normal merge path, which is worse than not requiring review at all — it trains the maintainer to click through the warning.

The rejected alternative worth naming is `required_approving_review_count: 0` with `dismiss_stale_reviews`, which some repositories use to get the *pull-request* requirement without the *approval* requirement. It is unnecessary here: declaring `protection` at all already forces changes through pull requests, which is the property actually wanted.

### `enforce_admins: false` — a deliberate escape hatch

This is the decision most likely to read as a weakening, so the reasoning is worth stating plainly. The realistic failure mode for this change is a required check name that no job reports — after a job rename, or a workflow refactor — and its symptom is that *every* pull request becomes unmergeable, including the one that would fix the file. With `enforce_admins: true`, recovery means editing or deleting the protection rule through the UI, which the App then fights on the next reconciliation. With `false`, recovery is one merge by the admin who is the only person here anyway.

It buys back very little risk. The threat model for a single-maintainer public repository is the maintainer's own mistake, and an override is a conscious, logged click past a visible warning — not something that happens by accident. `true` is the right value the day a second maintainer joins and the rule starts constraining someone other than its author.

### `restrictions: null` — do not restrict who may push

Push restrictions on classic protection are an organisation-only feature; on a user-owned repository the API rejects a non-null `restrictions` object. `null` is the only value that works, and it is also the value that means what is wanted: nobody may push directly to `main`, which the protection itself already enforces, and there is no allowlist to maintain.

### All five contexts, listed individually

The test job is a matrix, so it reports `Test (ubuntu-latest)`, `Test (macos-latest)`, and `Test (windows-latest)` as three separate checks. There is no `Test` context to require — requiring that name would require a check that never reports, which is the exact lockout described above. `design.md` in `add-ci-release-pipeline` predicted this consequence; listing the three legs explicitly is the whole of acting on it.

The names were taken from the comment in `ci.yml` and then confirmed against what a real run reported, rather than from the comment alone. The comment is maintained by hand and could have drifted from the workflow beneath it; the run cannot.

### The settings edit ships in the same pull request as the planning artifacts

The repository's earlier changes split proposal and implementation across two pull requests (`#7` then `#8`, `#2` then `#3`). This one is a single edit to a single file, and the artifacts are almost entirely *about* that edit — the values, and why each was chosen. Splitting would put the rationale in one pull request and the eleven lines it justifies in another, and a reviewer would need both open at once. One pull request, reviewed as a unit.

## Risks / Trade-offs

- **A job rename in `ci.yml` leaves a required name nothing reports, blocking every pull request.** → The likeliest failure, and the reason `enforce_admins: false` is not negotiable in this profile: an admin merges the fix rather than dismantling the rule. `ci.yml` already carries a comment stating that job names are an interface; the spec now makes "rename in both places" normative rather than advisory.
- **`enforce_admins: false` means the protection is advisory for the only person who can push.** → Accepted, and stated in the spec rather than hidden in the file. The rule still blocks the accident it exists to block — merging a red pull request without noticing — because that requires overriding a visible warning. It stops being the right value the moment a second maintainer exists.
- **`strict: false` allows a merge whose checks ran against an older base, so `main` can break from a combination that never ran together.** → Accepted for a single-contributor repository. Push-triggered CI on `main` reports the break within minutes, and the fix is forward. Flipping to `strict: true` is a one-word edit if concurrent work becomes normal.
- **A malformed `branches:` block fails inside the Settings App, which reports to its own delivery log and not as a check on the pull request.** → The same silent-failure mode `add-repo-settings-yml` recorded, now with a worse consequence: believing `main` is protected when it is not. Mitigated by verifying with `gh api /repos/timvw/wtx/branches/main/protection` after merge instead of assuming, and by the App's "all four keys or nothing" rule being handled explicitly — every one of the four is present, three of them `null`.
- **Omitting one of the four required top-level keys silently applies none of the others.** → Called out here because it is the least intuitive property of this API and the easiest thing to "tidy up" in a later edit. The three `null`s are load-bearing and are commented as such in the file.
- **`contexts:` carries a closing-down notice.** → It is what the Settings App documents and it is still accepted. The migration to `checks:` is mechanical when it comes, and the trigger is the App's docs changing, not GitHub's deprecation notice.
- **This pull request is not itself subject to the rule it introduces, so the profile's first real exercise is the next pull request.** → Inherent to bootstrapping protection, and the reason both earlier changes deferred it to here. The mitigation is verification after merge (task group 4), not before.

## Migration Plan

1. Edit `.github/settings.yml` on a feature branch, replacing the commented-out block and rewriting the comment above it so the file no longer claims protection is deliberately absent. Open a pull request.
2. Confirm the file parses as YAML and that the five contexts match a real CI run's reported job names character for character.
3. Merge to `main`. The Settings App reconciles on the push — no manual trigger.
4. Verify with `gh api /repos/timvw/wtx/branches/main/protection`, which returned 404 before this change. Confirm the five contexts, `strict: false`, `enforce_admins.enabled: false`, and that `required_pull_request_reviews` is absent from the response.
5. Confirm on the next pull request that the five checks appear as required and that the merge button is gated on them.

**Rollback:** revert the commit and push to `main`. The App's `branches` plugin treats an empty or null `protection` mapping as a request to *delete* the rule, but a reverted file has no `branches:` key at all — and the App only touches keys the file declares, so removing the key leaves the applied protection in place rather than removing it. Rolling back therefore takes two steps: revert the file, then delete the rule explicitly, either with `gh api -X DELETE /repos/timvw/wtx/branches/main/protection` or by committing `branches: [{name: main, protection: null}]` and letting the App delete it. This asymmetry — declaring protection is one step, undeclaring it is two — is worth knowing before it is needed.
