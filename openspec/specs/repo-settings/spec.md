# repo-settings Specification

## Purpose

Keeps the GitHub repository's own configuration — visibility, description, merge behaviour, and feature toggles — declared in source control and applied automatically, so repository settings are reviewable in pull requests instead of being changed silently through the GitHub web UI.
## Requirements
### Requirement: Repository configuration is declared in source control

The repository SHALL declare its GitHub configuration in a single tracked file, `.github/settings.yml`, in the format consumed by the `repository-settings` Settings GitHub App. That file SHALL be the source of truth for every setting it declares.

Settings the file does not declare are outside this contract and remain whatever GitHub currently holds.

#### Scenario: Configuration file is present and valid

- **WHEN** the repository is checked out at the default branch
- **THEN** `.github/settings.yml` exists
- **AND** it parses as valid YAML with a top-level `repository:` mapping

#### Scenario: Merged configuration is applied to the live repository

- **WHEN** a commit that changes `.github/settings.yml` lands on the default branch
- **THEN** the Settings GitHub App applies the declared values to the GitHub repository
- **AND** querying the repository's settings afterwards returns the declared values

#### Scenario: Manual UI edits do not persist

- **WHEN** a declared setting is changed by hand in the GitHub web UI
- **AND** the Settings App next reconciles the repository
- **THEN** the value declared in `.github/settings.yml` is restored

### Requirement: Repository identity is public and named wtx

The configuration SHALL declare the repository name as `wtx`, its visibility as public, and its description as the exact string `wtx`.

#### Scenario: Repository is public

- **WHEN** the applied configuration is read back from GitHub
- **THEN** the repository visibility is `PUBLIC`

#### Scenario: Description is the exact string wtx

- **WHEN** the applied configuration is read back from GitHub
- **THEN** the repository description is exactly `wtx`
- **AND** the previous longer description is no longer present

### Requirement: Merge behaviour keeps history linear and branches tidy

The configuration SHALL declare which merge strategies pull requests may use and SHALL declare that head branches are deleted after merge.

#### Scenario: Only squash merging is offered

- **WHEN** a contributor opens the merge dropdown on a pull request
- **THEN** squash merge is available
- **AND** merge commits and rebase merging are not available

#### Scenario: Head branch is removed after merge

- **WHEN** a pull request is merged
- **THEN** its head branch is deleted automatically

### Requirement: Repository features are explicitly toggled

The configuration SHALL declare each optional GitHub repository feature (issues, projects, wiki) as explicitly enabled or disabled, rather than leaving it at whatever GitHub defaulted to.

#### Scenario: Feature toggles match the declaration

- **WHEN** the applied configuration is read back from GitHub
- **THEN** issues, projects, and wiki are each in the state declared in `.github/settings.yml`

### Requirement: Automated workflows obtain credentials from the timvw-ci-bot App

Automation that must act on the repository beyond the permissions of the default workflow token — opening or updating pull requests, and publishing check results — SHALL obtain a short-lived installation token from the `timvw-ci-bot` GitHub App rather than from a long-lived personal access token. The configuration file SHALL record this so the contract is discoverable from the repository.

#### Scenario: Credential source is documented in the repository

- **WHEN** a contributor reads `.github/settings.yml`
- **THEN** it states that the `timvw-ci-bot` App is installed and is the credential source for automated pull-request and check workflows

#### Scenario: No long-lived personal token is introduced

- **WHEN** this change is applied
- **THEN** no personal access token secret is added to the repository

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

