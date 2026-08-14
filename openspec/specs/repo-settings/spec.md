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

### Requirement: Branch protection remains an open extension point

Branch protection and required status checks SHALL NOT be declared by this change, because the repository has no CI workflows yet and therefore no checks that could be required. The configuration file SHALL mark where they will be declared so a later change is an additive edit.

#### Scenario: No branch protection is asserted

- **WHEN** `.github/settings.yml` is read
- **THEN** it contains no active `branches:` protection rules
- **AND** the default branch's existing protection state is left unchanged

#### Scenario: Extension point is signposted

- **WHEN** a contributor opens `.github/settings.yml` to add required checks
- **THEN** a comment marks where branch protection and required status checks belong
