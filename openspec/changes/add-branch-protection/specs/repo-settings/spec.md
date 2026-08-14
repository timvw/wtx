## ADDED Requirements

### Requirement: The default branch is protected and accepts changes only through pull requests

The configuration SHALL declare branch protection for the default branch in `.github/settings.yml`, so that protection is reviewable in a pull request and is restored automatically if it is altered through the GitHub web UI.

While protection is in force, a commit SHALL reach the default branch only by merging a pull request. A direct push to the default branch SHALL be refused.

#### Scenario: Protection is declared in the configuration file

- **WHEN** `.github/settings.yml` is read
- **THEN** it contains an active `branches:` entry for `main` with a non-empty `protection` mapping
- **AND** no part of the file states that branch protection is deliberately absent

#### Scenario: Protection is applied on merge to the default branch

- **WHEN** a commit changing the `branches:` block lands on the default branch
- **THEN** the Settings GitHub App applies the declared protection to `main`
- **AND** querying the branch's protection afterwards returns the declared values rather than a 404

#### Scenario: Direct push to the default branch is refused

- **WHEN** a contributor pushes a commit directly to `main`
- **THEN** the push is rejected
- **AND** the change can only land by opening a pull request

### Requirement: Merging into the default branch requires every CI check to pass

The declared protection SHALL require, as a condition of merging into the default branch, that every status check reported by the repository's continuous-integration workflow has passed: the lint check, the test check for each supported operating system, and the cross-compile check.

A pull request SHALL NOT be mergeable while any required check is failing, or while any required check has not reported a result.

The protection SHALL NOT require the pull request branch to be up to date with the default branch before merging. A branch whose base has advanced since its checks passed SHALL remain mergeable without re-running them.

#### Scenario: Green pull request is mergeable

- **WHEN** every required check has reported success on a pull request targeting the default branch
- **THEN** the pull request is mergeable

#### Scenario: Failing check blocks the merge

- **WHEN** any required check reports failure on a pull request targeting the default branch
- **THEN** the pull request is not mergeable
- **AND** the merge button identifies the failing check

#### Scenario: Missing check blocks the merge

- **WHEN** a required check has not reported any result on a pull request targeting the default branch
- **THEN** the pull request is not mergeable

#### Scenario: An out-of-date branch is still mergeable

- **WHEN** a pull request's required checks have all passed
- **AND** another pull request merges into the default branch afterwards
- **THEN** the first pull request remains mergeable without rebasing and re-running its checks

### Requirement: Required check names track the CI workflow's job names

The check names listed as required in `.github/settings.yml` SHALL be exactly the names the CI workflow reports, character for character, including the matrix value in the name of each per-operating-system test leg.

Because GitHub matches required checks by name, renaming a workflow job SHALL be accompanied by the same rename in the configuration file. A required name that no job reports blocks every pull request indefinitely, since the check can never report success.

#### Scenario: Declared names match reported names

- **WHEN** the required check names in `.github/settings.yml` are compared with the check names a completed CI run reports
- **THEN** the two sets are identical
- **AND** each operating-system test leg is listed under its own name rather than under a single combined test name

#### Scenario: A rename is applied in both files

- **WHEN** a job in the CI workflow is renamed
- **THEN** the corresponding required check name in `.github/settings.yml` is renamed in the same change

### Requirement: Protection requires no approving review and does not bind administrators

The declared protection SHALL NOT require an approving pull-request review, because the repository has a single maintainer who would otherwise be unable to merge his own pull requests. It SHALL NOT restrict which users, teams, or apps may push to the default branch.

The protection SHALL NOT be enforced against repository administrators, so that a maintainer retains a documented escape hatch when a required check can never report — for example after a workflow job is renamed — without having to edit or delete the protection rule to recover.

#### Scenario: A pull request merges without a review

- **WHEN** the sole maintainer opens a pull request and its required checks pass
- **THEN** the pull request is mergeable with no approving review

#### Scenario: An administrator can merge past a stuck check

- **WHEN** a required check can never report a result
- **THEN** a repository administrator can still merge the pull request
- **AND** recovering does not require changing the protection declared in `.github/settings.yml`

## REMOVED Requirements

### Requirement: Branch protection remains an open extension point

**Reason**: The requirement's premise no longer holds. It withheld branch protection because the repository had no CI workflows and therefore no checks that could be required, and it asked instead for a comment marking where protection would later be declared. `add-ci-release-pipeline` has since added `.github/workflows/ci.yml`, which reports five checks and is green on the default branch. Keeping a requirement that forbids declaring protection would now contradict the requirements this change adds, and the signposting comment it mandates has been consumed — replaced by the live `branches:` block it was pointing at.

**Migration**: No migration is needed for anything outside this repository; the requirement described the deliberate absence of a setting, and nothing depended on that absence. Within the `repo-settings` capability it is superseded by "The default branch is protected and accepts changes only through pull requests", "Merging into the default branch requires every CI check to pass", and "Protection requires no approving review and does not bind administrators". The commented-out `branches:` block that satisfied its "extension point is signposted" scenario is replaced in place by the active block, so the file has no inert protection stanza left behind.
