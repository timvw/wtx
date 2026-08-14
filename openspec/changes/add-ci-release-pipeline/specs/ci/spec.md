## Purpose

Defines the automated verification every change to `wtx` must pass before it can land: which properties of the source tree are checked, on which operating systems the test suite is proven to pass, and how the outcome is reported back to a pull request so it can later be made a merge requirement.

## ADDED Requirements

### Requirement: Every pull request and every push to the default branch is verified automatically

The repository SHALL run an automated verification pipeline, hosted by GitHub Actions and declared in a tracked file under `.github/workflows/`, on each of these events:

- a pull request targeting the default branch is opened or updated,
- a commit lands on the default branch,
- a maintainer triggers the pipeline manually.

The pipeline SHALL report a pass/fail result to the triggering commit or pull request. It SHALL NOT require any credential beyond the workflow token GitHub provides by default, so that a pull request opened from a fork is verified identically to one opened from a branch in this repository.

#### Scenario: Pull request triggers verification

- **WHEN** a pull request targeting the default branch is opened, or a new commit is pushed to an open pull request's branch
- **THEN** the verification pipeline runs against the pull request's merge result
- **AND** its outcome is reported as a status check on the pull request

#### Scenario: Push to the default branch triggers verification

- **WHEN** a commit lands on the default branch
- **THEN** the verification pipeline runs against that commit

#### Scenario: Manual trigger

- **WHEN** a maintainer dispatches the pipeline by hand from the GitHub UI or API
- **THEN** it runs against the selected ref

#### Scenario: Fork pull request is verified without secrets

- **WHEN** a pull request is opened from a fork of the repository
- **THEN** the verification pipeline runs to completion
- **AND** it does not fail for want of a repository secret

### Requirement: Verification fails when the source is not gofmt-formatted

The pipeline SHALL check that every tracked Go source file is formatted as `gofmt` would format it, and SHALL fail when any file is not. The check SHALL report which files are unformatted. It SHALL NOT rewrite files.

#### Scenario: Unformatted source fails the pipeline

- **WHEN** a commit contains a Go source file that `gofmt` would reformat
- **THEN** the verification pipeline fails
- **AND** the failing file's path appears in the pipeline output

#### Scenario: Formatted source passes

- **WHEN** every tracked Go source file is already `gofmt`-formatted
- **THEN** the formatting check passes and the working tree is left unmodified

### Requirement: Verification runs the Go vet and lint suites

The pipeline SHALL run `go vet` over all packages, and SHALL run a static-analysis linter whose enabled checks are declared in a tracked configuration file at the repository root. Either reporting a problem SHALL fail the pipeline.

Declaring the linter's configuration in a tracked file makes the analysis reproducible: a contributor running the linter locally SHALL get the same set of checks the pipeline runs.

#### Scenario: Vet finding fails the pipeline

- **WHEN** `go vet` reports a finding for any package in the module
- **THEN** the verification pipeline fails

#### Scenario: Lint finding fails the pipeline

- **WHEN** the linter reports a finding under the repository's declared configuration
- **THEN** the verification pipeline fails
- **AND** the finding's file, line, and rule appear in the pipeline output

#### Scenario: Lint configuration is tracked

- **WHEN** the repository is checked out
- **THEN** a linter configuration file is present at the repository root
- **AND** running the linter locally against that configuration enables the same checks the pipeline enables

### Requirement: The module builds and its tests pass on Linux, macOS and Windows

The pipeline SHALL build the module and run its full unit-test suite on Linux, macOS and Windows runners. A build failure or a test failure on any one of those operating systems SHALL fail the pipeline.

The operating-system legs SHALL be independent: a failure on one SHALL NOT prevent the others from running to completion, so a single pipeline run reports every platform's result rather than only the first failure.

#### Scenario: Tests pass on every supported operating system

- **WHEN** the verification pipeline runs on a commit whose tests pass everywhere
- **THEN** the module builds and its tests pass on Linux, on macOS, and on Windows
- **AND** the pipeline succeeds

#### Scenario: A platform-specific test failure fails the pipeline

- **WHEN** the test suite fails on exactly one of the supported operating systems
- **THEN** the verification pipeline fails
- **AND** the output identifies which operating system failed

#### Scenario: One platform's failure does not mask the others

- **WHEN** the test suite fails on one operating system
- **THEN** the legs for the remaining operating systems still run to completion and report their own results

### Requirement: Verification proves the release target platforms still cross-compile

The pipeline SHALL cross-compile the executable for every operating-system and architecture pair that the release pipeline publishes, and SHALL fail if any one of them does not build. This check SHALL run on pull requests, so that a change breaking a platform the maintainer does not develop on is caught before merge rather than when a release tag is pushed.

The set of platforms checked here SHALL match the set the release pipeline builds; if they diverge, a platform can be published without ever having been compiled in verification.

#### Scenario: All release platforms cross-compile

- **WHEN** the verification pipeline runs
- **THEN** the executable is built for each operating-system and architecture pair the release pipeline publishes
- **AND** the pipeline succeeds

#### Scenario: A broken target platform fails the pipeline

- **WHEN** a change compiles on the runner's own platform but fails to compile for one of the release target platforms
- **THEN** the verification pipeline fails
- **AND** the output identifies the failing operating-system and architecture pair

### Requirement: Verification jobs are individually addressable as status checks

Each distinct verification concern SHALL be reported under a stable, human-readable check name, so that a later change can name specific checks as required for merging without having to restructure the pipeline.

Check names SHALL NOT change gratuitously: renaming a check silently removes it from any branch-protection rule that requires it.

#### Scenario: Checks are named and distinguishable

- **WHEN** the verification pipeline completes on a pull request
- **THEN** its results appear as named status checks on that pull request
- **AND** the names distinguish the lint, test, and cross-compile concerns from one another

#### Scenario: Check names can be required by branch protection

- **WHEN** a maintainer configures required status checks for the default branch
- **THEN** the names reported by the verification pipeline are selectable as required checks
